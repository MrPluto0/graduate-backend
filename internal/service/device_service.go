package service

import (
	"errors"
	"go-backend/internal/models"
	"go-backend/internal/repository"
)

type DeviceService struct {
	repo *repository.DeviceRepository
}

func NewDeviceService(repo *repository.DeviceRepository) *DeviceService {
	return &DeviceService{repo: repo}
}

// CreateDevice 创建新设备，包含业务逻辑验证
func (s *DeviceService) CreateDevice(device *models.Device) error {
	// 检查MAC地址是否已存在
	existingDevice, _ := s.repo.GetByMAC(device.MAC)
	if existingDevice != nil {
		return errors.New("设备MAC地址已存在")
	}

	return s.repo.Create(device)
}

// GetDevice 获取设备详情
func (s *DeviceService) GetDevice(id uint) (*models.Device, error) {
	return s.repo.GetByID(id)
}

// ListDevices 获取设备列表
func (s *DeviceService) ListDevices(page, pageSize int, filters map[string]interface{}) ([]models.Device, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(offset, pageSize, filters)
}

// UpdateDevice 更新设备信息
func (s *DeviceService) UpdateDevice(device *models.Device) error {
	// 检查设备是否存在
	existingDevice, err := s.repo.GetByID(device.ID)
	if err != nil {
		return errors.New("设备不存在")
	}

	// 如果MAC地址发生变化，检查新MAC地址是否已被使用
	if device.MAC != existingDevice.MAC {
		deviceWithMAC, _ := s.repo.GetByMAC(device.MAC)
		if deviceWithMAC != nil {
			return errors.New("新MAC地址已被其他设备使用")
		}
	}

	return s.repo.Update(device)
}

// DeleteDevice 删除设备
func (s *DeviceService) DeleteDevice(id uint) error {
	// 检查设备是否存在
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("设备不存在")
	}

	return s.repo.Delete(id)
}

// GetDeviceByMAC 根据MAC地址获取设备
func (s *DeviceService) GetDeviceByMAC(mac string) (*models.Device, error) {
	return s.repo.GetByMAC(mac)
}
