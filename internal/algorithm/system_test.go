package algorithm

import (
	"go-backend/internal/algorithm/define"
	"sync"
	"testing"
	"time"
)

// TestConcurrentTaskSubmission 测试并发提交任务
func TestConcurrentTaskSubmission(t *testing.T) {
	// 跳过测试,因为需要数据库连接
	t.Skip("需要数据库连接,手动测试")

	sys := NewSystem()
	defer sys.Stop()

	var wg sync.WaitGroup
	numGoroutines := 10
	tasksPerGoroutine := 5

	// 并发提交任务
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < tasksPerGoroutine; j++ {
				_, err := sys.SubmitTask(1, 100.0, "test")
				if err != nil {
					t.Errorf("Goroutine %d: 提交任务失败: %v", goroutineID, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// 验证任务数量
	expectedTasks := numGoroutines * tasksPerGoroutine
	actualTasks := sys.TaskManager.Count()
	if actualTasks != expectedTasks {
		t.Errorf("期望 %d 个任务,实际 %d 个", expectedTasks, actualTasks)
	}
}

// TestConcurrentReadWrite 测试并发读写
func TestConcurrentReadWrite(t *testing.T) {
	t.Skip("需要数据库连接,手动测试")

	sys := NewSystem()
	defer sys.Stop()

	// 提交一些任务
	for i := 0; i < 10; i++ {
		sys.SubmitTask(1, 100.0, "test")
	}

	var wg sync.WaitGroup

	// 启动多个读取goroutine
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				// 并发读取系统信息
				info := sys.GetSystemInfo()
				if info == nil {
					t.Error("GetSystemInfo 返回 nil")
				}
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	// 启动调度循环(会修改任务状态)
	sys.IsRunning = true
	go sys.runSchedulingLoop()

	wg.Wait()
}

// TestTaskManagerConcurrency 测试TaskManager的并发安全
func TestTaskManagerConcurrency(t *testing.T) {
	tm := NewTaskManager()

	var wg sync.WaitGroup
	numGoroutines := 10

	// 并发添加任务
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task := define.NewTask(uint(id), 100.0, "test")
			tm.AddTask(task)
		}(i)
	}

	// 并发读取任务
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				tm.GetActiveTasks()
				tm.Count()
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}

	wg.Wait()

	// 验证结果
	if tm.Count() != numGoroutines {
		t.Errorf("期望 %d 个任务,实际 %d 个", numGoroutines, tm.Count())
	}
}

// TestAssignmentManagerConcurrency 测试AssignmentManager的并发安全
func TestAssignmentManagerConcurrency(t *testing.T) {
	am := NewAssignmentManager()

	var wg sync.WaitGroup
	numTasks := 5
	assignmentsPerTask := 10

	// 并发添加分配记录
	for i := 0; i < numTasks; i++ {
		for j := 0; j < assignmentsPerTask; j++ {
			wg.Add(1)
			go func(taskID int, slot int) {
				defer wg.Done()
				assign := &define.Assignment{
					TimeSlot: uint(slot),
					TaskID:   string(rune(taskID)),
					CommID:   1,
				}
				am.AddAssignment(assign)
			}(i, j)
		}
	}

	// 并发读取
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				am.GetLastAssignment(string(rune(taskID)))
				am.GetHistory(string(rune(taskID)))
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
}
