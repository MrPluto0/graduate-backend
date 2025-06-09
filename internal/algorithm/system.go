package algorithm

import (
	"go-backend/internal/models"
)

type UserDevice struct {
	ID     uint `json:"id"`
	Device models.Device
	Task   uint
}

type CommDevice struct {
	ID     uint `json:"id"`
	Device models.Device
}

type System struct {
	UserNodes []UserDevice `json:"users"` // 用户设备列表
	CommNodes []CommDevice `json:"comm"`  // 通信设备列表

	T     uint   // 系统时间
	Q     []uint // 通信设备的队列长度
	QMid  uint   // 通信设备的中间队列长度
	QNext uint   // 通信设备的下一步队列长度

}

// // NewSystem 创建一个新的系统实例
// func NewSystem() *System {
// 	return &System{
// 		userNodes: make([]UserDevice, 0),
// 		commNodes: make([]CommDevice, 0),
// 	}
// }

// // AddUserDevice 添加用户设备
// func (s *System) addUserDevice(device models.Device, task uint) {
// 	s.userNodes = append(s.userNodes, UserDevice{
// 		ID:     device.ID,
// 		Device: device,
// 		Task:   task,
// 	})
// }
