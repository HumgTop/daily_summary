package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"humg.top/daily_summary/internal/models"
	"humg.top/daily_summary/internal/storage"
)

// RunAdd 添加工作记录
func RunAdd(store storage.Storage, content string) error {
	now := time.Now()

	entry := models.WorkEntry{
		Timestamp: now,
		Content:   content,
	}

	if err := store.SaveEntry(entry); err != nil {
		return fmt.Errorf("failed to save entry: %w", err)
	}

	log.Printf("Work entry added: %s", content)
	fmt.Printf("✓ 已记录：%s (%s)\n", content, now.Format("15:04"))

	// 发送重置信号给调度器
	if err := sendResetSignal(); err != nil {
		// 发送信号失败不影响主流程，只记录日志
		log.Printf("Failed to send reset signal: %v", err)
	} else {
		log.Println("Reset signal sent to scheduler")
	}

	return nil
}

// RunList 列出今日记录
func RunList(store storage.Storage) error {
	today := time.Now()

	dailyData, err := store.GetDailyData(today)
	if err != nil {
		return fmt.Errorf("failed to get daily data: %w", err)
	}

	if len(dailyData.Entries) == 0 {
		fmt.Println("今日暂无记录")
		return nil
	}

	fmt.Printf("📝 今日工作记录 (%s)：\n\n", today.Format("2006-01-02"))
	for _, entry := range dailyData.Entries {
		fmt.Printf("  • %s - %s\n", entry.Timestamp.Format("15:04"), entry.Content)
	}
	fmt.Printf("\n共 %d 条记录\n", len(dailyData.Entries))

	return nil
}

// CheckAndAcquireLock 检查并获取进程锁
func CheckAndAcquireLock() error {
	lockFile := getLockFilePath()

	// 读取现有锁文件
	if data, err := os.ReadFile(lockFile); err == nil {
		oldPID := strings.TrimSpace(string(data))

		// 检查进程是否还在运行
		if isProcessRunning(oldPID) {
			return fmt.Errorf("服务已在运行 (PID: %s)\n\n提示：\n  - 后台服务已启动，无需手动运行 serve 命令\n  - 如需查看日志: tail -f ~/daily_summary/logs/app.log\n  - 如需重启服务: ./scripts/install.sh\n  - 如需停止服务: ./scripts/uninstall.sh", oldPID)
		}

		// 进程已结束，删除旧锁文件
		os.Remove(lockFile)
	}

	// 确保目录存在
	lockDir := filepath.Dir(lockFile)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return fmt.Errorf("failed to create lock directory: %w", err)
	}

	// 创建新锁文件
	pid := fmt.Sprintf("%d", os.Getpid())
	if err := os.WriteFile(lockFile, []byte(pid), 0644); err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	return nil
}

// ReleaseLock 释放进程锁
func ReleaseLock() {
	lockFile := getLockFilePath()
	os.Remove(lockFile)
}

// getLockFilePath 获取锁文件路径
func getLockFilePath() string {
	return filepath.Join("run", "daily_summary.lock")
}

// isProcessRunning 检查进程是否在运行
func isProcessRunning(pidStr string) bool {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false
	}

	// 发送信号 0 检查进程是否存在（不实际发送信号）
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// sendResetSignal 发送重置信号给调度器
func sendResetSignal() error {
	signalFile := filepath.Join("run", ".reset_signal")

	// 确保目录存在
	signalDir := filepath.Dir(signalFile)
	if err := os.MkdirAll(signalDir, 0755); err != nil {
		return fmt.Errorf("failed to create signal directory: %w", err)
	}

	// 创建信号文件（空文件即可）
	if err := os.WriteFile(signalFile, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to create signal file: %w", err)
	}

	return nil
}
