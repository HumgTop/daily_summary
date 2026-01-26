package scheduler

import (
	"fmt"
	"log"
	"time"
)

// InitializeTasksFromConfig 从配置初始化任务注册表
// 如果 tasks.json 不存在，则从传统配置创建默认任务
func (s *Scheduler) InitializeTasksFromConfig(
	hourlyInterval int,
	minuteInterval int,
	summaryTime string,
	enableWeeklySummary bool,
	weeklySummaryTime string,
	weeklySummaryDay int,
) error {
	// 每次启动时都根据配置重新初始化任务，确保配置与代码保持一致
	log.Println("Initializing tasks from config...")

	// 创建工作记录提醒任务配置
	intervalMinutes := 60 // 默认 1 小时
	if minuteInterval > 0 {
		intervalMinutes = minuteInterval
	} else if hourlyInterval > 0 {
		intervalMinutes = hourlyInterval * 60
	}

	now := time.Now()
	nextReminderTime := calculateNextReminderTime(now, intervalMinutes)

	reminderTask := &TaskConfig{
		ID:              "work-reminder",
		Name:            "工作记录提醒",
		Type:            TaskTypeInterval,
		Enabled:         true,
		IntervalMinutes: intervalMinutes,
		NextRun:         nextReminderTime,
	}

	if err := s.upsertTask(reminderTask); err != nil {
		return err
	}
	log.Printf("Initialized task: %s (interval: %d minutes, next run: %s)",
		reminderTask.Name, intervalMinutes, nextReminderTime.Format("15:04:05"))

	// 创建每日总结任务配置
	nextSummaryTime := CalculateNextSummaryTime(now, summaryTime)

	summaryTask := &TaskConfig{
		ID:      "daily-summary",
		Name:    "每日总结生成",
		Type:    TaskTypeDaily,
		Enabled: true,
		Time:    summaryTime,
		NextRun: nextSummaryTime,
		Data:    make(map[string]interface{}),
	}

	if err := s.upsertTask(summaryTask); err != nil {
		return err
	}
	log.Printf("Initialized task: %s (time: %s, next run: %s)",
		summaryTask.Name, summaryTime, nextSummaryTime.Format("2006-01-02 15:04:05"))

	// 创建日志轮转任务配置（每3小时执行一次）
	nextLogRotateTime := now.Add(3 * time.Hour)

	logRotateTask := &TaskConfig{
		ID:              "log-rotate",
		Name:            "日志文件轮转",
		Type:            TaskTypeInterval,
		Enabled:         true,
		IntervalMinutes: 180, // 3小时 = 180分钟
		NextRun:         nextLogRotateTime,
	}

	if err := s.upsertTask(logRotateTask); err != nil {
		return err
	}
	log.Printf("Initialized task: %s (interval: 3 hours, next run: %s)",
		logRotateTask.Name, nextLogRotateTime.Format("2006-01-02 15:04:05"))

	// 创建周度总结任务配置（如果启用）
	if enableWeeklySummary {
		nextWeeklySummaryTime := calculateNextWeeklySummaryTime(now, weeklySummaryDay, weeklySummaryTime)

		weeklySummaryTask := &TaskConfig{
			ID:      "weekly-summary",
			Name:    "周度总结生成",
			Type:    TaskTypeDaily,
			Enabled: true,
			Time:    weeklySummaryTime,
			NextRun: nextWeeklySummaryTime,
			Data: map[string]interface{}{
				"weekday": weeklySummaryDay,
			},
		}

		if err := s.upsertTask(weeklySummaryTask); err != nil {
			return err
		}
		log.Printf("Initialized task: %s (weekday: %d, time: %s, next run: %s)",
			weeklySummaryTask.Name, weeklySummaryDay, weeklySummaryTime,
			nextWeeklySummaryTime.Format("2006-01-02 15:04:05"))
	}

	// 所有任务已通过 upsertTask 自动保存到文件
	log.Println("Tasks initialized and saved to registry")
	return nil
}

// upsertTask 添加或更新任务（如果已存在则更新，否则添加）
func (s *Scheduler) upsertTask(task *TaskConfig) error {
	existing := s.registry.GetTask(task.ID)
	if existing != nil {
		// 任务已存在，使用 PatchTask 增量更新静态配置，保留运行时状态（NextRun, LastRun, Data 等）
		return s.registry.PatchTask(task.ID, func(latest *TaskConfig) {
			latest.Name = task.Name
			latest.Type = task.Type
			latest.Enabled = task.Enabled
			latest.IntervalMinutes = task.IntervalMinutes
			latest.Time = task.Time
			// 注意：不更新 NextRun, LastRun, LastSuccess, LastError, Data
			// 从而在重启后保留任务的执行进度和状态
		})
	}
	// 任务不存在，添加
	return s.registry.AddTask(task)
}

// calculateNextReminderTime 计算下次提醒时间
func calculateNextReminderTime(from time.Time, intervalMinutes int) time.Time {
	interval := time.Duration(intervalMinutes) * time.Minute

	// 对齐到分钟边界
	next := from.Truncate(time.Minute).Add(interval)

	// 确保在未来
	for !next.After(from) {
		next = next.Add(interval)
	}

	return next
}

// CalculateNextSummaryTime 计算下次总结时间（导出函数，供 tasks 包调用）
func CalculateNextSummaryTime(from time.Time, summaryTime string) time.Time {
	// 解析总结时间
	var hour, minute int
	if _, err := fmt.Sscanf(summaryTime, "%d:%d", &hour, &minute); err != nil {
		// 解析失败，使用默认时间 00:00
		hour, minute = 0, 0
	}

	// 构造今天的总结时间点
	todaySummaryTime := time.Date(from.Year(), from.Month(), from.Day(),
		hour, minute, 0, 0, from.Location())

	// 如果当前时间已过今天的总结时间，返回明天的总结时间
	if from.After(todaySummaryTime) || from.Equal(todaySummaryTime) {
		return todaySummaryTime.AddDate(0, 0, 1)
	}

	// 否则返回今天的总结时间
	return todaySummaryTime
}

// calculateNextWeeklySummaryTime 计算下次周度总结时间
func calculateNextWeeklySummaryTime(from time.Time, weekday int, summaryTime string) time.Time {
	// 解析时间
	var hour, minute int
	if _, err := fmt.Sscanf(summaryTime, "%d:%d", &hour, &minute); err != nil {
		hour, minute = 9, 0 // 默认 09:00
	}

	// 转换 weekday（1-7）到 Go 的 Weekday（0-6）
	var targetWeekday time.Weekday
	if weekday == 7 {
		targetWeekday = time.Sunday
	} else {
		targetWeekday = time.Weekday(weekday)
	}

	// 计算距离目标星期几还有几天
	daysUntil := int(targetWeekday - from.Weekday())
	if daysUntil < 0 {
		daysUntil += 7
	}

	// 构造目标日期和时间
	targetDate := from.AddDate(0, 0, daysUntil)
	targetTime := time.Date(
		targetDate.Year(), targetDate.Month(), targetDate.Day(),
		hour, minute, 0, 0, targetDate.Location(),
	)

	// 如果目标时间已过（今天是目标星期几但时间已过），跳到下周
	if !targetTime.After(from) {
		targetTime = targetTime.AddDate(0, 0, 7)
	}

	return targetTime
}
