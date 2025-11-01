package service

import (
	"go-backend/internal/models"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type MonitorService struct{}

func NewMonitorService() *MonitorService {
	return &MonitorService{}
}

// GetSystemMetrics 获取系统监控指标
func (s *MonitorService) GetSystemMetrics() (*models.SystemMetrics, error) {
	metrics := &models.SystemMetrics{
		Timestamp: time.Now(),
	}

	// 获取CPU使用率
	cpuPercent, err := cpu.Percent(time.Millisecond*100, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPUUsage = cpuPercent[0]
	}

	// 获取内存信息
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		metrics.MemTotal = memInfo.Total
		metrics.MemUsed = memInfo.Used
		metrics.MemFree = memInfo.Free
		metrics.MemUsageRate = memInfo.UsedPercent
	}

	// 获取Goroutine数量
	metrics.GoroutineCount = runtime.NumGoroutine()

	return metrics, nil
}
