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

// WeeklyPromptData 周报模板数据结构
type WeeklyPromptData struct {
	WeekStartDate   string
	WeekEndDate     string
	EntryCount      int
	DailySummaries  []DailySummaryEntry
}

// DailySummaryEntry 单日总结条目
type DailySummaryEntry struct {
	Date       string
	Weekday    string
	HasSummary bool
	Summary    string
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

// GenerateWeeklySummary 生成周度总结
// weekEndDate: 周的最后一天（周日）
func (g *Generator) GenerateWeeklySummary(weekEndDate time.Time) error {
	// 计算周的开始日期（周一）
	weekStartDate := weekEndDate.AddDate(0, 0, -6)

	log.Printf("Generating weekly summary for week %s to %s",
		weekStartDate.Format("2006-01-02"),
		weekEndDate.Format("2006-01-02"))

	// 获取该周的所有每日总结
	dailySummaries, err := g.storage.GetDailySummariesInRange(weekStartDate, weekEndDate)
	if err != nil {
		return fmt.Errorf("get daily summaries: %w", err)
	}

	if len(dailySummaries) == 0 {
		return fmt.Errorf("no daily summaries found for week %s to %s",
			weekStartDate.Format("2006-01-02"),
			weekEndDate.Format("2006-01-02"))
	}

	log.Printf("Found %d daily summaries for the week", len(dailySummaries))

	// 构建周度总结的 prompt
	prompt := g.buildWeeklyPrompt(weekStartDate, weekEndDate, dailySummaries)

	// 调用 AI 生成周度总结
	summary, err := g.aiClient.GenerateSummary(prompt)
	if err != nil {
		return fmt.Errorf("generate weekly summary: %w", err)
	}

	// 保存周度总结
	metadata := models.SummaryMetadata{
		GeneratedAt: time.Now(),
		Date:        weekEndDate.Format("2006-01-02"),
		EntryCount:  len(dailySummaries),
	}

	if err := g.storage.SaveWeeklySummary(weekEndDate, summary, metadata); err != nil {
		return fmt.Errorf("save weekly summary: %w", err)
	}

	// 发送通知
	if g.notifier != nil {
		title := "周报已生成"
		message := fmt.Sprintf("%s 至 %s 的周报已完成",
			weekStartDate.Format("01月02日"),
			weekEndDate.Format("01月02日"))
		log.Printf("Sending notification: %s - %s", title, message)
		if err := g.notifier.ShowNotification(title, message); err != nil {
			log.Printf("Failed to send notification: %v", err)
		} else {
			log.Printf("Notification sent successfully")
		}
	}

	log.Printf("Weekly summary generated successfully")
	return nil
}

// buildWeeklyPrompt 构建周度总结的 prompt
func (g *Generator) buildWeeklyPrompt(
	weekStartDate, weekEndDate time.Time,
	dailySummaries map[string]string,
) string {
	// 准备模板数据
	summaries := make([]DailySummaryEntry, 0, 7)
	current := weekStartDate
	for !current.After(weekEndDate) {
		dateStr := current.Format("2006-01-02")
		weekdayStr := getWeekdayName(current)

		if summary, ok := dailySummaries[dateStr]; ok {
			summaries = append(summaries, DailySummaryEntry{
				Date:       dateStr,
				Weekday:    weekdayStr,
				HasSummary: true,
				Summary:    summary,
			})
		} else {
			summaries = append(summaries, DailySummaryEntry{
				Date:       dateStr,
				Weekday:    weekdayStr,
				HasSummary: false,
				Summary:    "",
			})
		}

		current = current.AddDate(0, 0, 1)
	}

	data := WeeklyPromptData{
		WeekStartDate:  weekStartDate.Format("2006-01-02"),
		WeekEndDate:    weekEndDate.Format("2006-01-02"),
		EntryCount:     len(dailySummaries),
		DailySummaries: summaries,
	}

	// 确定模板路径
	templatePath := "templates/weekly_summary_prompt.md"

	// 读取模板文件
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		log.Printf("Warning: failed to read weekly template file %s: %v, using fallback", templatePath, err)
		return g.buildWeeklyFallbackPrompt(weekStartDate, weekEndDate, dailySummaries)
	}

	// 解析并执行模板
	tmpl, err := template.New("weekly_prompt").Parse(string(templateContent))
	if err != nil {
		log.Printf("Warning: failed to parse weekly template: %v, using fallback", err)
		return g.buildWeeklyFallbackPrompt(weekStartDate, weekEndDate, dailySummaries)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Warning: failed to execute weekly template: %v, using fallback", err)
		return g.buildWeeklyFallbackPrompt(weekStartDate, weekEndDate, dailySummaries)
	}

	return buf.String()
}

// buildWeeklyFallbackPrompt 降级方案：使用原有的硬编码逻辑
func (g *Generator) buildWeeklyFallbackPrompt(
	weekStartDate, weekEndDate time.Time,
	dailySummaries map[string]string,
) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("请基于以下每日工作总结生成一份周报（%s 至 %s）\n\n",
		weekStartDate.Format("2006-01-02"),
		weekEndDate.Format("2006-01-02")))

	builder.WriteString("## 本周每日总结\n\n")

	// 按日期顺序遍历（周一到周日）
	current := weekStartDate
	for !current.After(weekEndDate) {
		dateStr := current.Format("2006-01-02")
		weekdayStr := getWeekdayName(current)

		if summary, ok := dailySummaries[dateStr]; ok {
			builder.WriteString(fmt.Sprintf("### %s (%s)\n\n", dateStr, weekdayStr))
			builder.WriteString(summary)
			builder.WriteString("\n\n")
		} else {
			builder.WriteString(fmt.Sprintf("### %s (%s)\n\n", dateStr, weekdayStr))
			builder.WriteString("*（当天无工作记录）*\n\n")
		}

		current = current.AddDate(0, 0, 1)
	}

	builder.WriteString("---\n\n")
	builder.WriteString("请基于以上每日总结，生成一份结构化的周报，包括以下部分：\n\n")
	builder.WriteString("## 本周完成情况\n")
	builder.WriteString("（汇总本周完成的主要任务，按项目或模块分类）\n\n")
	builder.WriteString("## 关键进展与成果\n")
	builder.WriteString("（突出本周的重要进展、里程碑和亮点）\n\n")
	builder.WriteString("## 遇到的问题与解决方案\n")
	builder.WriteString("（列出本周遇到的主要问题及解决情况）\n\n")
	builder.WriteString("## 下周计划\n")
	builder.WriteString("（基于本周情况和记录中的计划，规划下周重点）\n")

	return builder.String()
}

// getWeekdayName 获取星期几的中文名称
func getWeekdayName(t time.Time) string {
	weekdays := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	return weekdays[t.Weekday()]
}
