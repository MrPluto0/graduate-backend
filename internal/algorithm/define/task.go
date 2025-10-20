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

type Task struct {
	TaskID     string    `json:"task_id"`
	UserID     uint      `json:"user_id"`
	DataSize   float64   `json:"data_size"`
	TaskType   string    `json:"task_type"`
	CreateTime time.Time `json:"create_time"`

	Status         TaskStatus `json:"status"`
	AssignedCommID uint       `json:"assigned_comm_id"` // 分配的计算节点
	TransferPath   []uint     `json:"transfer_path"`    // 传输路径

	AllocResource float64 `json:"allocated_resource"` // 分配的计算资源比例
	QueuedData    float64 `json:"queued_data"`        // 队列中的数据量
	ProcessedData float64 `json:"processed_data"`     // 已处理的数据量

	ScheduledTime time.Time `json:"scheduled_time"` // 调度时间
	StartTime     time.Time `json:"start_time"`     // 开始处理时间
	CompleteTime  time.Time `json:"complete_time"`  // 完成时间

	UserIndex int `json:"-"` // 用户在Users数组中的索引
	CommIndex int `json:"-"` // 分配的通信设备索引
}

func NewTask(base TaskBase) *Task {
	return &Task{
		TaskID:     utils.GenerateTaskID(),
		UserID:     base.UserID,
		DataSize:   base.DataSize,
		TaskType:   base.TaskType,
		CreateTime: time.Now(),
		Status:     TaskPending,
		UserIndex:  -1,
		CommIndex:  -1,
	}
}

func (t *Task) Copy() *Task {
	newTask := *t
	if len(t.TransferPath) > 0 {
		newTask.TransferPath = make([]uint, len(t.TransferPath))
		copy(newTask.TransferPath, t.TransferPath)
	}
	return &newTask
}
