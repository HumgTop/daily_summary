package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"humg.top/daily_summary/internal/scheduler"
)

// TestUpdateTaskSchedule 测试更新任务调度
func TestUpdateTaskSchedule(t *testing.T) {
	// 使用临时目录进行测试
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	os.MkdirAll(dataDir, 0755)

	// 创建任务注册表并添加测试任务
	registry := scheduler.NewRegistry(tmpDir)
	now := time.Now()

	taskConfig := &scheduler.TaskConfig{
		ID:              "work-reminder",
		Name:            "工作记录提醒",
		Type:            scheduler.TaskTypeInterval,
		Enabled:         true,
		IntervalMinutes: 45,
		NextRun:         now.Add(30 * time.Minute),
	}

	if err := registry.AddTask(taskConfig); err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	if err := registry.Save(); err != nil {
		t.Fatalf("Failed to save registry: %v", err)
	}

	// 更新任务调度
	addTime := now.Add(5 * time.Minute)
	if err := updateTaskSchedule(dataDir, addTime); err != nil {
		t.Fatalf("Failed to update task schedule: %v", err)
	}

	// 重新加载并验证
	registry2 := scheduler.NewRegistry(tmpDir)
	if err := registry2.Load(); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	updatedTask := registry2.GetTask("work-reminder")
	if updatedTask == nil {
		t.Fatal("Task not found after update")
	}

	// 验证 next_run 已更新（应该比 addTime 晚 45 分钟）
	expectedNextRun := addTime.Truncate(time.Minute).Add(45 * time.Minute)
	if !updatedTask.NextRun.Equal(expectedNextRun) {
		t.Errorf("Expected next_run %s, got %s",
			expectedNextRun.Format("15:04:05"),
			updatedTask.NextRun.Format("15:04:05"))
	}
}

// TestUpdateTaskScheduleWithoutRegistry 测试当 tasks.json 不存在时
func TestUpdateTaskScheduleWithoutRegistry(t *testing.T) {
	// 使用临时目录进行测试
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	os.MkdirAll(dataDir, 0755)

	// tasks.json 不存在，updateTaskSchedule 应该正常返回（不报错）
	now := time.Now()
	if err := updateTaskSchedule(dataDir, now); err != nil {
		t.Errorf("Should not error when tasks.json doesn't exist: %v", err)
	}
}
