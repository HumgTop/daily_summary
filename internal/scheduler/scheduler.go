package scheduler

import (
	"fmt"
	"log"
	"strings"
	"time"

	"humg.top/daily_summary/internal/dialog"
	"humg.top/daily_summary/internal/models"
	"humg.top/daily_summary/internal/storage"
	"humg.top/daily_summary/internal/summary"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	config    *models.Config
	dialog    dialog.Dialog
	storage   storage.Storage
	generator *summary.Generator
	stopCh    chan struct{}
}

// NewScheduler 创建调度器
func NewScheduler(
	config *models.Config,
	dialog dialog.Dialog,
	storage storage.Storage,
	generator *summary.Generator,
) *Scheduler {
	return &Scheduler{
		config:    config,
		dialog:    dialog,
		storage:   storage,
		generator: generator,
		stopCh:    make(chan struct{}),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	log.Println("Scheduler started")

	// 启动小时任务
	go s.runHourlyTask()

	// 启动每日总结任务
	go s.runDailySummaryTask()

	// 等待停止信号
	<-s.stopCh
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

// runHourlyTask 定期弹窗任务（支持小时或分钟级）
func (s *Scheduler) runHourlyTask() {
	var interval time.Duration
	now := time.Now()

	// 检查是否使用分钟级调度
	if s.config.MinuteInterval > 0 {
		// 分钟级调度
		interval = time.Duration(s.config.MinuteInterval) * time.Minute
		log.Printf("Using minute-based scheduling: every %d minute(s)", s.config.MinuteInterval)
	} else {
		// 小时级调度（默认）
		interval = time.Duration(s.config.HourlyInterval) * time.Hour
		log.Printf("Using hour-based scheduling: every %d hour(s)", s.config.HourlyInterval)
	}

	for {
		// 每次循环都重新计算下一个触发时间
		now = time.Now()
		var nextTrigger time.Time

		if s.config.MinuteInterval > 0 {
			// 分钟级：对齐到分钟边界
			nextTrigger = now.Truncate(time.Minute).Add(interval)
			if nextTrigger.Before(now) || nextTrigger.Equal(now) {
				nextTrigger = nextTrigger.Add(interval)
			}
		} else {
			// 小时级：对齐到整点
			nextTrigger = now.Truncate(time.Hour).Add(interval)
			// 确保下一个触发时间在未来
			for !nextTrigger.After(now) {
				nextTrigger = nextTrigger.Add(interval)
			}
		}

		log.Printf("Next reminder scheduled at %s", nextTrigger.Format("15:04:05"))

		// 计算跳过阈值
		var skipThreshold time.Duration
		if s.config.MinuteInterval > 0 {
			// 分钟级：阈值为间隔的 50%
			skipThreshold = interval / 2
		} else {
			// 小时级：固定 5 分钟
			skipThreshold = 5 * time.Minute
		}

		// 等待到触发时间
		select {
		case <-time.After(time.Until(nextTrigger)):
			// 再次检查当前时间，确保没有严重延迟
			actualTime := time.Now()
			expectedTime := nextTrigger

			// 如果延迟超过阈值（比如从睡眠中唤醒），跳过本次调度
			delay := actualTime.Sub(expectedTime)
			if delay > skipThreshold {
				log.Printf("Skipped reminder due to delay (expected: %s, actual: %s, delay: %s, threshold: %s)",
					expectedTime.Format("15:04:05"),
					actualTime.Format("15:04:05"),
					delay,
					skipThreshold)
				continue
			}

			// 正常执行
			s.showWorkEntryDialog()
		case <-s.stopCh:
			return
		}
	}
}

// showWorkEntryDialog 显示工作记录对话框
func (s *Scheduler) showWorkEntryDialog() {
	now := time.Now()
	title := "工作记录"

	// 获取今日所有记录
	todayData, err := s.storage.GetDailyData(now)
	var message string
	if err != nil {
		log.Printf("Failed to get today's data: %v", err)
		message = fmt.Sprintf("请输入工作内容 (当前时间: %s):", now.Format("15:04"))
	} else {
		message = s.buildDialogMessage(now, todayData)
	}

	content, ok, err := s.dialog.ShowInput(title, message, "")
	if err != nil {
		log.Printf("Failed to show dialog: %v", err)
		return
	}

	if !ok || content == "" {
		log.Println("User cancelled or input is empty, skipping this entry")
		return
	}

	// 保存工作记录
	entry := models.WorkEntry{
		Timestamp: now,
		Content:   content,
	}

	if err := s.storage.SaveEntry(entry); err != nil {
		log.Printf("Failed to save entry: %v", err)
		return
	}

	log.Printf("Work entry saved: %s", content)
}

// buildDialogMessage 构建弹窗消息
func (s *Scheduler) buildDialogMessage(now time.Time, todayData *models.DailyData) string {
	currentTime := now.Format("15:04")

	if len(todayData.Entries) == 0 {
		return fmt.Sprintf("📝 工作记录 (当前时间: %s)\n\n今日暂无记录\n\n请输入当前工作内容:", currentTime)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("📝 工作记录 (当前时间: %s)\n\n", currentTime))
	builder.WriteString("今日已记录：\n")

	for _, entry := range todayData.Entries {
		entryTime := entry.Timestamp.Format("15:04")
		builder.WriteString(fmt.Sprintf("  %s - %s\n", entryTime, entry.Content))
	}

	builder.WriteString("\n请输入当前工作内容:")
	return builder.String()
}

// runDailySummaryTask 每日总结任务（支持 at least once 语义）
func (s *Scheduler) runDailySummaryTask() {
	// 解析总结时间（格式: "HH:MM"）
	summaryHour, summaryMin := parseSummaryTime(s.config.SummaryTime)

	for {
		now := time.Now()
		yesterday := now.AddDate(0, 0, -1)
		
		// 检查昨天是否已生成总结
		yesterdayData, err := s.storage.GetDailyData(yesterday)
		if err == nil && !yesterdayData.SummaryGenerated {
			// 昨天未生成总结，检查是否已过配置时间
			todaySummaryTime := time.Date(now.Year(), now.Month(), now.Day(), 
				summaryHour, summaryMin, 0, 0, now.Location())
			
			if now.After(todaySummaryTime) {
				// 已过配置时间，立即生成昨天的总结
				log.Printf("Missed scheduled time %s, generating summary immediately", 
					todaySummaryTime.Format("15:04"))
				s.generateSummary()
				// 生成后继续计算下一次时间
			}
		}

		// 计算下一个总结时间
		nextSummary := time.Date(now.Year(), now.Month(), now.Day(), summaryHour, summaryMin, 0, 0, now.Location())

		// 如果今天的时间已经过了，则等到明天
		if now.After(nextSummary) {
			nextSummary = nextSummary.Add(24 * time.Hour)
		}

		// 等待到总结时间
		waitDuration := time.Until(nextSummary)
		log.Printf("Next summary generation at: %s (in %s)", nextSummary.Format("2006-01-02 15:04:05"), waitDuration)

		select {
		case <-time.After(waitDuration):
			s.generateSummary()
		case <-s.stopCh:
			return
		}
	}
}

// generateSummary 生成前一天的工作总结
func (s *Scheduler) generateSummary() {
	// 生成前一天的总结
	yesterday := time.Now().AddDate(0, 0, -1)

	log.Printf("Generating summary for %s", yesterday.Format("2006-01-02"))

	if err := s.generator.GenerateDailySummary(yesterday); err != nil {
		log.Printf("Failed to generate summary: %v", err)
		return
	}

	// 标记总结已生成
	if err := s.storage.MarkSummaryGenerated(yesterday); err != nil {
		log.Printf("Failed to mark summary as generated: %v", err)
		// 不返回错误，因为总结已经成功生成
	}

	log.Printf("Summary generated successfully for %s", yesterday.Format("2006-01-02"))
}

// parseSummaryTime 解析总结时间
func parseSummaryTime(timeStr string) (hour, min int) {
	// 默认 00:00
	hour, min = 0, 0
	fmt.Sscanf(timeStr, "%d:%d", &hour, &min)
	return
}
