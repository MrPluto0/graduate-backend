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
	return s.nodeRepo.Create(node)
}

// UpdateNode 更新节点
func (s *NetworkService) UpdateNode(node *models.Node) error {
	// 检查节点是否存在
	_, err := s.nodeRepo.GetByID(node.ID)
	if err != nil {
		return errors.New("节点不存在")
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
