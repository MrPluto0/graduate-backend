package service

import (
	"errors"
	"go-backend/internal/models"
	"go-backend/internal/repository"
)

type DeviceService struct {
	deviceRepo *repository.DeviceRepository
	nodeRepo   *repository.NodeRepository
	linkRepo   *repository.LinkRepository
}

// OverviewStats 系统概览统计数据
type OverviewStats struct {
	DeviceCount int64 `json:"device_count"` // 设备总数
	ActiveCount int64 `json:"active_count"` // 活跃设备数
	NodeCount   int64 `json:"node_count"`   // 节点总数
	LinkCount   int64 `json:"link_count"`   // 链路总数
}

func NewDeviceService(deviceRepo *repository.DeviceRepository, nodeRepo *repository.NodeRepository, linkRepo *repository.LinkRepository) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		nodeRepo:   nodeRepo,
		linkRepo:   linkRepo,
	}
}

// CreateDevice 创建新设备，包含业务逻辑验证
func (s *DeviceService) CreateDevice(device *models.Device) error {
	// 检查MAC地址是否已存在
	existingDevice, _ := s.deviceRepo.GetByMAC(device.MAC)
	if existingDevice != nil {
		return errors.New("设备MAC地址已存在")
	}

	return s.deviceRepo.Create(device)
}

// GetDevice 获取设备详情
func (s *DeviceService) GetDevice(id uint) (*models.Device, error) {
	return s.deviceRepo.GetByID(id)
}

// ListDevices 获取设备列表
func (s *DeviceService) ListDevices(current, size int, filters map[string]interface{}) ([]models.Device, int64, error) {
	return s.deviceRepo.List(current, size, filters)
}

// UpdateDevice 更新设备信息
func (s *DeviceService) UpdateDevice(device *models.Device) error {
	// 检查设备是否存在
	existingDevice, err := s.deviceRepo.GetByID(device.ID)
	if err != nil {
		return errors.New("设备不存在")
	}

	// 如果MAC地址发生变化，检查新MAC地址是否已被使用
	if device.MAC != existingDevice.MAC {
		deviceWithMAC, _ := s.deviceRepo.GetByMAC(device.MAC)
		if deviceWithMAC != nil {
			return errors.New("新MAC地址已被其他设备使用")
		}
	}

	return s.deviceRepo.Update(device)
}

// DeleteDevice 删除设备
func (s *DeviceService) DeleteDevice(id uint) error {
	// 检查设备是否存在
	_, err := s.deviceRepo.GetByID(id)
	if err != nil {
		return errors.New("设备不存在")
	}

	return s.deviceRepo.Delete(id)
}

// GetDeviceByMAC 根据MAC地址获取设备
func (s *DeviceService) GetDeviceByMAC(mac string) (*models.Device, error) {
	return s.deviceRepo.GetByMAC(mac)
}

// GetOverviewStats 获取系统概览统计信息
func (s *DeviceService) GetOverviewStats() (*OverviewStats, error) {
	stats := &OverviewStats{}

	// 获取设备总数
	var err error
	stats.DeviceCount, err = s.deviceRepo.Count(nil)
	if err != nil {
		return nil, err
	}

	// 获取活跃（在线）设备数量
	stats.ActiveCount, err = s.deviceRepo.Count(map[string]interface{}{
		"status": "online",
	})
	if err != nil {
		return nil, err
	}

	// 获取节点总数
	stats.NodeCount, err = s.nodeRepo.Count(nil)
	if err != nil {
		return nil, err
	}

	// 获取链路总数
	stats.LinkCount, err = s.linkRepo.Count(nil)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
