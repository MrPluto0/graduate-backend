package algorithm

import (
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/utils"
	"math"
	"time"
)

type Graph struct {
	System     *System     // 所属系统
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

// 构建网络拓扑图
// 规则：用户设备连接最近的通信设备；通信设备之间根据距离阈值Radius连接
func (g *Graph) constructGraph() {
	// 1. 用户设备找到最近的通信设备
	for _, user := range g.System.Users {
		user.CalcNearest(g.System.Comms)
	}

	// 2. 初始化通信设备之间的速率矩阵
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
			// 跳过已计算的
			if speeds[i][j] != -1 {
				continue
			} else if i == j {
				// 自己到自己，速率为无穷大（表示无延迟）
				speeds[i][j] = math.Inf(1)
				continue
			} else {
				d := utils.Distance(g.System.Comms[i].X, g.System.Comms[i].Y, g.System.Comms[j].X, g.System.Comms[j].Y)
				// 超过通信半径，无法连接
				if d > constant.Radius {
					speeds[i][j] = 0
					speeds[j][i] = 0 // 对称矩阵
				} else {
					// 使用基站发射功率P_b计算传输速率
					speed := utils.TransferSpeed(constant.P_b, d)
					speeds[i][j] = speed
					speeds[j][i] = speed // 对称矩阵
				}
			}
		}
	}

	g.CommMat = speeds
}

// 使用Floyd算法计算最短路径
func (g *Graph) calcByFloyd() {
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

func (g *Graph) Scheduler(state State) State {
	U := len(g.System.Users) // 用户数量
	B := len(g.System.Comms) // 通信设备数量

	// 创建用户索引列表并打乱顺序
	userIndices := make([]int, U)
	for i := range userIndices {
		userIndices[i] = i
	}
	shuffleUserIndices(userIndices)

	nextState := state.Copy()

	// 随机遍历用户，找出每个用户的最佳决策
	for _, userIdx := range userIndices {
		user := g.System.Users[userIdx]

		// 如果用户的新数据为0，则跳过该用户
		if state.R[userIdx] == 0 {
			continue
		}

		bestCost := math.Inf(1) // 最佳成本
		start := user.Nearest   // 距离最近的通信设备索引
		var bestState State

		// 找到start在通信设备列表中的索引
		startIdx := -1
		for i, comm := range g.System.Comms {
			if comm.ID == start {
				startIdx = i
				break
			}
		}

		// 以end为终点，更新当前决策的状态，找出最佳计算节点
		for endIdx := range B {
			// start -> end 的最短传输路径
			var path []int
			if startIdx == endIdx {
				path = []int{startIdx}
			} else if startIdx < len(g.ShortPaths) && endIdx < len(g.ShortPaths[startIdx]) {
				path = g.ShortPaths[startIdx][endIdx]
			} else {
				continue // 无有效路径
			}

			// 如果路径为空或不可达，跳过
			if len(path) == 0 {
				continue
			}

			// 创建临时状态
			tempState := nextState.Copy()

			// 更新决策变量
			tempState.UpdateDelta(userIdx, endIdx)
			tempState.UpdateEpsilon(userIdx, path, user.Speed, g.CommMat)
			tempState.UpdateF()
			tempState.ComputeData()

			// 计算当前决策成本
			cost := tempState.Objective()

			// 更新最佳决策方案
			if cost < bestCost {
				bestCost = cost
				bestState = tempState
			}
		}

		// 如果找到了更好的决策，更新nextState
		if bestCost < math.Inf(1) {
			nextState = bestState
		}
	}

	// 如果没有新数据，路径决策不会执行，因此再执行一次资源分配
	nextState.UpdateF()
	nextState.ComputeData()

	return nextState
}

// shuffleUserIndices 使用Fisher-Yates算法打乱用户索引顺序
func shuffleUserIndices(indices []int) {
	for i := len(indices) - 1; i > 0; i-- {
		j := randInt(0, i+1)
		indices[i], indices[j] = indices[j], indices[i]
	}
}

// randInt 生成[min, max)范围内的随机整数
func randInt(min, max int) int {
	return min + int(math.Floor(math.Mod(float64(time.Now().UnixNano()), float64(max-min))))
}
