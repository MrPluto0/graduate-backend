package algorithm

import (
	"fmt"
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/define"
	"go-backend/internal/algorithm/utils"
	"go-backend/internal/models"
	"go-backend/internal/repository"
	"go-backend/pkg/database"
	"log"
	"math"
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

	// 网络拓扑
	NodeIDToIndex map[uint]int       // NodeID -> Floyd矩阵索引
	IndexToNodeID map[int]uint       // Floyd矩阵索引 -> NodeID
	FloydResult   *utils.FloydResult // Floyd最短路径结果

	// 核心组件
	TaskManager        *TaskManager
	AssignmentManager  *AssignmentManager
	Scheduler          *Scheduler          // 简单调度器 (已弃用)
	LyapunovScheduler  *LyapunovScheduler  // Lyapunov负载均衡调度器
	AlarmMonitor       *AlarmMonitor       // 告警监控器
	UseLyapunov        bool                // 是否使用Lyapunov调度器 (默认true)

	// 运行状态
	TimeSlot      uint
	IsRunning     bool
	IsInitialized bool
	CurrentState  *define.StateMetrics // 当前系统状态指标
	StopChan      chan bool
	mutex         sync.RWMutex
}

// NewSystem 创建新系统实例 (替代单例模式)
func NewSystem() *System {
	sys := &System{
		Users:         make([]*define.UserDevice, 0),
		Comms:         make([]*define.CommDevice, 0),
		UserMap:       make(map[uint]*define.UserDevice),
		CommMap:       make(map[uint]*define.CommDevice),
		LinkMap:       make(map[[2]uint]*models.Link),
		NodeIDToIndex: make(map[uint]int),
		IndexToNodeID: make(map[int]uint),
		CurrentState:  define.NewStateMetrics(),
		StopChan:      make(chan bool, 1),
	}

	// 加载设备数据
	if err := sys.loadNodesFromDB(); err != nil {
		log.Printf("⚠️  系统初始化失败: %v", err)
		// 不返回nil,而是返回部分初始化的系统(允许降级运行)
		sys.IsInitialized = false
		return sys
	}

	// 构建Floyd最短路径
	if err := sys.buildFloydPaths(); err != nil {
		log.Printf("⚠️  Floyd路径计算失败: %v", err)
		sys.IsInitialized = false
		return sys
	}

	// 初始化组件
	sys.TaskManager = NewTaskManager()
	sys.AssignmentManager = NewAssignmentManager()
	sys.Scheduler = NewScheduler(sys, sys.AssignmentManager)
	sys.LyapunovScheduler = NewLyapunovScheduler(sys, sys.AssignmentManager)
	sys.UseLyapunov = true // 默认使用Lyapunov调度器

	sys.IsInitialized = true
	log.Println("✓ 系统初始化完成 (使用Lyapunov调度器)")
	return sys
}

// SetAlarmMonitor 设置告警监控器（依赖注入）
func (s *System) SetAlarmMonitor(monitor *AlarmMonitor) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.AlarmMonitor = monitor
	log.Println("✓ 告警监控器已启用")
}

// loadNodesFromDB 从数据库加载设备数据
func (s *System) loadNodesFromDB() error {
	db := database.GetDB()
	nodeRepo := repository.NewNodeRepository(db)
	linkRepo := repository.NewLinkRepository(db)

	// 加载节点
	nodes, err := nodeRepo.List(nil)
	if err != nil {
		log.Printf("❌ 加载节点失败: %v", err)
		return fmt.Errorf("加载节点失败: %w", err)
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
		log.Printf("❌ 加载链路失败: %v", err)
		return fmt.Errorf("加载链路失败: %w", err)
	}

	for _, link := range links {
		s.LinkMap[[2]uint{link.SourceID, link.TargetID}] = &link

		// 填充用户设备的上行速度
		if user, exists := s.UserMap[link.TargetID]; exists {
			if comm, isComm := s.CommMap[link.SourceID]; isComm {
				// 计算用户到基站的上行速率 (bits/s)
				dist := utils.Distance(user.X, user.Y, comm.X, comm.Y)
				user.Speed = utils.TransferSpeed(constant.P_u, dist)
			}
		}
	}

	log.Printf("✓ 成功加载节点数据: %d个用户设备, %d个通信设备", len(s.Users), len(s.Comms))
	return nil
}

// buildFloydPaths 构建Floyd最短路径矩阵
func (s *System) buildFloydPaths() error {
	// 收集所有节点ID
	allNodeIDs := make([]uint, 0, len(s.UserMap)+len(s.CommMap))
	for id := range s.UserMap {
		allNodeIDs = append(allNodeIDs, id)
	}
	for id := range s.CommMap {
		allNodeIDs = append(allNodeIDs, id)
	}

	if len(allNodeIDs) == 0 {
		return fmt.Errorf("没有可用节点")
	}

	// 构建ID映射 (NodeID <-> Matrix Index)
	for idx, nodeID := range allNodeIDs {
		s.NodeIDToIndex[nodeID] = idx
		s.IndexToNodeID[idx] = nodeID
	}

	n := len(allNodeIDs)
	// 初始化邻接矩阵 (无穷大表示不连通)
	adjMatrix := make([][]float64, n)
	for i := range adjMatrix {
		adjMatrix[i] = make([]float64, n)
		for j := range adjMatrix[i] {
			if i == j {
				adjMatrix[i][j] = 0 // 自身到自身距离为0
			} else {
				adjMatrix[i][j] = math.Inf(1) // 初始为无穷大
			}
		}
	}

	// 填充链路权重 (使用传输延迟作为权重)
	for key, link := range s.LinkMap {
		srcID, dstID := key[0], key[1]
		srcIdx, srcOk := s.NodeIDToIndex[srcID]
		dstIdx, dstOk := s.NodeIDToIndex[dstID]

		if !srcOk || !dstOk {
			continue
		}

		// 使用传输延迟作为边权重
		delay := 1.0 // 默认延迟
		if bandwidth, ok := link.Properties["bandwidth"].(float64); ok && bandwidth > 0 {
			// 延迟 ≈ 1/带宽 (简化模型)
			delay = 1.0 / bandwidth
		}

		// 正向边
		adjMatrix[srcIdx][dstIdx] = delay

		// 自动添加反向边 (对称链路)
		// 1. Comm ↔ Comm (无人机/基站之间): 双向对称
		_, srcIsComm := s.CommMap[srcID]
		_, dstIsComm := s.CommMap[dstID]
		if srcIsComm && dstIsComm {
			adjMatrix[dstIdx][srcIdx] = delay
		}

		// 2. User ↔ Comm (用户设备与基站): 双向但上行速率不同
		_, srcIsUser := s.UserMap[srcID]
		_, dstIsUser := s.UserMap[dstID]
		if (srcIsComm && dstIsUser) || (srcIsUser && dstIsComm) {
			// 上行延迟假设与下行相同 (简化模型)
			// 实际上行速率更低,但延迟差异不大
			adjMatrix[dstIdx][srcIdx] = delay
		}
	}

	// 运行Floyd算法
	s.FloydResult = utils.Floyd(adjMatrix)

	log.Printf("✓ Floyd最短路径计算完成 (%d个节点)", n)
	return nil
}

// SubmitTask 提交任务
func (s *System) SubmitTask(userID uint, dataSize float64, taskType string) (*define.Task, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查系统是否已初始化
	if !s.IsInitialized {
		log.Printf("❌ 系统未初始化,无法提交任务")
		return nil, fmt.Errorf("系统未初始化")
	}

	// 验证用户
	if _, exists := s.UserMap[userID]; !exists {
		log.Printf("❌ 提交任务失败: 用户 %d 不存在", userID)
		return nil, fmt.Errorf("用户不存在: %d", userID)
	}

	// 验证数据大小
	if dataSize <= 0 {
		log.Printf("❌ 提交任务失败: 无效的数据大小 %.2f", dataSize)
		return nil, fmt.Errorf("无效的数据大小: %.2f", dataSize)
	}

	// 创建任务
	task := define.NewTask(userID, dataSize, taskType)
	s.TaskManager.AddTask(task)

	log.Printf("✓ 任务 %s 已提交 (用户:%d, 数据:%.2fMB, 类型:%s)",
		task.ID, userID, dataSize, taskType)

	// 启动调度循环
	if !s.IsRunning {
		s.IsRunning = true
		go s.runSchedulingLoop()
		log.Println("✓ 调度循环已启动")
	}

	return task, nil
}

// SubmitTaskWithPriority 提交带优先级的任务
func (s *System) SubmitTaskWithPriority(userID uint, dataSize float64, taskType string, priority int) (*define.Task, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查系统是否已初始化
	if !s.IsInitialized {
		log.Printf("❌ 系统未初始化,无法提交任务")
		return nil, fmt.Errorf("系统未初始化")
	}

	// 验证用户
	if _, exists := s.UserMap[userID]; !exists {
		log.Printf("❌ 提交任务失败: 用户 %d 不存在", userID)
		return nil, fmt.Errorf("用户不存在: %d", userID)
	}

	// 验证数据大小
	if dataSize <= 0 {
		log.Printf("❌ 提交任务失败: 无效的数据大小 %.2f", dataSize)
		return nil, fmt.Errorf("无效的数据大小: %.2f", dataSize)
	}

	// 创建带优先级的任务
	task := define.NewTaskWithPriority(userID, dataSize, taskType, priority)
	s.TaskManager.AddTask(task)

	log.Printf("✓ 任务 %s 已提交 (用户:%d, 数据:%.2fMB, 类型:%s, 优先级:%d)",
		task.ID, userID, dataSize, taskType, priority)

	// 启动调度循环
	if !s.IsRunning {
		s.IsRunning = true
		go s.runSchedulingLoop()
		log.Println("✓ 调度循环已启动")
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

	// 2. 检查超时任务
	s.checkTimeouts()

	// 3. 获取活跃任务（TaskManager有自己的锁）
	tasks := s.TaskManager.GetActiveTasks()
	if len(tasks) == 0 {
		log.Println("所有任务已完成，停止调度")
		s.mutex.Lock()
		s.IsRunning = false
		s.mutex.Unlock()
		return
	}

	// 3. 创建调度分配（根据配置选择调度器）
	var assignments []*define.Assignment
	s.mutex.RLock()
	useLyapunov := s.UseLyapunov
	s.mutex.RUnlock()

	if useLyapunov {
		assignments = s.LyapunovScheduler.Schedule(currentSlot, tasks)
	} else {
		assignments = s.Scheduler.Schedule(currentSlot, tasks)
	}

	// 4. 执行分配,计算传输和处理量（不需要System锁）
	taskMap := make(map[string]*define.Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	if useLyapunov {
		s.LyapunovScheduler.ExecuteAssignments(assignments, taskMap)
	} else {
		s.Scheduler.ExecuteAssignments(assignments, taskMap)
	}

	// 5. 更新任务状态（TaskManager内部有锁）
	s.updateTaskStates(assignments)

	// 6. 保存分配历史（AssignmentManager内部有锁）
	for _, assign := range assignments {
		s.AssignmentManager.AddAssignment(assign)
	}

	// 7. 更新系统状态指标（供前端Dashboard使用）
	s.updateStateMetrics(assignments, tasks)

	// 8. 检查系统状态并产生告警（如果告警监控器已启用）
	s.mutex.RLock()
	alarmMonitor := s.AlarmMonitor
	currentState := s.CurrentState
	s.mutex.RUnlock()

	if alarmMonitor != nil && currentState != nil {
		// 检查系统状态告警
		alarmMonitor.CheckSystemState(currentState, tasks)

		// 检查任务失败告警
		for _, task := range tasks {
			if task.Status == define.TaskFailed {
				alarmMonitor.CheckTaskFailures(task)
			}
		}
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

// CancelTask 取消任务
func (s *System) CancelTask(taskID string) error {
	// TaskManager.CancelTask 内部有锁保护
	err := s.TaskManager.CancelTask(taskID)
	if err != nil {
		log.Printf("❌ 取消任务失败: %v", err)
		return err
	}

	log.Printf("✓ 任务 %s 已取消", taskID)
	return nil
}

// checkTimeouts 在每个时隙检查超时任务
func (s *System) checkTimeouts() {
	timedOutTasks := s.TaskManager.CheckTimeouts()
	if len(timedOutTasks) > 0 {
		log.Printf("⚠️  检测到 %d 个超时任务: %v", len(timedOutTasks), timedOutTasks)
	}
}

// Stop 停止调度
func (s *System) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.IsRunning {
		s.StopChan <- true
		s.IsRunning = false
		log.Println("✓ 调度循环已停止")
	}
}

// SetSchedulerType 设置调度器类型
// schedulerType: "lyapunov" 或 "simple"
func (s *System) SetSchedulerType(schedulerType string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch schedulerType {
	case "lyapunov":
		s.UseLyapunov = true
		log.Println("✓ 切换到Lyapunov负载均衡调度器")
		return nil
	case "simple":
		s.UseLyapunov = false
		log.Println("✓ 切换到简单贪心调度器")
		return nil
	default:
		return fmt.Errorf("未知的调度器类型: %s (支持: lyapunov, simple)", schedulerType)
	}
}

// GetSchedulerType 获取当前调度器类型
func (s *System) GetSchedulerType() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.UseLyapunov {
		return "lyapunov"
	}
	return "simple"
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

	// 获取活跃任务数量
	activeTaskCount := len(s.TaskManager.GetActiveTasks())

	// 复制当前状态（避免并发问题）
	// 如果没有活跃任务，返回空状态
	s.mutex.RLock()
	var currentState interface{}
	if activeTaskCount > 0 && s.CurrentState != nil {
		currentState = s.CurrentState
	} else {
		currentState = nil
	}
	s.mutex.RUnlock()

	return &define.SystemInfo{
		UserCount:      userCount,
		CommCount:      commCount,
		IsRunning:      isRunning,
		IsInitialized:  isInitialized,
		TimeSlot:       timeSlot,
		TransferPath:   transferPaths,
		TaskCount:      s.TaskManager.Count(),
		ActiveTasks:    activeTaskCount,
		CompletedTasks: s.TaskManager.CountCompleted(),
		State:          currentState,
	}
}

// updateStateMetrics 更新系统全局状态指标
func (s *System) updateStateMetrics(assignments []*define.Assignment, tasks []*define.Task) {
	state := define.NewStateMetrics()

	// 1. 统计每个Comm的队列长度（从assignments）
	commQueues := make(map[uint]float64)
	for _, assign := range assignments {
		// QueueData表示本时隙结束后该任务在队列中的数据量
		queueData := assign.QueueData + assign.TransferredData - assign.ProcessedData
		if queueData > 0 {
			commQueues[assign.CommID] += queueData
			state.TotalQueue += queueData
		}
	}

	// 转换为string key (前端期望)
	for commID, queue := range commQueues {
		state.CommQueues[fmt.Sprintf("%d", commID)] = queue
	}

	// 2. 从Scheduler计算本时隙的延迟和能耗（使用简化估算）
	for _, assign := range assignments {
		// 传输延迟估算: 数据量 / 平均速率
		avgSpeed := 1.0
		if len(assign.Speeds) > 0 {
			for _, speed := range assign.Speeds {
				avgSpeed += speed
			}
			avgSpeed /= float64(len(assign.Speeds))
		}
		transferDelay := assign.TransferredData / avgSpeed

		// 计算延迟: ProcessedData × Rho / (ResourceFraction × C)
		computeDelay := 0.0
		if assign.ResourceFraction > 0 && assign.ProcessedData > 0 {
			computeDelay = assign.ProcessedData * constant.Rho / (assign.ResourceFraction * constant.C)
		}

		// 传输能耗: Power × TransmissionTime
		transferEnergy := 0.0
		for i, speed := range assign.Speeds {
			if speed > 0 && assign.TransferredData > 0 {
				transmissionTime := assign.TransferredData / speed
				power := 0.5 // 默认功率
				if i < len(assign.Powers) {
					power = assign.Powers[i]
				}
				transferEnergy += power * transmissionTime
			}
		}

		// 计算能耗: ResourceFraction × Kappa × C³ × Slot
		computeEnergy := assign.ResourceFraction * constant.Kappa * constant.C * constant.C * constant.C * constant.Slot

		state.TransferDelay += transferDelay
		state.ComputeDelay += computeDelay
		state.TransferEnergy += transferEnergy
		state.ComputeEnergy += computeEnergy
	}

	state.TotalDelay = state.TransferDelay + state.ComputeDelay
	state.TotalEnergy = state.TransferEnergy + state.ComputeEnergy

	// 3. 计算系统负载 (活跃任务数 / 通信设备数)
	if len(s.Comms) > 0 {
		state.Load = float64(len(tasks)) / float64(len(s.Comms))
	}

	// 4. 计算Cost (简化版: 加权和)
	state.Cost = state.TotalDelay*1.0 + state.TotalEnergy*0.1 + state.TotalQueue*0.05

	// 5. 计算Drift和Penalty (简化版: 基于队列和延迟)
	state.Drift = state.TotalQueue * 0.5
	state.Penalty = state.TotalDelay + state.TotalEnergy*0.1

	// 6. 原子更新CurrentState
	s.mutex.Lock()
	s.CurrentState = state
	s.mutex.Unlock()
}
