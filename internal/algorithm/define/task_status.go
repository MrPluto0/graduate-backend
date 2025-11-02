package define

import "time"

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskPending   TaskStatus = iota // 等待调度
	TaskQueued                      // 已分配，排队中
	TaskComputing                   // 计算中
	TaskCompleted                   // 已完成
	TaskFailed                      // 失败
)

// TaskBase 任务基本信息
type TaskBase struct {
	ID        string     `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Type      string     `json:"type,omitempty"`
	UserID    uint       `json:"user_id" binding:"required"`
	DataSize  float64    `json:"data_size" binding:"required"`
	Priority  int        `json:"priority,omitempty"`
	Status    TaskStatus `json:"status,omitempty"`
	CreatedAt time.Time  `json:"create_time,omitempty"`
}

// TaskWithMetrics 为了兼容旧API,提供带性能指标的扩展Task结构
type TaskWithMetrics struct {
	TaskBase

	// 调度结果 (从Assignment填充)
	AssignedCommID uint          `json:"assigned_comm_id"`
	TransferPath   *TransferPath `json:"transfer_path"`

	// 时间
	ScheduledTime time.Time `json:"scheduled_time"`
	CompleteTime  time.Time `json:"complete_time"`

	// 性能指标历史 (从Assignment转换)
	MetricsHistory []SlotMetrics `json:"metrics_history,omitempty"`
}

// TransferPath 传输路径 (从Assignment提取)
type TransferPath struct {
	Path   []uint    `json:"path"`
	Speeds []float64 `json:"speeds"`
	Powers []float64 `json:"powers"`
}
