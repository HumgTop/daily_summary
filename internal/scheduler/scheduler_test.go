package scheduler

import (
	"testing"
	"time"
)

// TestRegistryLoadAndSave 测试任务注册表的加载和保存
func TestRegistryLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)

	// 添加测试任务
	task := &TaskConfig{
		ID:              "test-task",
		Name:            "Test Task",
		Type:            TaskTypeInterval,
		Enabled:         true,
		IntervalMinutes: 30,
		NextRun:         time.Now().Add(30 * time.Minute),
	}

	if err := registry.AddTask(task); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// 保存
	if err := registry.Save(); err != nil {
		t.Fatalf("Failed to save registry: %v", err)
	}

	// 创建新的注册表并加载
	registry2 := NewRegistry(tmpDir)
	if err := registry2.Load(); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// 验证加载的任务
	loadedTask := registry2.GetTask("test-task")
	if loadedTask == nil {
		t.Fatal("Task not found after load")
	}

	if loadedTask.Name != task.Name {
		t.Errorf("Expected name %s, got %s", task.Name, loadedTask.Name)
	}

	if loadedTask.IntervalMinutes != task.IntervalMinutes {
		t.Errorf("Expected interval %d, got %d", task.IntervalMinutes, loadedTask.IntervalMinutes)
	}
}

// TestRegistryOperations 测试注册表的增删改查操作
func TestRegistryOperations(t *testing.T) {
	tmpDir := t.TempDir()
	registry := NewRegistry(tmpDir)

	// 测试添加任务
	task1 := &TaskConfig{
		ID:      "task-1",
		Name:    "Task 1",
		Type:    TaskTypeInterval,
		Enabled: true,
	}

	if err := registry.AddTask(task1); err != nil {
		t.Fatalf("Failed to add task1: %v", err)
	}

	// 测试重复添加应该失败
	if err := registry.AddTask(task1); err == nil {
		t.Error("Expected error when adding duplicate task")
	}

	// 测试获取任务
	retrieved := registry.GetTask("task-1")
	if retrieved == nil {
		t.Fatal("Failed to get task")
	}

	// 测试更新任务
	retrieved.Enabled = false
	if err := registry.UpdateTask(retrieved); err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	updated := registry.GetTask("task-1")
	if updated.Enabled {
		t.Error("Task should be disabled after update")
	}

	// 测试获取所有任务
	allTasks := registry.GetAllTasks()
	if len(allTasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(allTasks))
	}

	// 测试删除任务
	if err := registry.RemoveTask("task-1"); err != nil {
		t.Fatalf("Failed to remove task: %v", err)
	}

	if registry.GetTask("task-1") != nil {
		t.Error("Task should be removed")
	}

	// 测试删除不存在的任务应该失败
	if err := registry.RemoveTask("non-existent"); err == nil {
		t.Error("Expected error when removing non-existent task")
	}
}

// TestSchedulerInitialization 测试调度器初始化
func TestSchedulerInitialization(t *testing.T) {
	tmpDir := t.TempDir()
	sched := NewScheduler(tmpDir)

	if sched.registry == nil {
		t.Error("Registry should be initialized")
	}

	if sched.tasks == nil {
		t.Error("Tasks map should be initialized")
	}

	if sched.checkInterval != 1*time.Minute {
		t.Errorf("Expected check interval 1 minute, got %v", sched.checkInterval)
	}
}

// TestCalculateNextReminderTime 测试计算下次提醒时间
func TestCalculateNextReminderTime(t *testing.T) {
	tests := []struct {
		name            string
		from            time.Time
		intervalMinutes int
		expectedMinute  int
	}{
		{
			name:            "30 minutes from 14:45",
			from:            time.Date(2026, 1, 23, 14, 45, 0, 0, time.Local),
			intervalMinutes: 30,
			expectedMinute:  15, // 14:45 -> 15:15
		},
		{
			name:            "45 minutes from 14:30",
			from:            time.Date(2026, 1, 23, 14, 30, 0, 0, time.Local),
			intervalMinutes: 45,
			expectedMinute:  15, // 14:30 -> 15:15
		},
		{
			name:            "60 minutes from 14:30",
			from:            time.Date(2026, 1, 23, 14, 30, 0, 0, time.Local),
			intervalMinutes: 60,
			expectedMinute:  30, // 14:30 -> 15:30
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := calculateNextReminderTime(tt.from, tt.intervalMinutes)

			// 验证时间在未来
			if !next.After(tt.from) {
				t.Errorf("Next time %s should be after from time %s",
					next.Format("15:04"), tt.from.Format("15:04"))
			}

			// 验证分钟对齐
			if next.Minute() != tt.expectedMinute {
				t.Errorf("Expected minute %d, got %d", tt.expectedMinute, next.Minute())
			}
		})
	}
}

// TestCalculateNextSummaryTime 测试计算下次总结时间
func TestCalculateNextSummaryTime(t *testing.T) {
	tests := []struct {
		name        string
		from        time.Time
		summaryTime string
		expectedDay int    // 期望的日期（相对于 from）
		expectedH   int    // 期望的小时
		expectedM   int    // 期望的分钟
	}{
		{
			name:        "before summary time today",
			from:        time.Date(2026, 1, 23, 10, 30, 0, 0, time.Local),
			summaryTime: "11:00",
			expectedDay: 0, // 今天
			expectedH:   11,
			expectedM:   0,
		},
		{
			name:        "after summary time today",
			from:        time.Date(2026, 1, 23, 11, 30, 0, 0, time.Local),
			summaryTime: "11:00",
			expectedDay: 1, // 明天
			expectedH:   11,
			expectedM:   0,
		},
		{
			name:        "exactly at summary time",
			from:        time.Date(2026, 1, 23, 11, 0, 0, 0, time.Local),
			summaryTime: "11:00",
			expectedDay: 1, // 明天（因为 from.After(todaySummaryTime) || from.Equal(todaySummaryTime)）
			expectedH:   11,
			expectedM:   0,
		},
		{
			name:        "late night before midnight summary",
			from:        time.Date(2026, 1, 23, 23, 30, 0, 0, time.Local),
			summaryTime: "00:00",
			expectedDay: 1, // 明天
			expectedH:   0,
			expectedM:   0,
		},
		{
			name:        "early morning after midnight summary",
			from:        time.Date(2026, 1, 23, 0, 30, 0, 0, time.Local),
			summaryTime: "00:00",
			expectedDay: 1, // 明天
			expectedH:   0,
			expectedM:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := calculateNextSummaryTime(tt.from, tt.summaryTime)

			expectedTime := tt.from.AddDate(0, 0, tt.expectedDay)
			expectedTime = time.Date(expectedTime.Year(), expectedTime.Month(), expectedTime.Day(),
				tt.expectedH, tt.expectedM, 0, 0, time.Local)

			if !next.Equal(expectedTime) {
				t.Errorf("Expected %s, got %s",
					expectedTime.Format("2006-01-02 15:04:05"),
					next.Format("2006-01-02 15:04:05"))
			}

			// 验证时间在未来或等于当前（等于的情况会在下一个周期跳过）
			if next.Before(tt.from) {
				t.Errorf("Next time %s should not be before from time %s",
					next.Format("2006-01-02 15:04:05"),
					tt.from.Format("2006-01-02 15:04:05"))
			}
		})
	}
}

// TestTwoStageScheduling 测试两段式调度判断
func TestTwoStageScheduling(t *testing.T) {
	tmpDir := t.TempDir()
	sched := NewScheduler(tmpDir)

	// 创建一个测试任务配置，next_run 设置为未来
	futureTime := time.Now().Add(10 * time.Minute)
	config := &TaskConfig{
		ID:      "test-task",
		Name:    "Test Task",
		Type:    TaskTypeInterval,
		Enabled: true,
		NextRun: futureTime,
	}

	if err := sched.registry.AddTask(config); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// 注册一个模拟任务（ShouldRun 总是返回 true）
	mockTask := &mockAlwaysRunTask{}
	sched.RegisterTask(mockTask)

	// 执行检查
	sched.checkAndRunTasks()

	// 验证任务没有被执行（因为 next_run 在未来，第一段判断就跳过了）
	if mockTask.executed {
		t.Error("Task should not be executed when next_run is in the future")
	}
}

// mockAlwaysRunTask 模拟任务，ShouldRun 总是返回 true
type mockAlwaysRunTask struct {
	executed bool
}

func (m *mockAlwaysRunTask) ID() string                                              { return "test-task" }
func (m *mockAlwaysRunTask) Name() string                                            { return "Mock Task" }
func (m *mockAlwaysRunTask) ShouldRun(now time.Time, config *TaskConfig) bool        { return true }
func (m *mockAlwaysRunTask) Execute() error                                          { m.executed = true; return nil }
func (m *mockAlwaysRunTask) OnExecuted(now time.Time, config *TaskConfig, err error) {}
