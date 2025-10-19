package algorithm

import (
	"math"

	"go-backend/internal/algorithm/constant"
)

// State 系统状态结构体
type State struct {
	// 基本状态信息
	T uint      `json:"t"` // 系统时间
	R []float64 `json:"r"` // 用户待处理数据

	// 队列状态
	Q     [][]float64 `json:"Q"`      // 当前队列长度
	F     [][]float64 `json:"F"`      // 计算资源分配比例
	QNext [][]float64 `json:"Q_next"` // 下一时刻队列长度
	QMid  [][]float64 `json:"Q_mid"`  // 中间队列状态（Q + delta * r）

	// 决策变量
	Delta   [][]float64 `json:"delta"`   // 通信资源分配（0或1）
	Epsilon [][]float64 `json:"epsilon"` // 卸载决策（0或1）

	TransferPath [][]int `json:"transfer_path"` // 数据传输路径

	// 辅助变量
	V [][]float64 `json:"v"` // 传输速率矩阵
	P [][]float64 `json:"p"` // 传输功率矩阵

	// 性能指标
	Cost    float64 `json:"cost"`    // 总成本
	Drift   float64 `json:"drift"`   // 漂移值
	Penalty float64 `json:"penalty"` // 惩罚项

	// 惩罚项指标
	ComputeEnergy  float64 `json:"compute_energy"`  // 计算能耗
	TransferEnergy float64 `json:"transfer_energy"` // 传输能耗
	ComputeDelay   float64 `json:"compute_delay"`   // 计算延迟
	TransferDelay  float64 `json:"transfer_delay"`  // 传输延迟
	Load           float64 `json:"load"`            // 系统负载
}

func NewState(system *System) *State {
	numUsers := len(system.Users)
	numComms := len(system.Comms)

	state := &State{
		T:            system.T,
		R:            make([]float64, numUsers),
		Q:            make([][]float64, numUsers),
		F:            make([][]float64, numUsers),
		QNext:        make([][]float64, numUsers),
		QMid:         make([][]float64, numUsers),
		Delta:        make([][]float64, numUsers),
		Epsilon:      make([][]float64, numUsers),
		TransferPath: make([][]int, numUsers),
		V:            make([][]float64, numUsers),
		P:            make([][]float64, numUsers),
	}

	// 初始化矩阵
	for i := range numUsers {
		state.Q[i] = make([]float64, numComms)
		state.F[i] = make([]float64, numComms)
		state.QNext[i] = make([]float64, numComms)
		state.QMid[i] = make([]float64, numComms)
		state.Delta[i] = make([]float64, numComms)
		state.Epsilon[i] = make([]float64, numComms)
		state.V[i] = make([]float64, numComms)
		state.P[i] = make([]float64, numComms)
	}

	return state
}

// 计算资源利用率
func (s *State) CalcResourceUtil() float64 {
	totalF := 0.0
	if s.F == nil {
		return 0.0 // 避免除以零
	}
	for _, f := range s.F {
		for _, val := range f {
			totalF += val
		}
	}
	return totalF / float64(len(s.F[0]))
}

// 计算平均队列长度
func (s *State) CalcQueueAvg() float64 {
	totalQ := 0.0
	if s.Q == nil {
		return 0.0 // 避免除以零
	}
	for _, q := range s.Q {
		for _, val := range q {
			totalQ += val
		}
	}
	return totalQ / float64(len(s.Q)*len(s.Q[0]))
}

// 计算列的总队列长度
func (s *State) CalcRowQueue() []float64 {
	rowQ := make([]float64, len(s.Q[0]))
	if s.Q == nil {
		return rowQ // 避免除以零
	}

	for _, q := range s.Q {
		for j, val := range q {
			rowQ[j] += val
		}
	}

	return rowQ
}

// Copy 深拷贝状态
func (s *State) Copy() State {
	// 先浅拷贝结构体（拷贝所有基本类型字段）
	newState := *s

	// 深拷贝切片字段
	newState.R = append([]float64(nil), s.R...)

	// 深拷贝所有二维矩阵（宽度都相同）
	numRows := len(s.Q)
	newState.Q = make([][]float64, numRows)
	newState.F = make([][]float64, numRows)
	newState.QNext = make([][]float64, numRows)
	newState.QMid = make([][]float64, numRows)
	newState.Delta = make([][]float64, numRows)
	newState.Epsilon = make([][]float64, numRows)
	newState.V = make([][]float64, numRows)
	newState.P = make([][]float64, numRows)
	newState.TransferPath = make([][]int, numRows)

	for i := range numRows {
		newState.Q[i] = append([]float64(nil), s.Q[i]...)
		newState.F[i] = append([]float64(nil), s.F[i]...)
		newState.QNext[i] = append([]float64(nil), s.QNext[i]...)
		newState.QMid[i] = append([]float64(nil), s.QMid[i]...)
		newState.Delta[i] = append([]float64(nil), s.Delta[i]...)
		newState.Epsilon[i] = append([]float64(nil), s.Epsilon[i]...)
		newState.V[i] = append([]float64(nil), s.V[i]...)
		newState.P[i] = append([]float64(nil), s.P[i]...)
		newState.TransferPath[i] = append([]int(nil), s.TransferPath[i]...)
	}

	return newState
}

// UpdateDelta 更新通信资源分配（用户i选择通信设备j）
func (s *State) UpdateDelta(userIdx, commIdx int) {
	// 将资源分配给选定的通信设备
	s.Delta[userIdx][commIdx] = 1
}

// UpdateEpsilon 更新卸载决策（根据传输路径和速率）
func (s *State) UpdateEpsilon(userIdx int, jList []int, userSpeed float64, speeds [][]float64) {
	// 保存传输路径
	s.TransferPath[userIdx] = make([]int, len(jList))
	copy(s.TransferPath[userIdx], jList)

	// 根据路径更新卸载决策
	for idx, j := range jList {
		s.Epsilon[userIdx][j] = 1
		// 起始节点的上一条路径连接用户终端，后面的节点上一条路径连接UAV
		if idx == 0 {
			// 第一段：用户到第一个通信设备
			s.V[userIdx][j] = userSpeed
			s.P[userIdx][j] = constant.P_u
		} else {
			// 后续段：通信设备之间
			prevJ := jList[idx-1]
			s.V[userIdx][j] = speeds[prevJ][j]
			s.P[userIdx][j] = constant.P_b
		}
	}
}

// UpdateF 更新计算资源分配
// 更新中间队列状态：Q_mid = Q + delta * r
func (s *State) UpdateF() {
	numUsers := len(s.Q)
	numComms := len(s.Q[0])

	for i := range numUsers {
		for j := range numComms {
			s.QMid[i][j] = s.Q[i][j] + s.Delta[i][j]*s.R[i]
		}
	}

	// 计算每个通信设备的总队列需求
	for j := range numComms {
		totalQMid := 0.0
		for i := range numUsers {
			totalQMid += s.QMid[i][j]
		}

		// 按比例分配资源
		if totalQMid > 0 {
			for i := range numUsers {
				s.F[i][j] = s.QMid[i][j] / totalQMid
			}
		}
	}
}

// ComputeData 计算数据处理相关的指标
// R = delta * r（已收到的数据）
// F = s.F * C * Slot / Rho（能够处理的数据）
// Q_next = Q + R - F（下一时刻队列长度）
func (s *State) ComputeData() {
	numUsers := len(s.Q)
	numComms := len(s.Q[0])

	for i := range numUsers {
		for j := range numComms {
			// 计算下一时刻队列长度
			R_ij := s.Delta[i][j] * s.R[i]
			F_ij := s.F[i][j] * constant.C * constant.Slot / constant.Rho
			s.QNext[i][j] = s.Q[i][j] + R_ij - F_ij

			// 队列长度不能为负
			if s.QNext[i][j] < 0 {
				s.QNext[i][j] = 0
			}
		}
	}
}

// Objective 计算目标函数值（成本）
func (s *State) Objective() float64 {
	numUsers := len(s.Q)
	numComms := len(s.Q[0])

	// 计算漂移项: drift = Q * (R - F) / Shrink^2
	s.Drift = 0.0
	for i := range numUsers {
		for j := range numComms {
			R_ij := s.Delta[i][j] * s.R[i]
			F_ij := s.F[i][j] * constant.C * constant.Slot / constant.Rho
			drift := s.Q[i][j] * (R_ij - F_ij) / (constant.Shrink * constant.Shrink)
			s.Drift += drift
		}
	}

	// 计算传输时延: transfer_delay = Alpha * epsilon * r / v
	s.TransferDelay = 0.0
	for i := range numUsers {
		for j := range numComms {
			if s.V[i][j] != 0 {
				transferDelay := constant.Alpha * s.Epsilon[i][j] * s.R[i] / s.V[i][j]
				s.TransferDelay += transferDelay
			}
		}
	}

	// 计算计算时延: compute_delay = Alpha * Q_mid * Rho / (s.F * C)
	s.ComputeDelay = 0.0
	for i := range numUsers {
		for j := range numComms {
			if s.F[i][j] != 0 {
				computeDelay := constant.Alpha * s.QMid[i][j] * constant.Rho / (s.F[i][j] * constant.C)
				s.ComputeDelay += computeDelay
			}
		}
	}

	// 计算传输能耗: transfer_energy = Gamma * p * transfer_delay
	s.TransferEnergy = 0.0
	for i := range numUsers {
		for j := range numComms {
			if s.V[i][j] != 0 {
				transferDelay := constant.Alpha * s.Epsilon[i][j] * s.R[i] / s.V[i][j]
				transferEnergy := constant.Gamma * s.P[i][j] * transferDelay
				s.TransferEnergy += transferEnergy
			}
		}
	}

	// 计算计算能耗: compute_energy = Gamma * s.F * Kappa * C^3 * Slot
	s.ComputeEnergy = 0.0
	for i := range numUsers {
		for j := range numComms {
			computeEnergy := constant.Gamma * s.F[i][j] * constant.Kappa * math.Pow(constant.C, 3) * constant.Slot
			s.ComputeEnergy += computeEnergy
		}
	}

	// 计算负载: load = Beta * sum(Q_b) / Shrink
	// Q_b = sum((Q_next + Q) / 2, axis=0)
	Q_b := make([]float64, numComms)
	for j := range numComms {
		for i := range numUsers {
			Q_b[j] += (s.QNext[i][j] + s.Q[i][j]) / 2
		}
	}

	s.Load = 0.0
	for j := range numComms {
		s.Load += Q_b[j]
	}
	s.Load = constant.Beta * s.Load / constant.Shrink

	// 计算惩罚项
	s.Penalty = s.TransferDelay + s.ComputeDelay + s.TransferEnergy + s.ComputeEnergy + s.Load

	// 计算总成本
	s.Cost = s.Drift + constant.V*s.Penalty

	return s.Cost
}
