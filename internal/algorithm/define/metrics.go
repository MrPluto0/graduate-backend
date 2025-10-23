package define

// TaskMetrics 任务性能指标
type TaskMetrics struct {
	TransferDelay  float64 `json:"transfer_delay"`  // 传输延迟
	ComputeDelay   float64 `json:"compute_delay"`   // 计算延迟
	TotalDelay     float64 `json:"total_delay"`     // 总延迟
	TransferEnergy float64 `json:"transfer_energy"` // 传输能耗
	ComputeEnergy  float64 `json:"compute_energy"`  // 计算能耗
	TotalEnergy    float64 `json:"total_energy"`    // 总能耗
}

// SystemInfo 系统信息
type SystemInfo struct {
	UserCount      int               `json:"user_count"`      // 用户设备数量
	CommCount      int               `json:"comm_count"`      // 通信设备数量
	IsRunning      bool              `json:"is_running"`      // 是否运行中
	IsInitialized  bool              `json:"is_initialized"`  // 是否已初始化
	TimeSlot       uint              `json:"time_slot"`       // 当前时隙
	TransferPath   map[string][]uint `json:"transfer_path"`   // 任务ID -> 传输路径
	TaskCount      int               `json:"task_count"`      // 总任务数
	ActiveTasks    int               `json:"active_tasks"`    // 活跃任务数
	CompletedTasks int               `json:"completed_tasks"` // 已完成任务数
	State          interface{}       `json:"state,omitempty"` // 当前状态（可选）
}
