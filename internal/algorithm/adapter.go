package algorithm

import (
	"go-backend/internal/algorithm/define"
	"sync"
)

// SystemAdapter 适配器,让SystemV2兼容原System的API
type SystemAdapter struct {
	*SystemV2
}

// NewSystemAdapter 创建适配器
func NewSystemAdapter(sysV2 *SystemV2) *SystemAdapter {
	return &SystemAdapter{SystemV2: sysV2}
}

// SubmitBatchTasks 适配批量提交任务 (兼容旧API)
func (sa *SystemAdapter) SubmitBatchTasks(requests []define.TaskBase) ([]*define.Task, error) {
	tasks := make([]*define.Task, 0, len(requests))

	for _, req := range requests {
		taskV2, err := sa.SystemV2.SubmitTask(req.UserID, req.DataSize, req.Type)
		if err != nil {
			return nil, err
		}

		// 转换TaskV2为Task (兼容层)
		task := taskV2ToTask(taskV2)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// StopAlgorithm 停止算法 (兼容旧API)
func (sa *SystemAdapter) StopAlgorithm() {
	sa.SystemV2.Stop()
}

// ClearHistory 清除历史 (兼容旧API)
func (sa *SystemAdapter) ClearHistory() {
	sa.SystemV2.AssignmentManager.Clear()
	sa.SystemV2.TimeSlot = 0
}

// GetTasksWithPage 分页获取任务 (兼容旧API)
func (sa *SystemAdapter) GetTasksWithPage(offset, limit int, userID *uint, status *define.TaskStatus) ([]*define.Task, int64) {
	tasksV2, total := sa.SystemV2.TaskManager.GetTasksWithPage(offset, limit, userID, status)

	// 转换TaskV2为Task
	tasks := make([]*define.Task, len(tasksV2))
	for i, tv2 := range tasksV2 {
		tasks[i] = taskV2ToTask(tv2)
		// 填充调度信息
		sa.fillTaskAssignmentInfo(tasks[i])
	}

	return tasks, total
}

// GetTaskByID 获取单个任务 (兼容旧API)
func (sa *SystemAdapter) GetTaskByID(taskID string) *define.Task {
	taskV2 := sa.SystemV2.TaskManager.GetTask(taskID)
	if taskV2 == nil {
		return nil
	}

	task := taskV2ToTask(taskV2)
	sa.fillTaskAssignmentInfo(task)
	return task
}

// fillTaskAssignmentInfo 填充任务的分配信息 (TransferPath, MetricsHistory)
func (sa *SystemAdapter) fillTaskAssignmentInfo(task *define.Task) {
	history := sa.SystemV2.AssignmentManager.GetHistory(task.ID)

	if len(history) == 0 {
		return
	}

	// 使用最后一次分配填充TransferPath
	lastAssign := history[len(history)-1]
	task.AssignedCommID = lastAssign.CommID
	task.TransferPath = &define.TransferPath{
		Path:   lastAssign.Path,
		Speeds: lastAssign.Speeds,
		Powers: lastAssign.Powers,
	}

	// 转换Assignment历史为SlotMetrics (兼容旧格式)
	task.MetricsHistory = make([]define.SlotMetrics, len(history))
	for i, assign := range history {
		task.MetricsHistory[i] = define.SlotMetrics{
			TimeSlot:            assign.TimeSlot,
			TransferredData:     assign.TransferredData,
			ProcessedData:       assign.ProcessedData,
			QueuedData:          assign.QueueData,
			CumulativeProcessed: assign.CumulativeProcessed,
			ResourceFraction:    assign.ResourceFraction,
			TaskMetrics:         computeMetrics(assign),
		}
	}
}

// taskV2ToTask 转换TaskV2为Task
func taskV2ToTask(tv2 *define.TaskV2) *define.Task {
	return &define.Task{
		TaskBase: define.TaskBase{
			ID:        tv2.ID,
			Name:      tv2.Name,
			Type:      tv2.Type,
			UserID:    tv2.UserID,
			DataSize:  tv2.DataSize,
			Priority:  tv2.Priority,
			Status:    tv2.Status,
			CreatedAt: tv2.CreatedAt,
		},
		ScheduledTime: tv2.ScheduledTime,
		CompleteTime:  tv2.CompleteTime,
	}
}

// computeMetrics 根据Assignment计算性能指标
func computeMetrics(assign *define.Assignment) define.TaskMetrics {
	metrics := define.TaskMetrics{}

	// 传输延迟
	for _, speed := range assign.Speeds {
		if speed > 0 && assign.TransferredData > 0 {
			metrics.TransferDelay += assign.TransferredData / speed
		}
	}

	// 计算延迟 (简化版本,使用常量)
	if assign.ResourceFraction > 0 && assign.ProcessedData > 0 {
		// 这里需要导入constant包,暂时用估算值
		// metrics.ComputeDelay = assign.ProcessedData * Rho / (assign.ResourceFraction * C)
		metrics.ComputeDelay = assign.ProcessedData * 1e-6 // 简化估算
	}

	// 传输能耗
	for i, power := range assign.Powers {
		if i < len(assign.Speeds) && assign.Speeds[i] > 0 && assign.TransferredData > 0 {
			segmentDelay := assign.TransferredData / assign.Speeds[i]
			metrics.TransferEnergy += power * segmentDelay
		}
	}

	// 计算能耗 (简化估算)
	// Energy = ResourceFraction × Kappa × C³ × Slot
	if assign.ResourceFraction > 0 {
		metrics.ComputeEnergy = assign.ResourceFraction * 1e-12 // 简化估算
	}

	metrics.TotalDelay = metrics.TransferDelay + metrics.ComputeDelay
	metrics.TotalEnergy = metrics.TransferEnergy + metrics.ComputeEnergy

	return metrics
}

// DeleteTask 删除任务 (兼容旧API)
func (sa *SystemAdapter) DeleteTask(taskID string) error {
	// SystemV2暂不支持删除,返回nil
	return nil
}

// 全局SystemV2实例
var (
	globalSystemV2     *SystemV2
	globalSystemV2Once sync.Once
)

// GetSystemV2Instance 获取全局SystemV2实例
func GetSystemV2Instance() *SystemV2 {
	globalSystemV2Once.Do(func() {
		globalSystemV2 = NewSystemV2()
	})
	return globalSystemV2
}

// GetAdaptedSystem 获取适配后的系统 (供API handler使用)
func GetAdaptedSystem() *SystemAdapter {
	return NewSystemAdapter(GetSystemV2Instance())
}
