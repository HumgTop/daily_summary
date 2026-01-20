# Daily Summary Tool

一个用 Go 编写的自动化工作记录和总结工具，专为 macOS 设计。

## 功能特性

- **每小时自动提醒**：每小时整点弹出原生 macOS 对话框，提示记录过去1小时的工作内容
- **智能定时器重置** ⭐：手动添加记录后自动顺延提醒时间，避免重复提醒
- **手动记录**：随时通过 `add` 命令快速记录工作内容，无需等待定时弹窗
- **数据持久化**：工作记录自动保存为 JSON 文件，按日期组织
- **智能总结**：每天 0 点自动调用 Claude Code CLI 生成前一天的工作总结
- **后台运行**：使用 launchd 在后台持续运行，系统启动时自动启动
- **简单易用**：原生 macOS 对话框，无需切换应用

## 系统要求

- macOS 操作系统
- Go 1.19 或更高版本
- Claude Code CLI（可选，用于生成 AI 总结）

## 安装

1. 克隆或下载此项目到本地

2. 运行安装脚本：
```bash
./scripts/install.sh
```

安装脚本会：
- 编译程序
- 创建必要的目录
- 安装 launchd 服务
- 启动后台服务

## 使用

安装完成后，程序会自动在后台运行：

### 自动记录（定时弹窗）
1. **记录工作**：每小时整点会弹出对话框，输入过去1小时的工作内容后点击"确定"
2. **查看记录**：工作记录保存在 `./run/data/YYYY-MM-DD.json`
3. **查看总结**：每日总结保存在 `./run/summaries/YYYY-MM-DD.md`

### 手动记录（命令行）

**添加工作记录**：
```bash
daily_summary add "完成需求文档审查"
daily_summary add "参加技术评审会议"
```

**查看今日记录**：
```bash
daily_summary list
```

**智能定时器重置**：
执行 `add` 命令后，下一次提醒时间会自动顺延一个完整周期。例如：
```
10:00 - 服务启动，下一次提醒: 11:00
10:30 - 执行 add 命令添加记录
10:30 - 自动重置，下一次提醒: 11:30 ⏰ (避免在 11:00 重复提醒)
```

**获取帮助**：
```bash
daily_summary help
```

## 目录结构

```
项目目录/
├── run/                     # 运行时数据目录
│   ├── data/               # 工作记录（JSON 格式）
│   │   ├── 2026-01-19.json
│   │   └── 2026-01-20.json
│   ├── summaries/          # 工作总结（Markdown 格式）
│   │   └── 2026-01-19.md
│   ├── logs/               # 程序日志
│   │   ├── app.log
│   │   ├── stdout.log
│   │   └── stderr.log
│   ├── .reset_signal       # 重置信号文件（临时）
│   └── daily_summary.lock  # 进程锁文件
└── ~/.config/daily_summary/
    └── config.yaml         # 配置文件
```

## 配置

配置文件位于 `~/.config/daily_summary/config.yaml`（也支持 `.json` 格式）。

项目提供了配置文件示例 `config.example.yaml`，复制并修改即可：
```bash
cp config.example.yaml ~/.config/daily_summary/config.yaml
```

默认配置：
```yaml
# 数据存储目录（项目目录下的 run/data）
data_dir: ./run/data

# 工作总结目录（项目目录下的 run/summaries）
summary_dir: ./run/summaries

# 提醒间隔（支持小时或分钟级）
hourly_interval: 1        # 每小时提醒一次
# minute_interval: 30     # 或者每30分钟提醒一次（优先级更高）

# 每日总结生成时间（24小时制）
summary_time: "00:00"

# Claude Code CLI 路径
claude_code_path: claude-code

# 对话框超时时间（单位：秒）
dialog_timeout: 300

# 是否启用日志
enable_logging: true
```

### 配置项说明

- `data_dir`：工作记录的存储目录
- `summary_dir`：工作总结的保存目录
- `hourly_interval`：弹窗提醒间隔（小时），默认 1 小时
  - `1` = 每小时提醒一次
  - `2` = 每2小时提醒一次
- `minute_interval`：弹窗提醒间隔（分钟），**如果设置则优先使用**
  - `15` = 每15分钟提醒一次
  - `30` = 每30分钟提醒一次
  - `5` = 每5分钟提醒一次（适合测试）
  - 注意：设置后会覆盖 `hourly_interval`
- `summary_time`：每日生成总结的时间（HH:MM 格式）
  - `"00:00"` = 凌晨0点生成前一天的总结
  - `"23:00"` = 晚上11点生成当天的总结
- `claude_code_path`：Claude Code CLI 的路径或命令名
- `dialog_timeout`：对话框超时时间（秒），默认 300 秒（5分钟）
- `enable_logging`：是否将日志写入文件

## 常用命令

### 查看服务状态
```bash
launchctl list | grep daily_summary
```

### 查看日志
```bash
# 程序日志
tail -f ./run/logs/app.log

# 标准输出
tail -f ./run/logs/stdout.log

# 错误输出
tail -f ./run/logs/stderr.log
```

### 重启服务
```bash
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist
launchctl load ~/Library/LaunchAgents/com.humg.daily_summary.plist
```

### 手动运行（用于测试）
```bash
./daily_summary
```

## 卸载

运行卸载脚本：
```bash
./scripts/uninstall.sh
```

注意：卸载脚本只会移除 launchd 服务，数据文件会保留在项目的 `run/` 目录中。

如果要完全删除所有数据：
```bash
rm -rf ./run
rm -rf ~/.config/daily_summary
```

## Claude Code 集成

程序默认会尝试调用 `claude-code` CLI 生成工作总结。如果 Claude Code 未安装或不可用，程序会使用简单的模板生成总结。

要获得完整的 AI 生成总结功能，请确保：
1. 安装 Claude Code CLI
2. 在配置文件中设置正确的 `claude_code_path`

## 故障排除

### 对话框没有弹出
- 检查服务是否正在运行：`launchctl list | grep daily_summary`
- 查看日志文件：`tail -f ./run/logs/app.log`
- 确认下一次弹窗时间是否正确

### 总结生成失败
- 检查 Claude Code CLI 是否正确安装
- 查看错误日志：`tail -f ./run/logs/stderr.log`
- 确认前一天是否有工作记录

### 权限问题
- 确保安装脚本有执行权限：`chmod +x scripts/*.sh`
- 检查 macOS 是否授予了必要的权限

## 开发

### 编译
```bash
go build -o daily_summary
```

### 运行测试
```bash
go test ./...
```

## 许可证

本项目基于 MIT 许可证开源。

## 贡献

欢迎提交 Issue 和 Pull Request！
