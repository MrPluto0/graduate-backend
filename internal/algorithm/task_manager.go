package algorithm

import (
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/define"
	"math"
	"time"
)

// 任务管理器
type TaskManager struct {
	System    *System
	Tasks     map[string]*define.Task // 任务映射 taskID -> define.Task（快速查找）
	TaskList  []*define.Task          // 任务列表（按创建时间顺序）
	UserTasks map[uint][]string       // 用户任务列表 userID -> []taskID
}

func NewTaskManager(system *System) *TaskManager {
	return &TaskManager{
		System:    system,
		Tasks:     make(map[string]*define.Task),
		TaskList:  make([]*define.Task, 0),
		UserTasks: make(map[uint][]string),
	}
}

// 提交新任务到系统
func (tm *TaskManager) addTask(base define.TaskBase) (*define.Task, error) {
	task := define.NewTask(base)

	tm.Tasks[task.ID] = task
	tm.TaskList = append(tm.TaskList, task)
	tm.UserTasks[base.UserID] = append(tm.UserTasks[base.UserID], task.ID)

	return task, nil
}

// 获取任务列表（按创建时间顺序）
func (tm *TaskManager) getTasks(userID *uint, status *define.TaskStatus) []*define.Task {
	tasks := make([]*define.Task, 0)

	// 遍历 TaskList 保证顺序
	for _, task := range tm.TaskList {
		// 如果指定了用户ID，只返回该用户的任务
		if userID != nil && task.UserID != *userID {
			continue
		}

		// 如果指定了状态，只返回匹配状态的任务
		if status != nil && task.Status != *status {
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks
}

// 根据ID获取任务
func (tm *TaskManager) getTaskByID(taskID string) (*define.Task, bool) {
	task, exists := tm.Tasks[taskID]
	if !exists {
		return nil, false
	}

	return task, true
}

// 删除任务
func (tm *TaskManager) deleteTask(taskID string) bool {
	task, exists := tm.Tasks[taskID]
	if !exists {
		return false
	}

	// 从 Tasks map 中删除
	delete(tm.Tasks, taskID)

	// 从 TaskList 中删除
	newTaskList := make([]*define.Task, 0, len(tm.TaskList)-1)
	for _, t := range tm.TaskList {
		if t.ID != taskID {
			newTaskList = append(newTaskList, t)
		}
	}
	tm.TaskList = newTaskList

	// 从 UserTasks 中删除
	if taskIDs, ok := tm.UserTasks[task.UserID]; ok {
		newTaskIDs := make([]string, 0, len(taskIDs)-1)
		for _, id := range taskIDs {
			if id != taskID {
				newTaskIDs = append(newTaskIDs, id)
			}
		}
		tm.UserTasks[task.UserID] = newTaskIDs
	}

	return true
}

// 获取所有活跃任务（未完成的任务）
func (tm *TaskManager) getActiveTasks() []*define.Task {
	tasks := make([]*define.Task, 0)
	for _, task := range tm.TaskList {
		if task.Status != define.TaskCompleted && task.Status != define.TaskFailed {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// 从State快照同步结果到Task
func (tm *TaskManager) syncFromState(state *State, tasks []*define.Task, sys *System) {
	for _, task := range tasks {
		snap, ok := state.Snapshots[task.ID]
		if !ok {
			continue
		}

		if snap.AssignedCommID > 0 {
			// 获取当前队列和已处理数据（单位：bits）
			currentQueuedData := task.GetQueuedData()
			currentProcessedData := task.GetProcessedData()

			// 计算实际传输量（受时隙限制，使用等效速度）
			var actualReceivedData float64 // bits
			if snap.TransferPath != nil {
				equivalentSpeed := snap.TransferPath.CalcEquivalentSpeed()
				if equivalentSpeed > 0 {
					maxTransferInSlot := equivalentSpeed * constant.Slot // bits
					actualReceivedData = math.Min(snap.PendingTransferData, maxTransferInSlot)
				}
			}

			// 实际处理的数据量（计算完成的数据，单位：bits），不能超过队列中的数据量
			maxProcessedData := snap.ResourceFraction * constant.C * constant.Slot / constant.Rho
			processedData := math.Min(maxProcessedData, currentQueuedData+actualReceivedData)

			// 计算新的队列状态和计算累计已处理数据
			newQueuedData := currentQueuedData + actualReceivedData - processedData
			newProcessedData := currentProcessedData + processedData

			// 记录本时隙的执行历史
			slotMetrics := define.SlotMetrics{
				TimeSlot:            state.TimeSlot,
				TransferredData:     actualReceivedData,
				ProcessedData:       processedData,
				QueuedData:          newQueuedData,
				CumulativeProcessed: newProcessedData,
				ResourceFraction:    snap.ResourceFraction,
				TaskMetrics: snap.ComputeMetrics(
					actualReceivedData, // 实际传输的数据量 (bits)
					processedData,      // 实际处理的数据量 (bits)
				),
			}
			task.MetricsHistory = append(task.MetricsHistory, slotMetrics)

			// 同步其他字段
			task.AssignedCommID = snap.AssignedCommID
			task.TransferPath = snap.TransferPath
			task.TransferPath.Path = append([]uint{task.UserID}, task.TransferPath.Path...)

			// 更新任务状态
			if task.Status == define.TaskPending {
				task.Status = define.TaskQueued
				task.ScheduledTime = time.Now()
			}

			if newQueuedData > 0 && task.Status == define.TaskQueued {
				task.Status = define.TaskComputing
			}

			// 检查是否完成
			if newQueuedData < 0.001 && newProcessedData >= task.DataSize-0.001 {
				task.Status = define.TaskCompleted
				task.CompleteTime = time.Now()
			}
		}
	}
}
