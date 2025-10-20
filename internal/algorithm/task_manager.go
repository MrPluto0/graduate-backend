package algorithm

import (
	"go-backend/internal/algorithm/define"
	"time"
)

type TaskManager struct {
	System       *System
	Tasks        map[string]*define.Task // 任务映射 taskID -> define.Task
	UserTasks    map[uint][]string       // 用户任务列表 userID -> []taskID
	PendingTasks []string                // 待处理任务队列
	ActiveTasks  []string                // 活跃任务列表
}

func NewTaskManager(system *System) *TaskManager {
	return &TaskManager{
		System:       system,
		Tasks:        make(map[string]*define.Task),
		UserTasks:    make(map[uint][]string),
		PendingTasks: make([]string, 0),
		ActiveTasks:  make([]string, 0),
	}
}

// 提交新任务到系统
func (tm *TaskManager) AddTask(base define.TaskBase) (*define.Task, error) {
	// 找到用户索引
	userIdx := -1
	for i, user := range tm.System.Users {
		if user.ID == base.UserID {
			userIdx = i
			break
		}
	}

	task := define.NewTask(base)
	task.UserIndex = userIdx

	tm.Tasks[task.TaskID] = task
	tm.UserTasks[base.UserID] = append(tm.UserTasks[base.UserID], task.TaskID)
	tm.PendingTasks = append(tm.PendingTasks, task.TaskID)
	tm.ActiveTasks = append(tm.ActiveTasks, task.TaskID)

	return task.Copy(), nil
}

// getActiveTasks 获取所有活跃任务
func (tm *TaskManager) getActiveTasks() []*define.Task {
	tasks := make([]*define.Task, 0)
	for _, taskID := range tm.ActiveTasks {
		task, exists := tm.Tasks[taskID]
		if !exists {
			continue
		}
		if task.Status != define.TaskCompleted && task.Status != define.TaskFailed {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// updateFromTaskState 从TaskState更新任务状态
func (tm *TaskManager) updateFromTaskState(state *TaskState, tasks []*define.Task, sys *System) {
	for _, task := range tasks {
		alloc, ok := state.Allocations[task.TaskID]
		if !ok {
			continue
		}

		if alloc.AssignedCommID > 0 {
			task.AssignedCommID = alloc.AssignedCommID
			// 通过ID找到CommIndex
			for idx, comm := range sys.Comms {
				if comm.ID == alloc.AssignedCommID {
					task.CommIndex = idx
					break
				}
			}

			// 设置传输路径（已经是ID列表）
			task.TransferPath = append([]uint(nil), alloc.TransferPath...)

			// 更新队列和已处理数据
			// 本时隙传输了 R，处理了 (Q + R - QNext)
			processed := alloc.Q + alloc.R - alloc.QNext
			if processed > 0 {
				task.ProcessedData += processed
			}
			task.QueuedData = alloc.QNext
			task.AllocResource = alloc.F

			// 更新任务状态
			if task.Status == define.TaskPending {
				task.Status = define.TaskQueued
				task.ScheduledTime = time.Now()
			}

			if alloc.QNext > 0 && task.Status == define.TaskQueued {
				task.Status = define.TaskComputing
				if task.StartTime.IsZero() {
					task.StartTime = time.Now()
				}
			}

			// 检查是否完成
			if alloc.QNext < 0.001 && task.ProcessedData >= task.DataSize-0.001 {
				task.Status = define.TaskCompleted
				task.CompleteTime = time.Now()
				task.ProcessedData = task.DataSize
				task.QueuedData = 0
			}
		}
	}

	tm.updateActiveTasksList()
}

// updateActiveTasksList 更新活跃任务列表
func (tm *TaskManager) updateActiveTasksList() {
	newActive := make([]string, 0)

	for _, taskID := range tm.ActiveTasks {
		task, exists := tm.Tasks[taskID]
		if !exists {
			continue
		}
		if task.Status != define.TaskCompleted && task.Status != define.TaskFailed {
			newActive = append(newActive, taskID)
		}
	}

	tm.ActiveTasks = newActive

	newPending := make([]string, 0)
	for _, taskID := range tm.PendingTasks {
		task, exists := tm.Tasks[taskID]
		if !exists {
			continue
		}
		if task.Status == define.TaskPending {
			newPending = append(newPending, taskID)
		}
	}
	tm.PendingTasks = newPending
}

// hasActiveTasks 检查是否有活跃任务
func (tm *TaskManager) hasActiveTasks() bool {
	for _, taskID := range tm.ActiveTasks {
		task, exists := tm.Tasks[taskID]
		if !exists {
			continue
		}
		if task.Status != define.TaskCompleted && task.Status != define.TaskFailed {
			return true
		}
	}
	return false
}
