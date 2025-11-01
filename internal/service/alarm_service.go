package service

import (
	"errors"
	"go-backend/internal/models"
	"go-backend/internal/repository"
	"time"
)

type AlarmService struct {
	alarmRepo *repository.AlarmRepository
}

// AlarmStats 告警统计数据
type AlarmStats struct {
	TotalCount    int64 `json:"total_count"`    // 告警总数
	ActiveCount   int64 `json:"active_count"`   // 活跃告警数
	ResolvedCount int64 `json:"resolved_count"` // 已解决告警数
}

func NewAlarmService(alarmRepo *repository.AlarmRepository) *AlarmService {
	return &AlarmService{
		alarmRepo: alarmRepo,
	}
}

// CreateAlarm 创建新告警
func (s *AlarmService) CreateAlarm(alarm *models.Alarm) error {
	// 设置默认状态为活跃
	if alarm.Status == "" {
		alarm.Status = models.AlarmStatusActive
	}

	// 验证事件类型
	if !s.isValidEventType(alarm.EventType) {
		return errors.New("无效的事件类型")
	}

	// 验证状态
	if !s.isValidStatus(alarm.Status) {
		return errors.New("无效的告警状态")
	}

	return s.alarmRepo.Create(alarm)
}

// GetAlarm 获取告警详情
func (s *AlarmService) GetAlarm(id uint) (*models.Alarm, error) {
	return s.alarmRepo.GetByID(id)
}

// ListAlarms 获取告警列表
func (s *AlarmService) ListAlarms(current, size int, filters map[string]interface{}) ([]models.Alarm, int64, error) {
	return s.alarmRepo.List(current, size, filters)
}

// GetAlarmList 获取告警列表（简化版，只按状态过滤）
func (s *AlarmService) GetAlarmList(page, size int, status string) ([]models.Alarm, int64, error) {
	filters := make(map[string]interface{})
	if status != "" {
		filters["status"] = status
	}
	return s.alarmRepo.List(page, size, filters)
}

// UpdateAlarm 更新告警信息
func (s *AlarmService) UpdateAlarm(alarm *models.Alarm) error {
	// 检查告警是否存在
	existingAlarm, err := s.alarmRepo.GetByID(alarm.ID)
	if err != nil {
		return errors.New("告警不存在")
	}

	// 验证事件类型
	if alarm.EventType != "" && !s.isValidEventType(alarm.EventType) {
		return errors.New("无效的事件类型")
	}

	// 验证状态
	if alarm.Status != "" && !s.isValidStatus(alarm.Status) {
		return errors.New("无效的告警状态")
	}

	// 如果状态变为已解决，设置解决时间
	if alarm.Status == models.AlarmStatusResolved && existingAlarm.Status != models.AlarmStatusResolved {
		now := time.Now()
		alarm.ResolvedAt = &now
	}

	// 如果状态从已解决变为其他状态，清除解决时间
	if alarm.Status != models.AlarmStatusResolved && existingAlarm.Status == models.AlarmStatusResolved {
		alarm.ResolvedAt = nil
	}

	return s.alarmRepo.Update(alarm)
}

// DeleteAlarm 删除告警
func (s *AlarmService) DeleteAlarm(id uint) error {
	// 检查告警是否存在
	_, err := s.alarmRepo.GetByID(id)
	if err != nil {
		return errors.New("告警不存在")
	}

	return s.alarmRepo.Delete(id)
}

// ResolveAlarm 解决告警
func (s *AlarmService) ResolveAlarm(id uint) error {
	alarm, err := s.alarmRepo.GetByID(id)
	if err != nil {
		return errors.New("告警不存在")
	}

	if alarm.Status == models.AlarmStatusResolved {
		return errors.New("告警已经被解决")
	}

	alarm.Status = models.AlarmStatusResolved
	now := time.Now()
	alarm.ResolvedAt = &now
	alarm.UpdatedAt = now

	return s.alarmRepo.Update(alarm)
}

// ReactivateAlarm 重新激活告警
func (s *AlarmService) ReactivateAlarm(id uint) error {
	alarm, err := s.alarmRepo.GetByID(id)
	if err != nil {
		return errors.New("告警不存在")
	}

	if alarm.Status == models.AlarmStatusActive {
		return errors.New("告警已经是活跃状态")
	}

	alarm.Status = models.AlarmStatusActive
	alarm.ResolvedAt = nil
	alarm.UpdatedAt = time.Now()

	return s.alarmRepo.Update(alarm)
}

// GetActiveAlarms 获取所有活跃告警
func (s *AlarmService) GetActiveAlarms() ([]models.Alarm, error) {
	return s.alarmRepo.GetActiveAlarms()
}

// GetAlarmsByStatus 根据状态获取告警
func (s *AlarmService) GetAlarmsByStatus(status models.AlarmStatus) ([]models.Alarm, error) {
	if !s.isValidStatus(status) {
		return nil, errors.New("无效的告警状态")
	}
	return s.alarmRepo.GetByStatus(status)
}

// GetAlarmsByEventType 根据事件类型获取告警
func (s *AlarmService) GetAlarmsByEventType(eventType models.AlarmEvent) ([]models.Alarm, error) {
	if !s.isValidEventType(eventType) {
		return nil, errors.New("无效的事件类型")
	}
	return s.alarmRepo.GetByEventType(eventType)
}

// GetRecentAlarms 获取最近的告警
func (s *AlarmService) GetRecentAlarms(limit int) ([]models.Alarm, error) {
	if limit <= 0 {
		limit = 10 // 默认返回10条
	}
	return s.alarmRepo.GetRecentAlarms(limit)
}

// GetAlarmStats 获取告警统计信息
func (s *AlarmService) GetAlarmStats() (*AlarmStats, error) {
	stats := &AlarmStats{}

	// 获取告警总数
	var err error
	stats.TotalCount, err = s.alarmRepo.Count(nil)
	if err != nil {
		return nil, err
	}

	// 获取活跃告警数量
	stats.ActiveCount, err = s.alarmRepo.Count(map[string]interface{}{
		"status": models.AlarmStatusActive,
	})
	if err != nil {
		return nil, err
	}

	// 获取已解决告警数量
	stats.ResolvedCount, err = s.alarmRepo.Count(map[string]interface{}{
		"status": models.AlarmStatusResolved,
	})
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// BatchResolveAlarms 批量解决告警
func (s *AlarmService) BatchResolveAlarms(ids []uint) error {
	if len(ids) == 0 {
		return errors.New("告警ID列表不能为空")
	}

	for _, id := range ids {
		if err := s.ResolveAlarm(id); err != nil {
			return err
		}
	}

	return nil
}

// BatchDeleteAlarms 批量删除告警
func (s *AlarmService) BatchDeleteAlarms(ids []uint) error {
	if len(ids) == 0 {
		return errors.New("告警ID列表不能为空")
	}

	for _, id := range ids {
		if err := s.DeleteAlarm(id); err != nil {
			return err
		}
	}

	return nil
}

// isValidEventType 验证事件类型是否有效
func (s *AlarmService) isValidEventType(eventType models.AlarmEvent) bool {
	switch eventType {
	case models.AlarmEventHardware, models.AlarmEventNetwork, models.AlarmEventSecurity,
		models.AlarmEventPerformance, models.AlarmEventSystem:
		return true
	default:
		return false
	}
}

// isValidStatus 验证告警状态是否有效
func (s *AlarmService) isValidStatus(status models.AlarmStatus) bool {
	switch status {
	case models.AlarmStatusActive, models.AlarmStatusResolved:
		return true
	default:
		return false
	}
}
