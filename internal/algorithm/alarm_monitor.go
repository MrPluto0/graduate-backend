package algorithm

import (
	"fmt"
	"go-backend/internal/algorithm/define"
	"go-backend/internal/models"
	"go-backend/internal/service"
	"log"
	"sync"
	"time"
)

// AlarmThresholds 告警阈值配置
type AlarmThresholds struct {
	// 延迟阈值 (秒)
	MaxDelay float64
	// 能耗阈值 (焦耳)
	MaxEnergy float64
	// 系统负载阈值
	MaxLoad float64
	// 队列积压阈值 (bits)
	MaxQueueData float64
	// 传输速率下限 (bits/s)
	MinTransferSpeed float64
}

// DefaultAlarmThresholds 默认告警阈值
var DefaultAlarmThresholds = AlarmThresholds{
	MaxDelay:         10.0,  // 10秒
	MaxEnergy:        100.0, // 100焦耳
	MaxLoad:          5.0,   // 负载5倍
	MaxQueueData:     1e8,   // 100MB
	MinTransferSpeed: 1e5,   // 100 Kbps
}

// AlarmMonitor 告警监控器
type AlarmMonitor struct {
	alarmService *service.AlarmService
	thresholds   AlarmThresholds
	mutex        sync.RWMutex

	// 告警去重: 相同类型的告警在一定时间内只产生一次
	lastAlarmTime map[string]time.Time
	cooldown      time.Duration // 告警冷却时间
}

// NewAlarmMonitor 创建告警监控器
func NewAlarmMonitor(alarmService *service.AlarmService) *AlarmMonitor {
	return &AlarmMonitor{
		alarmService:  alarmService,
		thresholds:    DefaultAlarmThresholds,
		lastAlarmTime: make(map[string]time.Time),
		cooldown:      5 * time.Minute, // 同类型告警5分钟内只触发一次
	}
}

// SetThresholds 设置自定义阈值
func (m *AlarmMonitor) SetThresholds(thresholds AlarmThresholds) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.thresholds = thresholds
}

// CheckSystemState 检查系统状态并产生告警
func (m *AlarmMonitor) CheckSystemState(state *define.StateMetrics, tasks []*define.Task) {
	if state == nil {
		return
	}

	m.mutex.RLock()
	thresholds := m.thresholds
	m.mutex.RUnlock()

	// 1. 检查系统延迟
	if state.TotalDelay > thresholds.MaxDelay {
		m.createAlarm(
			"performance_delay",
			"系统延迟过高",
			models.AlarmEventPerformance,
			fmt.Sprintf("系统总延迟 %.2f 秒超过阈值 %.2f 秒 (传输延迟: %.2f秒, 计算延迟: %.2f秒)",
				state.TotalDelay, thresholds.MaxDelay, state.TransferDelay, state.ComputeDelay),
		)
	}

	// 2. 检查系统能耗
	if state.TotalEnergy > thresholds.MaxEnergy {
		m.createAlarm(
			"performance_energy",
			"系统能耗过高",
			models.AlarmEventPerformance,
			fmt.Sprintf("系统总能耗 %.2f 焦耳超过阈值 %.2f 焦耳 (传输能耗: %.2f J, 计算能耗: %.2f J)",
				state.TotalEnergy, thresholds.MaxEnergy, state.TransferEnergy, state.ComputeEnergy),
		)
	}

	// 3. 检查系统负载
	if state.Load > thresholds.MaxLoad {
		m.createAlarm(
			"performance_load",
			"系统负载过高",
			models.AlarmEventPerformance,
			fmt.Sprintf("系统负载 %.2f 倍超过阈值 %.2f 倍，可能导致任务处理缓慢",
				state.Load, thresholds.MaxLoad),
		)
	}

	// 4. 检查队列积压
	if state.TotalQueue > thresholds.MaxQueueData {
		m.createAlarm(
			"network_queue",
			"网络队列积压严重",
			models.AlarmEventNetwork,
			fmt.Sprintf("总队列数据量 %.2f MB超过阈值 %.2f MB，通信设备处理能力不足",
				state.TotalQueue/1e6, thresholds.MaxQueueData/1e6),
		)
	}

	// 5. 检查单个通信设备队列
	for commID, queueData := range state.CommQueues {
		if queueData > thresholds.MaxQueueData*0.5 {
			m.createAlarm(
				fmt.Sprintf("network_queue_comm_%s", commID),
				fmt.Sprintf("通信设备 %s 队列积压", commID),
				models.AlarmEventNetwork,
				fmt.Sprintf("通信设备 %s 队列数据量 %.2f MB，可能存在传输瓶颈",
					commID, queueData/1e6),
			)
		}
	}
}

// CheckTaskFailures 检查任务失败
func (m *AlarmMonitor) CheckTaskFailures(task *define.Task) {
	if task.Status == define.TaskFailed {
		m.createAlarm(
			fmt.Sprintf("task_failed_%s", task.ID),
			fmt.Sprintf("任务失败: %s", task.Name),
			models.AlarmEventSystem,
			fmt.Sprintf("任务 %s (用户ID: %d) 执行失败，数据大小: %.2f MB",
				task.Name, task.UserID, task.DataSize/1e6),
		)
	}
}

// createAlarm 创建告警（带去重）
func (m *AlarmMonitor) createAlarm(alarmKey, name string, eventType models.AlarmEvent, description string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查冷却时间
	if lastTime, exists := m.lastAlarmTime[alarmKey]; exists {
		if time.Since(lastTime) < m.cooldown {
			// 在冷却期内，跳过
			return
		}
	}

	// 创建告警
	alarm := &models.Alarm{
		Name:        name,
		EventType:   eventType,
		Status:      models.AlarmStatusActive,
		Description: description,
	}

	if err := m.alarmService.CreateAlarm(alarm); err != nil {
		log.Printf("[AlarmMonitor] 创建告警失败: %v", err)
		return
	}

	// 记录告警时间
	m.lastAlarmTime[alarmKey] = time.Now()
	log.Printf("[AlarmMonitor] 告警已创建: %s - %s", name, description)
}

// CleanupOldAlarms 清理旧的告警时间记录（定期清理，避免内存泄漏）
func (m *AlarmMonitor) CleanupOldAlarms() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for key, lastTime := range m.lastAlarmTime {
		if now.Sub(lastTime) > m.cooldown*2 {
			delete(m.lastAlarmTime, key)
		}
	}
}
