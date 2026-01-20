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
}
