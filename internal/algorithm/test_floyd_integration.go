package algorithm

import (
	"fmt"
	"log"
)

// TestFloydIntegration 手动测试Floyd路径选择
func TestFloydIntegration() {
	log.Println("=== Floyd路径集成测试 ===")

	// 初始化系统
	sys := NewSystem()
	if !sys.IsInitialized {
		log.Fatal("系统初始化失败")
	}

	// 提交测试任务
	userID := uint(5) // 假设用户5存在
	dataSize := 10.0  // 10 MB

	task, err := sys.SubmitTask(userID, dataSize, "test")
	if err != nil {
		log.Fatalf("提交任务失败: %v", err)
	}

	log.Printf("✓ 任务 %s 已提交", task.ID)

	// 等待一个时隙
	sys.executeOneSlot()

	// 检查分配
	lastAssign := sys.AssignmentManager.GetLastAssignment(task.ID)
	if lastAssign == nil {
		log.Fatal("❌ 未找到任务分配")
	}

	// 打印路径
	fmt.Printf("\n任务 %s 的路径:\n", task.ID)
	fmt.Printf("  CommID: %d\n", lastAssign.CommID)
	fmt.Printf("  Path: %v\n", lastAssign.Path)
	fmt.Printf("  Speeds: %v Mbps\n", lastAssign.Speeds)
	fmt.Printf("  Powers: %v W\n", lastAssign.Powers)
	fmt.Printf("  Cost: transmission_delay + energy_cost\n")

	// 验证路径不是 [12, 12, 12, ...]
	if len(lastAssign.Path) >= 2 && lastAssign.Path[0] == lastAssign.Path[1] {
		log.Fatal("❌ 路径错误: 出现重复节点")
	}

	log.Println("✓ Floyd路径集成测试通过")
}
