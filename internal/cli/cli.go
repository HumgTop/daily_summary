package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"humg.top/daily_summary/internal/dialog"
	"humg.top/daily_summary/internal/models"
	"humg.top/daily_summary/internal/scheduler"
	"humg.top/daily_summary/internal/storage"
)

// RunAdd 添加工作记录
func RunAdd(store storage.Storage, content string, dataDir string) error {
	now := time.Now()

	entry := models.WorkEntry{
		Timestamp: now,
		Content:   content,
	}

	if err := store.SaveEntry(entry); err != nil {
		return fmt.Errorf("failed to save entry: %w", err)
	}

	log.Printf("Work entry added: %s", content)
	fmt.Printf("✓ 已记录：%s (%s)\n", content, now.Format("15:04"))

	// 更新任务调度（重新计算下次提醒时间）
	if err := updateTaskSchedule(dataDir, now); err != nil {
		// 更新失败不影响主流程，只记录日志
		log.Printf("Failed to update task schedule: %v", err)
	} else {
		log.Println("Task schedule updated")
	}

	return nil
}

// RunList 列出今日记录
func RunList(store storage.Storage) error {
	today := time.Now()

	dailyData, err := store.GetDailyData(today)
	if err != nil {
		return fmt.Errorf("failed to get daily data: %w", err)
	}

	if len(dailyData.Entries) == 0 {
		fmt.Println("今日暂无记录")
		return nil
	}

	fmt.Printf("📝 今日工作记录 (%s)：\n\n", today.Format("2006-01-02"))
	for _, entry := range dailyData.Entries {
		fmt.Printf("  • %s - %s\n", entry.Timestamp.Format("15:04"), entry.Content)
	}
	fmt.Printf("\n共 %d 条记录\n", len(dailyData.Entries))

	return nil
}

// RunPopup 显示对话框让用户输入工作记录
func RunPopup(store storage.Storage, dlg dialog.Dialog, dataDir string) error {
	now := time.Now()

	// 获取今日所有记录
	todayData, err := store.GetDailyData(now)
	if err != nil {
		return fmt.Errorf("failed to get today's data: %w", err)
	}

	// 构建对话框消息
	message := buildDialogMessage(now, todayData)

	// 显示对话框
	content, ok, err := dlg.ShowInput("工作记录", message, "")
	if err != nil {
		return fmt.Errorf("failed to show dialog: %w", err)
	}

	if !ok || content == "" {
		fmt.Println("已取消或未输入内容")
		return nil
	}

	// 保存工作记录
	entry := models.WorkEntry{
		Timestamp: now,
		Content:   content,
	}

	if err := store.SaveEntry(entry); err != nil {
		return fmt.Errorf("failed to save entry: %w", err)
	}

	log.Printf("Work entry added via popup: %s", content)
	fmt.Printf("✓ 已记录：%s (%s)\n", content, now.Format("15:04"))

	// 更新任务调度（重新计算下次提醒时间）
	if err := updateTaskSchedule(dataDir, now); err != nil {
		// 更新失败不影响主流程，只记录日志
		log.Printf("Failed to update task schedule: %v", err)
	} else {
		log.Println("Task schedule updated")
	}

	return nil
}

// buildDialogMessage 构建弹窗消息
func buildDialogMessage(now time.Time, todayData *models.DailyData) string {
	currentTime := now.Format("15:04")

	if len(todayData.Entries) == 0 {
		return fmt.Sprintf("📝 当前时间: %s\n\n═════════════════════\n\n今日暂无记录\n\n═════════════════════\n\n请输入当前工作内容:", currentTime)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📝 当前时间: %s\n\n", currentTime))
	builder.WriteString("═════════════════════\n\n")
	builder.WriteString("今日已记录：\n\n")

	for _, entry := range todayData.Entries {
		entryTime := entry.Timestamp.Format("15:04")
		builder.WriteString(fmt.Sprintf("  ▸ %s    %s\n", entryTime, entry.Content))
	}

	builder.WriteString("\n═════════════════════\n\n")
	builder.WriteString("请输入当前工作内容:")
	return builder.String()
}

// CheckAndAcquireLock 检查并获取进程锁
// workDir: 工作目录（项目根目录），用于确定锁文件位置
func CheckAndAcquireLock(workDir string) error {
	lockFile := getLockFilePath(workDir)

	// 读取现有锁文件
	if data, err := os.ReadFile(lockFile); err == nil {
		oldPID := strings.TrimSpace(string(data))

		// 检查进程是否还在运行
		if isProcessRunning(oldPID) {
			return fmt.Errorf("服务已在运行 (PID: %s)\n\n提示：\n  - 后台服务已启动，无需手动运行 serve 命令\n  - 如需查看日志: tail -f %s/run/logs/app.log\n  - 如需重启服务: ./scripts/install.sh\n  - 如需停止服务: launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist", oldPID, workDir)
		}

		// 进程已结束，删除旧锁文件
		log.Printf("Cleaning up stale lock file (PID: %s)", oldPID)
		os.Remove(lockFile)
	}

	// 确保目录存在
	lockDir := filepath.Dir(lockFile)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// 创建新锁文件
	pid := fmt.Sprintf("%d", os.Getpid())
	if err := os.WriteFile(lockFile, []byte(pid), 0644); err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	log.Printf("Process lock acquired (PID: %s, lock file: %s)", pid, lockFile)
	return nil
}

// ReleaseLock 释放进程锁
func ReleaseLock(workDir string) {
	lockFile := getLockFilePath(workDir)
	os.Remove(lockFile)
	log.Printf("Process lock released: %s", lockFile)
}

// getLockFilePath 获取锁文件路径（基于 workDir 的 run/ 子目录）
// workDir: 工作目录（项目根目录），如果为空则使用当前目录
func getLockFilePath(workDir string) string {
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			// fallback 到临时目录
			log.Printf("Warning: failed to get working directory: %v, using temp dir", err)
			return filepath.Join(os.TempDir(), "daily_summary.lock")
		}
	}

	lockDir := filepath.Join(workDir, "run")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		log.Printf("Warning: failed to create lock directory: %v, using temp dir", err)
		return filepath.Join(os.TempDir(), "daily_summary.lock")
	}

	return filepath.Join(lockDir, "daily_summary.lock")
}

// isProcessRunning 检查进程是否在运行
func isProcessRunning(pidStr string) bool {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false
	}

	// 发送信号 0 检查进程是否存在（不实际发送信号）
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// updateTaskSchedule 更新任务调度时间（直接修改 tasks.json）
// dataDir: 数据目录的绝对路径
// addTime: 记录添加的时间
func updateTaskSchedule(dataDir string, addTime time.Time) error {
	// 使用 dataDir 的父目录（项目 run 目录）
	runDir := filepath.Dir(dataDir)

	// 加载任务注册表
	registry := scheduler.NewRegistry(runDir)
	if err := registry.Load(); err != nil {
		// 如果 tasks.json 不存在，说明调度器还未初始化，无需更新
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to load task registry: %w", err)
	}

	// 获取 work-reminder 任务配置
	config := registry.GetTask("work-reminder")
	if config == nil {
		// 任务不存在，可能还未初始化
		return nil
	}

	// 计算新的下次执行时间（从当前时间开始）
	intervalMinutes := config.IntervalMinutes
	if intervalMinutes <= 0 {
		intervalMinutes = 60 // 默认 1 小时
	}

	interval := time.Duration(intervalMinutes) * time.Minute
	newNextRun := addTime.Truncate(time.Minute).Add(interval)

	// 确保在未来
	for !newNextRun.After(addTime) {
		newNextRun = newNextRun.Add(interval)
	}

	oldNextRun := config.NextRun
	config.NextRun = newNextRun

	// 更新任务配置
	if err := registry.UpdateTask(config); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}



	log.Printf("Updated work-reminder schedule: %s -> %s",
		oldNextRun.Format("15:04:05"),
		newNextRun.Format("15:04:05"))

	return nil
}
