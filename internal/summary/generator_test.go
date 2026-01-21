package summary

import (
	"testing"
	"time"

	"humg.top/daily_summary/internal/models"
)

// TestBuildPrompt 测试模板渲染功能
func TestBuildPrompt(t *testing.T) {
	// 创建测试数据
	dailyData := &models.DailyData{
		Date: "2026-01-21",
		Entries: []models.WorkEntry{
			{
				Timestamp: time.Date(2026, 1, 21, 9, 0, 0, 0, time.Local),
				Content:   "开始日常工作，查看邮件和需求",
			},
			{
				Timestamp: time.Date(2026, 1, 21, 10, 30, 0, 0, time.Local),
				Content:   "完成 API 开发",
			},
			{
				Timestamp: time.Date(2026, 1, 21, 14, 0, 0, 0, time.Local),
				Content:   "代码评审和优化",
			},
		},
	}

	generator := &Generator{
		templatePath: "", // 使用默认模板
	}

	// 测试模板渲染
	prompt := generator.buildPrompt(dailyData)

	// 验证输出不为空
	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}

	// 验证包含关键信息
	if !contains(prompt, "2026-01-21") {
		t.Error("prompt should contain date")
	}
	if !contains(prompt, "09:00") {
		t.Error("prompt should contain first entry time")
	}
	if !contains(prompt, "API 开发") {
		t.Error("prompt should contain entry content")
	}

	t.Logf("Generated prompt:\n%s", prompt)
}

// TestBuildPromptWithCustomTemplate 测试自定义模板
func TestBuildPromptWithCustomTemplate(t *testing.T) {
	dailyData := &models.DailyData{
		Date: "2026-01-21",
		Entries: []models.WorkEntry{
			{
				Timestamp: time.Date(2026, 1, 21, 9, 0, 0, 0, time.Local),
				Content:   "测试内容",
			},
		},
	}

	generator := &Generator{
		templatePath: "templates/summary_prompt.md", // 使用详细版模板
	}

	prompt := generator.buildPrompt(dailyData)

	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}

	// 详细版模板应包含更多指导信息
	if !contains(prompt, "注意事项") {
		t.Error("detailed template should contain 注意事项 section")
	}

	t.Logf("Generated prompt with detailed template:\n%s", prompt)
}

// TestBuildPromptFallback 测试降级逻辑
func TestBuildPromptFallback(t *testing.T) {
	dailyData := &models.DailyData{
		Date: "2026-01-21",
		Entries: []models.WorkEntry{
			{
				Timestamp: time.Date(2026, 1, 21, 9, 0, 0, 0, time.Local),
				Content:   "测试内容",
			},
		},
	}

	generator := &Generator{
		templatePath: "non_existent_template.md", // 不存在的模板
	}

	// 应该降级到硬编码逻辑
	prompt := generator.buildPrompt(dailyData)

	if prompt == "" {
		t.Fatal("fallback prompt should not be empty")
	}

	// 验证降级逻辑的输出
	if !contains(prompt, "2026-01-21") {
		t.Error("fallback prompt should contain date")
	}

	t.Logf("Fallback prompt:\n%s", prompt)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
