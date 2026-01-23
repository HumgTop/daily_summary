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
	// 检查是否已有任务配置
	configs := s.registry.GetAllTasks()
	if len(configs) > 0 {
		log.Println("Tasks already initialized from registry")
		return nil
	}

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

	if err := s.registry.AddTask(reminderTask); err != nil {
		return err
	}
	log.Printf("Initialized task: %s (interval: %d minutes, next run: %s)",
		reminderTask.Name, intervalMinutes, nextReminderTime.Format("15:04:05"))

	// 创建每日总结任务配置
	nextSummaryTime := calculateNextSummaryTime(now, summaryTime)

	summaryTask := &TaskConfig{
		ID:      "daily-summary",
		Name:    "每日总结生成",
		Type:    TaskTypeDaily,
		Enabled: true,
		Time:    summaryTime,
		NextRun: nextSummaryTime,
		Data:    make(map[string]interface{}),
	}

	if err := s.registry.AddTask(summaryTask); err != nil {
		return err
	}
	log.Printf("Initialized task: %s (time: %s, next run: %s)",
		summaryTask.Name, summaryTime, nextSummaryTime.Format("2006-01-02 15:04:05"))

	// 保存到文件
	if err := s.registry.Save(); err != nil {
		return err
	}

	log.Println("Tasks initialized and saved to registry")
	return nil
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

// calculateNextSummaryTime 计算下次总结时间
func calculateNextSummaryTime(from time.Time, summaryTime string) time.Time {
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
