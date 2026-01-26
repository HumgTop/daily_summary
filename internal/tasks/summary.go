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
	storage           storage.Storage
	generator         *summary.Generator
	hour              int       // 执行时间（小时）
	minute            int       // 执行时间（分钟）
	ungeneratedDates  []time.Time // 待生成日报的日期列表（临时字段，由 ShouldRun 设置，Execute 使用）
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
func (t *SummaryTask) ShouldRun(now time.Time, config *scheduler.TaskConfig) (bool, func(*scheduler.TaskConfig)) {
	if !config.Enabled {
		return false, nil
	}

	// 构造今天的总结时间点
	todaySummaryTime := time.Date(now.Year(), now.Month(), now.Day(),
		t.hour, t.minute, 0, 0, now.Location())

	// 检查是否已过总结时间
	if !now.After(todaySummaryTime) {
		return false, nil
	}

	// 获取今天之前所有未生成日报的日期（不包括今天）
	// 传入今天的开始时间（00:00:00），确保只检查今天之前的日期
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	ungeneratedDates, err := t.storage.GetUngeneratedDates(today)
	if err != nil {
		log.Printf("SummaryTask: failed to get ungenerated dates: %v", err)
		return false, nil
	}

	// 如果没有未生成的日报,顺延到明天的总结时间
	if len(ungeneratedDates) == 0 {
		log.Printf("SummaryTask: no ungenerated summaries, delaying to next day")

		// 计算下一次执行时间（明天的总结时间）
		tomorrowSummaryTime := time.Date(now.Year(), now.Month(), now.Day()+1,
			t.hour, t.minute, 0, 0, now.Location())

		// 返回更新函数，顺延 NextRun
		return false, func(cfg *scheduler.TaskConfig) {
			cfg.NextRun = tomorrowSummaryTime
		}
	}

	// 存储未生成的日期列表到临时字段，供 Execute 使用
	t.ungeneratedDates = ungeneratedDates

	log.Printf("SummaryTask: found %d ungenerated summaries, will generate", len(ungeneratedDates))
	return true, nil
}

// Execute 执行任务
func (t *SummaryTask) Execute() error {
	// 从临时字段读取未生成的日期列表
	if len(t.ungeneratedDates) == 0 {
		log.Printf("SummaryTask.Execute: no dates to generate (this should not happen)")
		return nil
	}

	totalDates := len(t.ungeneratedDates)
	log.Printf("SummaryTask: generating summaries for %d dates", totalDates)

	// 批量生成所有未生成的日报
	var generatedCount int
	var lastError error

	for _, date := range t.ungeneratedDates {
		dateStr := date.Format("2006-01-02")
		log.Printf("Generating summary for %s", dateStr)

		if err := t.generator.GenerateDailySummary(date); err != nil {
			log.Printf("Failed to generate summary for %s: %v", dateStr, err)
			lastError = err
			continue // 继续生成其他日期的日报
		}

		// 标记总结已生成
		if err := t.storage.MarkSummaryGenerated(date); err != nil {
			log.Printf("Failed to mark summary as generated for %s: %v", dateStr, err)
			// 不返回错误，因为总结已经成功生成
		}

		generatedCount++
		log.Printf("Summary generated successfully for %s (%d/%d)",
			dateStr, generatedCount, totalDates)
	}

	// 清空临时字段
	t.ungeneratedDates = nil

	if generatedCount == 0 {
		return fmt.Errorf("failed to generate any summaries: %w", lastError)
	}

	if lastError != nil {
		log.Printf("Warning: some summaries failed to generate (succeeded: %d, failed: %d)",
			generatedCount, totalDates-generatedCount)
	}

	log.Printf("SummaryTask: batch generation completed (generated: %d)", generatedCount)
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
	}

	// 计算下次执行时间（明天的总结时间）
	tomorrowSummaryTime := time.Date(now.Year(), now.Month(), now.Day()+1,
		t.hour, t.minute, 0, 0, now.Location())
	config.NextRun = tomorrowSummaryTime

	log.Printf("SummaryTask: next run scheduled at %s", tomorrowSummaryTime.Format("2006-01-02 15:04:05"))
}

// parseSummaryTime 解析总结时间
func parseSummaryTime(timeStr string) (hour, min int) {
	// 默认 00:00
	hour, min = 0, 0
	fmt.Sscanf(timeStr, "%d:%d", &hour, &min)
	return
}
