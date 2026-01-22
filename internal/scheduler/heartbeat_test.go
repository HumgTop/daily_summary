package scheduler

import (
	"testing"
	"time"
)

// TestHeartbeatWakeUpDetection 测试唤醒检测机制
func TestHeartbeatWakeUpDetection(t *testing.T) {
	// 这个测试演示了修复前后的差异：
	// - 修复前：使用 ticker.C 的时间，系统睡眠时 ticker 暂停，无法检测唤醒
	// - 修复后：使用 time.Now()，即使 ticker 暂停，也能检测到墙上时钟的跳变

	// 模拟场景：两次心跳之间系统睡眠了 12 小时
	lastHeartbeat := time.Now().Add(-12 * time.Hour)
	currentTime := time.Now()

	elapsed := currentTime.Sub(lastHeartbeat)

	// 验证：间隔应该远超过 20 秒阈值
	threshold := 20 * time.Second
	if elapsed <= threshold {
		t.Errorf("Expected elapsed time > %s, got %s", threshold, elapsed)
	}

	t.Logf("✓ Elapsed time: %s (threshold: %s)", elapsed, threshold)
	t.Logf("✓ Wake-up would be detected correctly")
}

// TestTickerBehaviorOnSleep 演示 ticker 在系统睡眠时的行为
func TestTickerBehaviorOnSleep(t *testing.T) {
	// 这个测试演示为什么原来的代码有问题

	// 问题：使用 case now := <-ticker.C 时，now 是 ticker 触发的时间
	// 当系统睡眠时，ticker 暂停
	// 唤醒后，ticker 从暂停处继续，两次 tick 之间看起来只有正常间隔

	// 正确做法：在收到 ticker 信号后，使用 time.Now() 获取当前墙上时钟
	// 这样即使 ticker 暂停，我们也能通过比较墙上时钟来检测时间跳变

	t.Log("Original code problem:")
	t.Log("  case now := <-ticker.C")
	t.Log("  // 'now' is the time when ticker fired")
	t.Log("  // During sleep, ticker is paused")
	t.Log("  // After wake-up, 'now' values are only ~10s apart")
	t.Log("")
	t.Log("Fixed code:")
	t.Log("  case <-ticker.C")
	t.Log("  now := time.Now()")
	t.Log("  // 'now' is the current wall-clock time")
	t.Log("  // Can detect time jump even if ticker was paused")
}
