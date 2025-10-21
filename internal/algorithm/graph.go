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
	CommMat    map[uint]map[uint]float64              // 通信设备间速率矩阵 [commID][commID]
	ShortPaths map[uint]map[uint]*define.TransferPath // 最短路径 [startID][endID]TransferPath（包含路径、速率、功率）
}

func NewGraph(system *System) *Graph {
	graph := &Graph{
		System:     system,
		CommMat:    make(map[uint]map[uint]float64),
		ShortPaths: make(map[uint]map[uint]*define.TransferPath),
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

	// 初始化通信设备间速率矩阵 (使用ID作为key)
	for i, commI := range g.System.Comms {
		g.CommMat[commI.ID] = make(map[uint]float64)

		for j, commJ := range g.System.Comms {
			if i == j {
				g.CommMat[commI.ID][commJ.ID] = math.Inf(1)
			} else {
				d := utils.Distance(commI.X, commI.Y, commJ.X, commJ.Y)
				if d > constant.Radius {
					g.CommMat[commI.ID][commJ.ID] = 0
				} else {
					speed := utils.TransferSpeed(constant.P_b, d)
					g.CommMat[commI.ID][commJ.ID] = speed
				}
			}
		}
	}
}

// Floyd算法计算最短路径
func (g *Graph) calcByFloyd() {
	n := len(g.System.Comms)

	// 构建基于索引的权重矩阵（Floyd 算法需要）
	weight := make([][]float64, n)
	for i := range weight {
		weight[i] = make([]float64, n)
		commI := g.System.Comms[i]
		for j := range weight[i] {
			commJ := g.System.Comms[j]
			speed := g.CommMat[commI.ID][commJ.ID]
			if speed == 0 {
				weight[i][j] = math.Inf(1)
			} else {
				weight[i][j] = 1 / speed
			}
		}
	}

	// 执行 Floyd 算法（基于索引）
	result := utils.Floyd(weight)

	// 将结果转换为基于 ID 的 TransferPath 对象
	for i, commI := range g.System.Comms {
		g.ShortPaths[commI.ID] = make(map[uint]*define.TransferPath)
		for j, commJ := range g.System.Comms {
			indexPath := result.Paths[i][j]
			idPath := make([]uint, len(indexPath))
			for k, idx := range indexPath {
				idPath[k] = g.System.Comms[idx].ID
			}

			// 构建 TransferPath,预计算 Speeds 和 Powers
			transferPath := &define.TransferPath{
				Path:   idPath,
				Speeds: make([]float64, len(idPath)),
				Powers: make([]float64, len(idPath)),
			}

			// 预计算每段的传输速度和功率
			for k := 0; k < len(idPath); k++ {
				if k == 0 {
					// 第一段:用户到第一个通信设备，速度和功率在 AssignTask 时根据 user.Speed 计算
					transferPath.Speeds[k] = 0 // 占位,稍后填充
					transferPath.Powers[k] = 0 // 占位,稍后填充
				} else {
					// 后续段:通信设备之间
					prevID := idPath[k-1]
					currID := idPath[k]
					transferPath.Speeds[k] = g.CommMat[prevID][currID]
					transferPath.Powers[k] = constant.P_b
				}
			}

			g.ShortPaths[commI.ID][commJ.ID] = transferPath
		}
	}
}

// 任务维度调度器：为每个任务选择最优计算节点
func (g *Graph) Schedule(state *TaskState, tasks []*define.Task) *TaskState {
	// 创建任务切片的副本并打乱顺序
	shuffledTasks := make([]*define.Task, len(tasks))
	copy(shuffledTasks, tasks)
	rand.Shuffle(len(shuffledTasks), func(i, j int) {
		shuffledTasks[i], shuffledTasks[j] = shuffledTasks[j], shuffledTasks[i]
	})

	nextState := state.Copy()

	for _, task := range shuffledTasks {
		snap, ok := nextState.Snapshots[task.TaskID]
		if !ok || snap.PendingTransferData == 0 {
			continue
		}

		// 只为 Pending 状态的任务分配设备（锁定机制）
		if task.Status != define.TaskPending {
			continue
		}

		// 获取任务对应的用户
		user, ok := g.System.UserMap[task.UserID]
		if !ok {
			continue
		}

		// 获取用户最近的通信设备 ID
		startCommID := user.Nearest
		if _, ok := g.ShortPaths[startCommID]; !ok {
			continue
		}

		bestCost := math.Inf(1)
		var bestState *TaskState

		// 遍历所有可能的计算节点,选择成本最低的方案
		for _, endComm := range g.System.Comms {
			endCommID := endComm.ID

			// 获取从 startCommID 到 endCommID 的 TransferPath
			transferPath, ok := g.ShortPaths[startCommID][endCommID]
			if !ok || len(transferPath.Path) == 0 {
				continue
			}

			tempState := nextState.Copy()
			tempState.AssignTask(task.TaskID, endCommID, transferPath, user.Speed)
			cost := tempState.Objective()

			if cost < bestCost {
				bestCost = cost
				bestState = tempState
			}
		}

		if bestCost < math.Inf(1) && bestState != nil {
			nextState = bestState
		}
	}

	// 最终计算一次完整的指标
	nextState.Objective()

	return nextState
}
