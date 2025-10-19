package define

import "go-backend/internal/models"

type CommDevice struct {
	models.Node
}

func NewCommDevice(node models.Node) *CommDevice {
	return &CommDevice{
		Node: node,
	}
}
