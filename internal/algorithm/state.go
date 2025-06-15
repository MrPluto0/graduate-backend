package algorithm

// State 系统状态结构体
type State struct {
	// 基本状态信息
	T uint      `json:"t"` // 系统时间
	R []float64 `json:"r"` // 用户待处理数据

	// 队列状态
	Q     [][]float64 `json:"Q"`      // 当前队列长度
	F     [][]float64 `json:"f"`      // 计算资源分配
	QNext [][]float64 `json:"Q_next"` // 下一时刻队列长度

	// 决策变量
	Delta   [][]float64 `json:"delta"`   // 通信资源分配
	Epsilon [][]float64 `json:"epsilon"` // 卸载决策

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
