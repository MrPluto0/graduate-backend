package algorithm

import (
	"fmt"
	"go-backend/internal/algorithm/define"
	"sync"
	"time"
)

// TaskManager 简化的任务管理器 (只管理Task本身,不管理调度状态)
type TaskManager struct {
	Tasks    map[string]*define.Task // TaskID -> Task
	TaskList []*define.Task          // 按创建时间排序的任务列表
	mutex    sync.RWMutex
}

// NewTaskManager 创建任务管理器
func NewTaskManager() *TaskManager {
	return &TaskManager{
		Tasks:    make(map[string]*define.Task),
		TaskList: make([]*define.Task, 0),
	}
}

// AddTask 添加任务
func (tm *TaskManager) AddTask(task *define.Task) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.Tasks[task.ID] = task
	tm.TaskList = append(tm.TaskList, task)
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(taskID string) *define.Task {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.Tasks[taskID]
}

// GetActiveTasks 获取所有活跃任务(未完成且未失败)
func (tm *TaskManager) GetActiveTasks() []*define.Task {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tasks := make([]*define.Task, 0)
	for _, task := range tm.TaskList {
		if task.StateMachine().IsActive() {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// GetTasksByStatus 按状态获取任务
func (tm *TaskManager) GetTasksByStatus(status define.TaskStatus) []*define.Task {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tasks := make([]*define.Task, 0)
	for _, task := range tm.TaskList {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// GetTasksWithPage 分页获取任务
func (tm *TaskManager) GetTasksWithPage(offset, limit int, userID *uint, status *define.TaskStatus) ([]*define.Task, int64) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// 过滤
	filtered := make([]*define.Task, 0)
	for _, task := range tm.TaskList {
		if userID != nil && task.UserID != *userID {
			continue
		}
		if status != nil && task.Status != *status {
			continue
		}
		filtered = append(filtered, task)
	}

	total := int64(len(filtered))

	// 分页
	if offset >= len(filtered) {
		return []*define.Task{}, total
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], total
}

// UpdateTaskStatus 更新任务状态
func (tm *TaskManager) UpdateTaskStatus(taskID string, newStatus define.TaskStatus) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	task := tm.Tasks[taskID]
	if task == nil {
		return nil
	}

	sm := task.StateMachine()
	switch newStatus {
	case define.TaskQueued:
		return sm.ToQueued()
	case define.TaskComputing:
		return sm.ToComputing()
	case define.TaskCompleted:
		return sm.ToCompleted()
	case define.TaskFailed:
		return sm.ToFailed("")
	}
	return nil
}

// Count 获取任务总数
func (tm *TaskManager) Count() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return len(tm.TaskList)
}

// CountCompleted 获取已完成任务数
func (tm *TaskManager) CountCompleted() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	count := 0
	for _, task := range tm.TaskList {
		if task.StateMachine().IsCompleted() {
			count++
		}
	}
	return count
}

// CancelTask 取消任务
func (tm *TaskManager) CancelTask(taskID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	task := tm.Tasks[taskID]
	if task == nil {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	sm := task.StateMachine()

	// 只能取消活跃任务
	if !sm.IsActive() {
		return fmt.Errorf("任务已结束,无法取消 (状态: %s)", sm.GetStatusName())
	}

	// 标记为已取消
	now := time.Now()
	task.CancelledAt = &now
	task.FailureReason = "用户取消"

	// 转换到Failed状态
	if err := sm.ToFailed("用户取消"); err != nil {
		return fmt.Errorf("状态转换失败: %w", err)
	}

	return nil
}

// CheckTimeouts 检查超时任务并标记为失败
func (tm *TaskManager) CheckTimeouts() []string {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	timedOutTasks := make([]string, 0)

	for _, task := range tm.TaskList {
		if task.IsTimedOut() && task.StateMachine().IsActive() {
			// 标记超时
			task.FailureReason = fmt.Sprintf("超时 (限制: %v, 实际: %v)",
				task.Timeout, task.GetElapsedTime())

			// 转换到Failed状态
			if err := task.StateMachine().ToFailed(task.FailureReason); err == nil {
				timedOutTasks = append(timedOutTasks, task.ID)
			}
		}
	}

	return timedOutTasks
}
