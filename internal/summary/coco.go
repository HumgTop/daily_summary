package summary

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

// CocoClient Coco CLI 客户端
type CocoClient struct {
	cocoPath string
	workDir  string
}

// NewCocoClient 创建 Coco 客户端
func NewCocoClient(cocoPath, projectDir string) (*CocoClient, error) {
	return &CocoClient{
		cocoPath: cocoPath,
		workDir:  projectDir, // 使用项目目录
	}, nil
}

// GenerateSummary 调用 Coco 生成总结
func (c *CocoClient) GenerateSummary(prompt string) (string, error) {
	// 检查 coco 是否存在
	cocoPath := c.cocoPath
	if cocoPath == "" {
		cocoPath = "coco"
	}

	if _, err := exec.LookPath(cocoPath); err != nil {
		// 如果 coco 不存在，使用回退总结
		log.Println("Warning: coco not found, using fallback summary")
		return c.generateFallbackSummary(), nil
	}

	// 记录调用信息
	log.Printf("调用 Coco: %s -p", cocoPath)
	log.Printf("工作目录: %s", c.workDir)
	log.Printf("Prompt 长度: %d 字符", len(prompt))

	// 同时在控制台输出进度（方便 CLI 用户看到）
	fmt.Printf("调用 Coco 生成总结...\n")

	// 调用 coco -p "{prompt}"
	cmd := exec.Command(cocoPath, "-p", prompt)
	cmd.Dir = c.workDir // 设置命令执行目录为项目目录

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Println("等待 Coco 响应...")
	fmt.Println("正在等待 Coco 响应...")

	if err := cmd.Run(); err != nil {
		log.Printf("Coco 执行失败: %v", err)
		if stderr.Len() > 0 {
			log.Printf("错误输出: %s", stderr.String())
		}
		return "", fmt.Errorf("execute coco: %w, stderr: %s", err, stderr.String())
	}

	response := stdout.String()
	log.Printf("✓ Coco 响应成功，长度: %d 字符", len(response))
	fmt.Printf("✓ Coco 响应成功 (长度: %d 字符)\n", len(response))

	return response, nil
}

// generateFallbackSummary 生成回退总结（当 coco 不可用时）
func (c *CocoClient) generateFallbackSummary() string {
	return `## 主要完成的任务

（由于 Coco CLI 不可用，这是一个自动生成的简单总结）

根据今天的工作记录，完成了多项任务。

## 关键进展

请查看原始工作记录以了解详细信息。

## 遇到的问题

暂无记录

## 明日计划

根据今天的进展继续推进

---

注意：这是一个简化的总结。要获得更详细的 AI 生成总结，请安装 Coco CLI。
`
}
