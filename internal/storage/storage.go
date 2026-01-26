package storage

import (
	"time"

	"humg.top/daily_summary/internal/models"
)

// Storage 数据存储接口
type Storage interface {
	// SaveEntry 保存单条工作记录
	SaveEntry(entry models.WorkEntry) error

	// GetDailyData 获取指定日期的所有工作记录
	GetDailyData(date time.Time) (*models.DailyData, error)

	// GetLastEntry 获取最后一条工作记录（用于显示在弹窗中）
	GetLastEntry() (*models.WorkEntry, error)

	// SaveSummary 保存生成的总结
	SaveSummary(date time.Time, summary string, metadata models.SummaryMetadata) error

	// GetSummary 获取指定日期的总结
	GetSummary(date time.Time) (string, error)

	// MarkSummaryGenerated 标记指定日期的总结已生成
	MarkSummaryGenerated(date time.Time) error

	// GetDailySummariesInRange 获取日期范围内的每日总结
	// 返回 map[string]string，key 为日期（YYYY-MM-DD），value 为总结内容
	GetDailySummariesInRange(startDate, endDate time.Time) (map[string]string, error)

	// SaveWeeklySummary 保存周度总结
	SaveWeeklySummary(weekEndDate time.Time, summary string, metadata models.SummaryMetadata) error
}
