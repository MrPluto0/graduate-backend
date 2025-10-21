package define

import (
	"go-backend/internal/algorithm/utils"
	"time"
)

type TaskStatus int

const (
	TaskPending   TaskStatus = iota // 等待调度
	TaskQueued                      // 已分配，排队中
	TaskComputing                   // 计算中
	TaskCompleted                   // 已完成
	TaskFailed                      // 失败
)

type TaskBase struct {
	UserID   uint
	DataSize float64
	TaskType string
}

// TaskMetrics 任务性能指标
type TaskMetrics struct {
	TransferDelay  float64 `json:"transfer_delay"`  // 传输延迟
	ComputeDelay   float64 `json:"compute_delay"`   // 计算延迟
	TotalDelay     float64 `json:"total_delay"`     // 总延迟
	TransferEnergy float64 `json:"transfer_energy"` // 传输能耗
	ComputeEnergy  float64 `json:"compute_energy"`  // 计算能耗
	TotalEnergy    float64 `json:"total_energy"`    // 总能耗
}

// Task 任务（持久化对象）
type Task struct {
	// 基础信息
	TaskID     string    `json:"task_id"`
	UserID     uint      `json:"user_id"`
	DataSize   float64   `json:"data_size"`
	TaskType   string    `json:"task_type"`
	CreateTime time.Time `json:"create_time"`

	// 状态
	Status TaskStatus `json:"status"`

	// 调度结果
	AssignedCommID uint          `json:"assigned_comm_id"`   // 分配的计算节点
	TransferPath   *TransferPath `json:"transfer_path"`      // 传输路径
	AllocResource  float64       `json:"allocated_resource"` // 分配的计算资源比例
	QueuedData     float64       `json:"queued_data"`        // 队列中的数据量
	ProcessedData  float64       `json:"processed_data"`     // 已处理的数据量

	// 时间
	ScheduledTime time.Time `json:"scheduled_time"` // 调度时间
	StartTime     time.Time `json:"start_time"`     // 开始处理时间
	CompleteTime  time.Time `json:"complete_time"`  // 完成时间

	// 性能指标
	Metrics *TaskMetrics `json:"metrics,omitempty"`
}

// TaskSnapshot 调度快照（纯计算临时数据，不持久化）
type TaskSnapshot struct {
	TaskID string `json:"task_id"` // 关联的任务ID（通过此ID查询Task获取其他信息）

	// 本轮调度分配结果（会写回Task）
	AssignedCommID   uint          `json:"assigned_comm_id"`
	TransferPath     *TransferPath `json:"transfer_path"`
	ResourceFraction float64       `json:"-"` // 本轮分配的资源比例

	// 纯临时计算变量（不写回Task）
	PendingTransferData float64     `json:"-"`       // 本轮待传输数据
	CurrentQueue        float64     `json:"-"`       // 当前队列
	NextQueue           float64     `json:"-"`       // 下一时隙队列
	IntermediateQueue   float64     `json:"-"`       // 中间队列
	Metrics             TaskMetrics `json:"metrics"` // 本轮性能指标
}

func NewTask(base TaskBase) *Task {
	return &Task{
		TaskID:     utils.GenerateTaskID(),
		UserID:     base.UserID,
		DataSize:   base.DataSize,
		TaskType:   base.TaskType,
		CreateTime: time.Now(),
		Status:     TaskPending,
		Metrics:    &TaskMetrics{},
	}
}

func (t *Task) Copy() *Task {
	newTask := *t
	if t.TransferPath != nil {
		newTask.TransferPath = t.TransferPath.Copy()
	}
	return &newTask
}
