package algorithm

import (
	"go-backend/internal/algorithm/define"
	"sync"
)

// AssignmentManager 管理任务调度分配历史
type AssignmentManager struct {
	// 每个任务的分配历史: TaskID -> []Assignment (按时隙排序)
	History map[string][]*define.Assignment
	mutex   sync.RWMutex
}

// NewAssignmentManager 创建分配管理器
func NewAssignmentManager() *AssignmentManager {
	return &AssignmentManager{
		History: make(map[string][]*define.Assignment),
	}
}

// AddAssignment 添加分配记录
func (am *AssignmentManager) AddAssignment(assign *define.Assignment) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if am.History[assign.TaskID] == nil {
		am.History[assign.TaskID] = make([]*define.Assignment, 0)
	}
	am.History[assign.TaskID] = append(am.History[assign.TaskID], assign)
}

// GetLastAssignment 获取任务的最后一次分配
func (am *AssignmentManager) GetLastAssignment(taskID string) *define.Assignment {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	history := am.History[taskID]
	if len(history) == 0 {
		return nil
	}
	return history[len(history)-1]
}

// GetHistory 获取任务的完整分配历史
func (am *AssignmentManager) GetHistory(taskID string) []*define.Assignment {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	history := am.History[taskID]
	// 返回副本,避免外部修改
	result := make([]*define.Assignment, len(history))
	for i, a := range history {
		result[i] = a.Copy()
	}
	return result
}

// GetCumulativeProcessed 获取任务的累计处理数据量
func (am *AssignmentManager) GetCumulativeProcessed(taskID string) float64 {
	last := am.GetLastAssignment(taskID)
	if last == nil {
		return 0
	}
	return last.CumulativeProcessed
}

// GetCurrentQueue 获取任务的当前队列数据量
func (am *AssignmentManager) GetCurrentQueue(taskID string, dataSize float64) float64 {
	last := am.GetLastAssignment(taskID)
	if last == nil {
		return 0 // 新任务,队列为空
	}

	// 队列 = 上次队列 + 本次传输 - 本次处理
	queue := last.QueueData + last.TransferredData - last.ProcessedData
	if queue < 0 {
		queue = 0
	}
	return queue
}

// Clear 清空所有历史记录
func (am *AssignmentManager) Clear() {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.History = make(map[string][]*define.Assignment)
}
