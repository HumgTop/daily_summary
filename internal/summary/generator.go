package summary

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"humg.top/daily_summary/internal/models"
	"humg.top/daily_summary/internal/storage"
)

// Notifier 通知接口
type Notifier interface {
	ShowNotification(title, message string) error
}

// Generator 总结生成器
type Generator struct {
	storage      storage.Storage
	aiClient     AIClient
	notifier     Notifier
	templatePath string // 提示词模板路径
}

// NewGenerator 创建总结生成器
func NewGenerator(storage storage.Storage, aiClient AIClient, notifier Notifier) *Generator {
	return &Generator{
		storage:      storage,
		aiClient:     aiClient,
		notifier:     notifier,
		templatePath: "", // 默认使用内置模板
	}
}

// SetTemplatePath 设置自定义模板路径
func (g *Generator) SetTemplatePath(path string) {
	g.templatePath = path
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

	// 调用 AI 客户端生成总结
	summary, err := g.aiClient.GenerateSummary(prompt)
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

	// 发送通知
	if g.notifier != nil {
		notificationTitle := "工作总结已生成"
		notificationMessage := fmt.Sprintf("%s 的工作总结已完成", date.Format("2006年01月02日"))
		log.Printf("Sending notification: %s - %s", notificationTitle, notificationMessage)
		if err := g.notifier.ShowNotification(notificationTitle, notificationMessage); err != nil {
			// 通知失败不影响主流程，只记录日志
			log.Printf("Failed to send notification: %v", err)
		} else {
			log.Printf("Notification sent successfully")
		}
	} else {
		log.Printf("Warning: notifier is nil, cannot send notification")
	}

	return nil
}

// PromptData 模板数据结构
type PromptData struct {
	Date       string
	EntryCount int
	Entries    []PromptEntry
}

// PromptEntry 单条工作记录
type PromptEntry struct {
	Time    string
	Content string
}

// buildPrompt 构建发送给 Claude 的提示词
func (g *Generator) buildPrompt(dailyData *models.DailyData) string {
	// 准备模板数据
	entries := make([]PromptEntry, 0, len(dailyData.Entries))
	for _, entry := range dailyData.Entries {
		entries = append(entries, PromptEntry{
			Time:    entry.Timestamp.Format("15:04"),
			Content: entry.Content,
		})
	}

	data := PromptData{
		Date:       dailyData.Date,
		EntryCount: len(dailyData.Entries),
		Entries:    entries,
	}

	// 确定模板路径
	templatePath := g.templatePath
	if templatePath == "" {
		// 默认使用模板
		templatePath = "templates/summary_prompt.md"
	}

	// 读取模板文件
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		log.Printf("Warning: failed to read template file %s: %v, using fallback", templatePath, err)
		return g.buildFallbackPrompt(dailyData)
	}

	// 解析并执行模板
	tmpl, err := template.New("prompt").Parse(string(templateContent))
	if err != nil {
		log.Printf("Warning: failed to parse template: %v, using fallback", err)
		return g.buildFallbackPrompt(dailyData)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Warning: failed to execute template: %v, using fallback", err)
		return g.buildFallbackPrompt(dailyData)
	}

	return buf.String()
}

// buildFallbackPrompt 降级方案：使用原有的硬编码逻辑
func (g *Generator) buildFallbackPrompt(dailyData *models.DailyData) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("请为以下工作记录生成一份结构化的工作总结（日期：%s）\n\n", dailyData.Date))
	builder.WriteString("工作记录（每1条记录都是对前一个时间窗口工作内容的总结）：\n\n")

	for _, entry := range dailyData.Entries {
		timeStr := entry.Timestamp.Format("15:04")
		builder.WriteString(fmt.Sprintf("- **%s**: %s\n", timeStr, entry.Content))
	}

	builder.WriteString("\n请按照以下格式生成总结：\n")
	builder.WriteString("## 主要完成的任务\n")
	builder.WriteString("（列出完成的主要工作，按项目或模块分类，并估算工作实际耗时）\n\n")
	builder.WriteString("## 关键进展\n")
	builder.WriteString("（突出重要的进展和成果）\n\n")
	builder.WriteString("## 遇到的问题\n")
	builder.WriteString("（如果有记录到问题，列出来）\n\n")
	builder.WriteString("## 明日计划\n")
	builder.WriteString("（如果记录中有提及，整理出来）\n")

	return builder.String()
}
