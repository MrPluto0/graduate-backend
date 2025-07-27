package algorithm

import "go-backend/internal/models"

func ConstructTopoByNodes(nodes []models.Node) ([]models.Link, error) {
	// 这里可以实现具体的拓扑构建算法
	// 例如，基于节点的连接关系生成链路
	var links []models.Link

	// 假设每个节点与下一个节点相连
	for i := 0; i < len(nodes)-1; i++ {
		link := models.Link{
			SourceID: nodes[i].ID,
			TargetID: nodes[i+1].ID,
		}
		links = append(links, link)
	}

	return links, nil

}
