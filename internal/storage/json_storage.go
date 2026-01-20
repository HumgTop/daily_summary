package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"humg.top/daily_summary/internal/models"
)

// JSONStorage JSON 文件存储实现
type JSONStorage struct {
	dataDir    string
	summaryDir string
}

// NewJSONStorage 创建 JSON 存储实例
func NewJSONStorage(dataDir, summaryDir string) *JSONStorage {
	return &JSONStorage{
		dataDir:    dataDir,
		summaryDir: summaryDir,
	}
}

// SaveEntry 保存工作记录
func (s *JSONStorage) SaveEntry(entry models.WorkEntry) error {
	// 获取当天的数据文件路径
	date := entry.Timestamp.Format("2006-01-02")
	filePath := filepath.Join(s.dataDir, fmt.Sprintf("%s.json", date))

	// 读取现有数据
	var dailyData models.DailyData
	data, err := os.ReadFile(filePath)
	if err == nil {
		if err := json.Unmarshal(data, &dailyData); err != nil {
			return fmt.Errorf("unmarshal daily data: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read daily data file: %w", err)
	} else {
		// 文件不存在，创建新的
		dailyData = models.DailyData{
			Date:    date,
			Entries: []models.WorkEntry{},
		}
	}

	// 添加新记录
	dailyData.Entries = append(dailyData.Entries, entry)

	// 保存回文件
	data, err = json.MarshalIndent(dailyData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal daily data: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write daily data file: %w", err)
	}

	return nil
}

// GetDailyData 获取指定日期的工作记录
func (s *JSONStorage) GetDailyData(date time.Time) (*models.DailyData, error) {
	dateStr := date.Format("2006-01-02")
	filePath := filepath.Join(s.dataDir, fmt.Sprintf("%s.json", dateStr))

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.DailyData{
				Date:    dateStr,
				Entries: []models.WorkEntry{},
			}, nil
		}
		return nil, fmt.Errorf("read daily data file: %w", err)
	}

	var dailyData models.DailyData
	if err := json.Unmarshal(data, &dailyData); err != nil {
		return nil, fmt.Errorf("unmarshal daily data: %w", err)
	}

	return &dailyData, nil
}

// GetLastEntry 获取最后一条工作记录
func (s *JSONStorage) GetLastEntry() (*models.WorkEntry, error) {
	// 获取今天和昨天的数据
	now := time.Now()
	today := s.getDailyDataQuiet(now)
	yesterday := s.getDailyDataQuiet(now.AddDate(0, 0, -1))

	// 合并今天和昨天的记录
	var allEntries []models.WorkEntry
	if yesterday != nil {
		allEntries = append(allEntries, yesterday.Entries...)
	}
	if today != nil {
		allEntries = append(allEntries, today.Entries...)
	}

	// 找到最后一条记录（时间戳最大）
	if len(allEntries) == 0 {
		return nil, nil
	}

	lastEntry := &allEntries[0]
	for i := range allEntries {
		if allEntries[i].Timestamp.After(lastEntry.Timestamp) {
			lastEntry = &allEntries[i]
		}
	}

	return lastEntry, nil
}

// getDailyDataQuiet 静默获取日期数据，出错返回 nil
func (s *JSONStorage) getDailyDataQuiet(date time.Time) *models.DailyData {
	data, err := s.GetDailyData(date)
	if err != nil {
		return nil
	}
	return data
}

// SaveSummary 保存总结
func (s *JSONStorage) SaveSummary(date time.Time, summary string, metadata models.SummaryMetadata) error {
	dateStr := date.Format("2006-01-02")
	filePath := filepath.Join(s.summaryDir, fmt.Sprintf("%s.md", dateStr))

	// 构建 Markdown 内容
	content := fmt.Sprintf(`# 工作总结 - %s

生成时间: %s
记录条数: %d

---

%s
`,
		dateStr,
		metadata.GeneratedAt.Format("2006-01-02 15:04:05"),
		metadata.EntryCount,
		summary,
	)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write summary file: %w", err)
	}

	return nil
}

// GetSummary 获取总结
func (s *JSONStorage) GetSummary(date time.Time) (string, error) {
	dateStr := date.Format("2006-01-02")
	filePath := filepath.Join(s.summaryDir, fmt.Sprintf("%s.md", dateStr))

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read summary file: %w", err)
	}

	return string(data), nil
}

// MarkSummaryGenerated 标记指定日期的总结已生成
func (s *JSONStorage) MarkSummaryGenerated(date time.Time) error {
	dateStr := date.Format("2006-01-02")
	filePath := filepath.Join(s.dataDir, fmt.Sprintf("%s.json", dateStr))

	// 读取现有数据
	var dailyData models.DailyData
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，创建新的
			dailyData = models.DailyData{
				Date:             dateStr,
				Entries:          []models.WorkEntry{},
				SummaryGenerated: true,
			}
		} else {
			return fmt.Errorf("read daily data file: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &dailyData); err != nil {
			return fmt.Errorf("unmarshal daily data: %w", err)
		}
		// 标记为已生成
		dailyData.SummaryGenerated = true
	}

	// 保存回文件
	data, err = json.MarshalIndent(dailyData, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal daily data: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write daily data file: %w", err)
	}

	return nil
}
