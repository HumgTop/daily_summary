package scheduler

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"humg.top/daily_summary/internal/models"
)

// TestResetSignalPath 测试重置信号文件路径
func TestResetSignalPath(t *testing.T) {
	// 使用临时目录创建测试配置
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	
	s := &Scheduler{
		config: &models.Config{
			DataDir: dataDir,
		},
	}
	
	path := s.getResetSignalPath()
	expected := filepath.Join(tmpDir, ".reset_signal")

	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}
}

// TestCheckAndClearResetSignal 测试信号文件检测和清除
func TestCheckAndClearResetSignal(t *testing.T) {
	// 使用临时目录创建测试配置
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	os.MkdirAll(dataDir, 0755)
	
	s := &Scheduler{
		config: &models.Config{
			DataDir: dataDir,
		},
	}

	signalFile := s.getResetSignalPath()

	// 测试：文件不存在时返回 false
	if s.checkAndClearResetSignal() {
		t.Error("Expected false when signal file does not exist")
	}

	// 创建信号文件
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

// TestHeartbeatMonitor 测试心跳监控和唤醒检测
func TestHeartbeatMonitor(t *testing.T) {
	s := &Scheduler{
		resetCh:        make(chan struct{}, 1),
		summaryResetCh: make(chan struct{}, 1),
		stopCh:         make(chan struct{}),
		lastHeartbeat:  time.Now(),
	}

	// 模拟睡眠：手动设置 lastHeartbeat 为很久以前
	s.lastHeartbeat = time.Now().Add(-30 * time.Second)

	// 调用 handleWakeUp 测试唤醒处理
	s.handleWakeUp(30 * time.Second)

	// 验证重置信号已发送到 resetCh
	select {
	case <-s.resetCh:
		// 成功接收 hourly task 重置信号
	case <-time.After(100 * time.Millisecond):
		t.Error("Should receive reset signal for hourly task")
	}

	// 验证重置信号已发送到 summaryResetCh
	select {
	case <-s.summaryResetCh:
		// 成功接收 summary task 重置信号
	case <-time.After(100 * time.Millisecond):
		t.Error("Should receive reset signal for summary task")
	}
}

// TestWakeUpHandler 测试唤醒处理逻辑
func TestWakeUpHandler(t *testing.T) {
	s := &Scheduler{
		resetCh:        make(chan struct{}, 1),
		summaryResetCh: make(chan struct{}, 1),
		stopCh:         make(chan struct{}),
		lastHeartbeat:  time.Now(),
	}

	// 测试短睡眠（< 1小时）
	s.handleWakeUp(30 * time.Minute)

	// 清空通道
	<-s.resetCh
	<-s.summaryResetCh

	// 测试长睡眠（> 1小时）
	s.handleWakeUp(2 * time.Hour)

	// 验证两个通道都收到信号
	select {
	case <-s.resetCh:
	default:
		t.Error("Should send reset signal for hourly task")
	}

	select {
	case <-s.summaryResetCh:
	default:
		t.Error("Should send reset signal for summary task")
	}
}

// TestSummaryResetChannel 测试总结任务重置通道
func TestSummaryResetChannel(t *testing.T) {
	s := &Scheduler{
		summaryResetCh: make(chan struct{}, 1),
		stopCh:         make(chan struct{}),
	}

	// 测试非阻塞发送
	select {
	case s.summaryResetCh <- struct{}{}:
		// 成功发送
	default:
		t.Error("Should be able to send to summary reset channel")
	}

	// 测试接收
	select {
	case <-s.summaryResetCh:
		// 成功接收
	case <-time.After(100 * time.Millisecond):
		t.Error("Should receive from summary reset channel")
	}
}
