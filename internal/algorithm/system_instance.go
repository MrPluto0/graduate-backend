package algorithm

import "sync"

// 全局SystemV2实例 (保持单例模式兼容性,但允许创建多个实例用于测试)
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

// UseSystemV2 切换到使用SystemV2 (迁移开关)
var UseSystemV2 = true // 设为true启用新系统

// GetCurrentSystem 获取当前使用的系统
func GetCurrentSystem() interface{} {
	if UseSystemV2 {
		return GetSystemV2Instance()
	}
	return GetSystemInstance()
}
