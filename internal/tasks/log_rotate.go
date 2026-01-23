package tasks

import (
	"fmt"
	"log"
	"os"
	"time"

	"humg.top/daily_summary/internal/scheduler"
)

// LogRotateTask 日志轮转任务
type LogRotateTask struct {
	logFiles     []string // 需要轮转的日志文件路径列表
	maxLogSizeMB int      // 日志文件最大大小（MB）
}

// NewLogRotateTask 创建日志轮转任务
func NewLogRotateTask(logFiles []string, maxLogSizeMB int) *LogRotateTask {
	return &LogRotateTask{
		logFiles:     logFiles,
		maxLogSizeMB: maxLogSizeMB,
	}
}

// ID 返回任务 ID
func (t *LogRotateTask) ID() string {
	return "log-rotate"
}

// Name 返回任务名称
func (t *LogRotateTask) Name() string {
	return "日志文件轮转"
}

// ShouldRun 判断是否应该执行
func (t *LogRotateTask) ShouldRun(now time.Time, config *scheduler.TaskConfig) (bool, func(*scheduler.TaskConfig)) {
	if !config.Enabled {
		return false, nil
	}

	// 检查下次执行时间
	if config.NextRun.IsZero() {
		return false, nil
	}

	return !now.Before(config.NextRun), nil
}

// Execute 执行任务
func (t *LogRotateTask) Execute() error {
	if t.maxLogSizeMB <= 0 {
		// 未设置大小限制，跳过
		return nil
	}

	var rotatedFiles []string
	var errors []error

	for _, logFile := range t.logFiles {
		rotated, err := t.rotateIfNeeded(logFile)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", logFile, err))
		} else if rotated {
			rotatedFiles = append(rotatedFiles, logFile)
		}
	}

	if len(rotatedFiles) > 0 {
		log.Printf("Log rotation completed: %d file(s) rotated", len(rotatedFiles))
	}

	if len(errors) > 0 {
		return fmt.Errorf("log rotation errors: %v", errors)
	}

	return nil
}

// OnExecuted 任务执行后的回调
func (t *LogRotateTask) OnExecuted(now time.Time, config *scheduler.TaskConfig, err error) {
	// 更新最后执行时间
	config.LastRun = now

	if err != nil {
		config.LastError = err.Error()
		log.Printf("Task %s failed: %v", t.Name(), err)
	} else {
		config.LastSuccess = now
		config.LastError = ""
	}

	// 计算下次执行时间（3小时后）
	config.NextRun = now.Add(3 * time.Hour)
}

// rotateIfNeeded 检查并轮转单个日志文件
func (t *LogRotateTask) rotateIfNeeded(logFile string) (bool, error) {
	// 检查文件是否存在
	info, err := os.Stat(logFile)
	if os.IsNotExist(err) {
		// 文件不存在，无需轮转
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat log file: %w", err)
	}

	// 计算文件大小（字节转 MB）
	fileSizeMB := float64(info.Size()) / (1024 * 1024)

	// 如果文件大小未超过限制，无需轮转
	if fileSizeMB <= float64(t.maxLogSizeMB) {
		return false, nil
	}

	// 执行日志轮转：重命名为 .old
	oldLogFile := logFile + ".old"

	// 如果 .old 文件已存在，先删除
	if _, err := os.Stat(oldLogFile); err == nil {
		if err := os.Remove(oldLogFile); err != nil {
			return false, fmt.Errorf("remove old backup: %w", err)
		}
	}

	// 重命名当前日志文件为 .old
	if err := os.Rename(logFile, oldLogFile); err != nil {
		return false, fmt.Errorf("rename log file: %w", err)
	}

	log.Printf("Log file rotated: %s (%.2f MB) -> %s", logFile, fileSizeMB, oldLogFile)
	return true, nil
}
