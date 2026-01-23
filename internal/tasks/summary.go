package tasks

import (
	"fmt"
	"log"
	"time"

	"humg.top/daily_summary/internal/scheduler"
	"humg.top/daily_summary/internal/storage"
	"humg.top/daily_summary/internal/summary"
)

// SummaryTask 每日总结生成任务
type SummaryTask struct {
	storage   storage.Storage
	generator *summary.Generator
	hour      int // 执行时间（小时）
	minute    int // 执行时间（分钟）
}

// NewSummaryTask 创建每日总结任务
func NewSummaryTask(storage storage.Storage, generator *summary.Generator, summaryTime string) *SummaryTask {
	hour, minute := parseSummaryTime(summaryTime)
	return &SummaryTask{
		storage:   storage,
		generator: generator,
		hour:      hour,
		minute:    minute,
	}
}

// ID 返回任务 ID
func (t *SummaryTask) ID() string {
	return "daily-summary"
}

// Name 返回任务名称
func (t *SummaryTask) Name() string {
	return "每日总结生成"
}

// ShouldRun 判断是否应该执行
func (t *SummaryTask) ShouldRun(now time.Time, config *scheduler.TaskConfig) bool {
	if !config.Enabled {
		return false
	}

	// 今天的日期
	today := now.Format("2006-01-02")

	// 检查今天是否已经生成过
	if lastDate, ok := config.Data["last_generated_date"].(string); ok {
		if lastDate == today {
			return false
		}
	}

	// 构造今天的总结时间点
	todaySummaryTime := time.Date(now.Year(), now.Month(), now.Day(),
		t.hour, t.minute, 0, 0, now.Location())

	// 检查是否已过总结时间
	if !now.After(todaySummaryTime) {
		return false
	}

	// 检查昨天是否有记录且未生成总结
	yesterday := now.AddDate(0, 0, -1)
	yesterdayData, err := t.storage.GetDailyData(yesterday)
	if err != nil {
		// 没有昨天的数据，跳过
		return false
	}

	// 如果昨天有记录且未生成总结，返回 true
	return len(yesterdayData.Entries) > 0 && !yesterdayData.SummaryGenerated
}

// Execute 执行任务
func (t *SummaryTask) Execute() error {
	// 生成前一天的总结
	yesterday := time.Now().AddDate(0, 0, -1)

	log.Printf("Generating summary for %s", yesterday.Format("2006-01-02"))

	if err := t.generator.GenerateDailySummary(yesterday); err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	// 标记总结已生成
	if err := t.storage.MarkSummaryGenerated(yesterday); err != nil {
		log.Printf("Failed to mark summary as generated: %v", err)
		// 不返回错误，因为总结已经成功生成
	}

	log.Printf("Summary generated successfully for %s", yesterday.Format("2006-01-02"))
	return nil
}

// OnExecuted 任务执行后的回调
func (t *SummaryTask) OnExecuted(now time.Time, config *scheduler.TaskConfig, err error) {
	// 更新最后执行时间
	config.LastRun = now

	if err != nil {
		config.LastError = err.Error()
		log.Printf("Task %s failed: %v", t.Name(), err)
	} else {
		config.LastSuccess = now
		config.LastError = ""

		// 更新最后生成日期
		today := now.Format("2006-01-02")
		if config.Data == nil {
			config.Data = make(map[string]interface{})
		}
		config.Data["last_generated_date"] = today
	}

	// 计算下次执行时间（复用与初始化时相同的逻辑）
	summaryTime := fmt.Sprintf("%d:%d", t.hour, t.minute)
	config.NextRun = scheduler.CalculateNextSummaryTime(now, summaryTime)
}

// parseSummaryTime 解析总结时间
func parseSummaryTime(timeStr string) (hour, min int) {
	// 默认 00:00
	hour, min = 0, 0
	fmt.Sscanf(timeStr, "%d:%d", &hour, &min)
	return
}
