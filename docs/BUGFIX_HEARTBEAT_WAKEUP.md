# Bug 修复：系统唤醒检测失效

## 问题描述

在 2026-01-22 发现，当系统从睡眠状态唤醒后，心跳监控机制无法检测到唤醒事件，导致定时任务时间错乱。

### 症状

- 用户在 11 点唤醒电脑后，日志中没有 "=== System wake-up detected! ===" 记录
- 只能看到延迟跳过提醒的日志：`Skipped reminder due to delay (expected: 22:35:00, actual: 11:12:11, delay: 12h37m11s)`
- hourly task 能检测到延迟并跳过，但无法触发完整的唤醒处理流程

## 根本原因

### Go Timer/Ticker 的行为

**关键发现**：Go 的 `time.Ticker` 在系统睡眠时会暂停！

```go
ticker := time.NewTicker(10 * time.Second)
// ...
case now := <-ticker.C:  // ❌ 问题：now 是 ticker 触发的时间
    elapsed := now.Sub(s.lastHeartbeat)
```

**时间线分析**：

1. **睡眠前**（21:50）：
   - ticker 每 10 秒触发一次
   - lastHeartbeat 正常更新

2. **系统睡眠**（21:50 - 11:12）：
   - **ticker 暂停**，不触发任何 tick
   - 墙上时钟继续走（经过了 12+ 小时）

3. **唤醒后**（11:12）：
   - ticker 从暂停处继续
   - 下一次 tick 的 `now` 时间是 11:12:xx
   - 上一次 tick 的 `now` 时间是 21:50:xx
   - 但从 **ticker 的角度看**，两次 tick 之间只有正常的 10 秒间隔
   - `elapsed = 11:12:10 - 11:12:00 ≈ 10s` ✅ 小于 20s 阈值

### 为什么原代码失效

```go
case now := <-ticker.C:  // ticker 给的时间
    // now 在唤醒后的值：
    // - 第一次：11:12:00 (ticker 恢复后的第一次触发)
    // - 第二次：11:12:10 (10秒后)
    // elapsed = 11:12:10 - 11:12:00 = 10s
    // 20s 阈值检测失败！❌
```

ticker 的时间戳不反映墙上时钟的跳变，只反映 ticker 自己的触发间隔。

## 解决方案

### 修复代码

将 `case now := <-ticker.C:` 改为 `case <-ticker.C:` 并手动获取当前时间：

```go
case <-ticker.C:  // 只接收信号，不使用 ticker 的时间
    // 使用当前墙上时间，而不是 ticker 的时间
    // 因为系统睡眠时 ticker 会暂停，唤醒后继续，无法检测到时间跳变
    now := time.Now()  // ✅ 墙上时钟时间

    s.heartbeatMu.Lock()
    elapsed := now.Sub(s.lastHeartbeat)
    s.lastHeartbeat = now
    s.heartbeatMu.Unlock()

    if elapsed > 20*time.Second {
        // 现在可以正确检测到唤醒了！
    }
```

### 为什么修复有效

```go
case <-ticker.C:
    now := time.Now()  // 墙上时钟
    // now 在唤醒后的值：
    // - lastHeartbeat: 21:50:00 (睡眠前的最后一次)
    // - now: 11:12:00 (唤醒后的当前时间)
    // elapsed = 11:12:00 - 21:50:00 = 12h22m ≫ 20s
    // 检测成功！✅
```

使用 `time.Now()` 获取的是真实的墙上时钟，能反映系统睡眠期间经过的时间。

## 修复后的行为

唤醒后应该看到以下日志：

```
2026/01/22 11:12:xx === System wake-up detected! ===
2026/01/22 11:12:xx Last heartbeat: 12h22m ago (threshold: 20s)
2026/01/22 11:12:xx Handling system wake-up (sleep duration: 12h22m)...
2026/01/22 11:12:xx ✓ Hourly task reset signal sent
2026/01/22 11:12:xx ✓ Summary task reset signal sent
2026/01/22 11:12:xx === Wake-up handling completed ===
```

## 测试验证

新增测试文件 `internal/scheduler/heartbeat_test.go`：

```bash
go test -v ./internal/scheduler/ -run TestHeartbeat
```

## 部署步骤

1. 重新构建：
   ```bash
   go build -o daily_summary
   ```

2. 重新安装服务：
   ```bash
   ./scripts/install.sh
   ```

3. 验证修复：
   - 让电脑进入睡眠状态
   - 等待 10 分钟以上
   - 唤醒电脑
   - 查看日志：
     ```bash
     tail -50 run/logs/app.log | grep "wake-up"
     ```

## 相关信息

- **Issue**: 唤醒检测失效导致定时任务错乱
- **修复文件**: `internal/scheduler/scheduler.go:357`
- **测试文件**: `internal/scheduler/heartbeat_test.go`
- **修复日期**: 2026-01-22
- **修复人**: Claude Code

## 技术要点

### Go Timer/Ticker 特性

在 Go 中，`time.Timer` 和 `time.Ticker` 基于单调时钟（monotonic clock），而不是墙上时钟（wall clock）：

- **单调时钟**：只会向前走，不受系统时间调整影响，但会暂停于系统睡眠
- **墙上时钟**：`time.Now()` 返回的时间，反映真实的日期和时间，包括睡眠期间

### 最佳实践

当需要检测时间跳变（如睡眠/唤醒、NTP 调整）时：

✅ **推荐**：
```go
case <-ticker.C:
    now := time.Now()  // 墙上时钟
    elapsed := now.Sub(lastTime)
```

❌ **不推荐**：
```go
case now := <-ticker.C:  // ticker 的时间
    elapsed := now.Sub(lastTime)
```

### 参考

- Go time package: https://pkg.go.dev/time
- Monotonic vs Wall Clock: https://go.dev/blog/timers
