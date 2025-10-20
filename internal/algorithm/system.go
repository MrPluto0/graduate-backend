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
	Users []*define.UserDevice // 用户设备列表
	Comms []*define.CommDevice // 通信设备列表

	T           uint   // 当前时隙
	Graph       *Graph // 网络拓扑图
	TaskManager *TaskManager
	LatestState *TaskState // 最新的任务状态（包含所有计算好的指标）

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
			StopChan: make(chan bool, 1),
		}
		sys.loadNodesFromDB()
		sys.Graph = NewGraph(sys)
		sys.TaskManager = NewTaskManager(sys)
		sys.IsInitialized = true
	})
	return sys
}

// 从数据库加载节点信息
func (s *System) loadNodesFromDB() {
	db := database.GetDB()
	nodeRepo := repository.NewNodeRepository(db)
	nodes, err := nodeRepo.List(nil)
	if err != nil {
		return
	}

	for _, node := range nodes {
		switch node.NodeType {
		case models.NodeTypeUser:
			s.Users = append(s.Users, define.NewUserDevice(node))
		case models.NodeTypeComm:
			s.Comms = append(s.Comms, define.NewCommDevice(node))
		}
	}
}

func (s *System) SubmitTask(req define.TaskBase) (*define.Task, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	userIdx := -1
	for i, user := range s.Users {
		if user.ID == req.UserID {
			userIdx = i
			break
		}
	}
	if userIdx == -1 {
		return nil, fmt.Errorf("用户不存在: %d", req.UserID)
	}

	task, err := s.TaskManager.AddTask(req)

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
			if s.isFinished() {
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

	s.T++

	// 创建任务维度的状态
	taskState := NewTaskState(s.T, activeTasks, len(s.Comms))

	// 多次迭代寻找最优解
	taskCount := len(activeTasks)
	maxIter := min(constant.Iters, int(math.Pow(2, float64(taskCount))))

	bestCost := math.Inf(1)
	var bestState *TaskState

	for iter := 0; iter < maxIter; iter++ {
		tempState := taskState.Copy()
		newState := s.Graph.Scheduler(tempState, activeTasks)
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

	// 保存最新状态
	s.LatestState = bestState

	// 更新任务状态
	s.TaskManager.updateFromTaskState(bestState, activeTasks, s)

	log.Printf("时隙 %d 完成，成本: %.2f, 任务数: %d\n", s.T, bestCost, len(activeTasks))
	log.Printf("通信设备队列: %+v\n", bestState.CommQueues)
}

func (s *System) isFinished() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return !s.TaskManager.hasActiveTasks()
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

// 获取系统信息
func (s *System) GetSystemInfo() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	systemInfo := map[string]interface{}{
		"user_count":      len(s.Users),
		"comm_count":      len(s.Comms),
		"is_running":      s.IsRunning,
		"is_initialized":  s.IsInitialized,
		"time_slot":       s.T,
		"transfer_path":   map[string][]uint{}, // 任务ID -> 传输路径
		"each_queue":      make([]float64, len(s.Comms)),
		"queue":           0.0,
		"delay":           0.0,
		"energy":          0.0,
		"task_count":      0,
		"active_tasks":    0,
		"completed_tasks": 0,
	}

	// 从 TaskManager 获取任务统计
	if s.TaskManager != nil {
		systemInfo["task_count"] = len(s.TaskManager.Tasks)
		systemInfo["active_tasks"] = len(s.TaskManager.getActiveTasks())

		completedCount := 0
		for _, task := range s.TaskManager.Tasks {
			if task.Status == define.TaskCompleted {
				completedCount++
			}
		}
		systemInfo["completed_tasks"] = completedCount
	}

	// 从最新的 TaskState 获取计算好的指标
	if s.LatestState != nil {
		transferPaths := make(map[string][]uint)

		for taskID, alloc := range s.LatestState.Allocations {
			// 收集传输路径（已经是ID列表，直接使用）
			if len(alloc.TransferPath) > 0 {
				transferPaths[taskID] = alloc.TransferPath
			}
		}

		systemInfo["transfer_path"] = transferPaths
		systemInfo["each_queue"] = s.LatestState.CommQueues
		systemInfo["queue"] = s.LatestState.Load
		systemInfo["delay"] = s.LatestState.TotalDelay
		systemInfo["energy"] = s.LatestState.TotalEnergy
		systemInfo["cost"] = s.LatestState.Cost
		systemInfo["drift"] = s.LatestState.Drift
		systemInfo["penalty"] = s.LatestState.Penalty
	}

	return systemInfo
}

// 清除历史状态
func (s *System) ClearHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.T = 0
	s.LatestState = nil
	// 清除所有任务
	if s.TaskManager != nil {
		s.TaskManager.Tasks = make(map[string]*define.Task)
		s.TaskManager.UserTasks = make(map[uint][]string)
		s.TaskManager.PendingTasks = make([]string, 0)
		s.TaskManager.ActiveTasks = make([]string, 0)
	}
}
