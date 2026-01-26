package tasks

import (
	"fmt"
	"log"
	"time"

	"humg.top/daily_summary/internal/scheduler"
	"humg.top/daily_summary/internal/storage"
	"humg.top/daily_summary/internal/summary"
)

// WeeklySummaryTask 周度总结生成任务
type WeeklySummaryTask struct {
	storage   storage.Storage
	generator *summary.Generator
	weekday   time.Weekday // 周几执行（1=周一, 7=周日）
	hour      int          // 执行时间（小时）
	minute    int          // 执行时间（分钟）
}

// NewWeeklySummaryTask 创建周度总结任务
// weekday: 1=周一, 2=周二, ..., 7=周日
func NewWeeklySummaryTask(
	storage storage.Storage,
	generator *summary.Generator,
	weekday int,
	summaryTime string,
) *WeeklySummaryTask {
	hour, minute := parseSummaryTime(summaryTime)

	// 将 1-7 转换为 Go 的 Weekday（0=周日, 1=周一, ..., 6=周六）
	var wd time.Weekday
	if weekday == 7 {
		wd = time.Sunday
	} else {
		wd = time.Weekday(weekday)
	}

	return &WeeklySummaryTask{
		storage:   storage,
		generator: generator,
		weekday:   wd,
		hour:      hour,
		minute:    minute,
	}
}

// ID 返回任务 ID
func (t *WeeklySummaryTask) ID() string {
	return "weekly-summary"
}

// Name 返回任务名称
func (t *WeeklySummaryTask) Name() string {
	return "周度总结生成"
}

// ShouldRun 判断是否应该执行
func (t *WeeklySummaryTask) ShouldRun(now time.Time, config *scheduler.TaskConfig) (bool, func(*scheduler.TaskConfig)) {
	if !config.Enabled {
		return false, nil
	}

	// 检查今天是否是指定的星期几
	if now.Weekday() != t.weekday {
		return false, nil
	}

	// 获取本周的周标识（YYYY-WW）
	year, week := now.ISOWeek()
	currentWeekKey := fmt.Sprintf("%d-W%02d", year, week)

	// 检查本周是否已经生成过
	if lastWeek, ok := config.Data["last_generated_week"].(string); ok {
		if lastWeek == currentWeekKey {
			return false, nil
		}
	}

	// 构造今天的总结时间点
	todaySummaryTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		t.hour, t.minute, 0, 0, now.Location(),
	)

	// 检查是否已过总结时间
	if !now.After(todaySummaryTime) {
		return false, nil
	}

	// 满足所有条件，应该执行
	return true, nil
}

// Execute 执行任务
func (t *WeeklySummaryTask) Execute() error {
	now := time.Now()

	// 计算上周的周日日期（周末）
	// 今天是周一，上周日 = 今天 - 1 天
	// 今天是周二，上周日 = 今天 - 2 天
	// ...
	daysFromLastSunday := int(now.Weekday())
	if daysFromLastSunday == 0 {
		// 今天是周日，上周日 = 今天 - 7 天
		daysFromLastSunday = 7
	}
	lastSunday := now.AddDate(0, 0, -daysFromLastSunday)

	log.Printf("Generating weekly summary for week ending %s", lastSunday.Format("2006-01-02"))

	if err := t.generator.GenerateWeeklySummary(lastSunday); err != nil {
		return fmt.Errorf("failed to generate weekly summary: %w", err)
	}

	log.Printf("Weekly summary generated successfully")
	return nil
}

// OnExecuted 任务执行后的回调
func (t *WeeklySummaryTask) OnExecuted(now time.Time, config *scheduler.TaskConfig, err error) {
	config.LastRun = now

	if err != nil {
		config.LastError = err.Error()
		log.Printf("Task %s failed: %v", t.Name(), err)
	} else {
		config.LastSuccess = now
		config.LastError = ""

		// 标记本周已生成
		year, week := now.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)
		if config.Data == nil {
			config.Data = make(map[string]interface{})
		}
		config.Data["last_generated_week"] = weekKey
	}

	// 计算下次执行时间（下周的同一天同一时间）
	config.NextRun = calculateNextWeeklyTime(now, t.weekday, t.hour, t.minute)
}

// calculateNextWeeklyTime 计算下次周度总结时间
func calculateNextWeeklyTime(now time.Time, targetWeekday time.Weekday, hour, minute int) time.Time {
	// 计算距离目标星期几还有几天
	daysUntil := int(targetWeekday - now.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7 // 已经过了本周的目标日，跳到下周
	}

	nextDate := now.AddDate(0, 0, daysUntil)
	return time.Date(
		nextDate.Year(), nextDate.Month(), nextDate.Day(),
		hour, minute, 0, 0, nextDate.Location(),
	)
}
