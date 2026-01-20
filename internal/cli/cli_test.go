package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSendResetSignal 测试发送重置信号
func TestSendResetSignal(t *testing.T) {
	// 使用临时目录进行测试
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	os.MkdirAll(dataDir, 0755)
	
	signalFile := filepath.Join(tmpDir, ".reset_signal")

	// 发送信号
	if err := sendResetSignal(dataDir); err != nil {
		t.Fatalf("Failed to send reset signal: %v", err)
	}

	// 验证文件已创建
	if _, err := os.Stat(signalFile); os.IsNotExist(err) {
		t.Error("Signal file should be created")
	}
}

// TestSendResetSignalMultipleTimes 测试多次发送信号
func TestSendResetSignalMultipleTimes(t *testing.T) {
	// 使用临时目录进行测试
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	os.MkdirAll(dataDir, 0755)
	
	signalFile := filepath.Join(tmpDir, ".reset_signal")

	// 连续发送多次
	for i := 0; i < 3; i++ {
		if err := sendResetSignal(dataDir); err != nil {
			t.Fatalf("Failed to send reset signal (attempt %d): %v", i+1, err)
		}
	}

	// 验证文件存在
	if _, err := os.Stat(signalFile); os.IsNotExist(err) {
		t.Error("Signal file should exist after multiple sends")
	}
}
