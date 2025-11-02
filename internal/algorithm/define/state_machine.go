package define

import (
	"fmt"
	"time"
)

// TaskStateMachine 任务状态机
type TaskStateMachine struct {
	task *Task
}

// NewTaskStateMachine 创建任务状态机
func NewTaskStateMachine(task *Task) *TaskStateMachine {
	return &TaskStateMachine{task: task}
}

// 状态转换方法

// ToQueued 转换到Queued状态（任务被分配到通信设备）
func (sm *TaskStateMachine) ToQueued() error {
	if sm.task.Status != TaskPending {
		return fmt.Errorf("invalid state transition: %s -> Queued", sm.statusName())
	}
	sm.task.Status = TaskQueued
	sm.task.ScheduledTime = time.Now()
	return nil
}

// ToComputing 转换到Computing状态（任务开始处理数据）
func (sm *TaskStateMachine) ToComputing() error {
	if sm.task.Status != TaskQueued {
		return fmt.Errorf("invalid state transition: %s -> Computing", sm.statusName())
	}
	sm.task.Status = TaskComputing
	return nil
}

// ToCompleted 转换到Completed状态（任务完成）
func (sm *TaskStateMachine) ToCompleted() error {
	if sm.task.Status != TaskComputing && sm.task.Status != TaskQueued {
		return fmt.Errorf("invalid state transition: %s -> Completed", sm.statusName())
	}
	sm.task.Status = TaskCompleted
	sm.task.CompleteTime = time.Now()
	return nil
}

// ToFailed 转换到Failed状态（任务失败）
func (sm *TaskStateMachine) ToFailed(reason string) error {
	// 任何状态都可以转换到Failed（除了Completed）
	if sm.task.Status == TaskCompleted {
		return fmt.Errorf("cannot fail a completed task")
	}
	sm.task.Status = TaskFailed
	// TODO: 添加失败原因字段到Task
	return nil
}

// 状态查询方法

// CanTransitionTo 检查是否可以转换到目标状态
func (sm *TaskStateMachine) CanTransitionTo(target TaskStatus) bool {
	current := sm.task.Status
	switch target {
	case TaskQueued:
		return current == TaskPending
	case TaskComputing:
		return current == TaskQueued
	case TaskCompleted:
		return current == TaskComputing || current == TaskQueued
	case TaskFailed:
		return current != TaskCompleted
	default:
		return false
	}
}

// IsActive 任务是否处于活跃状态（未完成且未失败）
func (sm *TaskStateMachine) IsActive() bool {
	return sm.task.Status != TaskCompleted && sm.task.Status != TaskFailed
}

// IsPending 是否等待调度
func (sm *TaskStateMachine) IsPending() bool {
	return sm.task.Status == TaskPending
}

// IsQueued 是否已分配但排队中
func (sm *TaskStateMachine) IsQueued() bool {
	return sm.task.Status == TaskQueued
}

// IsComputing 是否正在计算
func (sm *TaskStateMachine) IsComputing() bool {
	return sm.task.Status == TaskComputing
}

// IsCompleted 是否已完成
func (sm *TaskStateMachine) IsCompleted() bool {
	return sm.task.Status == TaskCompleted
}

// IsFailed 是否失败
func (sm *TaskStateMachine) IsFailed() bool {
	return sm.task.Status == TaskFailed
}

// statusName 获取当前状态名称（用于错误消息）
func (sm *TaskStateMachine) statusName() string {
	switch sm.task.Status {
	case TaskPending:
		return "Pending"
	case TaskQueued:
		return "Queued"
	case TaskComputing:
		return "Computing"
	case TaskCompleted:
		return "Completed"
	case TaskFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// GetStatus 获取当前状态
func (sm *TaskStateMachine) GetStatus() TaskStatus {
	return sm.task.Status
}

// GetStatusName 获取当前状态的可读名称
func (sm *TaskStateMachine) GetStatusName() string {
	return sm.statusName()
}
