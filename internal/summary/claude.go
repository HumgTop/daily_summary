package summary

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ClaudeClient Claude Code CLI 客户端
type ClaudeClient struct {
	claudeCodePath string
	workDir        string // 临时工作目录
}

// NewClaudeClient 创建 Claude 客户端
func NewClaudeClient(claudeCodePath string) (*ClaudeClient, error) {
	// 创建临时工作目录
	homeDir, _ := os.UserHomeDir()
	workDir := filepath.Join(homeDir, ".daily_summary_temp")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, fmt.Errorf("create work dir: %w", err)
	}

	return &ClaudeClient{
		claudeCodePath: claudeCodePath,
		workDir:        workDir,
	}, nil
}

// GenerateSummary 调用 Claude Code 生成总结
func (c *ClaudeClient) GenerateSummary(prompt string) (string, error) {
	// 将提示词写入临时文件
	promptFile := filepath.Join(c.workDir, "prompt.txt")
	if err := os.WriteFile(promptFile, []byte(prompt), 0644); err != nil {
		return "", fmt.Errorf("write prompt file: %w", err)
	}
	defer os.Remove(promptFile)

	// 检查 claude-code 是否存在
	if _, err := exec.LookPath(c.claudeCodePath); err != nil {
		// 如果 claude-code 不存在，使用简单模板生成总结
		return c.generateFallbackSummary(prompt), nil
	}

	// 调用 claude-code CLI
	cmd := exec.Command(c.claudeCodePath, "--prompt", prompt)
	cmd.Dir = c.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("execute claude-code: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// generateFallbackSummary 生成回退总结（当 claude-code 不可用时）
func (c *ClaudeClient) generateFallbackSummary(prompt string) string {
	_ = prompt // 回退方案使用简单模板，不需要解析 prompt
	return `## 主要完成的任务

（由于 Claude Code CLI 不可用，这是一个自动生成的简单总结）

根据今天的工作记录，完成了多项任务。

## 关键进展

请查看原始工作记录以了解详细信息。

## 遇到的问题

暂无记录

## 明日计划

根据今天的进展继续推进

---

注意：这是一个简化的总结。要获得更详细的 AI 生成总结，请安装 Claude Code CLI。
`
}
