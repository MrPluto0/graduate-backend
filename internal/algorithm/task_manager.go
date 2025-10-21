package algorithm

import (
	"go-backend/internal/algorithm/define"
	"time"
)

type TaskManager struct {
	System      *System
	Tasks       map[string]*define.Task // 任务映射 taskID -> define.Task
	UserTasks   map[uint][]string       // 用户任务列表 userID -> []taskID
	ActiveTasks []string                // 活跃任务列表（所有未完成的任务）
}

func NewTaskManager(system *System) *TaskManager {
	return &TaskManager{
		System:      system,
		Tasks:       make(map[string]*define.Task),
		UserTasks:   make(map[uint][]string),
		ActiveTasks: make([]string, 0),
	}
}

// 提交新任务到系统
func (tm *TaskManager) AddTask(base define.TaskBase) (*define.Task, error) {
	task := define.NewTask(base)

	tm.Tasks[task.TaskID] = task
	tm.UserTasks[base.UserID] = append(tm.UserTasks[base.UserID], task.TaskID)
	tm.ActiveTasks = append(tm.ActiveTasks, task.TaskID)

	return task.Copy(), nil
}

// getActiveTasks 获取所有活跃任务（未完成的任务）
func (tm *TaskManager) getActiveTasks() []*define.Task {
	tasks := make([]*define.Task, 0, len(tm.ActiveTasks))
	for _, taskID := range tm.ActiveTasks {
		if task, exists := tm.Tasks[taskID]; exists {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// updateFromTaskState 从TaskState快照同步结果到Task
func (tm *TaskManager) updateFromTaskState(state *TaskState, tasks []*define.Task, sys *System) {
	for _, task := range tasks {
		snap, ok := state.Snapshots[task.TaskID]
		if !ok {
			continue
		}

		if snap.AssignedCommID > 0 {
			// 同步分配结果（直接使用ID）
			task.AssignedCommID = snap.AssignedCommID
			task.TransferPath = snap.TransferPath.Copy()

			// 更新队列和已处理数据
			processed := snap.CurrentQueue + snap.PendingTransferData - snap.NextQueue
			if processed > 0 {
				task.ProcessedData += processed
			}
			task.QueuedData = snap.NextQueue
			task.AllocResource = snap.ResourceFraction

			// 同步性能指标到 Task.Metrics（直接拷贝整个结构体）
			if task.Metrics == nil {
				task.Metrics = &define.TaskMetrics{}
			}
			*task.Metrics = snap.Metrics

			// 更新任务状态
			if task.Status == define.TaskPending {
				task.Status = define.TaskQueued
				task.ScheduledTime = time.Now()
			}

			if snap.NextQueue > 0 && task.Status == define.TaskQueued {
				task.Status = define.TaskComputing
			}

			// 检查是否完成
			if snap.NextQueue < 0.001 && task.ProcessedData >= task.DataSize-0.001 {
				task.Status = define.TaskCompleted
				task.CompleteTime = time.Now()
				task.ProcessedData = task.DataSize
				task.QueuedData = 0
			}
		}
	}

	tm.updateActiveTasksList()
}

// updateActiveTasksList 更新活跃任务列表（移除已完成和失败的任务）
func (tm *TaskManager) updateActiveTasksList() {
	newActive := make([]string, 0)

	for _, taskID := range tm.ActiveTasks {
		task, exists := tm.Tasks[taskID]
		if !exists {
			continue
		}
		// 只保留未完成的任务
		if task.Status != define.TaskCompleted && task.Status != define.TaskFailed {
			newActive = append(newActive, taskID)
		}
	}

	tm.ActiveTasks = newActive
}

// hasActiveTasks 检查是否有活跃任务
func (tm *TaskManager) hasActiveTasks() bool {
	return len(tm.ActiveTasks) > 0
}
