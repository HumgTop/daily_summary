# CLAUDE.md

本文件为 Claude Code 提供项目开发指导。

## Project Overview

**Daily Summary** 是一个 Go 编写的 macOS 工作记录和 AI 总结工具。

**核心功能：**
- 定时提醒记录工作（支持小时/分钟级间隔）
- 手动添加工作记录（CLI 命令）
- 智能重置（手动记录后自动顺延提醒时间）
- JSON 格式数据持久化（按日期组织）
- AI 自动生成每日/每周工作总结
- 多 AI 支持（Codex/Coco/Claude Code）
- launchd 后台服务（自动启动、进程锁保护）

**技术栈：**
- Go 1.19+，单一二进制文件
- macOS osascript 原生对话框
- 统一任务调度器 + 文件持久化注册表
- Markdown 模板驱动的 Prompt 系统

## Quick Start

```bash
# 编译
go build -o daily_summary

# 本地运行测试
go run main.go add "测试记录"
go run main.go list
go run main.go popup

# 安装为系统服务
./scripts/install.sh

# 查看日志
tail -f ./run/logs/app.log
tail -f ./run/logs/scheduler_check.log
```

**测试脚本**（位于 `scripts/`）：
- `quick_test.sh` - 快速功能测试（分钟级调度）
- `test_config.sh` - 配置加载测试
- `test_dialog.sh` - 对话框测试
- `test_minute_interval.sh` - 分钟级调度测试

**调试任务系统**：
```bash
# 查看任务注册表状态
cat run/tasks.json | jq '.'

# 重置任务（删除后会自动重新初始化）
rm run/tasks.json && go run main.go serve
```

**调试 launchd 服务**：
```bash
launchctl list | grep daily_summary
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist
launchctl load ~/Library/LaunchAgents/com.humg.daily_summary.plist
```

## Development Notes

**重要约定**：
- 对话过程中创建新文档需经我批准
- 产出的文档统一放在 `docs/` 目录

**配置优先级**：
1. `--config` 命令行参数
2. 项目根目录 `config.yaml`
3. `~/.config/daily_summary/config.yaml`
4. 默认配置（硬编码）

**注意事项**：
- Storage 操作无锁保护，进程锁保证单实例运行
- JSON 按日期分文件，无数据库，无事务保证
- AI 调用同步阻塞，失败时使用回退机制
- 所有时间使用系统本地时区，无转换逻辑

**扩展开发**：
- 添加新 AI 提供商：参考 `internal/summary/codex.go`，实现 `AIClient` 接口
- 平台扩展（Linux/Windows）：在 `internal/dialog/` 添加平台实现，使用 build tags 区分

## Troubleshooting

**服务无法启动**：
```bash
# 检查是否已有实例运行
launchctl list | grep daily_summary
cat run/daily_summary.lock

# 强制重启
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist
rm run/daily_summary.lock
./scripts/install.sh
```

**对话框不弹出**：
```bash
# 手动触发测试
./scripts/test_dialog.sh

# 查看下次提醒时间
tail -20 run/logs/app.log | grep "Next reminder"
```

**总结生成失败**：
```bash
# 检查 AI CLI 是否可用
which codex && which coco && which claude-code

# 手动测试
daily_summary summary --date 2026-01-20

# 查看任务错误
cat run/tasks.json | jq '.tasks[] | select(.id=="daily-summary")'
```

**任务调度异常**：
```bash
# 查看所有任务状态
cat run/tasks.json | jq '.tasks[] | {id, enabled, next_run, last_error}'

# 手动触发任务测试
daily_summary popup      # 测试提醒
daily_summary summary    # 测试每日总结
daily_summary weekly     # 测试周总结

# 重置任务注册表
mv run/tasks.json run/tasks.json.backup
./scripts/install.sh
```

更多问题排查：查看日志文件 `run/logs/*.log`

## Architecture

### 核心组件

**1. Scheduler（任务调度）** - `internal/scheduler/`
- 每分钟唤醒检查所有任务
- 任务注册表持久化到 `run/tasks.json`
- 支持三种任务类型：interval（间隔）、daily（每日）、once（一次性）
- 详见：`scheduler.go`, `registry.go`, `task.go`

**2. Tasks（任务实现）** - `internal/tasks/`
- `ReminderTask`: 工作记录提醒（小时级/分钟级间隔）
- `SummaryTask`: 每日总结生成（指定时间触发，支持批量补充）
- `WeeklySummaryTask`: 每周总结生成（聚合 7 天的每日总结）
- `LogRotateTask`: 日志轮转（间隔触发，检查多个日志文件）

**3. Dialog（对话框）** - `internal/dialog/`
- macOS osascript 实现
- 输入对话框（展示当日记录作为上下文）
- 系统通知

**4. Storage（数据存储）** - `internal/storage/`
- JSON 文件存储，按日期分文件 (`YYYY-MM-DD.json`)
- 数据结构：`WorkEntry`（单条记录）、`DailyData`（单日记录+总结状态）
- 详见：`json_storage.go`

**5. Summary（AI 总结）** - `internal/summary/`
- 多 AI 架构：`AIClient` 接口，三种实现（Codex/Coco/Claude）
- Prompt 模板系统：从 `templates/*.md` 加载，支持自定义
- 批量生成：自动生成所有未生成的历史总结
- 回退机制：AI 不可用时生成简单模板总结
- 详见：`generator.go`, `codex.go`, `coco.go`, `claude.go`

**6. CLI（命令行）** - `internal/cli/`
- 进程锁机制（防止重复启动）
- 命令：add, popup, list, summary, weekly
- 手动添加记录后调用 `UpdateTaskSchedule()` 顺延提醒时间

**7. Config（配置管理）** - `config/`
- 支持 YAML/JSON 自动识别
- 路径解析（相对/绝对路径）
- 默认配置 fallback

### 任务调度系统

**工作流程**：
```
服务启动 → 加载 tasks.json → 每分钟检查循环 → 遍历所有任务 →
调用 ShouldRun() → 执行 Execute() → 调用 OnExecuted() → 保存状态
```

**Task 接口**（详见 `internal/scheduler/task.go`）：
```go
type Task interface {
    ID() string
    Name() string
    ShouldRun(now time.Time) bool
    Execute() error
    OnExecuted(success bool, err error)
}
```

**任务注册表**（`run/tasks.json`）：
- 存储任务配置和状态
- 字段：ID、Name、Type、Enabled、NextRun、LastRun、LastSuccess、LastError、Data
- 示例结构见代码或生成的 `run/tasks.json` 文件

### AI 总结集成

**支持的 AI 提供商**：
1. **Codex**（默认）：`codex exec "{prompt}"`
2. **Coco**：`coco -p "{prompt}"`
3. **Claude Code**：`claude-code --prompt "{prompt}"`

**总结生成流程**：
1. 查找未生成的总结（批量补充）
2. 收集工作记录
3. 加载 Prompt 模板（`templates/summary_prompt.md` 或 `weekly_summary_prompt.md`）
4. 填充模板并调用 AI
5. 保存 Markdown 文件到 `run/summaries/daily/` 或 `weekly/`
6. 标记生成状态
7. 发送系统通知

**Prompt 模板**：
- 支持占位符：`{date}`, `{count}`, `{formatted_entries}`
- 每日总结模板：包含主要任务、工作时间分析（Mermaid 图表）、关键进展、问题、明日计划
- 每周总结模板：聚合 7 天的每日总结，更高层次的总结
- 详见：`templates/summary_prompt.md`, `templates/weekly_summary_prompt.md`

### 关键特性

**1. 智能定时器重置**：
- `add` 命令执行后直接更新任务注册表
- 从当前时间顺延一个完整周期，避免重复提醒
- 实现：`internal/cli/cli.go` 中的 `UpdateTaskSchedule()` 调用

**2. 进程锁机制**：
- 启动时检查 `run/daily_summary.lock` 文件
- 防止同时运行多个实例

**3. At-least-once 总结生成**：
- `DailyData` 包含 `summary_generated` 字段
- 启动时检查并批量补充所有未生成的总结

**4. 延迟检测与跳过**：
- 电脑休眠后唤醒检测延迟
- 超过阈值（小时级 5 分钟，分钟级 50%）跳过本次调度

**5. 日志轮转**：
- LogRotateTask 定期检查日志大小
- 超过 `max_log_size_mb` 时重命名为 `.old`，保留最近两个版本

**6. 任务注册表持久化**：
- 任务状态在进程重启后不丢失
- 支持任务特定数据存储（Data 字段）

### 数据结构

核心结构定义见 `internal/models/models.go`：

```go
// 单条工作记录
type WorkEntry struct {
    Timestamp time.Time
    Content   string
}

// 单日所有工作记录
type DailyData struct {
    Date             string      // YYYY-MM-DD
    Entries          []WorkEntry
    SummaryGenerated bool
}

// 任务配置（存储在 run/tasks.json）
type TaskConfig struct {
    ID              string
    Name            string
    Type            TaskType  // "interval", "daily", "once"
    Enabled         bool
    IntervalMinutes int       // interval 任务的分钟数
    Time            string    // daily 任务的时间（HH:MM）
    NextRun         time.Time
    LastRun         time.Time
    LastSuccess     time.Time
    LastError       string
    Data            map[string]interface{}
}

// 应用配置（完整定义见 models.go）
type Config struct {
    WorkDir              string
    DataDir              string
    SummaryDir           string
    HourlyInterval       int
    MinuteInterval       int
    SummaryTime          string
    AIProvider           string  // "codex", "coco", "claude"
    EnableWeeklySummary  bool
    WeeklySummaryTime    string
    WeeklySummaryDay     int     // 1=周一, 7=周日
    // ... 更多字段见 models.go
}
```

### 文件组织

```
daily_summary/
├── main.go                           # 程序入口，CLI 路由
├── config/config.go                  # 配置管理
├── internal/
│   ├── models/models.go              # 数据模型定义
│   ├── scheduler/                    # 任务调度
│   │   ├── scheduler.go              # 统一调度器
│   │   ├── task.go                   # Task 接口
│   │   ├── registry.go               # 任务注册表
│   │   └── init.go                   # 任务初始化
│   ├── tasks/                        # 具体任务实现
│   │   ├── reminder.go               # 工作提醒
│   │   ├── summary.go                # 每日总结
│   │   ├── weekly_summary.go         # 每周总结
│   │   └── log_rotate.go             # 日志轮转
│   ├── dialog/                       # 对话框
│   │   ├── dialog.go                 # 接口定义
│   │   └── osascript.go              # macOS 实现
│   ├── storage/                      # 数据存储
│   │   ├── storage.go                # 接口定义
│   │   └── json_storage.go           # JSON 实现
│   ├── summary/                      # AI 总结
│   │   ├── client.go                 # AIClient 接口
│   │   ├── codex.go                  # Codex 实现
│   │   ├── coco.go                   # Coco 实现
│   │   ├── claude.go                 # Claude Code 实现
│   │   └── generator.go              # 总结生成器
│   └── cli/cli.go                    # CLI 命令实现
├── templates/                        # Prompt 模板
│   ├── summary_prompt.md             # 每日总结模板
│   └── weekly_summary_prompt.md      # 每周总结模板
├── scripts/                          # 安装和测试脚本
│   ├── install.sh                    # 安装脚本
│   ├── uninstall.sh                  # 卸载脚本
│   └── *.sh                          # 各种测试脚本
├── launchd/                          # macOS 后台服务
│   └── com.humg.daily_summary.plist  # launchd 配置
├── docs/                             # 项目文档
│   ├── QUICK_START.md
│   ├── CONFIGURATION.md
│   └── MINUTE_SCHEDULING.md
├── run/                              # 运行时数据（.gitignore）
│   ├── data/                         # 工作记录 JSON
│   ├── summaries/
│   │   ├── daily/                    # 每日总结 MD
│   │   └── weekly/                   # 每周总结 MD
│   ├── logs/                         # 日志文件
│   ├── tasks.json                    # 任务注册表
│   └── daily_summary.lock            # 进程锁
├── config.yaml                       # 用户配置
└── config.example.yaml               # 配置示例
```

## Configuration

配置文件支持 YAML/JSON 格式，详见 `config.example.yaml`。

**关键配置项**：
```yaml
# 目录
work_dir: "."                         # 工作目录（项目根目录）
data_dir: "./run/data"                # 数据目录
summary_dir: "./run/summaries"        # 总结目录（包含 daily/ 和 weekly/）

# 提醒间隔（二选一，minute_interval 优先）
hourly_interval: 1                    # 小时间隔
minute_interval: 25                   # 分钟间隔（番茄工作法等场景）

# 每日总结
summary_time: "23:00"                 # 生成时间
ai_provider: "codex"                  # "codex", "coco", "claude"
codex_path: "codex"                   # CLI 路径
coco_path: "coco"
claude_code_path: "claude-code"

# 每周总结（可选）
enable_weekly_summary: true
weekly_summary_time: "11:00"
weekly_summary_day: 1                 # 1=周一, 7=周日

# 其他
dialog_timeout: 300                   # 对话框超时（秒）
enable_logging: true
log_file: "./run/logs/app.log"
max_log_size_mb: 10                   # 日志轮转阈值
```

详细配置说明见 `docs/CONFIGURATION.md`。

## CLI Usage

```bash
# 1. serve（默认命令）- 启动后台服务
daily_summary serve
daily_summary                         # 等同于 serve

# 2. add - 手动添加工作记录
daily_summary add "完成需求文档审查"
# 自动顺延下次提醒时间

# 3. list - 查看今日记录
daily_summary list

# 4. popup - 手动触发输入对话框
daily_summary popup

# 5. summary - 手动生成每日总结
daily_summary summary                 # 今天
daily_summary summary --date 2026-01-19

# 6. weekly - 手动生成每周总结
daily_summary weekly                  # 本周
daily_summary weekly --date 2026-01-20

# 7. help - 显示帮助
daily_summary help

# 全局选项
daily_summary --config ~/my-config.yaml serve
```

## Deployment

**安装为系统服务**（使用 launchd）：
```bash
./scripts/install.sh
```

**卸载**：
```bash
./scripts/uninstall.sh
# 数据和配置将保留
```

**launchd 配置**（`launchd/com.humg.daily_summary.plist`）：
- WorkingDirectory: 项目根目录（绝对路径）
- StandardOutPath: `run/logs/stdout.log`
- StandardErrorPath: `run/logs/stderr.log`
- RunAtLoad: 系统启动时自动运行
- KeepAlive: 崩溃后自动重启

**备份重要数据**：
- `run/data/*.json` - 工作记录
- `run/summaries/daily/*.md` - 每日总结
- `run/summaries/weekly/*.md` - 每周总结
- `run/tasks.json` - 任务状态
- `config.yaml` - 配置
- `templates/*.md` - 自定义模板

## Testing

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/scheduler/
go test -v ./internal/summary/

# 测试覆盖率
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Further Reading

- `README.md` - 项目总体介绍
- `docs/QUICK_START.md` - 快速开始指南
- `docs/CONFIGURATION.md` - 配置详解
- `docs/MINUTE_SCHEDULING.md` - 分钟级调度说明
- `CHANGELOG.md` - 更新日志
- `templates/*.md` - Prompt 模板示例

## Index for Exploration

当需要了解详细实现时，查看以下文件：

**任务系统**：
- `internal/scheduler/scheduler.go` - 调度器核心逻辑
- `internal/scheduler/registry.go` - 任务注册表持久化
- `internal/tasks/*.go` - 各任务的具体实现

**AI 集成**：
- `internal/summary/generator.go` - 总结生成流程
- `internal/summary/codex.go` - Codex 客户端实现
- `templates/summary_prompt.md` - Prompt 模板

**数据存储**：
- `internal/storage/json_storage.go` - JSON 存储实现
- `run/data/*.json` - 实际数据文件示例
- `run/tasks.json` - 任务注册表示例

**命令行**：
- `main.go` - CLI 路由和命令处理
- `internal/cli/cli.go` - 命令实现和进程锁

**配置**：
- `config/config.go` - 配置加载和解析
- `config.example.yaml` - 配置示例
