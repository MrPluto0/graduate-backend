package algorithm

import (
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/utils"
	"math"
	"math/rand"
)

type Graph struct {
	System     *System
	CommMat    [][]float64 // 通信设备间速率矩阵
	ShortPaths [][][]int   // 最短路径矩阵
}

func NewGraph(system *System) *Graph {
	graph := &Graph{
		System:  system,
		CommMat: [][]float64{},
	}

	graph.constructGraph()
	graph.calcByFloyd()

	return graph
}

// 构建网络拓扑图：用户连接最近通信设备，通信设备间根据距离阈值建立连接
func (g *Graph) constructGraph() {
	for _, user := range g.System.Users {
		user.CalcNearest(g.System.Comms)
	}

	// 初始化通信设备间速率矩阵
	B := len(g.System.Comms)
	speeds := make([][]float64, B)
	for i := range speeds {
		speeds[i] = make([]float64, B)
		for j := range speeds[i] {
			speeds[i][j] = -1
		}
	}

	for i := range B {
		for j := range B {
			if speeds[i][j] != -1 {
				continue
			} else if i == j {
				speeds[i][j] = math.Inf(1)
				continue
			} else {
				d := utils.Distance(g.System.Comms[i].X, g.System.Comms[i].Y, g.System.Comms[j].X, g.System.Comms[j].Y)
				if d > constant.Radius {
					speeds[i][j] = 0
					speeds[j][i] = 0
				} else {
					speed := utils.TransferSpeed(constant.P_b, d)
					speeds[i][j] = speed
					speeds[j][i] = speed
				}
			}
		}
	}

	g.CommMat = speeds
}

// Floyd算法计算最短路径
func (g *Graph) calcByFloyd() {
	// 将速率的倒数作为权重（和时间呈正相关）
	weight := make([][]float64, len(g.CommMat))
	for i := range g.CommMat {
		weight[i] = make([]float64, len(g.CommMat))
		for j := range g.CommMat[i] {
			if g.CommMat[i][j] == 0 {
				weight[i][j] = math.Inf(1)
			} else {
				weight[i][j] = 1 / g.CommMat[i][j]
			}
		}
	}

	result := utils.Floyd(weight)
	g.ShortPaths = result.Paths
}

// 调度器：为每个用户选择最优计算节点
func (g *Graph) Scheduler(state State) State {
	U := len(g.System.Users)
	B := len(g.System.Comms)

	// 随机打乱用户顺序
	userIndices := make([]int, U)
	for i := range userIndices {
		userIndices[i] = i
	}
	rand.Shuffle(U, func(i, j int) {
		userIndices[i], userIndices[j] = userIndices[j], userIndices[i]
	})

	nextState := state.Copy()

	for _, userIdx := range userIndices {
		user := g.System.Users[userIdx]

		if state.R[userIdx] == 0 {
			continue
		}

		bestCost := math.Inf(1)
		var bestState State

		// 找到用户最近的通信设备索引
		startIdx := -1
		for i, comm := range g.System.Comms {
			if comm.ID == user.Nearest {
				startIdx = i
				break
			}
		}

		if startIdx == -1 || startIdx >= len(g.ShortPaths) {
			continue
		}

		// 遍历所有可能的计算节点，选择成本最低的方案
		for endIdx := range B {
			var path []int
			if startIdx == endIdx {
				path = []int{startIdx}
			} else if endIdx < len(g.ShortPaths[startIdx]) {
				path = g.ShortPaths[startIdx][endIdx]
				if len(path) == 0 {
					continue
				}
			} else {
				continue
			}

			tempState := nextState.Copy()
			tempState.UpdateDelta(userIdx, endIdx)
			tempState.UpdateEpsilon(userIdx, path, user.Speed, g.CommMat)
			tempState.UpdateF()
			tempState.ComputeData()

			cost := tempState.Objective()

			if cost < bestCost {
				bestCost = cost
				bestState = tempState
			}
		}

		if bestCost < math.Inf(1) {
			nextState = bestState
		}
	}

	// 最终更新资源分配
	nextState.UpdateF()
	nextState.ComputeData()

	return nextState
}
