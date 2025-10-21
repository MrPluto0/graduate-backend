package algorithm

import (
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/define"
	"math"
)

// TaskState 系统在某个时刻的调度状态（临时计算用）
type TaskState struct {
	T         uint                            `json:"t"`         // 系统时间
	Snapshots map[string]*define.TaskSnapshot `json:"snapshots"` // 任务快照 taskID -> snapshot

	// 每个通信设备的总队列 (使用 map 而不是数组)
	CommQueues map[uint]float64 `json:"comm_queues"` // commID -> 队列长度

	// 全局性能指标
	Cost           float64 `json:"cost"`            // 总成本
	Drift          float64 `json:"drift"`           // 漂移值
	Penalty        float64 `json:"penalty"`         // 惩罚项
	TotalDelay     float64 `json:"total_delay"`     // 总延迟
	TotalEnergy    float64 `json:"total_energy"`    // 总能耗
	Load           float64 `json:"load"`            // 系统负载
	TransferDelay  float64 `json:"transfer_delay"`  // 总传输延迟
	ComputeDelay   float64 `json:"compute_delay"`   // 总计算延迟
	TransferEnergy float64 `json:"transfer_energy"` // 总传输能耗
	ComputeEnergy  float64 `json:"compute_energy"`  // 总计算能耗
}

// NewTaskState 创建任务维度状态（从Task创建快照）
func NewTaskState(t uint, tasks []*define.Task, sys *System) *TaskState {
	snapshots := make(map[string]*define.TaskSnapshot)

	for _, task := range tasks {
		// 计算本时隙需要传输的数据量
		pendingTransferData := 0.0
		if task.Status == define.TaskPending {
			// 未分配：需要传输全部剩余数据
			pendingTransferData = task.DataSize - task.ProcessedData
		}

		snapshot := &define.TaskSnapshot{
			TaskID:              task.TaskID,
			AssignedCommID:      task.AssignedCommID,      // 继承已有的分配
			TransferPath:        task.TransferPath.Copy(), // 继承已有路径（深拷贝）
			PendingTransferData: pendingTransferData,
			CurrentQueue:        task.QueuedData, // 继承当前队列
		}

		snapshots[task.TaskID] = snapshot
	}

	// 初始化每个通信设备的队列 (使用 map)
	commQueues := make(map[uint]float64)
	for _, comm := range sys.Comms {
		commQueues[comm.ID] = 0
	}

	return &TaskState{
		T:          t,
		Snapshots:  snapshots,
		CommQueues: commQueues,
	}
}

// Copy 深拷贝
func (ts *TaskState) Copy() *TaskState {
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
func (ts *TaskState) AssignTask(taskID string, assignedCommID uint, transferPath *define.TransferPath, userSpeed float64) {
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
func (ts *TaskState) UpdateResourceAlloc() {
	// 第一步：计算中间队列并累计每个通信设备的总队列需求
	commTotalQMid := make(map[uint]float64)
	for _, snap := range ts.Snapshots {
		if snap.AssignedCommID > 0 {
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
func (ts *TaskState) ComputeNextQueue() {
	for _, snap := range ts.Snapshots {
		if snap.AssignedCommID > 0 {
			// 下一时隙队列 = 当前队列 + 收到数据 - 处理数据
			receivedData := snap.PendingTransferData
			processedData := snap.ResourceFraction * constant.C * constant.Slot / constant.Rho
			snap.NextQueue = snap.CurrentQueue + receivedData - processedData

			if snap.NextQueue < 0 {
				snap.NextQueue = 0
			}
		}
	}

	// 更新每个通信设备的总队列
	for commID := range ts.CommQueues {
		ts.CommQueues[commID] = 0
	}
	for _, snap := range ts.Snapshots {
		if snap.AssignedCommID > 0 {
			ts.CommQueues[snap.AssignedCommID] += snap.NextQueue
		}
	}
}

// 计算性能指标
func (ts *TaskState) ComputeMetrics() {
	ts.TransferDelay = 0.0
	ts.ComputeDelay = 0.0
	ts.TransferEnergy = 0.0
	ts.ComputeEnergy = 0.0
	ts.Load = 0.0
	ts.Drift = 0.0

	for _, snap := range ts.Snapshots {
		if snap.AssignedCommID == 0 || snap.TransferPath == nil {
			continue
		}

		// 创建局部 metrics 变量
		metrics := define.TaskMetrics{}

		// 传输延迟：每段路径的延迟之和
		for _, speed := range snap.TransferPath.Speeds {
			if speed > 0 {
				metrics.TransferDelay += snap.PendingTransferData / speed
			}
		}
		ts.TransferDelay += metrics.TransferDelay

		// 计算延迟
		if snap.ResourceFraction > 0 {
			metrics.ComputeDelay = snap.IntermediateQueue * constant.Rho / (snap.ResourceFraction * constant.C)
			ts.ComputeDelay += metrics.ComputeDelay
		}

		// 传输能耗：每段路径的能耗之和
		for i, power := range snap.TransferPath.Powers {
			if i < len(snap.TransferPath.Speeds) && snap.TransferPath.Speeds[i] > 0 {
				segmentDelay := snap.PendingTransferData / snap.TransferPath.Speeds[i]
				metrics.TransferEnergy += power * segmentDelay
			}
		}
		ts.TransferEnergy += metrics.TransferEnergy

		// 计算能耗
		metrics.ComputeEnergy = snap.ResourceFraction * constant.Kappa * math.Pow(constant.C, 3) * constant.Slot
		ts.ComputeEnergy += metrics.ComputeEnergy

		// 总延迟和总能耗
		metrics.TotalDelay = metrics.TransferDelay + metrics.ComputeDelay
		metrics.TotalEnergy = metrics.TransferEnergy + metrics.ComputeEnergy

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

	// 总延迟和能耗
	ts.TotalDelay = ts.TransferDelay + ts.ComputeDelay
	ts.TotalEnergy = ts.TransferEnergy + ts.ComputeEnergy

	// 惩罚项（在这里统一添加系数）
	ts.Penalty = constant.Alpha*ts.TotalDelay + constant.Gamma*ts.TotalEnergy + constant.Beta*ts.Load

	// 总成本
	ts.Cost = ts.Drift + constant.V*ts.Penalty
}

// 计算并返回目标函数值
func (ts *TaskState) Objective() float64 {
	ts.UpdateResourceAlloc()
	ts.ComputeNextQueue()
	ts.ComputeMetrics()
	return ts.Cost
}
