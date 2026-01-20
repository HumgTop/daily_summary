package summary

import (
	"fmt"
	"strings"
	"time"

	"humg.top/daily_summary/internal/models"
	"humg.top/daily_summary/internal/storage"
)

// Generator 总结生成器
type Generator struct {
	storage      storage.Storage
	claudeClient *ClaudeClient
}

// NewGenerator 创建总结生成器
func NewGenerator(storage storage.Storage, claudeClient *ClaudeClient) *Generator {
	return &Generator{
		storage:      storage,
		claudeClient: claudeClient,
	}
}

// GenerateDailySummary 生成每日总结
func (g *Generator) GenerateDailySummary(date time.Time) error {
	// 获取当天的所有工作记录
	dailyData, err := g.storage.GetDailyData(date)
	if err != nil {
		return fmt.Errorf("get daily data: %w", err)
	}

	if len(dailyData.Entries) == 0 {
		return fmt.Errorf("no work entries for date %s", date.Format("2006-01-02"))
	}

	// 构建提示词
	prompt := g.buildPrompt(dailyData)

	// 调用 Claude Code 生成总结
	summary, err := g.claudeClient.GenerateSummary(prompt)
	if err != nil {
		return fmt.Errorf("generate summary: %w", err)
	}

	// 保存总结
	metadata := models.SummaryMetadata{
		GeneratedAt: time.Now(),
		Date:        date.Format("2006-01-02"),
		EntryCount:  len(dailyData.Entries),
	}

	if err := g.storage.SaveSummary(date, summary, metadata); err != nil {
		return fmt.Errorf("save summary: %w", err)
	}

	return nil
}

// buildPrompt 构建发送给 Claude 的提示词
func (g *Generator) buildPrompt(dailyData *models.DailyData) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("请为以下工作记录生成一份结构化的工作总结（日期：%s）\n\n", dailyData.Date))
	builder.WriteString("工作记录：\n\n")

	for _, entry := range dailyData.Entries {
		timeStr := entry.Timestamp.Format("15:04")
		builder.WriteString(fmt.Sprintf("- **%s**: %s\n", timeStr, entry.Content))
	}

	builder.WriteString("\n请按照以下格式生成总结：\n")
	builder.WriteString("## 主要完成的任务\n")
	builder.WriteString("（列出完成的主要工作，按项目或模块分类）\n\n")
	builder.WriteString("## 关键进展\n")
	builder.WriteString("（突出重要的进展和成果）\n\n")
	builder.WriteString("## 遇到的问题\n")
	builder.WriteString("（如果有记录到问题，列出来）\n\n")
	builder.WriteString("## 明日计划\n")
	builder.WriteString("（如果记录中有提及，整理出来）\n")

	return builder.String()
}
