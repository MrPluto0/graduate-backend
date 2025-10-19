package utils

import (
	"go-backend/internal/models"
	"math"
)

// FloydResult Floyd算法的结果
type FloydResult struct {
	Dist  [][]float64 // 最短距离矩阵
	Paths [][][]int   // 最短路径矩阵
}

// Floyd Floyd算法实现
// graph: 图的邻接矩阵表示，graph[i][j]表示顶点i到顶点j的距离
// 返回: 最短距离矩阵和最短路径矩阵
func Floyd(graph [][]float64) *FloydResult {
	numVertices := len(graph)
	if numVertices == 0 {
		return &FloydResult{
			Dist:  [][]float64{},
			Paths: [][][]int{},
		}
	}

	// 初始化距离矩阵，深拷贝图的邻接矩阵
	dist := make([][]float64, numVertices)
	for i := range dist {
		dist[i] = make([]float64, len(graph[i]))
		copy(dist[i], graph[i])
	}

	// 初始化路径矩阵，初始值为 -1 表示无中间顶点
	path := make([][]int, numVertices)
	for i := range path {
		path[i] = make([]int, numVertices)
		for j := range path[i] {
			path[i][j] = -1
		}
	}

	// Floyd核心算法
	for k := 0; k < numVertices; k++ {
		for i := 0; i < numVertices; i++ {
			for j := 0; j < numVertices; j++ {
				// 避免自身到自身的比较
				if i == j {
					continue
				}
				// 检查是否可以通过顶点 k 缩短从 i 到 j 的距离
				if dist[i][k]+dist[k][j] < dist[i][j] {
					dist[i][j] = dist[i][k] + dist[k][j]
					// 更新路径矩阵，k 为 i 到 j 的中间顶点
					path[i][j] = k
				}
			}
		}
	}

	// 构建所有顶点对之间的最短路径
	resultPaths := make([][][]int, numVertices)
	for i := 0; i < numVertices; i++ {
		resultPaths[i] = make([][]int, numVertices)
		for j := 0; j < numVertices; j++ {
			resultPaths[i][j] = getShortestPaths(i, j, dist, path)
		}
	}

	return &FloydResult{
		Dist:  dist,
		Paths: resultPaths,
	}
}

// getShortestPaths 构建从 start 到 end 的最短路径
// start: 起始顶点
// end: 结束顶点
// dist: 距离矩阵
// path: 路径矩阵
// 返回: 最短路径列表
func getShortestPaths(start, end int, dist [][]float64, path [][]int) []int {
	// 如果没有中间顶点
	if path[start][end] == -1 {
		// 如果距离是无穷大，说明不可达
		if math.IsInf(dist[start][end], 1) {
			return []int{}
		}
		// 直接连接
		return []int{start, end}
	}

	// 有中间顶点，递归构建路径
	intermediate := path[start][end]
	leftPath := getShortestPaths(start, intermediate, dist, path)
	rightPath := getShortestPaths(intermediate, end, dist, path)

	// 合并路径，去掉重复的中间节点
	result := make([]int, 0, len(leftPath)+len(rightPath)-1)
	result = append(result, leftPath...)
	if len(rightPath) > 0 {
		result = append(result, rightPath[1:]...)
	}

	return result
}

// ConstructTopoByNodes 根据节点列表构建网络拓扑
func ConstructTopoByNodes(nodes []models.Node) ([]models.Link, error) {
	// 这里可以实现具体的拓扑构建算法
	// 例如，基于节点的连接关系生成链路
	var links []models.Link

	return links, nil
}
