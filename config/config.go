package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"humg.top/daily_summary/internal/models"
)

// DefaultConfig 返回默认配置
func DefaultConfig() *models.Config {
	homeDir, _ := os.UserHomeDir()
	return &models.Config{
		DataDir:              filepath.Join(homeDir, "daily_summary", "data"),
		SummaryDir:           filepath.Join(homeDir, "daily_summary", "summaries"),
		HourlyInterval:       1,
		SummaryTime:          "00:00",
		ClaudeCodePath:       "claude-code",
		DialogTimeout:        300, // 5分钟
		EnableLogging:        true,
	}
}

// Load 从文件加载配置，如果不存在则使用默认配置
// 支持 YAML 和 JSON 格式，根据文件扩展名自动识别
func Load(configPath string) (*models.Config, error) {
	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()

	// 根据文件扩展名判断格式
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse JSON config: %w", err)
		}
	default:
	// 默认使用 YAML
		if err := yaml.Unmarshal(data, cfg); err != nil {
			// 如果 YAML 失败，尝试 JSON
			if jsonErr := json.Unmarshal(data, cfg); jsonErr != nil {
				return nil, fmt.Errorf("parse config (tried YAML and JSON): YAML error: %w, JSON error: %v", err, jsonErr)
			}
		}
	}

	// 解析路径
	resolvePaths(cfg)

	return cfg, nil
}

// resolvePaths 根据 WorkDir 解析配置中的路径
func resolvePaths(cfg *models.Config) {
	// 如果配置了 WorkDir，将其转换为绝对路径
	if cfg.WorkDir != "" {
		if absPath, err := filepath.Abs(cfg.WorkDir); err == nil {
			cfg.WorkDir = absPath
		}
	}

	// 辅助函数：将相对路径转换为基于 WorkDir 的绝对路径
	resolve := func(path string) string {
		if path == "" {
			return path
		}
		// 如果是绝对路径，直接返回
		if filepath.IsAbs(path) {
			return path
		}
		// 如果没有配置 WorkDir，保持原样（或者可以使用当前目录，视需求而定）
		if cfg.WorkDir == "" {
			return path
		}
		// 拼接路径
		return filepath.Join(cfg.WorkDir, path)
	}

	cfg.DataDir = resolve(cfg.DataDir)
	cfg.SummaryDir = resolve(cfg.SummaryDir)
	cfg.LogFile = resolve(cfg.LogFile)
}

// Save 保存配置到文件
// 根据文件扩展名自动选择 YAML 或 JSON 格式
func Save(cfg *models.Config, configPath string) error {
	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	var data []byte
	var err error

	// 根据文件扩展名判断格式
	ext := strings.ToLower(filepath.Ext(configPath))
	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshal YAML config: %w", err)
		}
	case ".json":
		data, err = json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal JSON config: %w", err)
		}
	default:
		// 默认使用 YAML
		data, err = yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshal config: %w", err)
		}
	}

	return os.WriteFile(configPath, data, 0644)
}

// EnsureDirectories 确保必要的目录存在
func EnsureDirectories(cfg *models.Config) error {
	homeDir, _ := os.UserHomeDir()
	dirs := []string{
		cfg.DataDir,
		cfg.SummaryDir,
		filepath.Join(homeDir, "daily_summary", "logs"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
