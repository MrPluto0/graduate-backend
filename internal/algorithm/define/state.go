package define

// StateMetrics 系统全局状态指标
type StateMetrics struct {
	CommQueues     map[string]float64 `json:"comm_queues"`     // 每个通信设备的队列长度
	TotalQueue     float64            `json:"total_queue"`     // 总队列长度
	TransferDelay  float64            `json:"transfer_delay"`  // 传输延迟
	ComputeDelay   float64            `json:"compute_delay"`   // 计算延迟
	TotalDelay     float64            `json:"total_delay"`     // 总延迟
	TransferEnergy float64            `json:"transfer_energy"` // 传输能耗
	ComputeEnergy  float64            `json:"compute_energy"`  // 计算能耗
	TotalEnergy    float64            `json:"total_energy"`    // 总能耗
	Load           float64            `json:"load"`            // 系统负载
	Cost           float64            `json:"cost"`            // 总成本
	Drift          float64            `json:"drift"`           // 漂移值
	Penalty        float64            `json:"penalty"`         // 惩罚项
}

// NewStateMetrics 创建空的状态指标
func NewStateMetrics() *StateMetrics {
	return &StateMetrics{
		CommQueues: make(map[string]float64),
	}
}
