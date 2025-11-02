package define

import (
	"go-backend/internal/algorithm/utils"
	"time"
)

// 任务优先级常量
const (
	PriorityLow      = 0  // 低优先级
	PriorityNormal   = 5  // 正常优先级（默认）
	PriorityHigh     = 10 // 高优先级
	PriorityUrgent   = 15 // 紧急优先级
	PriorityCritical = 20 // 关键优先级
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
	ScheduledTime time.Time  `json:"scheduled_time,omitempty"` // 首次分配时间
	CompleteTime  time.Time  `json:"complete_time,omitempty"`  // 完成时间
	CancelledAt   *time.Time `json:"cancelled_at,omitempty"`   // 取消时间

	// 超时和取消
	Timeout       time.Duration `json:"timeout,omitempty"`        // 超时时长 (0表示无超时)
	FailureReason string        `json:"failure_reason,omitempty"` // 失败原因
}

// NewTask 创建新任务
func NewTask(userID uint, dataSize float64, taskType string) *Task {
	return &Task{
		ID:        utils.GenerateTaskID(),
		UserID:    userID,
		DataSize:  dataSize,
		Type:      taskType,
		Priority:  PriorityNormal, // 默认正常优先级
		CreatedAt: time.Now(),
		Status:    TaskPending,
	}
}

// NewTaskWithPriority 创建带优先级的任务
func NewTaskWithPriority(userID uint, dataSize float64, taskType string, priority int) *Task {
	task := NewTask(userID, dataSize, taskType)
	task.Priority = priority
	return task
}

// StateMachine 获取任务的状态机
func (t *Task) StateMachine() *TaskStateMachine {
	return NewTaskStateMachine(t)
}

// IsCancelled 检查任务是否已取消
func (t *Task) IsCancelled() bool {
	return t.CancelledAt != nil
}

// IsTimedOut 检查任务是否超时
func (t *Task) IsTimedOut() bool {
	if t.Timeout == 0 {
		return false // 无超时限制
	}
	if t.Status == TaskCompleted || t.Status == TaskFailed {
		return false // 已结束的任务不算超时
	}
	elapsed := time.Since(t.CreatedAt)
	return elapsed > t.Timeout
}

// GetElapsedTime 获取任务已运行时间
func (t *Task) GetElapsedTime() time.Duration {
	if !t.CompleteTime.IsZero() {
		return t.CompleteTime.Sub(t.CreatedAt)
	}
	return time.Since(t.CreatedAt)
}

// GetWaitTime 获取任务等待时间 (从创建到调度)
func (t *Task) GetWaitTime() time.Duration {
	if !t.ScheduledTime.IsZero() {
		return t.ScheduledTime.Sub(t.CreatedAt)
	}
	return time.Since(t.CreatedAt) // 还在等待
}

// IsStarving 检查任务是否处于饥饿状态 (等待时间过长)
// 饥饿阈值: 低优先级10秒, 普通5秒, 高优先级2秒
func (t *Task) IsStarving() bool {
	if t.Status != TaskPending {
		return false // 已调度的任务不算饥饿
	}

	waitTime := t.GetWaitTime()
	switch {
	case t.Priority >= PriorityHigh:
		return waitTime > 2*time.Second
	case t.Priority >= PriorityNormal:
		return waitTime > 5*time.Second
	default:
		return waitTime > 10*time.Second
	}
}
