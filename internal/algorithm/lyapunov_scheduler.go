package algorithm

import (
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/define"
	"log"
	"math"
	"math/rand"
)

// LyapunovScheduler 真正的Lyapunov drift-plus-penalty调度器
// 实现动态负载均衡和任务重分配
type LyapunovScheduler struct {
	System            *System
	AssignmentManager *AssignmentManager

	// Lyapunov优化参数
	V float64 // 控制参数 (权衡队列稳定性与性能)

	// 上一时隙的队列状态 (用于计算drift)
	lastCommQueues map[uint]float64
}

// NewLyapunovScheduler 创建Lyapunov调度器
func NewLyapunovScheduler(system *System, assignmentManager *AssignmentManager) *LyapunovScheduler {
	return &LyapunovScheduler{
		System:            system,
		AssignmentManager: assignmentManager,
		V:                 constant.V, // 从常量读取
		lastCommQueues:    make(map[uint]float64),
	}
}

// Schedule 为所有活跃任务寻找最优分配 (Lyapunov drift-plus-penalty)
//
// 算法流程:
// 1. 迭代 constant.Iters 次，每次生成随机的任务→设备分配
// 2. 对每个分配方案计算:
//    - Drift = Σ[(Q_i(t+1))² - (Q_i(t))²] (队列长度变化)
//    - Penalty = α×Delay + β×Energy + γ×Load (性能指标)
//    - Cost = Drift + V × Penalty
// 3. 选择Cost最小的分配方案
func (ls *LyapunovScheduler) Schedule(timeSlot uint, tasks []*define.Task) []*define.Assignment {
	if len(tasks) == 0 {
		return nil
	}

	bestCost := math.MaxFloat64
	var bestAssignments []*define.Assignment

	// 多次迭代寻找最优分配
	for iter := 0; iter < constant.Iters; iter++ {
		// 生成一个候选分配方案
		candidateAssignments := ls.generateCandidateAssignments(timeSlot, tasks, iter)

		// 计算该方案的Lyapunov cost
		cost := ls.computeLyapunovCost(candidateAssignments, tasks)

		// 更新最优解
		if cost < bestCost {
			bestCost = cost
			bestAssignments = candidateAssignments
		}

		// 早停: 如果cost已经很小，提前结束
		if iter > 5 && bestCost < constant.Bias {
			break
		}
	}

	// 计算资源分配比例
	ls.allocateResources(bestAssignments)

	log.Printf("⚡ Lyapunov调度 (时隙%d): 找到最优cost=%.2f (迭代%d次)", timeSlot, bestCost, constant.Iters)
	return bestAssignments
}

// generateCandidateAssignments 生成一个候选分配方案
// 策略:
//   - iter=0: 优先复用上次分配 (减少任务迁移)
//   - iter>0: 随机分配 + 部分保留 (探索更优解)
func (ls *LyapunovScheduler) generateCandidateAssignments(timeSlot uint, tasks []*define.Task, iter int) []*define.Assignment {
	assignments := make([]*define.Assignment, 0, len(tasks))

	for _, task := range tasks {
		var assign *define.Assignment

		// 第一次迭代: 优先复用上次分配
		if iter == 0 {
			lastAssign := ls.AssignmentManager.GetLastAssignment(task.ID)
			if lastAssign != nil && (task.Status == define.TaskQueued || task.Status == define.TaskComputing) {
				assign = ls.reuseAssignment(timeSlot, task, lastAssign)
			}
		}

		// 如果没有复用，则随机分配到某个通信设备
		if assign == nil {
			assign = ls.randomAssignment(timeSlot, task)
		}

		if assign != nil {
			assignments = append(assignments, assign)
		}
	}

	return assignments
}

// reuseAssignment 复用上次的分配
func (ls *LyapunovScheduler) reuseAssignment(timeSlot uint, task *define.Task, lastAssign *define.Assignment) *define.Assignment {
	queue := ls.AssignmentManager.GetCurrentQueue(task.ID, task.DataSize)

	return &define.Assignment{
		TimeSlot:              timeSlot,
		TaskID:                task.ID,
		CommID:                lastAssign.CommID,
		Path:                  lastAssign.Path,
		Speeds:                lastAssign.Speeds,
		Powers:                lastAssign.Powers,
		QueueData:             queue,
		CumulativeTransferred: lastAssign.CumulativeTransferred,
		CumulativeProcessed:   lastAssign.CumulativeProcessed,
		ResourceFraction:      0, // 稍后计算
		TransferredData:       0, // 稍后计算
		ProcessedData:         0, // 稍后计算
	}
}

// randomAssignment 随机分配任务到某个通信设备
func (ls *LyapunovScheduler) randomAssignment(timeSlot uint, task *define.Task) *define.Assignment {
	user, ok := ls.System.UserMap[task.UserID]
	if !ok {
		return nil
	}

	// 随机选择一个通信设备
	commIDs := make([]uint, 0, len(ls.System.CommMap))
	for commID := range ls.System.CommMap {
		commIDs = append(commIDs, commID)
	}
	if len(commIDs) == 0 {
		return nil
	}
	commID := commIDs[rand.Intn(len(commIDs))]

	// 计算路径
	path := ls.getPath(task.UserID, commID)
	if len(path) < 2 {
		return nil
	}

	// 获取路径速率和功率
	speeds, powers := ls.getPathSpeedsAndPowers(path, user.Speed)

	// 获取当前队列状态
	lastAssign := ls.AssignmentManager.GetLastAssignment(task.ID)
	queue := 0.0
	transferred := 0.0
	processed := 0.0
	if lastAssign != nil {
		queue = ls.AssignmentManager.GetCurrentQueue(task.ID, task.DataSize)
		transferred = lastAssign.CumulativeTransferred
		processed = lastAssign.CumulativeProcessed
	}

	return &define.Assignment{
		TimeSlot:              timeSlot,
		TaskID:                task.ID,
		CommID:                commID,
		Path:                  path,
		Speeds:                speeds,
		Powers:                powers,
		QueueData:             queue,
		CumulativeTransferred: transferred,
		CumulativeProcessed:   processed,
		ResourceFraction:      0,
		TransferredData:       0,
		ProcessedData:         0,
	}
}

// computeLyapunovCost 计算Lyapunov cost = Drift + V × Penalty
func (ls *LyapunovScheduler) computeLyapunovCost(assignments []*define.Assignment, tasks []*define.Task) float64 {
	// 1. 预测执行后的状态
	predictedState := ls.predictState(assignments, tasks)

	// 2. 计算Drift (队列长度的二次变化)
	drift := ls.computeDrift(predictedState)

	// 3. 计算Penalty (性能指标: 延迟 + 能耗 + 负载)
	penalty := ls.computePenalty(predictedState)

	// 4. Lyapunov cost
	cost := drift + ls.V*penalty

	return cost
}

// predictState 预测执行assignments后的系统状态
func (ls *LyapunovScheduler) predictState(assignments []*define.Assignment, tasks []*define.Task) *define.StateMetrics {
	state := define.NewStateMetrics()

	// 为每个assignment计算传输量和处理量
	taskMap := make(map[string]*define.Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	// 先计算总队列，用于资源分配
	totalQueue := 0.0
	for _, assign := range assignments {
		totalQueue += assign.QueueData
	}

	// 临时分配资源 (简化版，后续会被allocateResources覆盖)
	for _, assign := range assignments {
		if totalQueue > 0 {
			assign.ResourceFraction = assign.QueueData / totalQueue
		} else {
			assign.ResourceFraction = 1.0 / float64(len(assignments))
		}
	}

	// 计算每个assignment的传输量和处理量
	for _, assign := range assignments {
		task := taskMap[assign.TaskID]
		if task == nil {
			continue
		}

		// 传输量 (用户以固定速率传输)
		userSpeed := 0.0
		if len(assign.Speeds) > 0 {
			userSpeed = assign.Speeds[0]
		}
		transferred := userSpeed * constant.Slot

		remaining := task.DataSize - assign.CumulativeTransferred
		if transferred > remaining {
			transferred = remaining
		}
		if transferred < 0 {
			transferred = 0
		}

		// 处理量
		processed := 0.0
		if assign.ResourceFraction > 0 && assign.QueueData > 0 {
			processingCapacity := assign.ResourceFraction * constant.C / constant.Rho * constant.Slot
			processed = math.Min(assign.QueueData, processingCapacity)
		}

		assign.TransferredData = transferred
		assign.ProcessedData = processed

		// 更新通信设备队列 (Queue_i(t+1) = Queue_i(t) + 传输 - 处理)
		newQueue := assign.QueueData + transferred - processed
		if newQueue < 0 {
			newQueue = 0
		}
		state.CommQueues[assign.CommID] += newQueue
		state.TotalQueue += newQueue

		// 累加延迟和能耗
		state.TransferDelay += ls.computeTransferDelay(assign)
		state.ComputeDelay += ls.computeComputeDelay(assign)
		state.TransferEnergy += ls.computeTransferEnergy(assign)
		state.ComputeEnergy += ls.computeComputeEnergy(assign)
	}

	state.TotalDelay = state.TransferDelay + state.ComputeDelay
	state.TotalEnergy = state.TransferEnergy + state.ComputeEnergy
	state.Load = state.TotalQueue

	return state
}

// computeDrift 计算Lyapunov drift = Σ[(Q_i(t+1))² - (Q_i(t))²]
func (ls *LyapunovScheduler) computeDrift(predictedState *define.StateMetrics) float64 {
	drift := 0.0

	// 遍历所有通信设备
	for commID, newQueue := range predictedState.CommQueues {
		oldQueue := ls.lastCommQueues[commID]

		// Drift_i = Q_i(t+1)² - Q_i(t)²
		drift += (newQueue*newQueue - oldQueue*oldQueue)
	}

	return drift / constant.Shrink // 归一化 (防止数值过大)
}

// computePenalty 计算penalty = α×Delay + β×Energy + γ×Load
func (ls *LyapunovScheduler) computePenalty(state *define.StateMetrics) float64 {
	// 使用constant中定义的权重
	penalty := constant.Alpha*state.TotalDelay +
		constant.Beta*state.TotalEnergy +
		constant.Gamma*state.Load

	return penalty / constant.Shrink // 归一化
}

// computeTransferDelay 计算传输延迟 (使用你的原公式)
func (ls *LyapunovScheduler) computeTransferDelay(assign *define.Assignment) float64 {
	delay := 0.0
	for _, speed := range assign.Speeds {
		if speed > 0 && assign.TransferredData > 0 {
			delay += assign.TransferredData / speed
		}
	}
	return delay
}

// computeComputeDelay 计算计算延迟 (使用你的原公式)
func (ls *LyapunovScheduler) computeComputeDelay(assign *define.Assignment) float64 {
	if assign.ResourceFraction > 0 && assign.ProcessedData > 0 {
		return assign.ProcessedData * constant.Rho / (assign.ResourceFraction * constant.C)
	}
	return 0
}

// computeTransferEnergy 计算传输能耗 (使用你的原公式)
func (ls *LyapunovScheduler) computeTransferEnergy(assign *define.Assignment) float64 {
	energy := 0.0
	for i, power := range assign.Powers {
		if i < len(assign.Speeds) && assign.Speeds[i] > 0 && assign.TransferredData > 0 {
			transmissionTime := assign.TransferredData / assign.Speeds[i]
			energy += power * transmissionTime
		}
	}
	return energy
}

// computeComputeEnergy 计算计算能耗 (使用你的原公式)
func (ls *LyapunovScheduler) computeComputeEnergy(assign *define.Assignment) float64 {
	if assign.ResourceFraction > 0 {
		return assign.ResourceFraction * constant.Kappa * constant.C * constant.C * constant.C * constant.Slot
	}
	return 0
}

// allocateResources 分配资源比例 (基于队列长度和优先级)
func (ls *LyapunovScheduler) allocateResources(assignments []*define.Assignment) {
	if len(assignments) == 0 {
		return
	}

	// 按通信设备分组
	commGroups := make(map[uint][]*define.Assignment)
	for _, assign := range assignments {
		commGroups[assign.CommID] = append(commGroups[assign.CommID], assign)
	}

	// 对每个通信设备，按权重分配资源
	for _, group := range commGroups {
		totalWeight := 0.0
		weights := make([]float64, len(group))

		for i, assign := range group {
			task := ls.AssignmentManager.GetTask(assign.TaskID, ls.System.TaskManager)
			if task == nil {
				continue
			}

			// 权重 = 优先级因子 × 队列因子
			priorityFactor := float64(task.Priority)/10.0 + 1.0
			queueFactor := assign.QueueData + 1.0

			// 饥饿提升
			if task.IsStarving() {
				waitTime := task.GetWaitTime().Seconds()
				priorityFactor *= (1.0 + waitTime/10.0)
			}

			weight := priorityFactor * queueFactor
			weights[i] = weight
			totalWeight += weight
		}

		// 分配资源
		if totalWeight < 0.001 {
			// 平均分配
			for _, assign := range group {
				assign.ResourceFraction = 1.0 / float64(len(group))
			}
		} else {
			for i, assign := range group {
				assign.ResourceFraction = weights[i] / totalWeight
			}
		}
	}
}

// ExecuteAssignments 执行分配，计算实际传输和处理量，更新队列状态
func (ls *LyapunovScheduler) ExecuteAssignments(assignments []*define.Assignment, tasks map[string]*define.Task) {
	// 清空上次队列状态
	ls.lastCommQueues = make(map[uint]float64)

	for _, assign := range assignments {
		task := tasks[assign.TaskID]
		if task == nil {
			continue
		}

		// 计算传输量
		userSpeed := 0.0
		if len(assign.Speeds) > 0 {
			userSpeed = assign.Speeds[0]
		}
		transferred := userSpeed * constant.Slot

		remaining := task.DataSize - assign.CumulativeTransferred
		if transferred > remaining {
			transferred = remaining
		}
		if transferred < 0 {
			transferred = 0
		}

		// 计算处理量
		processed := 0.0
		if assign.ResourceFraction > 0 && assign.QueueData > 0 {
			processingCapacity := assign.ResourceFraction * constant.C / constant.Rho * constant.Slot
			processed = math.Min(assign.QueueData, processingCapacity)
		}

		// 更新assignment
		assign.TransferredData = transferred
		assign.ProcessedData = processed
		assign.CumulativeTransferred += transferred
		assign.CumulativeProcessed += processed

		// 更新队列状态 (用于下次drift计算)
		newQueue := assign.QueueData + transferred - processed
		if newQueue < 0 {
			newQueue = 0
		}
		ls.lastCommQueues[assign.CommID] += newQueue
	}
}

// getPath 获取从用户到通信设备的最短路径
func (ls *LyapunovScheduler) getPath(userID, commID uint) []uint {
	if ls.System.FloydResult == nil {
		return []uint{userID, commID}
	}

	srcIdx, srcOk := ls.System.NodeIDToIndex[userID]
	dstIdx, dstOk := ls.System.NodeIDToIndex[commID]

	if !srcOk || !dstOk {
		return []uint{userID, commID}
	}

	pathIndices := ls.System.FloydResult.Paths[srcIdx][dstIdx]
	if len(pathIndices) == 0 {
		return nil
	}

	path := make([]uint, len(pathIndices))
	for i, idx := range pathIndices {
		path[i] = ls.System.IndexToNodeID[idx]
	}

	return path
}

// getPathSpeedsAndPowers 获取路径的速率和功率
func (ls *LyapunovScheduler) getPathSpeedsAndPowers(path []uint, userSpeed float64) ([]float64, []float64) {
	speeds := make([]float64, len(path)-1)
	powers := make([]float64, len(path)-1)

	for i := 0; i < len(path)-1; i++ {
		srcID := path[i]
		dstID := path[i+1]

		if i == 0 {
			// 第一段: 用户 → 设备
			speeds[i] = userSpeed
			powers[i] = constant.P_u
		} else {
			// 后续段: 设备 → 设备
			link, exists := ls.System.LinkMap[[2]uint{srcID, dstID}]
			if !exists {
				speeds[i] = 10.0 // 默认值 (Mbps)
				powers[i] = constant.P_b
				continue
			}

			bandwidth := 10.0
			if bw, ok := link.Properties["bandwidth"].(float64); ok && bw > 0 {
				bandwidth = bw
			}
			speeds[i] = bandwidth

			power := constant.P_b
			if pw, ok := link.Properties["power"].(float64); ok && pw > 0 {
				power = pw
			}
			powers[i] = power
		}
	}

	return speeds, powers
}
