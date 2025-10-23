package define

import (
	"go-backend/internal/algorithm/constant"
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
	ID        string     `json:"id,omitempty"`
	Name      string     `json:"name,omitempty"`
	Type      string     `json:"type,omitempty"`
	UserID    uint       `json:"user_id" binding:"required"`
	DataSize  float64    `json:"data_size" binding:"required"`
	Priority  int        `json:"priority,omitempty"`
	Status    TaskStatus `json:"status,omitempty"`
	CreatedAt time.Time  `json:"create_time,omitempty"`
}

// Task 任务（持久化对象）
type Task struct {
	TaskBase

	// 调度结果
	AssignedCommID uint          `json:"assigned_comm_id"`   // 分配的计算节点
	TransferPath   *TransferPath `json:"transfer_path"`      // 传输路径
	AllocResource  float64       `json:"allocated_resource"` // 分配的计算资源比例
	QueuedData     float64       `json:"queued_data"`        // 队列中的数据量
	ProcessedData  float64       `json:"processed_data"`     // 已处理的数据量

	// 时间
	ScheduledTime time.Time `json:"scheduled_time"` // 调度时间
	CompleteTime  time.Time `json:"complete_time"`  // 完成时间

	// 性能指标
	Metrics *TaskMetrics `json:"metrics,omitempty"`
}

// TaskSnapshot 调度快照（纯计算临时数据，不持久化）
type TaskSnapshot struct {
	ID string `json:"id"` // 关联的任务ID（通过此ID查询Task获取其他信息）

	// 本轮调度分配结果（会写回Task）
	AssignedCommID   uint          `json:"assigned_comm_id"`
	TransferPath     *TransferPath `json:"transfer_path"`
	ResourceFraction float64       `json:"-"` // 本轮分配的资源比例

	// 纯临时计算变量（不写回Task）
	PendingTransferData float64     `json:"-"`       // 任务剩余待传输数据总量（DataSize - ProcessedData）
	CurrentQueue        float64     `json:"-"`       // 当前队列
	NextQueue           float64     `json:"-"`       // 下一时隙队列
	IntermediateQueue   float64     `json:"-"`       // 中间队列（预测阶段：CurrentQueue + PendingTransferData）
	Metrics             TaskMetrics `json:"metrics"` // 本轮性能指标
}

// ComputeMetrics 计算快照的性能指标
// transferredData: 实际传输的数据量
// queueData: 队列中的数据量
func (snap *TaskSnapshot) ComputeMetrics(transferredData, queueData float64) TaskMetrics {
	metrics := TaskMetrics{}

	// 传输延迟：传输数据通过各段路径的延迟之和
	if transferredData > 0 && snap.TransferPath != nil {
		for _, speed := range snap.TransferPath.Speeds {
			if speed > 0 {
				metrics.TransferDelay += transferredData / speed
			}
		}
	}

	// 计算延迟：队列数据的计算时间
	if snap.ResourceFraction > 0 {
		metrics.ComputeDelay = queueData * constant.Rho / (snap.ResourceFraction * constant.C)
	}

	// 传输能耗：各段路径的能耗之和
	if transferredData > 0 && snap.TransferPath != nil {
		for i, power := range snap.TransferPath.Powers {
			if i < len(snap.TransferPath.Speeds) && snap.TransferPath.Speeds[i] > 0 {
				segmentDelay := transferredData / snap.TransferPath.Speeds[i]
				metrics.TransferEnergy += power * segmentDelay
			}
		}
	}

	// 计算能耗
	metrics.ComputeEnergy = snap.ResourceFraction * constant.Kappa * constant.C * constant.C * constant.C * constant.Slot

	// 总延迟和总能耗
	metrics.TotalDelay = metrics.TransferDelay + metrics.ComputeDelay
	metrics.TotalEnergy = metrics.TransferEnergy + metrics.ComputeEnergy

	return metrics
}

func NewTask(base TaskBase) *Task {
	return &Task{
		TaskBase: TaskBase{
			ID:        utils.GenerateTaskID(),
			UserID:    base.UserID,
			DataSize:  base.DataSize,
			Type:      base.Type,
			CreatedAt: time.Now(),
			Status:    TaskPending,
		},
		Metrics: &TaskMetrics{},
	}
}

func (t *Task) Copy() *Task {
	newTask := *t
	if t.TransferPath != nil {
		newTask.TransferPath = t.TransferPath.Copy()
	}
	return &newTask
}
