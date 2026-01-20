# 分钟级调度功能

Daily Summary 工具现在支持分钟级的提醒调度，让你可以更灵活地控制工作记录的频率。

## 功能概述

除了原有的小时级调度（`hourly_interval`），现在新增了分钟级调度（`minute_interval`）：

- **小时级调度**：每 N 小时提醒一次（如 1小时、2小时、3小时）
- **分钟级调度**：每 N 分钟提醒一次（如 15分钟、30分钟、45分钟）

## 使用方法

### 启用分钟级调度

在配置文件中添加 `minute_interval` 参数：

```yaml
# 启用分钟级调度
minute_interval: 30  # 每30分钟提醒一次
```

### 优先级

如果同时设置了 `hourly_interval` 和 `minute_interval`：

```yaml
hourly_interval: 1    # 会被忽略
minute_interval: 30   # 优先使用
```

程序会优先使用 `minute_interval`，`hourly_interval` 将被忽略。

### 禁用分钟级调度

如果要使用小时级调度，只需：

1. 删除或注释掉 `minute_interval`：
```yaml
# minute_interval: 30
hourly_interval: 1
```

2. 或者不在配置文件中添加 `minute_interval`

## 常用配置

### 番茄工作法（25分钟）
```yaml
minute_interval: 25
summary_time: "23:00"
```

### 高频记录（每30分钟）
```yaml
minute_interval: 30
summary_time: "00:00"
```

### 每15分钟（适合短时间任务）
```yaml
minute_interval: 15
summary_time: "18:00"
```

### 每45分钟
```yaml
minute_interval: 45
summary_time: "00:00"
```

### 测试用（每5分钟）
```yaml
minute_interval: 5
summary_time: "23:59"
dialog_timeout: 60
```

## 快速测试

项目提供了预配置的测试文件：

### 1. 5分钟测试配置
```bash
./daily_summary --config config.test-5min.yaml
```

这将每5分钟弹窗一次，方便快速测试功能。

### 2. 番茄工作法配置
```bash
./daily_summary --config config.pomodoro.yaml
```

每25分钟提醒一次，配合番茄工作法使用。

### 3. 运行测试脚本
```bash
./scripts/test_minute_interval.sh
```

自动创建测试配置并提示如何测试。

## 工作原理

### 小时级调度
- 在整点触发（如 9:00, 10:00, 11:00）
- 使用 `time.Hour` 作为间隔单位
- 适合粗粒度的工作记录

### 分钟级调度
- 在分钟边界触发（如 9:05, 9:10, 9:15）
- 使用 `time.Minute` 作为间隔单位
- 适合细粒度的工作记录

### 日志输出

程序启动时会在日志中显示使用的调度模式：

**小时级调度：**
```
Using hour-based scheduling: every 1 hour(s)
Next reminder at 10:00:00
```

**分钟级调度：**
```
Using minute-based scheduling: every 30 minute(s)
Next reminder at 09:30:00
```

## 使用场景

### 场景1：番茄工作法
每25分钟工作，5分钟休息。配合番茄钟使用：

```yaml
minute_interval: 25
```

### 场景2：敏捷开发
短周期迭代，每30分钟记录一次：

```yaml
minute_interval: 30
```

### 场景3：详细追踪
需要详细记录每个时间段的工作：

```yaml
minute_interval: 15
```

### 场景4：会议密集
会议较多，每45分钟记录一次：

```yaml
minute_interval: 45
```

### 场景5：功能测试
快速验证程序功能：

```yaml
minute_interval: 5
dialog_timeout: 60
```

## 注意事项

### 1. 避免过短的间隔
不建议设置过小的值（如 1-2 分钟），可能会：
- 频繁打断工作
- 影响专注度
- 增加记录负担

**建议最小值：5 分钟**（仅用于测试）

### 2. 合理的间隔值
推荐的分钟间隔：
- **5 分钟**：测试用
- **10 分钟**：非常详细的记录
- **15 分钟**：详细记录
- **20-30 分钟**：平衡的记录频率
- **45 分钟**：较粗粒度

### 3. 电池消耗
分钟级调度比小时级调度更频繁，在笔记本上使用时可能会稍微增加电池消耗。

### 4. 配置更新
修改配置后需要重启服务才能生效：
```bash
./scripts/install.sh
```

## 故障排除

### 对话框没有按分钟级弹出

1. **检查配置文件**
```bash
cat ~/.config/daily_summary/config.yaml
```

确认 `minute_interval` 已设置且大于 0。

2. **查看日志确认调度模式**
```bash
tail -f ~/daily_summary/logs/app.log
```

应该看到类似：
```
Using minute-based scheduling: every 30 minute(s)
```

3. **重启服务**
```bash
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist
launchctl load ~/Library/LaunchAgents/com.humg.daily_summary.plist
```

### 仍然使用小时级调度

检查是否正确设置了 `minute_interval`：
```yaml
# ✓ 正确
minute_interval: 30

# ✗ 错误（注释掉了）
# minute_interval: 30

# ✗ 错误（设置为 0）
minute_interval: 0
```

### 触发时间不准确

分钟级调度会对齐到分钟边界：
- 设置 `minute_interval: 15`
- 可能的触发时间：00:00, 00:15, 00:30, 00:45, 01:00, ...

如果当前时间是 10:08，下次触发将在 10:15，而不是 10:23。

## 示例配置文件

### 完整的番茄工作法配置
```yaml
# 番茄工作法配置
data_dir: ~/daily_summary/data
summary_dir: ~/daily_summary/summaries

# 每25分钟记录
minute_interval: 25

# 晚上11点生成总结
summary_time: "23:00"

# Claude Code 路径
claude_code_path: claude-code

# 3分钟超时（短一点）
dialog_timeout: 180

# 启用日志
enable_logging: true
```

### 测试配置
```yaml
# 快速测试配置
data_dir: ~/daily_summary_test/data
summary_dir: ~/daily_summary_test/summaries

# 每5分钟触发
minute_interval: 5

# 测试总结生成
summary_time: "23:59"

claude_code_path: claude-code

# 1分钟超时
dialog_timeout: 60

enable_logging: true
```

## 与小时级调度的对比

| 特性 | 小时级调度 | 分钟级调度 |
|------|-----------|-----------|
| 配置参数 | `hourly_interval` | `minute_interval` |
| 最小间隔 | 1 小时 | 1 分钟 |
| 触发时间 | 整点（9:00, 10:00） | 分钟边界（9:05, 9:10） |
| 使用场景 | 粗粒度记录 | 细粒度记录 |
| 适合人群 | 一般用户 | 番茄钟用户、敏捷团队 |
| 优先级 | 较低 | 较高（覆盖小时级） |

## 更多信息

- 完整配置说明：[CONFIGURATION.md](CONFIGURATION.md)
- 快速开始：[QUICK_START.md](QUICK_START.md)
- 项目文档：[README.md](../README.md)
