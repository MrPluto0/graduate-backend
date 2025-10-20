package algorithm

import (
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/define"
	"math"
)

// TaskAllocation 单个任务的资源分配方案
type TaskAllocation struct {
	TaskID         string `json:"task_id"`       // 任务ID
	UserIndex      int    `json:"-"`             // 对应的用户索引（内部使用）
	AssignedCommID uint   `json:"assigned_comm"` // 分配的计算节点ID
	TransferPath   []uint `json:"transfer_path"` // 传输路径（comm ID列表）

	// 内部使用的索引（不序列化）
	assignedCommIdx int   `json:"-"` // 内部使用的comm索引
	transferPathIdx []int `json:"-"` // 内部使用的path索引

	// 队列状态
	Q     float64 `json:"q"`      // 当前队列长度
	QNext float64 `json:"q_next"` // 下一时隙队列长度
	QMid  float64 `json:"q_mid"`  // 中间队列状态

	// 资源分配
	R float64 `json:"r"` // 待处理数据量
	F float64 `json:"f"` // 分配的计算资源比例

	// 传输参数
	V []float64 `json:"v"` // 每段路径的传输速率
	P []float64 `json:"p"` // 每段路径的传输功率

	// 性能指标
	TransferDelay  float64 `json:"transfer_delay"`  // 传输延迟
	ComputeDelay   float64 `json:"compute_delay"`   // 计算延迟
	TransferEnergy float64 `json:"transfer_energy"` // 传输能耗
	ComputeEnergy  float64 `json:"compute_energy"`  // 计算能耗
}

// TaskState 任务维度的系统状态
type TaskState struct {
	T           uint                       `json:"t"`           // 系统时间
	Allocations map[string]*TaskAllocation `json:"allocations"` // 任务分配映射 taskID -> allocation

	// 每个通信设备的总队列
	CommQueues []float64 `json:"comm_queues"` // 每个comm的总队列长度

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

// NewTaskState 创建任务维度状态
func NewTaskState(t uint, tasks []*define.Task, numComms int) *TaskState {
	allocations := make(map[string]*TaskAllocation)

	for _, task := range tasks {
		allocations[task.TaskID] = &TaskAllocation{
			TaskID:          task.TaskID,
			UserIndex:       task.UserIndex,
			assignedCommIdx: -1,
			AssignedCommID:  0,
			TransferPath:    make([]uint, 0),
			transferPathIdx: make([]int, 0),
			R:               task.DataSize - task.ProcessedData,
			Q:               task.QueuedData,
			V:               make([]float64, 0),
			P:               make([]float64, 0),
		}
	}

	return &TaskState{
		T:           t,
		Allocations: allocations,
		CommQueues:  make([]float64, numComms),
	}
}

// Copy 深拷贝
func (ts *TaskState) Copy() *TaskState {
	newState := &TaskState{
		T:              ts.T,
		Allocations:    make(map[string]*TaskAllocation),
		CommQueues:     append([]float64(nil), ts.CommQueues...),
		Cost:           ts.Cost,
		Drift:          ts.Drift,
		Penalty:        ts.Penalty,
		TotalDelay:     ts.TotalDelay,
		TotalEnergy:    ts.TotalEnergy,
		Load:           ts.Load,
		TransferDelay:  ts.TransferDelay,
		ComputeDelay:   ts.ComputeDelay,
		TransferEnergy: ts.TransferEnergy,
		ComputeEnergy:  ts.ComputeEnergy,
	}

	for taskID, alloc := range ts.Allocations {
		newState.Allocations[taskID] = &TaskAllocation{
			TaskID:          alloc.TaskID,
			UserIndex:       alloc.UserIndex,
			assignedCommIdx: alloc.assignedCommIdx,
			AssignedCommID:  alloc.AssignedCommID,
			transferPathIdx: append([]int(nil), alloc.transferPathIdx...),
			TransferPath:    append([]uint(nil), alloc.TransferPath...),
			Q:               alloc.Q,
			QNext:           alloc.QNext,
			QMid:            alloc.QMid,
			R:               alloc.R,
			F:               alloc.F,
			V:               append([]float64(nil), alloc.V...),
			P:               append([]float64(nil), alloc.P...),
			TransferDelay:   alloc.TransferDelay,
			ComputeDelay:    alloc.ComputeDelay,
			TransferEnergy:  alloc.TransferEnergy,
			ComputeEnergy:   alloc.ComputeEnergy,
		}
	}

	return newState
}

// AssignTask 为任务分配通信设备和路径（内部使用索引，对外暴露ID）
func (ts *TaskState) AssignTask(taskID string, commIdx int, path []int, userSpeed float64, commSpeeds [][]float64, sys *System) {
	alloc, ok := ts.Allocations[taskID]
	if !ok {
		return
	}

	// 保存索引（内部使用）
	alloc.assignedCommIdx = commIdx
	alloc.transferPathIdx = append([]int(nil), path...)

	// 同步ID（对外暴露）
	if commIdx >= 0 && commIdx < len(sys.Comms) {
		alloc.AssignedCommID = sys.Comms[commIdx].ID
	}
	alloc.TransferPath = make([]uint, len(path))
	for i, idx := range path {
		if idx >= 0 && idx < len(sys.Comms) {
			alloc.TransferPath[i] = sys.Comms[idx].ID
		}
	}

	// 计算每段路径的传输速率和功率
	alloc.V = make([]float64, len(path))
	alloc.P = make([]float64, len(path))

	for idx, j := range path {
		if idx == 0 {
			// 第一段：用户到第一个通信设备
			alloc.V[idx] = userSpeed
			alloc.P[idx] = constant.P_u
		} else {
			// 后续段：通信设备之间
			prevJ := path[idx-1]
			alloc.V[idx] = commSpeeds[prevJ][j]
			alloc.P[idx] = constant.P_b
		}
	}
}

// 更新计算资源分配
func (ts *TaskState) UpdateResourceAlloc() {
	numComms := len(ts.CommQueues)

	// 第一步：计算 Q_mid 并累计每个通信设备的总队列需求
	commTotalQMid := make([]float64, numComms)
	for _, alloc := range ts.Allocations {
		if alloc.assignedCommIdx >= 0 {
			alloc.QMid = alloc.Q + alloc.R
			commTotalQMid[alloc.assignedCommIdx] += alloc.QMid
		}
	}

	// 第二步：按比例分配计算资源
	for _, alloc := range ts.Allocations {
		if alloc.assignedCommIdx >= 0 && commTotalQMid[alloc.assignedCommIdx] > 0 {
			alloc.F = alloc.QMid / commTotalQMid[alloc.assignedCommIdx]
		}
	}
}

// ComputeNextQueue 计算下一时刻队列长度
func (ts *TaskState) ComputeNextQueue() {
	for _, alloc := range ts.Allocations {
		if alloc.assignedCommIdx >= 0 {
			// R: 收到的数据
			R := alloc.R
			// F: 能够处理的数据
			F := alloc.F * constant.C * constant.Slot / constant.Rho
			// Q_next = Q + R - F
			alloc.QNext = alloc.Q + R - F

			if alloc.QNext < 0 {
				alloc.QNext = 0
			}
		}
	}

	// 更新每个通信设备的总队列
	for i := range ts.CommQueues {
		ts.CommQueues[i] = 0
	}
	for _, alloc := range ts.Allocations {
		if alloc.assignedCommIdx >= 0 {
			ts.CommQueues[alloc.assignedCommIdx] += alloc.QNext
		}
	}
}

// ComputeMetrics 计算性能指标
func (ts *TaskState) ComputeMetrics() {
	ts.TransferDelay = 0.0
	ts.ComputeDelay = 0.0
	ts.TransferEnergy = 0.0
	ts.ComputeEnergy = 0.0
	ts.Drift = 0.0

	for _, alloc := range ts.Allocations {
		if alloc.assignedCommIdx < 0 {
			continue
		}

		// 传输延迟：每段路径的延迟之和
		alloc.TransferDelay = 0.0
		for _, v := range alloc.V {
			if v > 0 {
				alloc.TransferDelay += alloc.R / v
			}
		}
		ts.TransferDelay += constant.Alpha * alloc.TransferDelay

		// 计算延迟
		if alloc.F > 0 {
			alloc.ComputeDelay = alloc.QMid * constant.Rho / (alloc.F * constant.C)
			ts.ComputeDelay += constant.Alpha * alloc.ComputeDelay
		}

		// 传输能耗：每段路径的能耗之和
		alloc.TransferEnergy = 0.0
		for i, p := range alloc.P {
			if i < len(alloc.V) && alloc.V[i] > 0 {
				segmentDelay := alloc.R / alloc.V[i]
				alloc.TransferEnergy += p * segmentDelay
			}
		}
		ts.TransferEnergy += constant.Gamma * alloc.TransferEnergy

		// 计算能耗
		alloc.ComputeEnergy = alloc.F * constant.Kappa * math.Pow(constant.C, 3) * constant.Slot
		ts.ComputeEnergy += constant.Gamma * alloc.ComputeEnergy

		// 漂移项: Q * (R - F)
		R := alloc.R
		F := alloc.F * constant.C * constant.Slot / constant.Rho
		drift := alloc.Q * (R - F) / (constant.Shrink * constant.Shrink)
		ts.Drift += drift
	}

	// 计算负载：所有通信设备的平均队列
	ts.Load = 0.0
	for _, alloc := range ts.Allocations {
		if alloc.assignedCommIdx >= 0 {
			avgQ := (alloc.QNext + alloc.Q) / 2
			ts.Load += avgQ
		}
	}
	ts.Load = constant.Beta * ts.Load / constant.Shrink

	// 总延迟和能耗
	ts.TotalDelay = ts.TransferDelay + ts.ComputeDelay
	ts.TotalEnergy = ts.TransferEnergy + ts.ComputeEnergy

	// 惩罚项
	ts.Penalty = ts.TransferDelay + ts.ComputeDelay + ts.TransferEnergy + ts.ComputeEnergy + ts.Load

	// 总成本
	ts.Cost = ts.Drift + constant.V*ts.Penalty
}

// Objective 计算并返回目标函数值
func (ts *TaskState) Objective() float64 {
	ts.UpdateResourceAlloc()
	ts.ComputeNextQueue()
	ts.ComputeMetrics()
	return ts.Cost
}
