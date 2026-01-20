package scheduler

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestResetSignalPath 测试重置信号文件路径
func TestResetSignalPath(t *testing.T) {
	path := getResetSignalPath()

	expected := filepath.Join("run", ".reset_signal")

	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

// TestCheckAndClearResetSignal 测试信号文件检测和清除
func TestCheckAndClearResetSignal(t *testing.T) {
	// 创建临时调度器（只需要测试方法）
	s := &Scheduler{}

	signalFile := getResetSignalPath()

	// 确保开始时文件不存在
	os.Remove(signalFile)

	// 测试：文件不存在时返回 false
	if s.checkAndClearResetSignal() {
		t.Error("Expected false when signal file does not exist")
	}

	// 创建信号文件
	signalDir := filepath.Dir(signalFile)
	os.MkdirAll(signalDir, 0755)
	if err := os.WriteFile(signalFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create signal file: %v", err)
	}

	// 测试：文件存在时返回 true 并删除文件
	if !s.checkAndClearResetSignal() {
		t.Error("Expected true when signal file exists")
	}

	// 验证文件已被删除
	if _, err := os.Stat(signalFile); !os.IsNotExist(err) {
		t.Error("Signal file should be deleted after check")
	}
}

// TestResetChannel 测试重置通道机制
func TestResetChannel(t *testing.T) {
	// 创建调度器（不启动）
	s := &Scheduler{
		resetCh: make(chan struct{}, 1),
		stopCh:  make(chan struct{}),
	}

	// 测试非阻塞发送
	select {
	case s.resetCh <- struct{}{}:
		// 成功发送
	default:
		t.Error("Should be able to send to reset channel")
	}

	// 测试接收
	select {
	case <-s.resetCh:
		// 成功接收
	case <-time.After(100 * time.Millisecond):
		t.Error("Should receive from reset channel")
	}
}
