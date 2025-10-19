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

// UserData 用户新产生的数据
type UserData struct {
	UserID   uint    `json:"user_id"`   // 用户ID
	DataSize float64 `json:"data_size"` // 数据大小（比特）
	// TaskType string  `json:"task_type"` // 任务类型
}

type System struct {
	Users []*define.UserDevice // 用户设备列表
	Comms []*define.CommDevice // 通信设备列表

	T      uint      // 当前系统时隙
	R      []float64 // 用户待处理数据列表
	States []State   // 系统状态
	Graph  *Graph    // 系统图结构

	IsRunning     bool         // 是否正在运行
	IsInitialized bool         // 算法是否已初始化
	StopChan      chan bool    // 停止信号
	mutex         sync.RWMutex // 读写锁保护并发访问
}

var (
	sys  *System
	once sync.Once
)

// GetSystemInstance 初始化并获取系统单例实例
func GetSystemInstance() *System {
	once.Do(func() {
		sys = &System{
			Users:         make([]*define.UserDevice, 0),
			Comms:         make([]*define.CommDevice, 0),
			T:             0,
			R:             nil, // 初始化为0长度切片
			States:        make([]State, 0),
			IsRunning:     false,
			IsInitialized: false,
			StopChan:      make(chan bool, 1),
		}
		sys.loadNodesFromDB()
		sys.Graph = NewGraph(sys)
		sys.R = make([]float64, len(sys.Users))
		sys.IsInitialized = true
	})
	return sys
}

// loadNodesFromDB 从数据库加载节点信息
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

// Start 启动算法，处理用户数据
func (s *System) Start(userDataList []UserData) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查算法服务是否已初始化
	if !s.IsInitialized {
		return fmt.Errorf("资源调度算法服务未初始化")
	}

	// 根据UserID找到对应的用户索引并增加数据
	for _, userData := range userDataList {
		for i, user := range s.Users {
			if user.ID == userData.UserID {
				s.R[i] += userData.DataSize
				break
			}
		}
	}

	// 如果已经在运行，在现有数据基础上增加数据，否则启动轮询
	if !s.IsRunning {
		// 启动轮询
		s.IsRunning = true
		go s.runAlgorithmLoop()

	}

	return nil
}

// runAlgorithmLoop 运行算法轮询
func (s *System) runAlgorithmLoop() {
	ticker := time.NewTicker(1 * time.Second) // 每秒轮询
	defer ticker.Stop()

	for {
		select {
		case <-s.StopChan:
			s.mutex.Lock()
			s.IsRunning = false
			s.mutex.Unlock()
			return
		case <-ticker.C:
			// 执行一次算法
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

// executeOneIteration 执行一次算法迭代
func (s *System) executeOneIteration() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 提取当前待处理数据，并清空；有可能执行过程中会有新的数据产生
	R := make([]float64, len(s.R))
	copy(R, s.R)
	for i := range s.R {
		s.R[i] = 0
	}

	// 获取最新状态的队列
	var Q [][]float64
	if len(s.States) > 0 {
		Q = s.States[len(s.States)-1].QNext
	} else {
		Q = make([][]float64, len(s.Users))
		for i := range Q {
			Q[i] = make([]float64, len(s.Comms))
		}
	}

	// 调用算法服务
	iter := 0
	U := len(s.Users)
	s.T += 1

	// 初始化状态
	bestCost := math.Inf(1)
	bestState := State{}

	for iter < constant.Iters && iter < int(math.Pow(2, float64(U))) {
		iter++

		// 调度并获取新状态
		state := NewState(s)
		state.R = R
		state.Q = Q
		newState := s.Graph.Scheduler(*state)
		cost := newState.Objective()

		// 更新最佳决策方案
		prevCost := bestCost
		if cost < bestCost {
			bestCost = cost
			bestState = newState
		}

		// 如果改进很小，提前退出
		if math.Abs(cost-prevCost) < constant.Bias {
			break
		}
	}

	s.States = append(s.States, bestState)

	log.Println("时隙", s.T, "算法迭代完成，最佳成本:", bestCost)
	totalQ := make([]float64, len(s.Comms))
	for i := range bestState.QNext {
		for j := range bestState.QNext[i] {
			totalQ[j] += bestState.QNext[i][j]
		}
	}
	log.Printf("各通信设备队列长度: %+v\n", totalQ)
	log.Printf("各用户传输路径: %+v\n", bestState.TransferPath)
}

// isFinished 判断算法是否完成（没有待处理的数据）
func (s *System) isFinished() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 1. 检查是否有新产生的待处理数据
	for _, r := range s.R {
		if r > 0 {
			return false // 还有新数据要处理
		}
	}

	// 2. 检查队列中是否还有积压的数据
	if len(s.States) > 0 {
		lastState := s.States[len(s.States)-1]
		for i := range lastState.QNext {
			for j := range lastState.QNext[i] {
				if lastState.QNext[i][j] > 0.001 { // 使用小阈值避免浮点数精度问题
					return false // 队列中还有数据
				}
			}
		}
	}

	// 所有数据都已处理完成
	return true
}

// StopAlgorithm 外部调用停止算法
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

// GetSystemInfo 获取系统信息
func (s *System) GetSystemInfo() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	systemInfo := map[string]interface{}{
		"user_count":     len(s.Users),
		"comm_count":     len(s.Comms),
		"is_running":     s.IsRunning,
		"is_initialized": s.IsInitialized,
		"time_slot":      s.T,
		"transfer_path":  map[uint][]uint{},             // 如果没有运行，传输路径为空
		"each_queue":     make([]float64, len(s.Comms)), // 如果没有运行，队列长度为0
		"queue":          0,
		"delay":          0,
		"energy":         0,
		"utilization":    0,
		"drift":          0,
		"penalty":        0,
		"cost":           0,
	}

	// 如果有状态历史，获取最新状态
	if len(s.States) > 0 && s.IsRunning {
		state := &s.States[len(s.States)-1]
		for userIdx, uavs := range state.TransferPath {
			userId := s.Users[userIdx].ID
			if len(uavs) == 0 {
				continue
			}
			paths := make([]uint, 0)
			paths = append(paths, userId)
			for _, uav := range uavs {
				paths = append(paths, s.Comms[uav].ID)
			}
			systemInfo["transfer_path"].(map[uint][]uint)[uint(userId)] = paths
		}
		systemInfo["each_queue"] = state.CalcRowQueue()
		systemInfo["queue"] = state.CalcQueueAvg()
		systemInfo["delay"] = state.ComputeDelay + state.TransferDelay
		systemInfo["energy"] = state.ComputeEnergy + state.TransferEnergy
		systemInfo["utilization"] = state.CalcResourceUtil()
		systemInfo["drift"] = state.Drift
		systemInfo["penalty"] = state.Penalty
		systemInfo["cost"] = state.Cost
	}

	return systemInfo
}

// GetStateHistory 获取状态历史
func (s *System) GetStateHistory() []State {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 深拷贝，避免外部修改
	history := make([]State, len(s.States))
	for i := range s.States {
		history[i] = s.States[i].Copy()
	}
	return history
}

// ClearHistory 清除历史状态
func (s *System) ClearHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.States = make([]State, 0)
	s.T = 0
}
