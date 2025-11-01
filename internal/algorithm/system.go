package algorithm

import (
	"fmt"
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/define"
	"go-backend/internal/models"
	"go-backend/internal/repository"
	"go-backend/pkg/database"
	"log"
	"math"
	"sync"
	"time"
)

type System struct {
	Users   []*define.UserDevice        // 用户设备列表
	Comms   []*define.CommDevice        // 通信设备列表
	UserMap map[uint]*define.UserDevice // 用户ID -> 用户设备（快速查找）
	CommMap map[uint]*define.CommDevice // 通信设备ID -> 通信设备（快速查找）

	TimeSlot     uint         // 当前时隙
	Graph        *Graph       // 网络拓扑图
	TaskManager  *TaskManager // 任务管理器
	CurrentState *State       // 当前最优状态（包含所有性能指标）

	IsRunning     bool         // 是否运行中
	IsInitialized bool         // 是否已初始化
	StopChan      chan bool    // 停止信号通道
	mutex         sync.RWMutex // 并发锁
}

var (
	sys  *System
	once sync.Once
)

func GetSystemInstance() *System {
	once.Do(func() {
		sys = &System{
			Users:    make([]*define.UserDevice, 0),
			Comms:    make([]*define.CommDevice, 0),
			UserMap:  make(map[uint]*define.UserDevice),
			CommMap:  make(map[uint]*define.CommDevice),
			StopChan: make(chan bool, 1),
		}
		sys.loadNodesFromDB()
		sys.Graph = NewGraph(sys)
		sys.TaskManager = NewTaskManager(sys)
		sys.IsInitialized = true
	})
	return sys
}

func (s *System) SubmitTask(req define.TaskBase) (*define.Task, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 使用 Map 快速查找用户
	if _, exists := s.UserMap[req.UserID]; !exists {
		return nil, fmt.Errorf("用户不存在: %d", req.UserID)
	}

	task, err := s.TaskManager.addTask(req)
	if err != nil {
		return nil, err
	}

	if !s.IsRunning {
		s.IsRunning = true
		go s.runAlgorithmLoop()
	}

	return task, nil
}

func (s *System) SubmitBatchTasks(requests []define.TaskBase) ([]*define.Task, error) {
	tasks := make([]*define.Task, 0, len(requests))

	for _, req := range requests {
		task, err := s.SubmitTask(req)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// 停止算法
func (s *System) StopAlgorithm() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.IsRunning {
		select {
		case s.StopChan <- true:
		default:
		}
	}
}

// GetTasksWithPage 获取任务列表（分页）
func (s *System) GetTasksWithPage(offset, limit int, userID *uint, status *define.TaskStatus) ([]*define.Task, int64) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.TaskManager == nil {
		return []*define.Task{}, 0
	}

	// 获取过滤后的所有任务
	allTasks := s.TaskManager.getTasks(userID, status)
	total := int64(len(allTasks))

	// 分页
	if offset >= len(allTasks) {
		return []*define.Task{}, total
	}

	end := offset + limit
	if end > len(allTasks) {
		end = len(allTasks)
	}

	return allTasks[offset:end], total
}

// GetTaskByID 根据ID获取任务
func (s *System) GetTaskByID(taskID string) (*define.Task, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.TaskManager == nil {
		return nil, false
	}

	return s.TaskManager.getTaskByID(taskID)
}

// DeleteTask 删除任务
func (s *System) DeleteTask(taskID string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.TaskManager == nil {
		return false
	}

	return s.TaskManager.deleteTask(taskID)
}

// 清除历史状态
func (s *System) ClearHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.TimeSlot = 0
	s.CurrentState = nil
	// 清除所有任务
	if s.TaskManager != nil {
		s.TaskManager.Tasks = make(map[string]*define.Task)
		s.TaskManager.TaskList = make([]*define.Task, 0)
		s.TaskManager.UserTasks = make(map[uint][]string)
	}
}

// 获取系统信息
func (s *System) GetSystemInfo() *define.SystemInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	systemInfo := &define.SystemInfo{
		UserCount:     len(s.Users),
		CommCount:     len(s.Comms),
		IsRunning:     s.IsRunning,
		IsInitialized: s.IsInitialized,
		TimeSlot:      s.TimeSlot,
		TransferPath:  make(map[string][]uint),
	}

	// 从 TaskManager 获取任务统计
	if s.TaskManager != nil {
		systemInfo.TaskCount = len(s.TaskManager.TaskList)
		systemInfo.ActiveTasks = len(s.TaskManager.getActiveTasks())

		completedCount := 0
		transferPaths := make(map[string][]uint)

		for _, task := range s.TaskManager.Tasks {
			if task.StateMachine().IsCompleted() {
				completedCount++
			}
			// 收集传输路径
			if task.TransferPath != nil && len(task.TransferPath.Path) > 0 {
				transferPaths[task.ID] = task.TransferPath.Path
			}
		}

		systemInfo.CompletedTasks = completedCount
		systemInfo.TransferPath = transferPaths
	}

	// 从 CurrentState 获取系统级指标
	if s.CurrentState != nil && s.IsRunning {
		systemInfo.State = s.CurrentState
	}

	return systemInfo
}

// 从数据库加载节点信息
func (s *System) loadNodesFromDB() {
	db := database.GetDB()
	nodeRepo := repository.NewNodeRepository(db)
	nodes, err := nodeRepo.List(nil)
	if err != nil {
		log.Fatalf("无法从数据库加载节点数据: %v", err)
	}

	for _, node := range nodes {
		switch node.NodeType {
		case models.NodeTypeUser:
			userDevice := define.NewUserDevice(node)
			s.Users = append(s.Users, userDevice)
			s.UserMap[userDevice.ID] = userDevice
		case models.NodeTypeComm:
			commDevice := define.NewCommDevice(node)
			s.Comms = append(s.Comms, commDevice)
			s.CommMap[commDevice.ID] = commDevice
		}
	}

	// 检查是否成功加载了节点
	if len(s.Users) == 0 && len(s.Comms) == 0 {
		log.Fatalf("数据库中没有任何节点数据，请先添加用户设备和通信设备")
	}

	if len(s.Users) == 0 {
		log.Printf("警告: 没有加载到任何用户设备")
	}

	if len(s.Comms) == 0 {
		log.Fatalf("没有加载到任何通信设备，系统无法运行")
	}

	log.Printf("成功加载节点数据: %d个用户设备, %d个通信设备", len(s.Users), len(s.Comms))
}

// 运行算法轮询
func (s *System) runAlgorithmLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.StopChan:
			s.mutex.Lock()
			s.IsRunning = false
			s.mutex.Unlock()
			return
		case <-ticker.C:
			s.executeOneIteration()
			if len(s.TaskManager.getActiveTasks()) == 0 {
				log.Println("所有数据处理完成，算法停止")
				s.mutex.Lock()
				s.IsRunning = false
				s.mutex.Unlock()
				return
			}
		}
	}
}

// 执行一次算法迭代
func (s *System) executeOneIteration() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 收集所有活跃任务
	activeTasks := s.TaskManager.getActiveTasks()
	if len(activeTasks) == 0 {
		return
	}

	s.TimeSlot++

	// 多次迭代寻找最优解（优化版：无深拷贝）
	var bestState *State
	bestCost := math.Inf(1)
	maxIter := min(constant.Iters, len(activeTasks))

	for iter := 0; iter < maxIter; iter++ {
		// 关键优化：每次迭代重新创建state（轻量级），而不是深拷贝
		iterState := NewState(s.TimeSlot, activeTasks, s)
		newState := s.Graph.schedule(iterState, activeTasks)
		cost := newState.Cost

		prevCost := bestCost
		if cost < bestCost {
			bestCost = cost
			bestState = newState
		}

		if math.Abs(cost-prevCost) < constant.Bias {
			break
		}
	}

	if bestState == nil {
		return
	}

	// 保存当前最优状态
	s.CurrentState = bestState

	// 更新任务状态（同步到 Task.MetricsHistory）
	s.TaskManager.syncFromState(bestState, activeTasks, s)

	log.Printf("[时隙 %d]\n成本: %.2f, 任务数: %d", s.TimeSlot, bestCost, len(activeTasks))
	log.Printf("通信设备队列: %+v", bestState.CommQueues)
}
