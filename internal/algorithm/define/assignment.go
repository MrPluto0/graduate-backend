package define

// Assignment 任务调度分配 (每个时隙的调度决策)
type Assignment struct {
	TimeSlot uint   `json:"time_slot"` // 时隙编号
	TaskID   string `json:"task_id"`   // 任务ID

	// 调度决策
	CommID uint     `json:"comm_id"` // 分配的通信设备ID
	Path   []uint   `json:"path"`    // 传输路径(设备ID序列: user → ... → comm)
	Speeds []float64 `json:"speeds"`  // 路径每段的传输速率
	Powers []float64 `json:"powers"`  // 路径每段的传输功率

	// 资源分配
	ResourceFraction float64 `json:"resource_fraction"` // 分配的计算资源比例

	// 队列状态
	QueueData       float64 `json:"queue_data"`       // 时隙开始时的队列数据量
	TransferredData float64 `json:"transferred_data"` // 本时隙传输的数据量
	ProcessedData   float64 `json:"processed_data"`   // 本时隙处理的数据量

	// 累计进度
	CumulativeTransferred float64 `json:"cumulative_transferred"` // 累计已传输数据量
	CumulativeProcessed   float64 `json:"cumulative_processed"`   // 累计已处理数据量
}

// NewAssignment 创建新的分配记录
func NewAssignment(timeSlot uint, taskID string, commID uint, path []uint, speeds, powers []float64) *Assignment {
	return &Assignment{
		TimeSlot: timeSlot,
		TaskID:   taskID,
		CommID:   commID,
		Path:     path,
		Speeds:   speeds,
		Powers:   powers,
	}
}

// Copy 深拷贝Assignment
func (a *Assignment) Copy() *Assignment {
	if a == nil {
		return nil
	}
	cp := *a
	cp.Path = make([]uint, len(a.Path))
	copy(cp.Path, a.Path)
	cp.Speeds = make([]float64, len(a.Speeds))
	copy(cp.Speeds, a.Speeds)
	cp.Powers = make([]float64, len(a.Powers))
	copy(cp.Powers, a.Powers)
	return &cp
}
