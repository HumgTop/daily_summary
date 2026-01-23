package tasks

import (
	"fmt"
	"log"
	"strings"
	"time"

	"humg.top/daily_summary/internal/dialog"
	"humg.top/daily_summary/internal/models"
	"humg.top/daily_summary/internal/scheduler"
	"humg.top/daily_summary/internal/storage"
)

// ReminderTask 工作记录提醒任务
type ReminderTask struct {
	dialog  dialog.Dialog
	storage storage.Storage
}

// NewReminderTask 创建工作记录提醒任务
func NewReminderTask(dialog dialog.Dialog, storage storage.Storage) *ReminderTask {
	return &ReminderTask{
		dialog:  dialog,
		storage: storage,
	}
}

// ID 返回任务 ID
func (t *ReminderTask) ID() string {
	return "work-reminder"
}

// Name 返回任务名称
func (t *ReminderTask) Name() string {
	return "工作记录提醒"
}

// ShouldRun 判断是否应该执行
func (t *ReminderTask) ShouldRun(now time.Time, config *scheduler.TaskConfig) (bool, *scheduler.TaskConfig) {
	if !config.Enabled {
		return false, nil
	}

	// 检查下次执行时间
	if config.NextRun.IsZero() {
		return false, nil
	}

	// 如果还未到执行时间，跳过
	if now.Before(config.NextRun) {
		return false, nil
	}

	// 延迟检测：如果距离预定执行时间过长，说明任务失效（如电脑休眠）
	// 计算延迟时间
	delay := now.Sub(config.NextRun)
	maxDelay := time.Duration(config.IntervalMinutes/2) * time.Minute

	if delay > maxDelay {
		// 延迟过长，重新计算下次执行时间，跳过本次执行
		log.Printf("Task %s delayed too long (%v > %v), rescheduling...",
			config.ID, delay, maxDelay)

		// 创建新配置（不修改原配置）
		newConfig := *config
		newConfig.NextRun = t.calculateNextRun(now, config.IntervalMinutes)

		return false, &newConfig
	}

	return true, nil
}

// Execute 执行任务
func (t *ReminderTask) Execute() error {
	now := time.Now()
	title := "工作记录"

	// 获取今日所有记录
	todayData, err := t.storage.GetDailyData(now)
	var message string
	if err != nil {
		log.Printf("Failed to get today's data: %v", err)
		message = fmt.Sprintf("请输入工作内容 (当前时间: %s):", now.Format("15:04"))
	} else {
		message = t.buildDialogMessage(now, todayData)
	}

	// 显示对话框
	content, ok, err := t.dialog.ShowInput(title, message, "")
	if err != nil {
		return fmt.Errorf("failed to show dialog: %w", err)
	}

	if !ok || content == "" {
		log.Println("User cancelled or input is empty, skipping this entry")
		return nil
	}

	// 保存工作记录
	entry := models.WorkEntry{
		Timestamp: now,
		Content:   content,
	}

	if err := t.storage.SaveEntry(entry); err != nil {
		return fmt.Errorf("failed to save entry: %w", err)
	}

	log.Printf("Work entry saved: %s", content)
	return nil
}

// OnExecuted 任务执行后的回调
func (t *ReminderTask) OnExecuted(now time.Time, config *scheduler.TaskConfig, err error) {
	// 更新最后执行时间
	config.LastRun = now

	if err != nil {
		config.LastError = err.Error()
		log.Printf("Task %s failed: %v", t.Name(), err)
	} else {
		config.LastSuccess = now
		config.LastError = ""
	}

	// 计算下次执行时间
	config.NextRun = t.calculateNextRun(now, config.IntervalMinutes)
	log.Printf("Next %s at: %s", t.Name(), config.NextRun.Format("15:04:05"))
}

// calculateNextRun 计算下次执行时间
func (t *ReminderTask) calculateNextRun(from time.Time, intervalMinutes int) time.Time {
	interval := time.Duration(intervalMinutes) * time.Minute

	// 对齐到分钟边界
	next := from.Truncate(time.Minute).Add(interval)

	// 确保在未来
	for !next.After(from) {
		next = next.Add(interval)
	}

	return next
}

// buildDialogMessage 构建对话框消息
func (t *ReminderTask) buildDialogMessage(now time.Time, todayData *models.DailyData) string {
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
