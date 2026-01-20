package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSendResetSignal 测试发送重置信号
func TestSendResetSignal(t *testing.T) {
	signalFile := filepath.Join("run", ".reset_signal")

	// 清理可能存在的旧文件
	os.Remove(signalFile)

	// 发送信号
	if err := sendResetSignal(); err != nil {
		t.Fatalf("Failed to send reset signal: %v", err)
	}

	// 验证文件已创建
	if _, err := os.Stat(signalFile); os.IsNotExist(err) {
		t.Error("Signal file should be created")
	}

	// 清理
	os.Remove(signalFile)
}

// TestSendResetSignalMultipleTimes 测试多次发送信号
func TestSendResetSignalMultipleTimes(t *testing.T) {
	signalFile := filepath.Join("run", ".reset_signal")

	// 清理
	os.Remove(signalFile)

	// 连续发送多次
	for i := 0; i < 3; i++ {
		if err := sendResetSignal(); err != nil {
			t.Fatalf("Failed to send reset signal (attempt %d): %v", i+1, err)
		}
	}

	// 验证文件存在
	if _, err := os.Stat(signalFile); os.IsNotExist(err) {
		t.Error("Signal file should exist after multiple sends")
	}

	// 清理
	os.Remove(signalFile)
}
