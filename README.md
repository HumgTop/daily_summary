# Daily Summary

一个智能的工作记录和总结工具，帮助你更好地追踪日常工作进展。

## 💡 解决什么问题

**工作记录困境**：
- 写日报周报时想不起来做了什么
- 每天忙忙碌碌，却说不清具体产出
- 想记录工作内容，但总是忘记或懒得手动记录
- 工作记录散落各处，难以整理和回顾

**本工具的解决方案**：
- ⏰ **自动提醒**：定时弹窗提醒记录工作，不会遗漏
- 🤖 **AI 总结**：自动将工作记录生成结构化的每日/每周总结
- 📝 **随时记录**：命令行快速记录，无需打开应用
- 📊 **数据可视化**：生成工作时间分配图表，了解时间都花在哪
- 🔄 **智能调度**：手动记录后自动调整提醒时间，避免重复打扰

## ✨ 核心特性

- **定时提醒**：每小时或自定义间隔弹窗提醒（支持番茄工作法）
- **手动记录**：通过 CLI 命令随时添加工作记录
- **每日总结**：AI 自动生成结构化的工作总结（任务、进展、问题、计划）
- **每周总结**：自动聚合一周的工作内容，生成周报
- **多 AI 支持**：支持 Codex、Coco、Claude Code 三种 AI 提供商
- **模板驱动**：可自定义总结格式的 Markdown 模板
- **后台服务**：macOS launchd 持续运行，开机自启
- **智能重置**：手动添加记录后自动顺延提醒时间

## 🚀 快速开始

### 系统要求

- macOS 操作系统
- Go 1.19+（用于编译）
- AI CLI 工具（任选其一，用于生成总结）：
  - [Codex](https://github.com/codex-cli/codex)（推荐）
  - Coco
  - [Claude Code](https://claude.ai/code)

### 安装

1. **克隆项目**
```bash
git clone <repository-url>
cd daily_summary
```

2. **运行安装脚本**
```bash
./scripts/install.sh
```

安装脚本会自动完成：
- 编译程序
- 创建运行目录
- 安装 launchd 服务
- 启动后台服务

3. **验证安装**
```bash
# 查看服务状态
launchctl list | grep daily_summary

# 查看日志
tail -f run/logs/app.log
```

## 📖 使用指南

### 自动记录（推荐）

安装后，工具会在后台运行，每小时整点弹出对话框：

1. 在对话框中输入过去一段时间的工作内容
2. 点击"确定"保存记录
3. 第二天会自动生成前一天的工作总结

**上下文提示**：对话框会显示当天已有的记录，帮助你回忆工作内容。

### 手动记录

**快速添加记录**：
```bash
daily_summary add "完成用户认证模块的代码审查"
daily_summary add "修复了登录页面的3个bug"
```

**显示输入对话框**：
```bash
daily_summary popup
```

**查看今日记录**：
```bash
daily_summary list
```

输出示例：
```
📝 今日工作记录 (2026-02-02)：

  • 09:30 - 完成用户认证模块的代码审查
  • 11:00 - 参加团队技术分享会议
  • 14:00 - 修复了登录页面的3个bug

共 3 条记录
```

### 生成总结

**生成每日总结**：
```bash
# 生成今天的总结
daily_summary summary

# 生成指定日期的总结
daily_summary summary --date 2026-01-30
```

**生成每周总结**：
```bash
# 生成本周的总结
daily_summary weekly

# 生成指定日期所在周的总结
daily_summary weekly --date 2026-01-30
```

## ⚙️ 配置

配置文件：项目根目录的 `config.yaml`

**基础配置示例**：
```yaml
# 工作目录（项目根目录）
work_dir: /Users/xxx/daily_summary

# 数据存储
data_dir: run/data
summary_dir: run/summaries

# 提醒间隔（二选一）
hourly_interval: 1          # 每小时提醒
# minute_interval: 25       # 或每25分钟提醒（番茄工作法）

# 每日总结
summary_time: "23:00"       # 晚上11点生成总结
ai_provider: "codex"        # AI 提供商：codex/coco/claude
codex_path: "codex"

# 每周总结（可选）
enable_weekly_summary: true
weekly_summary_time: "11:00"
weekly_summary_day: 1       # 1=周一，7=周日

# 其他
dialog_timeout: 3600        # 对话框超时（秒）
enable_logging: true
max_log_size_mb: 10         # 日志自动轮转
```

**配置说明**：
- `minute_interval`：如果设置则优先于 `hourly_interval`
- `ai_provider`：可选 `codex`、`coco`、`claude`
- 周总结会自动聚合该周的所有每日总结

更多配置选项请参考 `config.example.yaml`。

## 📁 目录结构

```
daily_summary/
├── run/
│   ├── data/                    # 工作记录（JSON）
│   │   ├── 2026-02-01.json
│   │   └── 2026-02-02.json
│   ├── summaries/               # 生成的总结
│   │   ├── daily/               # 每日总结
│   │   │   ├── 2026-02-01.md
│   │   │   └── 2026-02-02.md
│   │   └── weekly/              # 每周总结
│   │       └── 2026-W05.md
│   ├── logs/                    # 日志文件
│   │   ├── app.log
│   │   ├── scheduler_check.log
│   │   ├── stdout.log
│   │   └── stderr.log
│   ├── tasks.json               # 任务调度状态
│   └── daily_summary.lock       # 进程锁
├── templates/                   # Prompt 模板（可自定义）
│   ├── summary_prompt.md
│   └── weekly_summary_prompt.md
└── config.yaml                  # 配置文件
```

## 🔧 常用命令

**服务管理**：
```bash
# 查看服务状态
launchctl list | grep daily_summary

# 重启服务
launchctl unload ~/Library/LaunchAgents/com.humg.daily_summary.plist
launchctl load ~/Library/LaunchAgents/com.humg.daily_summary.plist
```

**查看日志**：
```bash
# 应用日志
tail -f run/logs/app.log

# 调度器日志
tail -f run/logs/scheduler_check.log
```

**调试任务**：
```bash
# 查看任务状态
cat run/tasks.json | jq '.'

# 查看特定任务
cat run/tasks.json | jq '.tasks[] | select(.id=="work-reminder")'
```

## 🏗️ 实现原理

### 核心架构

```
┌─────────────────────────────────────────────┐
│          Scheduler (调度器)                  │
│   每分钟检查 tasks.json 中的任务            │
└─────────┬───────────────────────────────────┘
          │
          ├──> ReminderTask (工作提醒)
          │    - 按间隔触发对话框
          │    - 延迟检测和跳过
          │
          ├──> SummaryTask (每日总结)
          │    - 每日指定时间触发
          │    - 批量生成未生成的总结
          │
          ├──> WeeklySummaryTask (每周总结)
          │    - 每周指定时间触发
          │    - 聚合7天的每日总结
          │
          └──> LogRotateTask (日志轮转)
               - 定期检查日志大小
               - 自动轮转大文件
```

### 关键特性

**1. 任务注册表**（`run/tasks.json`）
- 持久化所有任务的状态（下次运行时间、上次运行时间、错误信息）
- 进程重启后自动恢复任务调度
- 支持任务特定数据存储

**2. 智能重置机制**
```
用户手动添加记录 → 更新任务注册表 → 顺延下次提醒时间 → 避免重复提醒
```

**3. 模板驱动的 Prompt**
- Prompt 从 `templates/*.md` 加载
- 支持自定义格式和要求
- 模板加载失败时使用内置回退

**4. 批量总结生成**
- SummaryTask 启动时扫描所有未生成总结的日期
- 自动补充生成遗漏的总结
- 保证 At-least-once 语义

**5. 延迟检测**
- 系统休眠唤醒后，检测任务延迟
- 如果延迟过大（>50%间隔），跳过本次执行并重新调度
- 避免唤醒后连续弹窗

### 技术栈

- **语言**：Go 1.19+
- **对话框**：macOS osascript（原生）
- **调度**：统一调度器 + 文件持久化注册表
- **存储**：JSON 文件（按日期分文件）
- **AI 集成**：通过 CLI 调用（Codex/Coco/Claude Code）
- **服务管理**：macOS launchd

## 🐛 故障排除

**对话框不弹出**：
1. 检查服务状态：`launchctl list | grep daily_summary`
2. 查看任务状态：`cat run/tasks.json | jq '.tasks[] | select(.id=="work-reminder")'`
3. 查看日志错误：`tail -100 run/logs/app.log`

**总结生成失败**：
1. 检查 AI CLI 是否安装：`which codex` / `which coco` / `which claude-code`
2. 查看任务错误信息：`cat run/tasks.json | jq '.tasks[] | select(.id=="daily-summary")'`
3. 手动测试：`daily_summary summary --date 2026-02-01`
4. 检查模板文件：`ls templates/`

**任务调度异常**：
1. 备份任务状态：`cp run/tasks.json run/tasks.json.backup`
2. 删除任务文件：`rm run/tasks.json`
3. 重启服务（会自动重新初始化）：`./scripts/install.sh`

更多问题请查看 [CLAUDE.md](CLAUDE.md) 的 Troubleshooting 章节。

## 🗑️ 卸载

```bash
# 运行卸载脚本（保留数据）
./scripts/uninstall.sh

# 完全删除（包括数据）
./scripts/uninstall.sh
rm -rf run/
rm config.yaml
```

## 📚 相关文档

- [CLAUDE.md](CLAUDE.md) - 完整的开发者文档
- [config.example.yaml](config.example.yaml) - 配置示例
- [CHANGELOG.md](CHANGELOG.md) - 更新日志

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
