# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Daily Summary 是一个用 Go 编写的 macOS 桌面工具，用于自动化工作记录和 AI 总结生成：

**核心功能：**
- **定时提醒**：每小时（或自定义分钟间隔）弹窗提醒用户记录工作内容
- **手动记录**：通过 CLI 命令随时添加工作记录
- **智能重置**：手动添加记录后自动顺延下次提醒时间，避免重复
- **数据持久化**：JSON 格式存储工作记录，按日期组织
- **AI 总结**：每天指定时间自动调用 AI 生成前一天的工作总结
- **多 AI 支持**：支持 Codex (默认) 和 Claude Code 两种 AI 提供商
- **后台服务**：使用 launchd 持续运行，系统启动时自动启动
- **进程管理**：通过文件锁防止重复启动

**技术栈：**
- Go 1.19+ 开发
- 编译为单一二进制可执行文件
- macOS osascript 实现原生对话框
- 定时任务调度（支持小时/分钟级）
- 信号文件机制实现进程间通信
- 多 AI 提供商架构（Codex/Claude Code）

## Development Guide

### Building

```bash
# 编译二进制文件
go build -o daily_summary

# 交叉编译（如需要）
GOOS=darwin GOARCH=amd64 go build -o daily_summary
GOOS=darwin GOARCH=arm64 go build -o daily_summary
```

### Testing

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/scheduler/

# 运行特定测试函数
go test -run TestFunctionName ./internal/scheduler/

# 显示详细输出
go test -v ./...

# 测试覆盖率
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 依赖管理

```bash
# 下载依赖
go mod download

# 清理未使用的依赖
go mod tidy

# 查看依赖图
go mod graph

# 更新依赖
go get -u ./...
```

### 快速测试脚本

项目提供了多个测试脚本方便开发：

```bash
# 快速功能测试（分钟级调度）
./scripts/quick_test.sh

# 测试配置加载
./scripts/test_config.sh

# 测试对话框
./scripts/test_dialog.sh

# 测试分钟级调度
./scripts/test_minute_interval.sh
```

### 开发调试

**本地运行**：
```bash
# 直接运行（使用默认配置）
go run main.go

# 使用自定义配置
go run main.go --config ./config.yaml

# 测试特定命令
go run main.go add "测试记录"
go run main.go list
go run main.go summary --date 2026-01-20
```

**查看日志**：
```bash
# 实时查看应用日志
tail -f ./run/logs/app.log

# 查看 launchd 输出
tail -f ./run/logs/stdout.log
tail -f ./run/logs/stderr.log
```

**调试 launchd 服务**：
```bash
# 查看服务状态
launchctl list | grep daily_summary

# 查看详细信息
launchctl list com.humg.daily_summary

# 停止服务
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist

# 启动服务
launchctl load ~/Library/LaunchAgents/com.humg.daily_summary.plist

# 重启服务
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist && \
launchctl load ~/Library/LaunchAgents/com.humg.daily_summary.plist
```

### 代码规范

**接口设计**：
- 优先定义接口，便于测试和扩展
- 接口命名使用动词或能力描述（如 `Dialog`, `Storage`, `AIClient`）
- 保持接口精简，单一职责

**错误处理**：
- 使用 `fmt.Errorf` 包装错误，添加上下文信息
- 日志记录错误详情，返回给调用者简洁的错误信息
- 区分致命错误和非致命错误

**日志规范**：
- 使用 `log.Printf` 记录重要操作和状态
- 错误日志包含足够的上下文信息
- 避免在循环中大量打印日志

**配置优先级**：
1. 命令行参数
2. 配置文件
3. 默认值

### 添加新 AI 提供商

如需添加新的 AI 提供商支持：

1. 在 `internal/summary/` 创建新的客户端文件（如 `openai.go`）
2. 实现 `AIClient` 接口：
   ```go
   type NewAIClient struct {
       // 配置字段
   }

   func (c *NewAIClient) GenerateSummary(prompt string) (string, error) {
       // 实现总结生成逻辑
   }
   ```
3. 在 `main.go` 中添加新提供商的初始化逻辑
4. 更新 `models.Config` 添加相关配置字段
5. 更新 `config.example.yaml` 和文档

### 平台扩展

当前仅支持 macOS，如需扩展到其他平台：

**Linux 支持**：
- 对话框：使用 `zenity` 或 `kdialog`
- 通知：使用 `notify-send`
- 在 `internal/dialog/` 添加 Linux 实现

**Windows 支持**：
- 对话框：使用 PowerShell 或 WinForms
- 通知：使用 Windows Toast
- 在 `internal/dialog/` 添加 Windows 实现

**跨平台抽象**：
- 保持 `Dialog` 接口不变
- 使用构建标签（build tags）区分平台实现
- 编译时自动选择对应平台的实现

## Troubleshooting

### 常见问题

**1. 服务无法启动**

检查项：
- 是否已有实例在运行：`launchctl list | grep daily_summary`
- 查看进程锁文件：`cat run/daily_summary.lock`
- 检查日志：`tail -f run/logs/stderr.log`

解决方案：
```bash
# 强制停止旧实例
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist
rm run/daily_summary.lock

# 重新启动
./scripts/install.sh
```

**2. 对话框不弹出**

检查项：
- 服务是否运行：`ps aux | grep daily_summary`
- 查看下次提醒时间：`tail -20 run/logs/app.log | grep "Next reminder"`
- 确认时间对齐是否正确

调试：
```bash
# 手动触发对话框测试
./scripts/test_dialog.sh

# 使用分钟级调度快速测试
cp config.pomodoro.yaml config.yaml
./scripts/quick_test.sh
```

**3. 总结生成失败**

检查项：
- AI CLI 是否安装：`which codex` 或 `which claude-code`
- 是否有工作记录：`daily_summary list`
- 配置是否正确：`cat config.yaml`

解决方案：
```bash
# 手动测试总结生成
daily_summary summary --date 2026-01-20

# 使用回退机制（即使 AI 不可用也能生成简单总结）
# 程序会自动检测并使用回退机制
```

**4. 时间对齐问题**

现象：提醒时间不在整点或预期的分钟边界

原因：
- 服务启动时间不在边界上
- 系统时间跳变（NTP 同步、手动调整）

解决方案：
```bash
# 重启服务，重新对齐
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist
launchctl load ~/Library/LaunchAgents/com.humg.daily_summary.plist

# 查看日志确认对齐
tail -f run/logs/app.log
```

**5. 日志文件过大**

配置 `max_log_size_mb` 启用日志轮转：

```yaml
enable_logging: true
max_log_size_mb: 10  # 超过 10MB 自动轮转
```

手动清理：
```bash
# 删除旧日志
rm run/logs/app.log.old

# 清空当前日志
: > run/logs/app.log
```

### 重要注意事项
工作过程产出的文档，统一放在 docs/ 目录下

**1. 配置文件路径**

优先级顺序：
1. `--config` 参数指定的路径
2. 当前目录的 `config.yaml`
3. `~/.config/daily_summary/config.yaml`
4. 默认配置（硬编码）

**2. 工作目录**

launchd 服务的工作目录由 plist 中的 `WorkingDirectory` 指定，必须是项目根目录的绝对路径。

相对路径（如 `./run/data`）会基于这个目录解析。

**3. 时区处理**

所有时间都使用系统本地时区，无时区转换逻辑。

**4. 并发安全**

- Storage 操作是文件级的，无锁保护
- 同一时间只应有一个服务实例运行（通过进程锁保证）
- 手动命令（add/list/summary）与服务实例可能有竞态，但影响很小

**5. 数据持久化**

- JSON 文件按日期分文件，易于备份和迁移
- 没有数据库，没有事务保证
- 文件写入失败会丢失数据（罕见）

**6. AI 调用**

- AI 调用是同步阻塞的，生成总结期间调度器会等待
- 如果 AI 响应超时或失败，使用回退总结
- AI 调用失败不会导致程序崩溃

**7. macOS 权限**

首次运行时 macOS 可能会请求以下权限：
- 辅助功能权限（osascript 对话框）
- 通知权限（系统通知）

如果对话框无法显示，检查"系统偏好设置 > 安全性与隐私 > 隐私 > 辅助功能"。

## Best Practices

### 配置建议

**生产环境**：
```yaml
hourly_interval: 1           # 每小时记录一次
summary_time: "23:00"        # 晚上 11 点生成总结
ai_provider: "codex"         # 使用 Codex
enable_logging: true
max_log_size_mb: 10          # 限制日志大小
```

**开发/测试**：
```yaml
minute_interval: 5           # 每 5 分钟测试一次
summary_time: "00:00"
ai_provider: "codex"
enable_logging: true
```

**番茄工作法**：
```yaml
minute_interval: 25          # 每 25 分钟记录
summary_time: "18:00"        # 下班时生成总结
```

### 备份建议

重要数据：
- `run/data/*.json` - 所有工作记录
- `run/summaries/*.md` - 所有总结文件
- `config.yaml` - 配置文件

备份脚本示例：
```bash
#!/bin/bash
tar -czf backup-$(date +%Y%m%d).tar.gz run/data run/summaries config.yaml
```

### 监控建议

监控项：
- 进程是否运行：`launchctl list | grep daily_summary`
- 日志文件大小：`du -sh run/logs/`
- 磁盘空间：`df -h .`
- 最近一次提醒时间：`tail run/logs/app.log | grep "Work entry saved"`

可以使用 cron 或其他监控工具定期检查。

## Key Features

### 1. 智能定时器重置机制

**问题**：用户手动添加记录后，如果下次定时提醒在不久后触发，会导致重复提醒。

**解决方案**：
- 用户执行 `add` 命令时，CLI 创建重置信号文件（`.reset_signal`）
- Scheduler 每秒检查该文件是否存在
- 检测到信号后，重新计算下一次提醒时间（从当前时间顺延一个完整周期）
- 删除信号文件

**示例**：
```
10:00 - 服务启动，下次提醒: 11:00
10:30 - 用户执行 add 添加记录
10:30 - 检测到信号，下次提醒: 11:30（避免 11:00 重复提醒）
```

### 2. 进程锁机制

**目的**：防止同时运行多个 daily_summary 实例。

**实现**：
- 启动时检查 `run/daily_summary.lock` 文件
- 如果存在且 PID 进程仍在运行，拒绝启动并提示
- 如果进程已结束，清理旧锁文件并创建新锁
- 退出时自动释放锁文件

### 3. At-least-once 总结生成

**保证**：即使错过了配置的总结时间，也会补充生成。

**机制**：
- `DailyData` 包含 `summary_generated` 字段跟踪状态
- Scheduler 启动时检查昨天是否已生成总结
- 如果未生成且已过配置时间，立即生成
- 之后继续正常调度下一次生成

### 4. 延迟检测与跳过

**问题**：电脑休眠后唤醒，timer 触发时间会严重延迟。

**解决方案**：
- 定时器触发时，比较实际时间和预期时间
- 如果延迟超过阈值（小时级 5 分钟，分钟级 50%），跳过本次调度
- 重新计算下一次触发时间

### 5. 日志轮转

**机制**：
- 启动时检查日志文件大小
- 如果超过 `max_log_size_mb` 限制，将当前日志重命名为 `.old`
- 创建新的日志文件
- 保留最近两个日志文件（current + .old）

## Architecture

### 主要组件

1. **定时器模块（Scheduler）** - `internal/scheduler/`
   - **双定时器架构**：
     - 工作记录提醒：支持小时级或分钟级调度
     - 总结生成：每日指定时间触发
   - **智能时间对齐**：
     - 小时级：对齐到整点（如 10:00, 11:00）
     - 分钟级：对齐到分钟边界（如 10:30, 11:00）
   - **延迟检测**：检测系统唤醒导致的延迟，跳过过期的提醒
   - **信号监控**：监控文件信号，实现智能重置
   - **At-least-once 语义**：启动时检查是否有遗漏的总结，补充生成

2. **对话框模块（Dialog）** - `internal/dialog/`
   - **平台**：macOS osascript 实现
   - **功能**：
     - 显示输入对话框收集工作记录
     - 支持超时自动关闭
     - 显示通知（总结生成完成）
   - **上下文展示**：弹窗中展示当日已有记录，帮助用户回忆

3. **数据存储模块（Storage）** - `internal/storage/`
   - **格式**：JSON 文件存储
   - **组织方式**：按日期分文件（`YYYY-MM-DD.json`）
   - **数据结构**：
     - WorkEntry：单条记录（时间戳 + 内容）
     - DailyData：单日所有记录 + 总结生成状态
   - **元数据跟踪**：记录总结是否已生成，避免重复

4. **总结生成模块（Summary Generator）** - `internal/summary/`
   - **多 AI 架构**：
     - AIClient 接口抽象
     - CodexClient 实现（默认）
     - ClaudeClient 实现（可选）
   - **Prompt 构建**：格式化工作记录为结构化 prompt
   - **回退机制**：AI 不可用时生成简单模板总结
   - **通知集成**：总结生成后发送系统通知

5. **CLI 命令模块（CLI）** - `internal/cli/`
   - **进程管理**：
     - 文件锁机制防止重复启动
     - PID 跟踪和进程检测
   - **手动操作**：
     - `add`：手动添加工作记录
     - `list`：查看今日记录
     - `summary`：手动触发总结生成
   - **信号机制**：添加记录后发送重置信号给 Scheduler

6. **配置管理模块（Config）** - `config/`
   - 支持 YAML/JSON 格式自动识别
   - 路径解析（相对/绝对路径处理）
   - 默认配置 fallback
   - 目录创建和权限管理

### AI 总结集成

程序支持两种 AI 提供商生成工作总结：

#### 1. Codex（默认提供商）
- **调用方式**：`codex exec "{prompt}"`
- **工作目录**：在项目目录（work_dir）中执行
- **回退机制**：如果 Codex 不可用，生成简单模板总结
- **配置项**：`ai_provider: "codex"`, `codex_path: "codex"`

#### 2. Claude Code（可选）
- **调用方式**：`claude-code --prompt "{prompt}"`
- **工作目录**：临时目录 `~/.daily_summary_temp`
- **回退机制**：如果 Claude Code 不可用，生成简单模板总结
- **配置项**：`ai_provider: "claude"`, `claude_code_path: "claude-code"`

#### 总结生成流程

当触发总结生成时（每日指定时间或手动命令），程序会：

1. **收集数据**：读取指定日期的所有工作记录
2. **检查记录**：如果没有记录，跳过生成并报错
3. **构建 Prompt**：按照固定格式组织工作记录
4. **调用 AI**：根据 `ai_provider` 配置选择对应客户端
5. **保存结果**：将 AI 生成的总结保存为 Markdown 文件
6. **标记状态**：在 DailyData 中标记 `summary_generated: true`
7. **发送通知**：通过 osascript 显示系统通知

#### AI 收到的 Prompt 格式

```
请为以下工作记录生成一份结构化的工作总结（日期：YYYY-MM-DD）

工作记录（每1条记录都是对前一个时间窗口工作内容的总结）：

- **09:00**: 完成了用户认证模块的代码审查
- **10:00**: 修复了3个前端 bug
- **11:00**: 参加团队站会
...

请按照以下格式生成总结：
## 主要完成的任务
（列出完成的主要工作，按项目或模块分类，并估算工作实际耗时）

## 关键进展
（突出重要的进展和成果）

## 遇到的问题
（如果有记录到问题，列出来）

## 明日计划
（如果记录中有提及，整理出来）
```

#### AI 期望输出

AI 应该返回符合上述格式的 Markdown 文本，包括：
- 按项目/模块分类的任务列表
- 工作时长估算
- 重要进展和成果
- 遇到的问题和解决方案
- 明日计划（如有）

### 数据结构

```go
// WorkEntry 表示单次工作记录
type WorkEntry struct {
    Timestamp time.Time `json:"timestamp"` // 记录时间
    Content   string    `json:"content"`   // 工作内容
}

// DailyData 表示一天的所有工作记录
type DailyData struct {
    Date             string      `json:"date"`               // 格式: YYYY-MM-DD
    Entries          []WorkEntry `json:"entries"`            // 工作记录列表
    SummaryGenerated bool        `json:"summary_generated"`  // 是否已生成总结
}

// SummaryMetadata 总结的元数据
type SummaryMetadata struct {
    GeneratedAt time.Time `json:"generated_at"` // 生成时间
    Date        string    `json:"date"`         // 总结对应的日期
    EntryCount  int       `json:"entry_count"`  // 记录条数
}

// Config 应用配置（完整定义见 internal/models/models.go）
type Config struct {
    WorkDir        string // 工作目录（项目根目录）
    DataDir        string // 数据目录
    SummaryDir     string // 总结目录
    HourlyInterval int    // 小时间隔
    MinuteInterval int    // 分钟间隔（优先级更高）
    SummaryTime    string // 生成总结的时间
    AIProvider     string // AI 提供商："codex" 或 "claude"
    CodexPath      string // Codex CLI 路径
    ClaudeCodePath string // Claude Code CLI 路径
    DialogTimeout  int    // 对话框超时（秒）
    EnableLogging  bool   // 是否启用日志
    LogFile        string // 日志文件路径
    MaxLogSizeMB   int    // 日志文件最大大小（MB）
}
```

### 文件组织

项目采用标准 Go 项目结构：

```
daily_summary/
├── main.go                           # 程序入口，CLI 路由和命令处理
├── go.mod / go.sum                   # Go 模块依赖
│
├── config/                           # 配置管理包
│   └── config.go                     # 配置加载、保存、路径解析
│
├── internal/                         # 内部包（不可被外部导入）
│   ├── models/                       # 数据模型定义
│   │   └── models.go                 # WorkEntry, DailyData, Config 等
│   │
│   ├── scheduler/                    # 定时任务调度
│   │   ├── scheduler.go              # 双定时器、信号监控、延迟检测
│   │   └── scheduler_test.go         # 单元测试
│   │
│   ├── dialog/                       # 对话框接口和实现
│   │   ├── dialog.go                 # Dialog 接口定义
│   │   └── osascript.go              # macOS osascript 实现
│   │
│   ├── storage/                      # 数据存储
│   │   ├── storage.go                # Storage 接口定义
│   │   └── json_storage.go           # JSON 文件存储实现
│   │
│   ├── summary/                      # AI 总结生成
│   │   ├── client.go                 # AIClient 接口定义
│   │   ├── codex.go                  # Codex 客户端实现
│   │   ├── claude.go                 # Claude Code 客户端实现
│   │   └── generator.go              # 总结生成器（Prompt 构建、流程编排）
│   │
│   └── cli/                          # CLI 命令实现
│       ├── cli.go                    # add, list 命令，进程锁，信号机制
│       └── cli_test.go               # 单元测试
│
├── scripts/                          # 安装和测试脚本
│   ├── install.sh                    # 安装脚本（编译、配置、launchd）
│   ├── uninstall.sh                  # 卸载脚本
│   ├── quick_test.sh                 # 快速测试脚本
│   ├── test_config.sh                # 配置测试
│   ├── test_dialog.sh                # 对话框测试
│   └── test_minute_interval.sh       # 分钟级调度测试
│
├── launchd/                          # macOS 后台服务
│   └── com.humg.daily_summary.plist  # launchd 配置文件
│
├── docs/                             # 项目文档
│   ├── QUICK_START.md                # 快速开始指南
│   ├── CONFIGURATION.md              # 配置详解
│   └── MINUTE_SCHEDULING.md          # 分钟级调度说明
│
├── run/                              # 运行时数据（.gitignore）
│   ├── data/                         # 工作记录 JSON 文件
│   │   ├── 2026-01-19.json
│   │   └── 2026-01-20.json
│   ├── summaries/                    # Markdown 总结文件
│   │   └── 2026-01-19.md
│   ├── logs/                         # 日志文件
│   │   ├── app.log                   # 应用日志
│   │   ├── app.log.old               # 轮转的旧日志
│   │   ├── stdout.log                # 标准输出（launchd）
│   │   └── stderr.log                # 标准错误（launchd）
│   ├── .reset_signal                 # 重置信号文件（临时）
│   └── daily_summary.lock            # 进程锁文件
│
├── config.yaml                       # 用户配置文件
├── config.example.yaml               # 配置示例
├── config.pomodoro.yaml              # 番茄钟配置示例
│
├── CLAUDE.md                         # 本文件：Claude Code 开发指南
├── README.md                         # 项目说明文档
└── CHANGELOG.md                      # 更新日志
```

### 配置文件

配置文件支持 YAML 和 JSON 两种格式（根据扩展名自动识别）：
- 默认路径：`~/.config/daily_summary/config.yaml` 或项目根目录 `config.yaml`
- 示例文件：`config.example.yaml`

**关键配置项：**

**目录配置：**
- `work_dir`：工作目录（项目根目录），相对路径会基于此目录解析
- `data_dir`：工作记录存储目录（默认 `./run/data`）
- `summary_dir`：总结文件存储目录（默认 `./run/summaries`）

**提醒间隔：**
- `hourly_interval`：小时级提醒间隔（默认 1，表示每小时提醒一次）
- `minute_interval`：分钟级提醒间隔（如果设置则优先使用，覆盖 hourly_interval）
  - 示例：30 表示每 30 分钟提醒一次
  - 适用于番茄工作法等场景

**总结生成：**
- `summary_time`：每日生成总结的时间，格式 "HH:MM"（默认 "00:00"）
- `ai_provider`：AI 提供商，可选 "codex" 或 "claude"（默认 "codex"）
- `codex_path`：Codex CLI 路径（默认 "codex"）
- `claude_code_path`：Claude Code CLI 路径（默认 "claude-code"）

**其他配置：**
- `dialog_timeout`：对话框超时时间（秒，默认 300）
- `enable_logging`：是否启用日志记录（默认 true）
- `log_file`：日志文件路径（默认 `./run/logs/app.log`）
- `max_log_size_mb`：日志文件最大大小（MB），超过后自动轮转（0 表示不限制）

配置文件会在程序启动时自动加载，如果不存在则使用默认配置。

## CLI Usage

程序支持以下命令：

### 1. serve（默认命令）

启动后台服务，运行定时提醒和总结生成任务。

```bash
daily_summary serve
# 或直接运行
daily_summary
```

**特性**：
- 检查并获取进程锁，防止重复启动
- 加载配置文件
- 设置日志记录
- 初始化所有组件（对话框、存储、AI 客户端、生成器、调度器）
- 监听系统信号优雅退出（Ctrl+C）

### 2. add - 手动添加工作记录

立即添加一条工作记录，无需等待定时弹窗。

```bash
daily_summary add "完成需求文档审查"
daily_summary add "修复登录页面的 bug"
```

**特性**：
- 记录当前时间戳
- 保存到当日 JSON 文件
- 发送重置信号，自动顺延下次提醒时间
- 打印确认信息

### 3. list - 查看今日记录

显示今天已记录的所有工作内容。

```bash
daily_summary list
```

**输出示例**：
```
📝 今日工作记录 (2026-01-20)：

  • 09:30 - 完成需求文档审查
  • 11:00 - 参加技术评审会议
  • 14:00 - 修复登录页面的 bug

共 3 条记录
```

### 4. summary - 手动生成总结

手动触发指定日期的工作总结生成。

```bash
# 生成今天的总结
daily_summary summary

# 生成指定日期的总结
daily_summary summary --date 2026-01-19
```

**特性**：
- 支持 `--date` 参数指定日期（格式：YYYY-MM-DD）
- 检查是否有工作记录
- 调用 AI 生成总结
- 保存 Markdown 文件到 summary_dir
- 标记总结生成状态
- 打印文件路径

### 5. help - 显示帮助信息

```bash
daily_summary help
# 或
daily_summary --help
```

### 全局选项

**--config**：指定配置文件路径

```bash
daily_summary --config ~/my-config.yaml serve
daily_summary --config ./custom.yaml add "完成测试"
```

## Deployment

### macOS launchd 集成

项目提供脚本实现 launchd 后台服务：

**安装脚本**：`scripts/install.sh`
- 编译二进制文件
- 创建运行目录（run/data, run/summaries, run/logs）
- 复制配置文件
- 安装 launchd plist 文件
- 启动服务

**卸载脚本**：`scripts/uninstall.sh`
- 停止并卸载 launchd 服务
- 删除 plist 文件
- 保留数据和配置

**plist 配置**：`launchd/com.humg.daily_summary.plist`
- WorkingDirectory：项目根目录
- StandardOutPath：`run/logs/stdout.log`
- StandardErrorPath：`run/logs/stderr.log`
- RunAtLoad：系统启动时自动运行
- KeepAlive：崩溃后自动重启 