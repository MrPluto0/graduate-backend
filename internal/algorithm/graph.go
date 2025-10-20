package algorithm

import (
	"go-backend/internal/algorithm/constant"
	"go-backend/internal/algorithm/define"
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

// 任务维度调度器：为每个任务选择最优计算节点
func (g *Graph) Scheduler(state *TaskState, tasks []*define.Task) *TaskState {
	B := len(g.System.Comms)
	taskCount := len(tasks)

	// 随机打乱任务顺序
	taskIndices := make([]int, taskCount)
	for i := range taskIndices {
		taskIndices[i] = i
	}
	rand.Shuffle(taskCount, func(i, j int) {
		taskIndices[i], taskIndices[j] = taskIndices[j], taskIndices[i]
	})

	nextState := state.Copy()

	for _, taskIdx := range taskIndices {
		task := tasks[taskIdx]
		alloc, ok := nextState.Allocations[task.TaskID]
		if !ok || alloc.R == 0 {
			continue
		}

		// 获取任务对应的用户
		userIdx := task.UserIndex
		if userIdx < 0 || userIdx >= len(g.System.Users) {
			continue
		}
		user := g.System.Users[userIdx]

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

		bestCost := math.Inf(1)
		var bestState *TaskState

		// 遍历所有可能的计算节点,选择成本最低的方案
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
			tempState.AssignTask(task.TaskID, endIdx, path, user.Speed, g.CommMat, g.System)
			cost := tempState.Objective()

			if cost < bestCost {
				bestCost = cost
				bestState = tempState
			}
		}

		if bestCost < math.Inf(1) && bestState != nil {
			nextState = bestState
			// 最佳分配已经在 bestState 中设置好了，不需要再手动赋值
		}
	}

	// 最终计算一次完整的指标
	nextState.Objective()

	return nextState
}
