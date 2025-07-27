package service

import (
	"errors"
	"go-backend/internal/models"
	"go-backend/internal/repository"
)

type NetworkService struct {
	nodeRepo *repository.NodeRepository
	linkRepo *repository.LinkRepository
}

func NewNetworkService(nodeRepo *repository.NodeRepository, linkRepo *repository.LinkRepository) *NetworkService {
	return &NetworkService{
		nodeRepo: nodeRepo,
		linkRepo: linkRepo,
	}
}

// ListNodesWithPage 获取分页的节点列表
func (s *NetworkService) ListNodesWithPage(offset, size int, filters map[string]interface{}) ([]models.Node, int64, error) {
	return s.nodeRepo.ListWithPage(offset, size, filters)
}

// GetNode 获取单个节点
func (s *NetworkService) GetNode(id uint) (*models.Node, error) {
	return s.nodeRepo.GetByID(id)
}

// CreateNode 创建节点
func (s *NetworkService) CreateNode(node *models.Node) error {
	// 如果设置了设备ID，检查是否重复
	if node.DeviceID != nil {
		existingNode, err := s.nodeRepo.GetByDeviceID(*node.DeviceID)
		if err == nil && existingNode != nil {
			return errors.New("该设备已经被其他节点关联")
		}
	}
	return s.nodeRepo.Create(node)
}

// UpdateNode 更新节点
func (s *NetworkService) UpdateNode(node *models.Node) error {
	// 检查节点是否存在
	existingNode, err := s.nodeRepo.GetByID(node.ID)
	if err != nil {
		return errors.New("节点不存在")
	}

	// 如果设备ID发生变化且不为空，检查新设备ID是否已被其他节点使用
	if node.DeviceID != nil && (existingNode.DeviceID == nil || *existingNode.DeviceID != *node.DeviceID) {
		nodeWithDevice, err := s.nodeRepo.GetByDeviceID(*node.DeviceID)
		if err == nil && nodeWithDevice != nil && nodeWithDevice.ID != node.ID {
			return errors.New("该设备已经被其他节点关联")
		}
	}

	return s.nodeRepo.Update(node)
}

// DeleteNode 删除节点
func (s *NetworkService) DeleteNode(id uint) error {
	// 检查是否存在关联的链路
	links, err := s.linkRepo.List(map[string]interface{}{
		"source_id": id,
		"target_id": id,
	})
	if err != nil {
		return err
	}
	if len(links) > 0 {
		return errors.New("请先删除与该节点关联的链路")
	}
	return s.nodeRepo.Delete(id)
}

// ListLinksWithPage 获取分页的链路列表
func (s *NetworkService) ListLinks(offset, size int, filters map[string]interface{}) ([]models.Link, int64, error) {
	return s.linkRepo.ListWithPage(offset, size, filters)
}

// GetLink 获取单个链路
func (s *NetworkService) GetLink(id uint) (*models.Link, error) {
	return s.linkRepo.GetByID(id)
}

// CreateLink 创建链路
func (s *NetworkService) CreateLink(link *models.Link) error {
	// 检查链路名称是否已存在
	if link.SourceID == link.TargetID {
		return errors.New("源节点和目标节点不能相同")
	}

	// 检查源节点和目标节点是否存在
	_, err := s.nodeRepo.GetByID(link.SourceID)
	if err != nil {
		return errors.New("源节点不存在")
	}
	_, err = s.nodeRepo.GetByID(link.TargetID)
	if err != nil {
		return errors.New("目标节点不存在")
	}

	// 检查是否已存在相同的链路
	existingLink, _ := s.linkRepo.GetByNodes(link.SourceID, link.TargetID)
	if existingLink != nil {
		return errors.New("链路已存在")
	}

	return s.linkRepo.Create(link)
}

// UpdateLink 更新链路
func (s *NetworkService) UpdateLink(link *models.Link) error {
	// 检查链路是否存在
	existing, err := s.linkRepo.GetByID(link.ID)
	if err != nil {
		return errors.New("链路不存在")
	}

	// 如果更改了源节点或目标节点，需要验证节点是否存在
	if existing.SourceID != link.SourceID {
		if _, err := s.nodeRepo.GetByID(link.SourceID); err != nil {
			return errors.New("源节点不存在")
		}
	}
	if existing.TargetID != link.TargetID {
		if _, err := s.nodeRepo.GetByID(link.TargetID); err != nil {
			return errors.New("目标节点不存在")
		}
	}

	return s.linkRepo.Update(link)
}

// DeleteLink 删除链路
func (s *NetworkService) DeleteLink(id uint) error {
	// 检查链路是否存在
	_, err := s.linkRepo.GetByID(id)
	if err != nil {
		return errors.New("链路不存在")
	}
	return s.linkRepo.Delete(id)
}

// TopologyData 网络拓扑数据结构
type TopologyData struct {
	Nodes []models.Node `json:"nodes"`
	Links []models.Link `json:"links"`
}

// GetTopology 获取完整的网络拓扑数据
func (s *NetworkService) GetTopology() (*TopologyData, error) {
	// 获取所有节点（不分页）
	nodes, err := s.nodeRepo.List(nil)
	if err != nil {
		return nil, err
	}

	// 获取所有链路（不分页）
	links, err := s.linkRepo.List(nil)
	if err != nil {
		return nil, err
	}

	return &TopologyData{
		Nodes: nodes,
		Links: links,
	}, nil
}

// BatchUpdateNodesPosition 批量更新节点位置
func (s *NetworkService) BatchUpdateNodesPosition(nodes []models.Node) error {
	if len(nodes) == 0 {
		return errors.New("节点位置列表不能为空")
	}

	// 提取所有节点ID
	var nodeIDs []uint
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.ID)
	}

	// 检查所有节点是否存在
	existingNodes, err := s.nodeRepo.GetByIDs(nodeIDs)
	if err != nil {
		return err
	}

	// 验证所有请求的节点都存在
	if len(existingNodes) != len(nodes) {
		existingIDMap := make(map[uint]bool)
		for _, node := range existingNodes {
			existingIDMap[node.ID] = true
		}

		var missingIDs []uint
		for _, node := range nodes {
			if !existingIDMap[node.ID] {
				missingIDs = append(missingIDs, node.ID)
			}
		}

		if len(missingIDs) > 0 {
			return errors.New("部分节点不存在")
		}
	}

	// 执行批量更新
	return s.nodeRepo.BatchUpdatePositions(nodes)
}

// 依据拓扑连接规则，对网络结构进行更新：
// 1. 用户节点只连接距离最近的UAV
// 2. UAV只连接距离在给定范围内的其他UAV

const Radius = 1000

// func (s *NetworkService) RefreshNetworkTopo(nodes []models.Node) error {
// 	links, err := s.linkRepo.List(nil)
// 	if err != nil {
// 		return errors.New("获取网络链路失败")
// 	}

// 	newLinks := make([]models.Link, 0)

// }
