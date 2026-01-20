# 快速开始

## 5分钟快速配置

### 1. 安装程序

```bash
cd /Users/bytedance/go/src/humg.top/daily_summary
./scripts/install.sh
```

### 2. 配置（可选）

如果需要自定义配置，复制示例配置文件并修改：

```bash
# 复制配置文件
cp config.example.yaml ~/.config/daily_summary/config.yaml

# 编辑配置
vim ~/.config/daily_summary/config.yaml
```

### 3. 常用配置调整

#### 修改提醒频率

**小时级调度（默认）：**
```yaml
# 每小时提醒一次（默认）
hourly_interval: 1

# 每2小时提醒一次（减少打断）
hourly_interval: 2
```

**分钟级调度（更灵活）：**
```yaml
# 每30分钟提醒一次
minute_interval: 30

# 每15分钟提醒一次（番茄工作法）
minute_interval: 15

# 每5分钟提醒一次（测试用）
minute_interval: 5
```

注意：设置 `minute_interval` 后会优先使用，`hourly_interval` 会被忽略。

#### 修改总结生成时间

```yaml
# 晚上11点生成今天的总结
summary_time: "23:00"

# 凌晨0点生成昨天的总结（默认）
summary_time: "00:00"

# 早上9点生成昨天的总结
summary_time: "09:00"
```

#### 修改对话框超时

```yaml
# 10分钟超时
dialog_timeout: 600

# 5分钟超时（默认）
dialog_timeout: 300
```

### 4. 重启服务

修改配置后需要重启：

```bash
./scripts/install.sh
```

或者手动重启：

```bash
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist
launchctl load ~/Library/LaunchAgents/com.humg.daily_summary.plist
```

## 常用命令

### 查看服务状态
```bash
launchctl list | grep daily_summary
```

### 查看日志
```bash
# 程序日志
tail -f ~/daily_summary/logs/app.log

# 标准输出
tail -f ~/daily_summary/logs/stdout.log
```

### 查看今天的工作记录
```bash
cat ~/daily_summary/data/$(date +%Y-%m-%d).json
```

### 查看昨天的工作总结
```bash
cat ~/daily_summary/summaries/$(date -v-1d +%Y-%m-%d).md
```

### 卸载程序
```bash
./scripts/uninstall.sh
```

## 配置示例场景

### 场景1：标准上班族
朝九晚六，每小时记录，凌晨生成总结

```yaml
hourly_interval: 1
summary_time: "00:00"
dialog_timeout: 300
```

### 场景2：专注工作者
需要长时间专注，减少打断

```yaml
hourly_interval: 3
summary_time: "23:00"
dialog_timeout: 600
```

### 场景3：自由职业者
灵活工作时间，晚上回顾

```yaml
hourly_interval: 2
summary_time: "22:00"
dialog_timeout: 300
```

### 场景4：番茄工作法
每25分钟记录一次（配合番茄钟）

```yaml
minute_interval: 25
summary_time: "23:00"
dialog_timeout: 180
```

### 场景5：高频记录
每30分钟记录一次，适合详细追踪

```yaml
minute_interval: 30
summary_time: "00:00"
dialog_timeout: 300
```

### 场景6：测试配置
快速测试，每5分钟提醒

```yaml
minute_interval: 5
summary_time: "23:59"
dialog_timeout: 60
```

## 故障排除

### 对话框没弹出？
1. 检查服务是否运行：`launchctl list | grep daily_summary`
2. 查看日志：`tail -f ~/daily_summary/logs/app.log`
3. 确认下次弹窗时间（程序会在整点弹窗）

### 总结没生成？
1. 检查是否有工作记录：`cat ~/daily_summary/data/$(date -v-1d +%Y-%m-%d).json`
2. 查看错误日志：`tail -f ~/daily_summary/logs/stderr.log`
3. 确认 Claude Code 是否可用：`which claude-code`

### 配置文件无效？
1. 检查 YAML 语法：https://www.yamllint.com/
2. 确认文件路径：`~/.config/daily_summary/config.yaml`
3. 重启服务使配置生效

## 更多信息

- 完整配置说明：查看 `docs/CONFIGURATION.md`
- 项目架构：查看 `CLAUDE.md`
- 使用手册：查看 `README.md`
