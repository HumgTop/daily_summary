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

	// 保存到文件
	if err := s.registry.Save(); err != nil {
		return err
	}

	log.Println("Tasks initialized and saved to registry")
	return nil
}

// upsertTask 添加或更新任务（如果已存在则更新，否则添加）
func (s *Scheduler) upsertTask(task *TaskConfig) error {
	existing := s.registry.GetTask(task.ID)
	if existing != nil {
		// 任务已存在，更新配置
		return s.registry.UpdateTask(task)
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
