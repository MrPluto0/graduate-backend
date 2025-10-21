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

// TaskSnapshot 任务快照（调度计算的临时数据）
type TaskSnapshot struct {
	TaskID string `json:"task_id"` // 任务ID

	// 当前状态（从 Task 拷贝）
	Status        TaskStatus `json:"status"`         // 任务状态
	QueuedData    float64    `json:"queued_data"`    // 当前队列
	ProcessedData float64    `json:"processed_data"` // 已处理数据

	// 调度分配结果（直接使用ID）
	AssignedCommID uint          `json:"assigned_comm_id"` // 分配的通信设备ID
	TransferPath   *TransferPath `json:"transfer_path"`    // 传输路径（包含预计算的速率和功率）

	// 调度过程的临时变量
	PendingTransferData float64 `json:"-"` // 本轮待传输数据
	CurrentQueue        float64 `json:"-"` // 当前队列（计算用）
	NextQueue           float64 `json:"-"` // 下一时隙队列
	IntermediateQueue   float64 `json:"-"` // 中间队列
	ResourceFraction    float64 `json:"-"` // 分配的资源比例

	// 计算出的性能指标（复用 TaskMetrics 类型）
	Metrics TaskMetrics `json:"metrics"`
}

type Task struct {
	TaskID     string    `json:"task_id"`
	UserID     uint      `json:"user_id"`
	DataSize   float64   `json:"data_size"`
	TaskType   string    `json:"task_type"`
	CreateTime time.Time `json:"create_time"`

	Status         TaskStatus    `json:"status"`
	AssignedCommID uint          `json:"assigned_comm_id"` // 分配的计算节点
	TransferPath   *TransferPath `json:"transfer_path"`    // 传输路径（包含预计算信息）

	AllocResource float64 `json:"allocated_resource"` // 分配的计算资源比例
	QueuedData    float64 `json:"queued_data"`        // 队列中的数据量
	ProcessedData float64 `json:"processed_data"`     // 已处理的数据量

	ScheduledTime time.Time `json:"scheduled_time"` // 调度时间
	StartTime     time.Time `json:"start_time"`     // 开始处理时间
	CompleteTime  time.Time `json:"complete_time"`  // 完成时间

	// 性能指标
	Metrics *TaskMetrics `json:"metrics,omitempty"` // 性能指标（可选）
}

func NewTask(base TaskBase) *Task {
	return &Task{
		TaskID:     utils.GenerateTaskID(),
		UserID:     base.UserID,
		DataSize:   base.DataSize,
		TaskType:   base.TaskType,
		CreateTime: time.Now(),
		Status:     TaskPending,
	}
}

func (t *Task) Copy() *Task {
	newTask := *t
	if t.TransferPath != nil {
		newTask.TransferPath = t.TransferPath.Copy()
	}
	return &newTask
}
