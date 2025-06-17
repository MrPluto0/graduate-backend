package algorithm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-backend/internal/models"
	"go-backend/internal/repository"
	"go-backend/pkg/database"
	"net/http"
	"sync"
	"time"
)

// UserData 用户新产生的数据
type UserData struct {
	UserID   uint    `json:"user_id"`   // 用户ID
	DataSize float64 `json:"data_size"` // 数据大小（比特）
	// TaskType string  `json:"task_type"` // 任务类型
}

// AlgInitRequest 算法初始化请求
type AlgInitRequest struct {
	Users []models.Node `json:"users"`
	UAVs  []models.Node `json:"uavs"`
}

// AlgStartRequest 算法启动请求
type AlgStartRequest struct {
	T uint        `json:"t"`
	R []float64   `json:"r"` // 新的数据
	Q [][]float64 `json:"Q"` // 队列
}

// AlgResponse 算法服务响应
type AlgResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type System struct {
	Users []models.Node // 用户设备列表
	Comms []models.Node // 通信设备列表

	T      uint      // 当前系统时隙
	R      []float64 // 用户待处理数据列表
	States []State   // 系统状态

	AlgServiceURL string       // 算法服务URL
	IsRunning     bool         // 是否正在运行
	IsInitialized bool         // 算法服务是否已初始化
	StopChan      chan bool    // 停止信号
	mutex         sync.RWMutex // 读写锁保护并发访问
	httpClient    *http.Client // HTTP客户端
}

var (
	systemInst *System
	once       sync.Once
)

// GetSystemInstance 获取系统单例实例
func GetSystemInstance() *System {
	once.Do(func() {
		systemInst = &System{
			Users:         make([]models.Node, 0),
			Comms:         make([]models.Node, 0),
			T:             0,
			R:             nil, // 初始化为0长度切片
			States:        make([]State, 0),
			AlgServiceURL: "http://localhost:8001", // 默认算法服务地址
			IsRunning:     false,
			IsInitialized: false,
			StopChan:      make(chan bool, 1),
			httpClient:    &http.Client{Timeout: 30 * time.Second},
		}
		systemInst.loadNodesFromDB()
		// 初始化待处理数据切片
		systemInst.R = make([]float64, len(systemInst.Users))
		// 在系统初始化时调用一次算法服务初始化
		if err := systemInst.initAlgService(); err != nil {
			fmt.Printf("算法服务初始化失败: %v\n", err)
		} else {
			systemInst.IsInitialized = true
		}
	})
	return systemInst
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
		if node.NodeType == models.NodeTypeUser {
			s.Users = append(s.Users, node)
		} else if node.NodeType == models.NodeTypeComm {
			s.Comms = append(s.Comms, node)
		}
	}
}

// Start 启动算法，处理用户数据
func (s *System) Start(userDataList []UserData) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查算法服务是否已初始化
	if !s.IsInitialized {
		return fmt.Errorf("算法服务未初始化")
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

	// 如果已经在运行，在现有数据基础上增加数据，继续执行
	if !s.IsRunning {
		// 启动轮询
		s.IsRunning = true
		go s.runAlgorithmLoop()

	}

	return nil
}

// initAlgService 初始化算法服务
func (s *System) initAlgService() error {
	// 调用算法服务初始化接口
	jsonData, err := json.Marshal(AlgInitRequest{
		Users: s.Users,
		UAVs:  s.Comms,
	})
	if err != nil {
		return fmt.Errorf("序列化初始化请求失败: %v", err)
	}

	resp, err := s.httpClient.Post(
		s.AlgServiceURL+"/alg/init",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("调用初始化接口失败: %v", err)
	}
	defer resp.Body.Close()

	var algResp AlgResponse
	if err := json.NewDecoder(resp.Body).Decode(&algResp); err != nil {
		return fmt.Errorf("解析初始化响应失败: %v", err)
	}

	if algResp.Code != 0 {
		return fmt.Errorf("算法服务初始化失败: %s", algResp.Message)
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
			if completed := s.executeOneIteration(); completed {
				s.mutex.Lock()
				s.IsRunning = false
				s.mutex.Unlock()
				return
			}
		}
	}
}

// executeOneIteration 执行一次算法迭代
func (s *System) executeOneIteration() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 提取当前待处理数据，并清空；有可能执行过程中会有新的数据产生
	r := make([]float64, len(s.Users))
	copy(r, s.R)
	for i := range s.R {
		s.R[i] = 0
	}

	// 获取最新状态的队列
	lastIdx := len(s.States) - 1
	var Q [][]float64
	if lastIdx >= 0 {
		Q = s.States[lastIdx].QNext
	} else {
		Q = make([][]float64, len(s.Users))
		for i := range Q {
			Q[i] = make([]float64, len(s.Comms))
		}
	}

	fmt.Println("当前时隙:", s.T)
	fmt.Println("用户数据:", r)

	jsonData, err := json.Marshal(AlgStartRequest{
		T: s.T,
		R: r,
		Q: Q,
	})
	if err != nil {
		fmt.Printf("序列化算法请求失败: %v\n", err)
		return true // 错误时停止
	}

	// 调用算法服务
	resp, err := s.httpClient.Post(
		s.AlgServiceURL+"/alg/start",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("调用算法接口失败: %v\n", err)
		return true // 错误时停止
	}
	defer resp.Body.Close()

	var algResp AlgResponse
	if err := json.NewDecoder(resp.Body).Decode(&algResp); err != nil {
		fmt.Printf("解析算法响应失败: %v\n", err)
		return true // 错误时停止
	}

	if algResp.Code != 0 {
		fmt.Printf("算法执行失败: %s\n", algResp.Message)
		return true // 错误时停止
	}

	// 解析状态数据
	if algResp.Data != nil {
		stateData, err := json.Marshal(algResp.Data)
		if err != nil {
			fmt.Printf("序列化状态数据失败: %v\n", err)
			return true
		}

		var state State
		if err := json.Unmarshal(stateData, &state); err != nil {
			fmt.Printf("解析状态数据失败: %v\n", err)
			return true
		}

		// 存储状态
		s.States = append(s.States, state)
		s.T++

		QMid := make([]float64, len(s.Comms))
		for i := range state.QNext {
			for j := range state.QNext[i] {
				QMid[j] += state.QNext[i][j]
			}
		}
		fmt.Printf("状态已更新, 当前通信队列: %+v\n", QMid)

		// 检查是否处理完毕：所有队列且待处理数据均为空，则认为处理完毕
		allQEmpty := true
		for _, qCol := range state.QNext {
			for _, q := range qCol {
				if q > 0.001 { // 考虑浮点精度
					allQEmpty = false
					break
				}
			}
			if !allQEmpty {
				break
			}
		}

		allREmpty := true
		for _, rVal := range r {
			if rVal > 0.001 { // 考虑浮点精度
				allREmpty = false
				break
			}
		}

		if allQEmpty && allREmpty {
			fmt.Println("所有队列已处理完毕，算法停止")
			return true // 处理完毕，停止轮询
		}
	}

	return false // 继续轮询
}

// stopAlgorithm 停止算法
func (s *System) stopAlgorithm() {
	if s.IsRunning {
		select {
		case s.StopChan <- true:
		default:
		}
	}
}

// StopAlgorithm 外部调用停止算法
func (s *System) StopAlgorithm() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stopAlgorithm()
}

// GetCurrentState 获取当前状态
func (s *System) GetCurrentState() *State {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.States) == 0 {
		return nil
	}
	return &s.States[len(s.States)-1]
}

// GetSystemInfo 获取系统信息
func (s *System) GetSystemInfo() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 当前状态
	state := s.GetCurrentState()

	if s.IsRunning {
		return map[string]interface{}{
			"user_count":     len(s.Users),
			"comm_count":     len(s.Comms),
			"time_slot":      s.T,
			"each_queue":     state.CalcRowQueue(),
			"queue":          state.CalcQueueAvg(),
			"delay":          state.ComputeDelay + state.TransferDelay,
			"energy":         state.ComputeEnergy + state.TransferEnergy,
			"utilization":    state.CalcResourceUtil(),
			"drift":          state.Drift,
			"penalty":        state.Penalty,
			"cost":           state.Cost,
			"is_running":     s.IsRunning,
			"is_initialized": s.IsInitialized,
		}
	} else {
		return map[string]interface{}{
			"user_count":     len(s.Users),
			"comm_count":     len(s.Comms),
			"time_slot":      s.T,
			"each_queue":     make([]float64, len(s.Comms)), // 如果没有运行，队列长度为0
			"queue":          0,
			"delay":          0,
			"energy":         0,
			"utilization":    0,
			"drift":          0,
			"penalty":        0,
			"cost":           0,
			"is_running":     s.IsRunning,
			"is_initialized": s.IsInitialized,
		}
	}
}

// SetAlgServiceURL 设置算法服务URL
func (s *System) SetAlgServiceURL(url string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.AlgServiceURL = url
}

// GetStateHistory 获取状态历史
func (s *System) GetStateHistory() []State {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 返回副本，避免外部修改
	history := make([]State, len(s.States))
	copy(history, s.States)
	return history
}

// ClearHistory 清除历史状态
func (s *System) ClearHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.States = make([]State, 0)
	s.T = 0
}
