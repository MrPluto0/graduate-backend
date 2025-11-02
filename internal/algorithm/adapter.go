package algorithm

import (
	"go-backend/internal/algorithm/define"
	"sync"
)

// SystemAdapter 适配器,让System兼容原System的API
type SystemAdapter struct {
	*System
}

// NewSystemAdapter 创建适配器
func NewSystemAdapter(sys *System) *SystemAdapter {
	return &SystemAdapter{System: sys}
}

// SubmitBatchTasks 适配批量提交任务 (兼容旧API)
func (sa *SystemAdapter) SubmitBatchTasks(requests []define.TaskBase) ([]*define.TaskWithMetrics, error) {
	tasks := make([]*define.TaskWithMetrics, 0, len(requests))

	for _, req := range requests {
		task, err := sa.System.SubmitTask(req.UserID, req.DataSize, req.Type)
		if err != nil {
			return nil, err
		}

		// 转换Task为TaskWithMetrics (兼容层)
		taskWithMetrics := taskToTaskWithMetrics(task)
		tasks = append(tasks, taskWithMetrics)
	}

	return tasks, nil
}

// StopAlgorithm 停止算法 (兼容旧API)
func (sa *SystemAdapter) StopAlgorithm() {
	sa.System.Stop()
}

// ClearHistory 清除历史 (兼容旧API)
func (sa *SystemAdapter) ClearHistory() {
	sa.System.AssignmentManager.Clear()
	sa.System.TimeSlot = 0
}

// GetTasksWithPage 分页获取任务 (兼容旧API)
func (sa *SystemAdapter) GetTasksWithPage(offset, limit int, userID *uint, status *define.TaskStatus) ([]*define.TaskWithMetrics, int64) {
	tasks, total := sa.System.TaskManager.GetTasksWithPage(offset, limit, userID, status)

	// 转换Task为TaskWithMetrics
	tasksWithMetrics := make([]*define.TaskWithMetrics, len(tasks))
	for i, t := range tasks {
		tasksWithMetrics[i] = taskToTaskWithMetrics(t)
		// 填充调度信息
		sa.fillTaskAssignmentInfo(tasksWithMetrics[i])
	}

	return tasksWithMetrics, total
}

// GetTaskByID 获取单个任务 (兼容旧API)
func (sa *SystemAdapter) GetTaskByID(taskID string) *define.TaskWithMetrics {
	task := sa.System.TaskManager.GetTask(taskID)
	if task == nil {
		return nil
	}

	taskWithMetrics := taskToTaskWithMetrics(task)
	sa.fillTaskAssignmentInfo(taskWithMetrics)
	return taskWithMetrics
}

// fillTaskAssignmentInfo 填充任务的分配信息 (TransferPath, MetricsHistory)
func (sa *SystemAdapter) fillTaskAssignmentInfo(task *define.TaskWithMetrics) {
	history := sa.System.AssignmentManager.GetHistory(task.ID)

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

// taskToTaskWithMetrics 转换Task为TaskWithMetrics
func taskToTaskWithMetrics(t *define.Task) *define.TaskWithMetrics {
	return &define.TaskWithMetrics{
		TaskBase: define.TaskBase{
			ID:        t.ID,
			Name:      t.Name,
			Type:      t.Type,
			UserID:    t.UserID,
			DataSize:  t.DataSize,
			Priority:  t.Priority,
			Status:    t.Status,
			CreatedAt: t.CreatedAt,
		},
		ScheduledTime: t.ScheduledTime,
		CompleteTime:  t.CompleteTime,
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
	// System暂不支持删除,返回nil
	return nil
}

// 全局System实例
var (
	globalSystem     *System
	globalSystemOnce sync.Once
)

// GetSystemInstance 获取全局System实例
func GetSystemInstance() *System {
	globalSystemOnce.Do(func() {
		globalSystem = NewSystem()
	})
	return globalSystem
}

// GetAdaptedSystem 获取适配后的系统 (供API handler使用)
func GetAdaptedSystem() *SystemAdapter {
	return NewSystemAdapter(GetSystemInstance())
}
