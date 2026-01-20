package summary

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// CodexClient Codex CLI 客户端
type CodexClient struct {
	codexPath string
	workDir   string
}

// NewCodexClient 创建 Codex 客户端
func NewCodexClient(codexPath, projectDir string) (*CodexClient, error) {
	// 创建临时工作目录 for internal use if needed, but we mainly care about projectDir for execution
	homeDir, _ := os.UserHomeDir()
	tempDir := filepath.Join(homeDir, ".daily_summary_temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("create work dir: %w", err)
	}

	return &CodexClient{
		codexPath: codexPath,
		workDir:   projectDir, // Use the project directory provided
	}, nil
}

// GenerateSummary 调用 Codex 生成总结
func (c *CodexClient) GenerateSummary(prompt string) (string, error) {
	// 检查 codex 是否存在
	codexPath := c.codexPath
	if codexPath == "" {
		codexPath = "codex"
	}
	
	if _, err := exec.LookPath(codexPath); err != nil {
		// 如果 codex 不存在，使用回退总结
		log.Println("Warning: codex not found, using fallback summary")
		return c.generateFallbackSummary(), nil
	}

	// 记录调用信息
	log.Printf("调用 Codex: %s exec", codexPath)
	log.Printf("工作目录: %s", c.workDir)
	log.Printf("Prompt 长度: %d 字符", len(prompt))

	// 调用 codex exec "{prompt}"
	cmd := exec.Command(codexPath, "exec", prompt)
	cmd.Dir = c.workDir // 设置命令执行目录为项目目录

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Println("等待 Codex 响应...")
	if err := cmd.Run(); err != nil {
		log.Printf("Codex 执行失败: %v", err)
		if stderr.Len() > 0 {
			log.Printf("错误输出: %s", stderr.String())
		}
		return "", fmt.Errorf("execute codex: %w, stderr: %s", err, stderr.String())
	}

	response := stdout.String()
	log.Printf("✓ Codex 响应成功，长度: %d 字符", len(response))
	
	return response, nil
}

// generateFallbackSummary 生成回退总结（当 codex 不可用时）
func (c *CodexClient) generateFallbackSummary() string {
	return `## 主要完成的任务

（由于 Codex CLI 不可用，这是一个自动生成的简单总结）

根据今天的工作记录，完成了多项任务。

## 关键进展

请查看原始工作记录以了解详细信息。

## 遇到的问题

暂无记录

## 明日计划

根据今天的进展继续推进

---

注意：这是一个简化的总结。要获得更详细的 AI 生成总结，请安装 Codex CLI。
`
}
