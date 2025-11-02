package algorithm

import (
	"fmt"
	"go-backend/internal/algorithm/define"
	"go-backend/internal/models"
	"go-backend/internal/repository"
	"go-backend/pkg/database"
	"log"
	"sync"
	"time"
)

// System 重构后的系统 (使用简化的数据结构)
type System struct {
	// 设备信息
	Users   []*define.UserDevice
	Comms   []*define.CommDevice
	UserMap map[uint]*define.UserDevice
	CommMap map[uint]*define.CommDevice
	LinkMap map[[2]uint]*models.Link // [源ID, 目标ID] -> Link

	// 核心组件
	TaskManager       *TaskManager
	AssignmentManager *AssignmentManager
	Scheduler         *Scheduler

	// 运行状态
	TimeSlot      uint
	IsRunning     bool
	IsInitialized bool
	StopChan      chan bool
	mutex         sync.RWMutex
}

// NewSystem 创建新系统实例 (替代单例模式)
func NewSystem() *System {
	sys := &System{
		Users:    make([]*define.UserDevice, 0),
		Comms:    make([]*define.CommDevice, 0),
		UserMap:  make(map[uint]*define.UserDevice),
		CommMap:  make(map[uint]*define.CommDevice),
		LinkMap:  make(map[[2]uint]*models.Link),
		StopChan: make(chan bool, 1),
	}

	// 加载设备数据
	sys.loadNodesFromDB()

	// 初始化组件
	sys.TaskManager = NewTaskManager()
	sys.AssignmentManager = NewAssignmentManager()
	sys.Scheduler = NewScheduler(sys, sys.AssignmentManager)

	sys.IsInitialized = true
	return sys
}

// loadNodesFromDB 从数据库加载设备数据
func (s *System) loadNodesFromDB() {
	db := database.GetDB()
	nodeRepo := repository.NewNodeRepository(db)
	linkRepo := repository.NewLinkRepository(db)

	// 加载节点
	nodes, err := nodeRepo.List(nil)
	if err != nil {
		log.Fatalf("加载节点失败: %v", err)
	}

	for _, node := range nodes {
		if node.NodeType == models.NodeTypeUser {
			user := &define.UserDevice{
				Node:  node,
				Speed: 0, // 稍后从Link中填充
			}
			s.Users = append(s.Users, user)
			s.UserMap[node.ID] = user
		} else if node.NodeType == models.NodeTypeComm {
			comm := &define.CommDevice{
				Node: node,
			}
			s.Comms = append(s.Comms, comm)
			s.CommMap[node.ID] = comm
		}
	}

	// 加载链路
	links, err := linkRepo.List(nil)
	if err != nil {
		log.Fatalf("加载链路失败: %v", err)
	}

	for _, link := range links {
		s.LinkMap[[2]uint{link.SourceID, link.TargetID}] = &link
	}

	log.Printf("成功加载节点数据: %d个用户设备, %d个通信设备", len(s.Users), len(s.Comms))
}

// SubmitTask 提交任务
func (s *System) SubmitTask(userID uint, dataSize float64, taskType string) (*define.Task, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证用户
	if _, exists := s.UserMap[userID]; !exists {
		return nil, fmt.Errorf("用户不存在: %d", userID)
	}

	// 创建任务
	task := define.NewTask(userID, dataSize, taskType)
	s.TaskManager.AddTask(task)

	// 启动调度循环
	if !s.IsRunning {
		s.IsRunning = true
		go s.runSchedulingLoop()
	}

	return task, nil
}

// runSchedulingLoop 调度循环 (简化的单一职责流程)
func (s *System) runSchedulingLoop() {
	ticker := time.NewTicker(1 * time.Second) // constant.Slot是float64,这里用1秒
	defer ticker.Stop()

	for {
		select {
		case <-s.StopChan:
			log.Println("调度循环停止")
			return
		case <-ticker.C:
			s.executeOneSlot()
		}
	}
}

// executeOneSlot 执行一个时隙的调度
func (s *System) executeOneSlot() {
	// 细化锁粒度: 只在必要时持有锁

	// 1. 原子递增时隙
	s.mutex.Lock()
	s.TimeSlot++
	currentSlot := s.TimeSlot
	s.mutex.Unlock()

	// 2. 获取活跃任务（TaskManager有自己的锁）
	tasks := s.TaskManager.GetActiveTasks()
	if len(tasks) == 0 {
		log.Println("所有任务已完成，停止调度")
		s.mutex.Lock()
		s.IsRunning = false
		s.mutex.Unlock()
		return
	}

	// 3. 创建调度分配（调度计算不需要持有System锁）
	assignments := s.Scheduler.Schedule(currentSlot, tasks)

	// 4. 执行分配,计算传输和处理量（不需要System锁）
	taskMap := make(map[string]*define.Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}
	s.Scheduler.ExecuteAssignments(assignments, taskMap)

	// 5. 更新任务状态（TaskManager内部有锁）
	s.updateTaskStates(assignments)

	// 6. 保存分配历史（AssignmentManager内部有锁）
	for _, assign := range assignments {
		s.AssignmentManager.AddAssignment(assign)
	}

	log.Printf("时隙 %d: 调度了 %d 个任务", currentSlot, len(assignments))
}

// updateTaskStates 根据分配结果更新任务状态
func (s *System) updateTaskStates(assignments []*define.Assignment) {
	// 批量更新任务状态（使用TaskManager的锁保护）
	for _, assign := range assignments {
		// 先读取任务信息（带读锁）
		task := s.TaskManager.GetTask(assign.TaskID)
		if task == nil {
			continue
		}

		currentStatus := task.Status
		dataSize := task.DataSize

		// 确定目标状态
		var targetStatus define.TaskStatus
		var shouldUpdate bool

		switch currentStatus {
		case define.TaskPending:
			// Pending → Queued (首次分配)
			targetStatus = define.TaskQueued
			shouldUpdate = true

		case define.TaskQueued:
			// Queued → Computing (开始处理数据)
			if assign.ProcessedData > 0 {
				targetStatus = define.TaskComputing
				shouldUpdate = true
			}
			// Queued → Completed (数据处理完成,未经Computing状态)
			if assign.CumulativeProcessed >= dataSize-0.001 {
				targetStatus = define.TaskCompleted
				shouldUpdate = true
			}

		case define.TaskComputing:
			// Computing → Completed (数据处理完成)
			if assign.CumulativeProcessed >= dataSize-0.001 {
				targetStatus = define.TaskCompleted
				shouldUpdate = true
			}
		}

		// 执行状态转换（使用TaskManager的写锁保护）
		if shouldUpdate {
			if err := s.TaskManager.UpdateTaskStatus(assign.TaskID, targetStatus); err != nil {
				log.Printf("状态转换失败 (%s→%d): %v", assign.TaskID, targetStatus, err)
			}
		}
	}
}

// Stop 停止调度
func (s *System) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.IsRunning {
		s.StopChan <- true
		s.IsRunning = false
	}
}

// GetSystemInfo 获取系统信息
func (s *System) GetSystemInfo() *define.SystemInfo {
	s.mutex.RLock()
	userCount := len(s.Users)
	commCount := len(s.Comms)
	isRunning := s.IsRunning
	isInitialized := s.IsInitialized
	timeSlot := s.TimeSlot
	s.mutex.RUnlock()

	// 收集传输路径 - 使用TaskManager的锁保护
	transferPaths := make(map[string][]uint)
	s.TaskManager.mutex.RLock()
	taskIDs := make([]string, 0, len(s.TaskManager.Tasks))
	for taskID := range s.TaskManager.Tasks {
		taskIDs = append(taskIDs, taskID)
	}
	s.TaskManager.mutex.RUnlock()

	// 获取每个任务的最后分配（AssignmentManager有自己的锁）
	for _, taskID := range taskIDs {
		lastAssign := s.AssignmentManager.GetLastAssignment(taskID)
		if lastAssign != nil && len(lastAssign.Path) > 0 {
			transferPaths[taskID] = lastAssign.Path
		}
	}

	return &define.SystemInfo{
		UserCount:      userCount,
		CommCount:      commCount,
		IsRunning:      isRunning,
		IsInitialized:  isInitialized,
		TimeSlot:       timeSlot,
		TransferPath:   transferPaths,
		TaskCount:      s.TaskManager.Count(),
		ActiveTasks:    len(s.TaskManager.GetActiveTasks()),
		CompletedTasks: s.TaskManager.CountCompleted(),
	}
}
