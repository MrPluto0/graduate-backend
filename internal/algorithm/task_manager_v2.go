package algorithm

import (
	"go-backend/internal/algorithm/define"
	"sync"
)

// TaskManagerV2 简化的任务管理器 (只管理Task本身,不管理调度状态)
type TaskManagerV2 struct {
	Tasks    map[string]*define.TaskV2 // TaskID -> Task
	TaskList []*define.TaskV2          // 按创建时间排序的任务列表
	mutex    sync.RWMutex
}

// NewTaskManagerV2 创建任务管理器
func NewTaskManagerV2() *TaskManagerV2 {
	return &TaskManagerV2{
		Tasks:    make(map[string]*define.TaskV2),
		TaskList: make([]*define.TaskV2, 0),
	}
}

// AddTask 添加任务
func (tm *TaskManagerV2) AddTask(task *define.TaskV2) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.Tasks[task.ID] = task
	tm.TaskList = append(tm.TaskList, task)
}

// GetTask 获取任务
func (tm *TaskManagerV2) GetTask(taskID string) *define.TaskV2 {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.Tasks[taskID]
}

// GetActiveTasks 获取所有活跃任务(未完成且未失败)
func (tm *TaskManagerV2) GetActiveTasks() []*define.TaskV2 {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tasks := make([]*define.TaskV2, 0)
	for _, task := range tm.TaskList {
		if task.StateMachine().IsActive() {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// GetTasksByStatus 按状态获取任务
func (tm *TaskManagerV2) GetTasksByStatus(status define.TaskStatus) []*define.TaskV2 {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tasks := make([]*define.TaskV2, 0)
	for _, task := range tm.TaskList {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// GetTasksWithPage 分页获取任务
func (tm *TaskManagerV2) GetTasksWithPage(offset, limit int, userID *uint, status *define.TaskStatus) ([]*define.TaskV2, int64) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// 过滤
	filtered := make([]*define.TaskV2, 0)
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
		return []*define.TaskV2{}, total
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], total
}

// UpdateTaskStatus 更新任务状态
func (tm *TaskManagerV2) UpdateTaskStatus(taskID string, newStatus define.TaskStatus) error {
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
func (tm *TaskManagerV2) Count() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return len(tm.TaskList)
}

// CountCompleted 获取已完成任务数
func (tm *TaskManagerV2) CountCompleted() int {
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
