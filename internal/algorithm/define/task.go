package define

import (
	"go-backend/internal/algorithm/utils"
	"time"
)

// Task 简化的任务对象 (纯持久化,不包含调度状态)
type Task struct {
	// 基本信息
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"`
	Type      string    `json:"type,omitempty"`
	UserID    uint      `json:"user_id"`
	DataSize  float64   `json:"data_size"`
	Priority  int       `json:"priority,omitempty"`
	CreatedAt time.Time `json:"created_at"`

	// 状态
	Status TaskStatus `json:"status"`

	// 时间戳
	ScheduledTime time.Time `json:"scheduled_time,omitempty"` // 首次分配时间
	CompleteTime  time.Time `json:"complete_time,omitempty"`  // 完成时间
}

// NewTask 创建新任务
func NewTask(userID uint, dataSize float64, taskType string) *Task {
	return &Task{
		ID:        utils.GenerateTaskID(),
		UserID:    userID,
		DataSize:  dataSize,
		Type:      taskType,
		CreatedAt: time.Now(),
		Status:    TaskPending,
	}
}

// StateMachine 获取任务的状态机
func (t *Task) StateMachine() *TaskStateMachine {
	return NewTaskStateMachine(t)
}
