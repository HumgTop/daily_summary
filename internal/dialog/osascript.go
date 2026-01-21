package dialog

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// OSAScriptDialog 使用 osascript 实现的对话框
type OSAScriptDialog struct {
	timeout time.Duration
}

// NewOSAScriptDialog 创建新的 osascript 对话框
func NewOSAScriptDialog(timeout time.Duration) *OSAScriptDialog {
	return &OSAScriptDialog{
		timeout: timeout,
	}
}

// ShowInput 显示文本输入对话框
func (d *OSAScriptDialog) ShowInput(title, message, defaultText string) (string, bool, error) {
	// 在消息末尾添加友好提示
	enhancedMessage := message
	if !strings.HasSuffix(message, ":") && !strings.HasSuffix(message, "：") {
		enhancedMessage = message + "\n\n💡 提示：可输入任意长度的文本内容"
	}

	// 构建优化的 AppleScript 命令
	// 移除 with icon note，使弹窗更加简洁
	script := fmt.Sprintf(`display dialog "%s" default answer "%s" with title "%s" buttons {"取消", "确定"} default button "确定"`,
		escapeString(enhancedMessage),
		escapeString(defaultText),
		escapeString(title),
	)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()

	// 执行 osascript
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// 用户取消或超时
		if ctx.Err() == context.DeadlineExceeded {
			return "", false, fmt.Errorf("dialog timeout")
		}
		// 用户点击取消按钮
		if strings.Contains(stderr.String(), "User canceled") {
			return "", false, nil
		}
		return "", false, fmt.Errorf("osascript error: %w, stderr: %s", err, stderr.String())
	}

	// 解析输出: "button returned:确定, text returned:用户输入的内容"
	output := stdout.String()
	text := parseOSAScriptOutput(output)

	return text, true, nil
}

// escapeString 转义 AppleScript 字符串中的特殊字符
func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// parseOSAScriptOutput 解析 osascript 的输出
func parseOSAScriptOutput(output string) string {
	// 输出格式: "button returned:确定, text returned:内容"
	parts := strings.Split(output, "text returned:")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// ShowNotification 显示弹窗提醒（改用 dialog 而非系统通知）
func (d *OSAScriptDialog) ShowNotification(title, message string) error {
	// 使用 display dialog 显示弹窗，只有一个"确定"按钮
	script := fmt.Sprintf(`display dialog "%s" with title "%s" buttons {"确定"} default button "确定" with icon note`,
		escapeString(message),
		escapeString(title),
	)

	// 创建带超时的上下文（30秒足够用户看到）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 执行 osascript
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// 超时或其他错误
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("notification dialog timeout")
		}
		return fmt.Errorf("osascript notification error: %w, stderr: %s", err, stderr.String())
	}

	return nil
}
