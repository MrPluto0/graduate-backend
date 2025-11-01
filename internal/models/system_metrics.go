package models

import "time"

// SystemMetrics 系统监控指标
type SystemMetrics struct {
	Timestamp time.Time `json:"timestamp"` // 采集时间

	// CPU相关
	CPUUsage float64 `json:"cpu_usage"` // CPU使用率 (0-100)

	// 内存相关
	MemTotal     uint64  `json:"mem_total"`      // 总内存 (bytes)
	MemUsed      uint64  `json:"mem_used"`       // 已用内存 (bytes)
	MemFree      uint64  `json:"mem_free"`       // 空闲内存 (bytes)
	MemUsageRate float64 `json:"mem_usage_rate"` // 内存使用率 (0-100)

	// Goroutine数量
	GoroutineCount int `json:"goroutine_count"`
}
