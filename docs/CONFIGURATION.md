# 配置指南

Daily Summary 工具支持灵活的配置，可以根据个人习惯调整工作记录的频率和总结生成的时间。

## 配置文件格式

支持两种配置文件格式：
- **YAML** (推荐): `config.yaml`
- **JSON**: `config.json`

程序会根据文件扩展名自动识别格式。

## 配置文件位置

默认配置文件路径：`~/.config/daily_summary/config.yaml`

可以通过命令行参数指定自定义配置文件：
```bash
./daily_summary --config /path/to/custom/config.yaml
```

## 完整配置示例

### YAML 格式（推荐）

```yaml
# 数据存储目录（工作记录 JSON 文件）
data_dir: ~/daily_summary/data

# 工作总结目录（Markdown 总结文件）
summary_dir: ~/daily_summary/summaries

# 每小时提醒间隔（单位：小时）
hourly_interval: 1

# 每日总结生成时间（24小时制，格式：HH:MM）
summary_time: "00:00"

# Claude Code CLI 可执行文件路径
claude_code_path: claude-code

# 对话框超时时间（单位：秒）
dialog_timeout: 300

# 是否启用日志记录
enable_logging: true
```

### JSON 格式

```json
{
  "data_dir": "~/daily_summary/data",
  "summary_dir": "~/daily_summary/summaries",
  "hourly_interval": 1,
  "summary_time": "00:00",
  "claude_code_path": "claude-code",
  "dialog_timeout": 300,
  "enable_logging": true
}
```

## 配置项详解

### data_dir (必填)
**工作记录的存储目录**

- 类型：字符串
- 默认值：`~/daily_summary/data`
- 说明：所有工作记录的 JSON 文件都会保存在这个目录下，文件名格式为 `YYYY-MM-DD.json`

示例：
```yaml
data_dir: ~/work_logs/data
```

### summary_dir (必填)
**工作总结的保存目录**

- 类型：字符串
- 默认值：`~/daily_summary/summaries`
- 说明：每日生成的工作总结 Markdown 文件保存位置，文件名格式为 `YYYY-MM-DD.md`

示例：
```yaml
summary_dir: ~/Documents/work_summaries
```

### hourly_interval (必填)
**弹窗提醒的时间间隔（小时级）**

- 类型：整数
- 默认值：`1`（每小时）
- 单位：小时
- 说明：控制弹窗提醒的频率（小时级调度）
- 注意：如果设置了 `minute_interval`，则此参数会被忽略

常用配置：
```yaml
# 每小时提醒一次（默认，适合详细记录）
hourly_interval: 1

# 每2小时提醒一次（适合不想频繁打断的场景）
hourly_interval: 2

# 每3小时提醒一次
hourly_interval: 3

# 每4小时提醒一次（适合粗粒度记录）
hourly_interval: 4
```

### minute_interval (可选)
**弹窗提醒的时间间隔（分钟级）**

- 类型：整数
- 默认值：未设置（使用 `hourly_interval`）
- 单位：分钟
- 说明：控制弹窗提醒的频率（分钟级调度）
- **优先级**：如果设置了此参数，将优先使用分钟级调度，`hourly_interval` 会被忽略

常用配置：
```yaml
# 每30分钟提醒一次（适合高频记录）
minute_interval: 30

# 每15分钟提醒一次（适合番茄工作法）
minute_interval: 15

# 每45分钟提醒一次
minute_interval: 45

# 每5分钟提醒一次（适合测试）
minute_interval: 5

# 每10分钟提醒一次
minute_interval: 10
```

**使用场景示例：**

**番茄工作法（25分钟工作 + 5分钟休息）**
```yaml
minute_interval: 25
```

**短间隔记录**
```yaml
minute_interval: 20  # 每20分钟记录一次
```

**测试配置**
```yaml
minute_interval: 5   # 每5分钟弹窗，用于测试
```

**注意事项：**
- 不要设置过小的值（如1分钟），可能会影响工作效率
- 建议最小值为 5 分钟
- 如果不需要分钟级调度，请注释掉或删除此配置项

### summary_time (必填)
**每日总结生成的时间点**

- 类型：字符串
- 默认值：`"00:00"`（凌晨0点）
- 格式：`"HH:MM"`（24小时制）
- 说明：每天在指定时间自动生成前一天的工作总结

常用场景：

**场景1：凌晨生成前一天总结（默认）**
```yaml
summary_time: "00:00"
```
适合：希望第二天早上就能看到昨天的总结

**场景2：工作日结束时生成当天总结**
```yaml
summary_time: "18:00"
```
适合：下班前查看今天的工作总结

**场景3：睡前生成今天的总结**
```yaml
summary_time: "23:00"
```
适合：晚上睡觉前回顾今天的工作

**场景4：早晨生成昨天的总结**
```yaml
summary_time: "09:00"
```
适合：早上上班时回顾昨天的工作

### claude_code_path (必填)
**Claude Code CLI 的路径**

- 类型：字符串
- 默认值：`claude-code`
- 说明：Claude Code 命令行工具的路径或命令名

如果 `claude-code` 在系统 PATH 中：
```yaml
claude_code_path: claude-code
```

如果需要指定完整路径：
```yaml
claude_code_path: /usr/local/bin/claude-code
```

如果使用自定义安装位置：
```yaml
claude_code_path: /opt/tools/claude-code
```

### dialog_timeout (可选)
**对话框的超时时间**

- 类型：整数
- 默认值：`300`（5分钟）
- 单位：秒
- 说明：用户未在指定时间内响应对话框时，对话框会自动关闭

常用配置：
```yaml
# 5分钟超时（默认）
dialog_timeout: 300

# 10分钟超时（适合可能暂时离开的场景）
dialog_timeout: 600

# 2分钟超时（适合快速记录）
dialog_timeout: 120

# 30分钟超时（适合会议等长时间场景）
dialog_timeout: 1800
```

### enable_logging (可选)
**是否启用日志文件记录**

- 类型：布尔值
- 默认值：`true`
- 说明：控制是否将日志写入文件

```yaml
# 启用日志（会写入 ~/daily_summary/logs/app.log）
enable_logging: true

# 禁用日志文件（日志只输出到 stdout）
enable_logging: false
```

## 使用场景示例

### 场景1：标准办公场景
每小时记录一次，凌晨生成总结

```yaml
data_dir: ~/daily_summary/data
summary_dir: ~/daily_summary/summaries
hourly_interval: 1
summary_time: "00:00"
claude_code_path: claude-code
dialog_timeout: 300
enable_logging: true
```

### 场景2：长时间专注工作
每3小时记录一次，晚上11点生成总结

```yaml
data_dir: ~/work_logs/data
summary_dir: ~/work_logs/summaries
hourly_interval: 3
summary_time: "23:00"
claude_code_path: claude-code
dialog_timeout: 600
enable_logging: true
```

### 场景3：灵活工作时间
每2小时记录一次，下午6点生成总结

```yaml
data_dir: ~/daily_summary/data
summary_dir: ~/daily_summary/summaries
hourly_interval: 2
summary_time: "18:00"
claude_code_path: /opt/claude/claude-code
dialog_timeout: 300
enable_logging: true
```

### 场景4：简化记录
每4小时记录一次，早上9点生成总结

```yaml
data_dir: ~/work_notes/data
summary_dir: ~/work_notes/summaries
hourly_interval: 4
summary_time: "09:00"
claude_code_path: claude-code
dialog_timeout: 180
enable_logging: false
```

## 配置更新

修改配置文件后，需要重启程序才能生效：

```bash
# 重启 launchd 服务
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist
launchctl load ~/Library/LaunchAgents/com.humg.daily_summary.plist
```

或者使用安装脚本重新安装：
```bash
./scripts/install.sh
```

## 验证配置

创建配置文件后，可以手动运行程序验证配置是否正确：

```bash
./daily_summary --config ~/.config/daily_summary/config.yaml
```

程序启动后会在日志中显示加载的配置信息。

## 故障排除

### 配置文件未加载
- 检查配置文件路径是否正确
- 检查 YAML/JSON 格式是否有语法错误
- 查看日志文件：`tail -f ~/daily_summary/logs/app.log`

### YAML 语法错误
常见错误：
- 字符串包含特殊字符时未加引号
- 缩进使用了 Tab 而不是空格
- 布尔值使用了引号（应该是 `true` 而不是 `"true"`）

使用在线 YAML 验证器检查语法：https://www.yamllint.com/

### 时间格式错误
确保使用 24 小时制，格式为 `"HH:MM"`，例如：
- ✅ `"00:00"`, `"09:30"`, `"18:00"`, `"23:59"`
- ❌ `"0:0"`, `"9:30am"`, `"6:00 PM"`
