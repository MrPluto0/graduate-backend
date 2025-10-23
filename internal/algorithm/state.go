package algorithm

import (
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/define"
)

// 系统在某个时刻的调度状态（临时计算用）
type State struct {
	TimeSlot  uint                            `json:"-"` // 系统时间
	Snapshots map[string]*define.TaskSnapshot `json:"-"` // 任务快照 taskID -> snapshot

	// 全局性能指标（匿名嵌入 TaskMetrics 避免重复）
	define.TaskMetrics
	CommQueues map[uint]float64 `json:"comm_queues"` // 每个通信设备的总队列
	TotalQueue float64          `json:"total_queue"` // 总队列长度
	Load       float64          `json:"load"`        // 系统负载
	Cost       float64          `json:"cost"`        // 总成本
	Drift      float64          `json:"drift"`       // 漂移值
	Penalty    float64          `json:"penalty"`     // 惩罚项
}

// 创建任务维度状态（从Task创建快照）
func NewState(t uint, tasks []*define.Task, sys *System) *State {
	snapshots := make(map[string]*define.TaskSnapshot)

	for _, task := range tasks {
		// 计算剩余待传输数据 = 总数据 - 已计算 - 队列中
		pendingTransferData := task.DataSize - task.ProcessedData - task.QueuedData
		if pendingTransferData < 0 {
			pendingTransferData = 0
		}

		snapshot := &define.TaskSnapshot{
			ID:                  task.ID,
			AssignedCommID:      task.AssignedCommID,      // 继承已有的分配
			TransferPath:        task.TransferPath.Copy(), // 继承已有路径（深拷贝）
			PendingTransferData: pendingTransferData,
			CurrentQueue:        task.QueuedData, // 继承当前队列
		}

		snapshots[task.ID] = snapshot
	}

	return &State{
		TimeSlot:   t,
		Snapshots:  snapshots,
		CommQueues: make(map[uint]float64),
	}
}

// 深拷贝
func (ts *State) copy() *State {
	newState := *ts
	newState.CommQueues = make(map[uint]float64, len(ts.CommQueues))
	for id, q := range ts.CommQueues {
		newState.CommQueues[id] = q
	}

	// 深拷贝 Snapshots map
	newState.Snapshots = make(map[string]*define.TaskSnapshot, len(ts.Snapshots))
	for taskID, snap := range ts.Snapshots {
		newSnap := *snap                                // 浅拷贝所有基本类型
		newSnap.TransferPath = snap.TransferPath.Copy() // 深拷贝 TransferPath
		newState.Snapshots[taskID] = &newSnap
	}

	return &newState
}

// 为任务分配计算设备,并设置传输路径
func (ts *State) assignTask(taskID string, assignedCommID uint, transferPath *define.TransferPath, userSpeed float64) {
	snap, ok := ts.Snapshots[taskID]
	if !ok {
		return
	}

	// 保存分配的通信设备
	snap.AssignedCommID = assignedCommID

	// 复制 TransferPath 并填充第一段的用户速度
	snap.TransferPath = transferPath.Copy()
	if len(snap.TransferPath.Speeds) > 0 {
		snap.TransferPath.Speeds[0] = userSpeed
		snap.TransferPath.Powers[0] = constant.P_u
	}
}

// 更新计算资源分配
func (ts *State) updateResourceAlloc() {
	// 第一步：计算中间队列并累计每个通信设备的总队列需求
	commTotalQMid := make(map[uint]float64)
	for _, snap := range ts.Snapshots {
		if snap.AssignedCommID > 0 {
			// 预测阶段：假设数据瞬间到达（不限时隙）
			snap.IntermediateQueue = snap.CurrentQueue + snap.PendingTransferData
			commTotalQMid[snap.AssignedCommID] += snap.IntermediateQueue
		}
	}

	// 第二步：按比例分配计算资源
	for _, snap := range ts.Snapshots {
		if snap.AssignedCommID > 0 {
			if totalQMid := commTotalQMid[snap.AssignedCommID]; totalQMid > 0 {
				snap.ResourceFraction = snap.IntermediateQueue / totalQMid
			}
		}
	}
}

// 计算下一时刻队列长度
func (ts *State) computeNextQueue() {
	for _, snap := range ts.Snapshots {
		if snap.AssignedCommID > 0 {
			// 预测阶段：假设数据瞬间到达（不限时隙）
			receivedData := snap.PendingTransferData
			processedData := snap.ResourceFraction * constant.C * constant.Slot / constant.Rho
			snap.NextQueue = snap.CurrentQueue + receivedData - processedData

			if snap.NextQueue < 0 {
				snap.NextQueue = 0
			}
		}
	}

	// 更新每个通信设备的总队列
	for _, snap := range ts.Snapshots {
		if snap.AssignedCommID > 0 {
			ts.CommQueues[snap.AssignedCommID] += snap.NextQueue
			ts.TotalQueue += snap.NextQueue
		}
	}
}

// 计算性能指标
func (ts *State) computeMetrics() {
	for _, snap := range ts.Snapshots {
		if snap.AssignedCommID == 0 || snap.TransferPath == nil {
			continue
		}

		// 使用 snapshot 自身的方法计算 metrics（预测阶段）
		metrics := snap.ComputeMetrics(
			snap.PendingTransferData, // 传输的数据量
			snap.IntermediateQueue,   // 队列数据量
		)

		// 累加到状态总指标
		ts.TransferDelay += metrics.TransferDelay
		ts.ComputeDelay += metrics.ComputeDelay
		ts.TransferEnergy += metrics.TransferEnergy
		ts.ComputeEnergy += metrics.ComputeEnergy

		// 赋值给 snapshot
		snap.Metrics = metrics

		ts.Load += (snap.NextQueue + snap.CurrentQueue) / 2

		// Lyapunov 漂移项: 使用队列长度的平方差
		// Drift = (Q_next^2 - Q^2) / 2
		queueDiff := (snap.NextQueue*snap.NextQueue - snap.CurrentQueue*snap.CurrentQueue) / 2.0
		ts.Drift += queueDiff / (constant.Shrink * constant.Shrink)
	}

	// 计算负载：所有通信设备的平均队列
	ts.Load = ts.Load / constant.Shrink

	// 总延迟和能耗（直接访问嵌入字段）
	ts.TotalDelay = ts.TransferDelay + ts.ComputeDelay
	ts.TotalEnergy = ts.TransferEnergy + ts.ComputeEnergy

	// 惩罚项（在这里统一添加系数）
	ts.Penalty = constant.Alpha*ts.TotalDelay + constant.Gamma*ts.TotalEnergy + constant.Beta*ts.Load

	// 总成本
	ts.Cost = ts.Drift + constant.V*ts.Penalty
}

// 计算并返回目标函数值
func (ts *State) objective() float64 {
	ts.updateResourceAlloc()
	ts.computeNextQueue()
	ts.computeMetrics()
	return ts.Cost
}
