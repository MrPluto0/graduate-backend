package define

import (
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/utils"
	"go-backend/internal/models"
	"math"
)

type UserDevice struct {
	models.Node

	Nearest uint    // 最近的通信设备ID
	Speed   float64 // 到最近通信设备的传输速率
}

func NewUserDevice(node models.Node) *UserDevice {
	return &UserDevice{
		Node: node,
	}
}

// CalcNearest 计算最近的通信设备及传输速率
func (u *UserDevice) CalcNearest(commDevices []*CommDevice) {
	if len(commDevices) == 0 {
		u.Nearest = 0
		u.Speed = 0
		return
	}

	minDist := math.Inf(1)
	nearestID := uint(0)
	nearestSpeed := 0.0

	// 计算到每个通信设备的距离
	for _, comm := range commDevices {
		// 计算距离
		d := utils.Distance(u.X, u.Y, comm.X, comm.Y)

		// 找到最近的通信设备
		if d < minDist {
			minDist = d
			nearestID = comm.ID
			// 计算传输速率（使用用户设备发射功率P_u）
			nearestSpeed = utils.TransferSpeed(constant.P_u, d)
		}
	}

	u.Nearest = nearestID
	u.Speed = nearestSpeed
}
