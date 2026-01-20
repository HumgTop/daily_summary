package models

import "time"

// WorkEntry 表示单次工作记录
type WorkEntry struct {
	Timestamp time.Time `json:"timestamp"` // 记录时间
	Content   string    `json:"content"`   // 工作内容
}

// DailyData 表示一天的所有工作记录
type DailyData struct {
	Date             string      `json:"date"`               // 格式: YYYY-MM-DD
	Entries          []WorkEntry `json:"entries"`            // 工作记录列表
	SummaryGenerated bool        `json:"summary_generated"`  // 是否已生成总结
}

// SummaryMetadata 总结的元数据
type SummaryMetadata struct {
	GeneratedAt time.Time `json:"generated_at"` // 生成时间
	Date        string    `json:"date"`         // 总结对应的日期
	EntryCount  int       `json:"entry_count"`  // 记录条数
}

// Config 应用配置
type Config struct {
	DataDir        string `yaml:"data_dir" json:"data_dir"`                 // 数据目录
	SummaryDir     string `yaml:"summary_dir" json:"summary_dir"`           // 总结目录
	HourlyInterval int    `yaml:"hourly_interval" json:"hourly_interval"`   // 小时间隔（默认1）
	MinuteInterval int    `yaml:"minute_interval" json:"minute_interval"`   // 分钟间隔（如果设置则优先使用）
	SummaryTime    string `yaml:"summary_time" json:"summary_time"`         // 生成总结的时间（默认"00:00"）
	
	// AI 总结生成配置
	AIProvider     string `yaml:"ai_provider" json:"ai_provider"`           // AI 提供商："codex" 或 "claude"（默认 codex）
	CodexPath      string `yaml:"codex_path" json:"codex_path"`             // Codex CLI 路径
	ClaudeCodePath string `yaml:"claude_code_path" json:"claude_code_path"` // Claude Code CLI 路径
	
	DialogTimeout  int    `yaml:"dialog_timeout" json:"dialog_timeout"`     // 对话框超时（秒）
	EnableLogging  bool   `yaml:"enable_logging" json:"enable_logging"`     // 是否启用日志
}
