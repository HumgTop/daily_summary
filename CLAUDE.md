# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Daily Summary 是一个用 Go 编写的桌面工具，用于自动化工作记录和总结：

**核心功能：**
- 每小时弹窗提醒用户记录过去1小时的工作内容
- 收集并存储用户输入的工作记录
- 每天0点自动调用 Claude Code 生成前一天的工作总结

**技术栈：**
- Go 语言开发
- 编译为单一二进制可执行文件
- 需要桌面通知/弹窗功能
- 定时任务调度
- 与 Claude Code CLI 集成

## Development Commands

### Build
```bash
go build -o daily_summary
```

### Run
```bash
./daily_summary
```

### Test
```bash
go test ./...
```

### Run specific test
```bash
go test -run TestFunctionName ./path/to/package
```

### Install dependencies
```bash
go mod download
```

### Tidy dependencies
```bash
go mod tidy
```

## Architecture

### 主要组件

1. **定时器模块（Timer/Scheduler）**
   - 每小时触发弹窗提醒
   - 每天0点触发总结生成
   - 使用 Go 的 time.Ticker 或 cron 库

2. **弹窗模块（Notification/Dialog）**
   - 跨平台桌面通知功能
   - 文本输入界面
   - 可能使用的库：
     - Linux: 使用 notify-send 或 zenity
     - macOS: 使用 osascript 或 terminal-notifier
     - Windows: 使用 PowerShell 或 Windows Toast

3. **数据存储模块（Storage）**
   - 存储每小时的工作记录
   - 可能使用 JSON 文件或 SQLite 数据库
   - 按日期组织数据便于生成总结

4. **总结生成模块（Summary Generator）**
   - 读取前一天的所有工作记录
   - 调用 Claude Code CLI 生成总结
   - 保存生成的总结（markdown 格式）

### Claude Code 集成

当在每天0点生成总结时，程序会：

1. 收集前一天的所有工作记录条目
2. 调用 Claude Code CLI，传入工作记录数据
3. Claude Code 应该：
   - 分析所有工作记录
   - 按项目/任务分类整理
   - 识别重要进展和完成的工作
   - 生成结构化的工作总结（markdown 格式）
   - 突出关键成果和待办事项

### 数据结构建议

```go
// 工作记录条目
type WorkEntry struct {
    Timestamp time.Time
    Content   string
}

// 每日总结
type DailySummary struct {
    Date    time.Time
    Entries []WorkEntry
    Summary string  // Claude Code 生成的总结
}
```

### 文件组织

实际项目结构：
```
.
├── main.go                    # 程序入口
├── go.mod                     # Go 模块配置
├── config.example.yaml        # 配置文件示例
├── config/                    # 配置管理模块
├── internal/
│   ├── models/               # 数据模型
│   ├── scheduler/            # 定时任务调度
│   ├── dialog/               # 弹窗对话框（macOS osascript）
│   ├── storage/              # JSON 数据存储
│   └── summary/              # 总结生成（调用 Claude Code）
├── scripts/                   # 安装/卸载脚本
└── launchd/                  # macOS 后台服务配置
```

### 配置文件

配置文件支持 YAML 和 JSON 两种格式（根据扩展名自动识别）：
- 默认路径：`~/.config/daily_summary/config.yaml`
- 示例文件：`config.example.yaml`

**关键配置项：**
- `hourly_interval`：每隔 N 小时弹窗提醒（默认 1）
- `summary_time`：每日生成总结的时间，格式 "HH:MM"（默认 "00:00"）
- `dialog_timeout`：对话框超时时间（秒，默认 300）
- `claude_code_path`：Claude Code CLI 路径（默认 "claude-code"）

配置文件会在程序启动时自动加载，如果不存在则使用默认配置。

## Claude Code Usage

当 Claude Code 被调用生成工作总结时，会收到如下格式的数据：

```
日期：2026-01-19
工作记录：

09:00 - 完成了用户认证模块的代码审查
10:00 - 修复了3个前端 bug
11:00 - 参加团队站会
...
18:00 - 编写了新功能的单元测试
```

期望输出一份结构化的工作总结，包括：
- 主要完成的任务
- 关键进展
- 遇到的问题
- 明天的计划（如果记录中有提及） 