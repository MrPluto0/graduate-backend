package algorithm

import (
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/define"
	"log"
	"math"
)

// Scheduler 任务调度器 (替代原来Graph.schedule的复杂逻辑)
type Scheduler struct {
	System            *System
	AssignmentManager *AssignmentManager
}

// NewScheduler 创建调度器
func NewScheduler(system *System, assignmentManager *AssignmentManager) *Scheduler {
	return &Scheduler{
		System:            system,
		AssignmentManager: assignmentManager,
	}
}

// Schedule 为所有活跃任务创建本时隙的调度分配
func (s *Scheduler) Schedule(timeSlot uint, tasks []*define.Task) []*define.Assignment {
	assignments := make([]*define.Assignment, 0, len(tasks))

	for _, task := range tasks {
		assign := s.scheduleTask(timeSlot, task)
		if assign != nil {
			assignments = append(assignments, assign)
		}
	}

	// 计算资源分配比例
	s.allocateResources(assignments)

	return assignments
}

// scheduleTask 为单个任务创建调度分配
func (s *Scheduler) scheduleTask(timeSlot uint, task *define.Task) *define.Assignment {
	sm := task.StateMachine()
	lastAssign := s.AssignmentManager.GetLastAssignment(task.ID)

	// 情况1: 任务已经有分配 (Queued/Computing状态) - 复用路径!
	if lastAssign != nil && (sm.IsQueued() || sm.IsComputing()) {
		return s.reuseAssignment(timeSlot, task, lastAssign)
	}

	// 情况2: 新任务 (Pending状态) - 寻找最佳路径
	if sm.IsPending() {
		return s.findBestAssignment(timeSlot, task)
	}

	// 情况3: 已完成或失败的任务 - 不需要调度
	return nil
}

// reuseAssignment 复用上次的分配 (解决12→12路径问题!)
func (s *Scheduler) reuseAssignment(timeSlot uint, task *define.Task, lastAssign *define.Assignment) *define.Assignment {
	queue := s.AssignmentManager.GetCurrentQueue(task.ID, task.DataSize)
	processed := s.AssignmentManager.GetCumulativeProcessed(task.ID)

	return &define.Assignment{
		TimeSlot:            timeSlot,
		TaskID:              task.ID,
		CommID:              lastAssign.CommID,              // 复用通信设备
		Path:                lastAssign.Path,                // 复用路径!
		Speeds:              lastAssign.Speeds,              // 复用速率
		Powers:              lastAssign.Powers,              // 复用功率
		QueueData:           queue,                          // 当前队列
		CumulativeProcessed: processed,                      // 累计进度
		ResourceFraction:    0,                              // 稍后由allocateResources计算
		TransferredData:     0,                              // 稍后由executeAssignment计算
		ProcessedData:       0,                              // 稍后由executeAssignment计算
	}
}

// findBestAssignment 为新任务寻找最佳分配 (简化的调度算法)
func (s *Scheduler) findBestAssignment(timeSlot uint, task *define.Task) *define.Assignment {
	user, ok := s.System.UserMap[task.UserID]
	if !ok {
		log.Printf("用户不存在: %d", task.UserID)
		return nil
	}

	var bestAssign *define.Assignment
	bestCost := math.MaxFloat64

	// 遍历所有通信设备,找到最低cost的分配
	for commID := range s.System.CommMap {
		// 计算路径
		path := s.getPath(task.UserID, commID)
		if len(path) < 2 {
			continue
		}

		// 获取路径的速率和功率
		speeds, powers := s.getPathSpeedsAndPowers(path, user.Speed)

		// 创建临时分配
		assign := &define.Assignment{
			TimeSlot: timeSlot,
			TaskID:   task.ID,
			CommID:   commID,
			Path:     path,
			Speeds:   speeds,
			Powers:   powers,
			QueueData: 0, // 新任务,队列为空
			CumulativeProcessed: 0,
		}

		// 计算cost (简化: 只考虑传输延迟)
		cost := s.computeTransferCost(assign, task.DataSize)

		if cost < bestCost {
			bestCost = cost
			bestAssign = assign
		}
	}

	return bestAssign
}

// computeTransferCost 计算传输cost (综合延迟、能耗、队列)
func (s *Scheduler) computeTransferCost(assign *define.Assignment, dataSize float64) float64 {
	// 1. 传输延迟 (秒)
	transmissionDelay := 0.0
	for _, speed := range assign.Speeds {
		if speed > 0 {
			transmissionDelay += dataSize / speed
		}
	}

	// 2. 能耗成本 (焦耳 = 功率 × 时间)
	energyCost := 0.0
	for i, power := range assign.Powers {
		if i < len(assign.Speeds) && assign.Speeds[i] > 0 {
			// 每段链路的传输时间
			segmentTime := dataSize / assign.Speeds[i]
			energyCost += power * segmentTime
		}
	}

	// 3. 队列延迟 (comm设备的当前负载)
	queueDelay := 0.0
	if lastAssign := s.AssignmentManager.GetLastAssignment(assign.TaskID); lastAssign != nil {
		// 如果设备上已有任务队列,增加等待成本
		queueDelay = lastAssign.QueueData
	}

	// 加权组合: cost = α×delay + β×energy + γ×queue
	// 权重选择: 延迟优先,能耗其次,队列最后
	alpha := 1.0    // 延迟权重
	beta := 0.1     // 能耗权重
	gamma := 0.05   // 队列权重

	totalCost := alpha*transmissionDelay + beta*energyCost + gamma*queueDelay
	return totalCost
}

// allocateResources 为所有分配计算资源比例 (基于优先级和队列长度,带饥饿保护)
func (s *Scheduler) allocateResources(assignments []*define.Assignment) {
	if len(assignments) == 0 {
		return
	}

	// 获取每个任务的优先级权重
	taskWeights := make(map[string]float64)
	totalWeight := 0.0

	for _, assign := range assignments {
		task := s.AssignmentManager.GetTask(assign.TaskID, s.System.TaskManager)
		if task == nil {
			continue
		}

		// 优先级因子: priority/10 + 1 (priority=0→1, priority=10→2, priority=20→3)
		priorityFactor := float64(task.Priority)/10.0 + 1.0

		// 饥饿提升: 如果任务等待过久,提升优先级
		if task.IsStarving() {
			waitTime := task.GetWaitTime().Seconds()
			starvationBoost := 1.0 + (waitTime / 10.0) // 每10秒增加1倍权重
			priorityFactor *= starvationBoost
			// log.Printf("⚠️  任务 %s 饥饿提升: %.2fx (等待%.1fs)", task.ID, starvationBoost, waitTime)
		}

		// 队列因子: 队列越长，需要更多资源
		queueFactor := assign.QueueData + 1.0 // +1避免除零

		weight := priorityFactor * queueFactor
		taskWeights[assign.TaskID] = weight
		totalWeight += weight
	}

	// 按权重分配资源
	if totalWeight < 0.001 {
		// 总权重太小,平均分配
		fraction := 1.0 / float64(len(assignments))
		for _, assign := range assignments {
			assign.ResourceFraction = fraction
		}
		return
	}

	// 按权重比例分配
	for _, assign := range assignments {
		weight := taskWeights[assign.TaskID]
		assign.ResourceFraction = weight / totalWeight
	}
}

// ExecuteAssignments 执行分配,计算传输和处理的数据量
func (s *Scheduler) ExecuteAssignments(assignments []*define.Assignment, tasks map[string]*define.Task) {
	for _, assign := range assignments {
		task := tasks[assign.TaskID]
		if task == nil {
			continue
		}

		// 计算传输量 (简化: 假设用户以固定速率传输)
		userSpeed := 0.0
		if len(assign.Speeds) > 0 {
			userSpeed = assign.Speeds[0]
		}
		transferred := userSpeed * constant.Slot // bits

		// 剩余待传输数据
		remaining := task.DataSize - assign.CumulativeProcessed
		if transferred > remaining {
			transferred = remaining
		}

		// 计算处理量
		processed := 0.0
		if assign.ResourceFraction > 0 && assign.QueueData > 0 {
			// 处理能力 = ResourceFraction × C (周期/秒) / Rho (周期/bit) × Slot (秒)
			processingCapacity := assign.ResourceFraction * constant.C / constant.Rho * constant.Slot
			processed = math.Min(assign.QueueData, processingCapacity)
		}

		// 更新分配
		assign.TransferredData = transferred
		assign.ProcessedData = processed
		assign.CumulativeProcessed += processed

		// 更新队列 (下一时隙的队列 = 当前队列 + 传输 - 处理)
		// 注意: 这里不修改assign.QueueData,因为它是本时隙开始时的值
		// 下一时隙调度时会通过GetCurrentQueue重新计算
	}
}

// getPathSpeedsAndPowers 获取路径的速率和功率
func (s *Scheduler) getPathSpeedsAndPowers(path []uint, userSpeed float64) ([]float64, []float64) {
	speeds := make([]float64, len(path)-1)
	powers := make([]float64, len(path)-1)

	for i := 0; i < len(path)-1; i++ {
		srcID := path[i]
		dstID := path[i+1]

		if i == 0 {
			// 第一段: 用户 → 设备 (使用用户上行速率)
			speeds[i] = userSpeed
			powers[i] = 0.5 // 默认功率
		} else {
			// 后续段: 设备 → 设备 (从Link.Properties解析)
			link, exists := s.System.LinkMap[[2]uint{srcID, dstID}]
			if !exists {
				// 链路不存在,使用默认值
				speeds[i] = 10.0 // Mbps
				powers[i] = 1.0  // W
				continue
			}

			// 解析带宽 (bandwidth in Mbps)
			bandwidth := 10.0 // 默认值
			if bw, ok := link.Properties["bandwidth"].(float64); ok && bw > 0 {
				bandwidth = bw
			}
			speeds[i] = bandwidth

			// 解析功率 (power in W)
			power := 1.0 // 默认值
			if pw, ok := link.Properties["power"].(float64); ok && pw > 0 {
				power = pw
			}
			powers[i] = power
		}
	}

	return speeds, powers
}

// getPath 获取从用户到通信设备的最短路径
func (s *Scheduler) getPath(userID, commID uint) []uint {
	// 使用Floyd算法的最短路径结果
	if s.System.FloydResult == nil {
		// 降级: 返回直连路径
		log.Printf("⚠️  Floyd结果未初始化,使用直连路径")
		return []uint{userID, commID}
	}

	// 获取节点在Floyd矩阵中的索引
	srcIdx, srcOk := s.System.NodeIDToIndex[userID]
	dstIdx, dstOk := s.System.NodeIDToIndex[commID]

	if !srcOk || !dstOk {
		log.Printf("⚠️  节点ID映射失败 (user:%d, comm:%d)", userID, commID)
		return []uint{userID, commID}
	}

	// 获取Floyd路径 (索引数组)
	pathIndices := s.System.FloydResult.Paths[srcIdx][dstIdx]
	if len(pathIndices) == 0 {
		// 路径不可达
		log.Printf("⚠️  路径不可达 (user:%d -> comm:%d)", userID, commID)
		return nil
	}

	// 转换索引为节点ID
	path := make([]uint, len(pathIndices))
	for i, idx := range pathIndices {
		path[i] = s.System.IndexToNodeID[idx]
	}

	return path
}
