# 更新日志

## [v1.4.0] - 2026-01-20

### 重要变更

#### 运行时目录重构 📁
- **统一运行时数据目录**：所有运行时数据现在统一存储在项目的 `run/` 目录下
  - 数据文件：`./run/data/`
  - 总结文件：`./run/summaries/`
  - 日志文件：`./run/logs/`
  - 锁文件：`./run/daily_summary.lock`
  - 信号文件：`./run/.reset_signal`

#### 配置更新
- **默认路径变更**：
  - 旧: `~/daily_summary/data` → 新: `./run/data`
  - 旧: `~/daily_summary/summaries` → 新: `./run/summaries`
  - 旧: `~/daily_summary/logs` → 新: `./run/logs`

### 优势

**项目自包含**：
- 所有运行时数据都在项目目录下，便于管理和备份
- 不再在用户主目录下创建额外目录
- 更容易迁移和部署

**开发友好**：
- 数据和代码放在一起，便于调试
- `.gitignore` 已配置忽略 `run/` 目录，避免提交用户数据
- 多个项目副本可以独立运行，互不干扰

**清晰的目录结构**：
```
项目目录/
├── run/                    # 运行时数据（被 .gitignore 忽略）
│   ├── data/              # 工作记录
│   ├── summaries/         # 每日总结
│   ├── logs/              # 所有日志
│   ├── .reset_signal      # 信号文件
│   └── daily_summary.lock # 进程锁
└── ~/.config/daily_summary/
    └── config.yaml        # 配置文件
```

### 升级说明

**自动升级**（推荐）：
```bash
# 重新安装服务
./scripts/install.sh

# 配置文件会自动使用新路径
```

**手动升级**：
如果你自定义了配置文件，需要手动更新路径：
```yaml
# 旧配置
data_dir: ~/daily_summary/data
summary_dir: ~/daily_summary/summaries

# 新配置
data_dir: ./run/data
summary_dir: ./run/summaries
```

**迁移已有数据**（可选）：
```bash
# 如果你想保留旧数据
mkdir -p ./run
cp -r ~/daily_summary/data ./run/
cp -r ~/daily_summary/summaries ./run/

# 清理旧目录（可选）
rm -rf ~/daily_summary
```

### 兼容性

- ⚠️ **不兼容 v1.3.0**：路径发生变化
- ✅ **配置文件格式兼容**：只需更新路径即可
- ✅ **数据格式兼容**：可以直接迁移旧数据

### 相关文件

- [config.example.yaml](config.example.yaml) - 更新默认路径
- [launchd/com.humg.daily_summary.plist](launchd/com.humg.daily_summary.plist) - 更新日志路径
- [scripts/install.sh](scripts/install.sh) - 更新目录创建逻辑
- [main.go](main.go#L92-L97) - 更新日志路径
- [internal/cli/cli.go](internal/cli/cli.go) - 更新锁文件和信号文件路径
- [internal/scheduler/scheduler.go](internal/scheduler/scheduler.go) - 更新信号文件路径

---

## [v1.3.0] - 2026-01-20

### 新增功能

#### 智能定时器重置 🔄
- **自动顺延提醒时间**：执行 `add` 命令后自动重置定时器
  - 手动添加记录后，下一次提醒时间自动顺延一个完整周期
  - 避免刚添加完记录后立即收到提醒的尴尬
  - 提升用户体验，让工作流程更加自然流畅

#### 工作原理
- **进程间通信机制**：基于文件的轻量级信号传递
  - 信号文件：`./run/.reset_signal`（v1.4.0 更新）
  - add 命令创建信号文件，serve 进程监控并响应
  - 简单、跨平台、易于调试
- **实时响应**：1秒内检测到信号并重置计时器
- **非阻塞设计**：使用带缓冲通道，确保系统稳定性

#### 使用场景

**小时级调度示例**：
```
10:00 - 服务启动，下一次提醒: 11:00
10:30 - 执行 add 命令添加记录
10:30 - 自动重置，下一次提醒: 11:30 ⏰ (顺延 1 小时)
11:30 - 弹窗提醒
```

**分钟级调度示例**：
```
10:00 - 服务启动，下一次提醒: 10:15
10:10 - 执行 add 命令添加记录
10:10 - 自动重置，下一次提醒: 10:25 ⏰ (顺延 15 分钟)
10:25 - 弹窗提醒
```

### 技术改进

#### 调度器增强
- **新增 resetCh 通道**：接收重置信号
- **watchResetSignal()**：监控信号文件（每秒检查）
- **checkAndClearResetSignal()**：原子性检测并删除信号文件
- **runHourlyTask()**：支持接收重置信号并重新调度

#### CLI 增强
- **sendResetSignal()**：在 RunAdd 完成后发送重置信号
- **容错设计**：信号发送失败不影响主流程
- **日志完善**：详细记录重置过程

### 测试验证

#### 单元测试
- ✅ `TestResetSignalPath` - 信号文件路径验证
- ✅ `TestCheckAndClearResetSignal` - 信号检测和清除
- ✅ `TestResetChannel` - 通道机制验证
- ✅ `TestSendResetSignal` - 信号发送功能
- ✅ `TestSendResetSignalMultipleTimes` - 多次发送测试

#### 集成测试
- 提供 `test_reset.sh` 测试脚本
- 支持手动验证重置功能
- 详细的日志输出验证

### 故障处理

**信号文件创建失败**：
- 只记录日志，不影响 add 命令执行
- 最坏情况：提醒时间不重置，记录仍正常保存

**信号文件删除失败**：
- 记录日志并跳过本次重置
- 下次检查时重试
- 避免无限循环发送信号

### 技术细节

**为什么使用文件信号？**
- 简单性：无需额外依赖
- 跨平台：所有系统都支持
- 轻量级：适合低频信号
- 可观察：便于调试

**带缓冲通道设计**：
```go
resetCh: make(chan struct{}, 1)  // 容量为 1
```
- 避免阻塞
- 多个信号效果相同，只需一个

**1秒监控间隔**：
- 平衡响应速度和系统开销
- 对用户体验影响极小

### 相关文件
- [internal/scheduler/scheduler.go](internal/scheduler/scheduler.go#L270-L315) - 重置机制实现
- [internal/cli/cli.go](internal/cli/cli.go#L119-L140) - 信号发送
- [internal/scheduler/scheduler_test.go](internal/scheduler/scheduler_test.go) - 单元测试
- [internal/cli/cli_test.go](internal/cli/cli_test.go) - CLI 测试
- [test_reset.sh](test_reset.sh) - 集成测试脚本

### 兼容性

- ✅ **完全向后兼容** v1.2.0
- ✅ **配置文件不变**：无需修改配置
- ✅ **数据格式不变**：与现有记录完全兼容
- ✅ **自动启用**：无需配置，开箱即用

### 升级说明

从 v1.2.0 升级到 v1.3.0：

```bash
# 1. 更新代码
git pull
go build -o daily_summary

# 2. 重新安装服务
./scripts/install.sh

# 3. 测试新功能
./daily_summary add "测试智能重置功能"
# 观察日志，下一次提醒时间应该自动顺延
```

---

## [v1.2.0] - 2026-01-20

### 新增功能

#### 命令行手动记录 🚀
- **add 子命令**：随时手动添加工作记录
  - 用法：`daily_summary add "工作内容"`
  - 无需等待定时弹窗，立即记录完成的工作
  - 记录会立即保存并在下次弹窗中显示
- **list 子命令**：快速查看今日所有记录
  - 用法：`daily_summary list`
  - 清晰的列表格式显示：序号、时间、内容
  - 显示记录总数
- **serve 子命令**：明确启动后台服务（可选）
  - 用法：`daily_summary serve`
  - 保持向后兼容：直接运行 `daily_summary` 仍默认启动服务

#### 进程锁保护机制 🔒
- **防止多实例运行**：确保同一时间只有一个服务实例
  - 使用 PID 文件锁（`~/daily_summary/daily_summary.lock`）
  - 自动检测已运行的服务进程
  - 清晰的错误提示和解决建议
- **智能进程检测**：区分正常运行和僵死进程
  - 自动清理僵死进程的锁文件
  - 使用系统信号检测进程状态
- **友好提示**：当尝试重复启动时给出明确指引
  - 提示如何查看日志
  - 提示如何重启/停止服务

### 技术改进

#### 代码架构优化
- **新增 CLI 包**：`internal/cli/cli.go`
  - 统一的命令行处理逻辑
  - 清晰的职责分离
  - 便于测试和维护
- **重构 main.go**：子命令架构
  - 支持多种调用方式（无参数、子命令、flag）
  - 完整的帮助信息
  - 向后兼容性保证

#### 用户体验提升
- **即时反馈**：添加记录后立即显示确认信息
  - 显示记录内容和时间
  - 成功标识 ✓
- **清晰的帮助信息**：`daily_summary help`
  - 完整的用法说明
  - 示例命令
  - 使用建议

### 使用示例

**手动记录工作**：
```bash
# 添加工作记录
daily_summary add "完成需求文档审查"
daily_summary add "参加技术评审会议"

# 查看今日记录
daily_summary list
```

**后台服务**：
```bash
# 方式 1：默认启动（无变化）
daily_summary

# 方式 2：明确指定（新增）
daily_summary serve

# 方式 3：带配置文件（兼容）
daily_summary --config /path/to/config.yaml
```

**协同工作**：
```bash
# 安装后台服务（定时提醒）
./scripts/install.sh

# 随时手动添加记录
daily_summary add "临时会议讨论"

# 下次定时弹窗会显示所有记录（包括手动添加的）
```

### 兼容性

- ✅ **完全向后兼容** v1.1.0
- ✅ **launchd plist 无需修改**：现有后台服务继续正常工作
- ✅ **数据格式不变**：与现有记录完全兼容
- ✅ **配置文件不变**：无需修改配置

### 升级说明

从 v1.1.0 升级到 v1.2.0：

```bash
# 1. 更新代码
git pull
go build -o daily_summary

# 2. 重新安装服务（更新二进制文件）
./scripts/install.sh

# 3. 开始使用新功能
daily_summary add "升级到 v1.2.0"
daily_summary list
```

---

## [v1.1.0] - 2026-01-20

### 新增功能

#### 分钟级调度支持 ⭐
- 新增 `minute_interval` 配置参数，支持分钟级的提醒调度
- 可以设置每 N 分钟提醒一次（如 15分钟、30分钟）
- 优先级高于 `hourly_interval`，设置后会覆盖小时级调度
- 适合番茄工作法、敏捷开发等需要高频记录的场景

#### YAML 配置文件支持
- 支持 YAML 格式的配置文件（默认 `config.yaml`）
- 同时兼容 JSON 格式（`config.json`）
- 根据文件扩展名自动识别格式
- 提供详细的配置文件示例和注释

### 功能改进

#### 定时调度准确性优化 🎯
- **动态延迟检测阈值**：根据调度间隔智能调整跳过阈值
  - 小时级调度：延迟超过 5 分钟自动跳过
  - 分钟级调度：延迟超过调度间隔的 50% 自动跳过
- **Mac 睡眠场景优化**：设备睡眠导致的延迟调度会被正确跳过
  - 示例：14:00 调度时 Mac 睡眠，14:30 唤醒会跳过该次调度，15:00 正常调度
- **增强日志**：跳过日志现在显示延迟时长和阈值，方便排查

#### 工作记录弹窗增强 📝
- **显示今日历史记录**：弹窗自动展示今日所有工作记录
  - 清晰的列表格式：时间 + 内容
  - 方便快速回顾今天的工作进度
- **跳过空输入**：用户直接按回车或取消时不保存空记录
  - 避免产生无效的空记录
  - 提升数据质量
- **优化消息构建**：使用 `strings.Builder` 高效构建弹窗消息

### 配置增强

#### 新增配置参数
```yaml
minute_interval: 30  # 分钟级调度间隔
```

#### 预配置文件
- `config.example.yaml` - 标准配置示例
- `config.test-5min.yaml` - 5分钟测试配置
- `config.pomodoro.yaml` - 番茄工作法配置

### 文档更新

#### 新增文档
- `docs/CONFIGURATION.md` - 完整配置指南
- `docs/QUICK_START.md` - 快速开始指南
- `docs/MINUTE_SCHEDULING.md` - 分钟级调度详细说明

#### 更新文档
- `README.md` - 添加分钟级调度说明
- `CLAUDE.md` - 更新项目架构说明

### 测试工具

#### 新增测试脚本
- `scripts/test_minute_interval.sh` - 分钟级调度测试
- `scripts/test_config.sh` - 配置文件加载测试

### 技术改进

- 优化调度器逻辑，支持小时和分钟双模式
- 改进日志输出，显示当前使用的调度模式
- 添加 YAML 依赖：`gopkg.in/yaml.v3`

### 使用示例

**番茄工作法（每25分钟）：**
```yaml
minute_interval: 25
summary_time: "23:00"
```

**高频记录（每30分钟）：**
```yaml
minute_interval: 30
summary_time: "00:00"
```

**标准模式（每小时）：**
```yaml
hourly_interval: 1
summary_time: "00:00"
```

## [v1.0.0] - 2026-01-20

### 初始版本

#### 核心功能
- 每小时弹窗提醒记录工作内容
- macOS 原生 osascript 对话框
- JSON 格式数据存储
- 每日自动生成工作总结
- 集成 Claude Code CLI
- launchd 后台服务支持

#### 配置选项
- 小时级提醒间隔（`hourly_interval`）
- 每日总结生成时间（`summary_time`）
- 对话框超时设置（`dialog_timeout`）
- Claude Code CLI 路径配置

#### 文档
- README.md - 项目说明
- CLAUDE.md - 开发指南
- 安装/卸载脚本

#### 部署
- macOS launchd 集成
- 自动安装脚本
- 日志记录支持

---

## 升级指南

### 从 v1.0.0 升级到 v1.1.0

1. **更新代码**
```bash
git pull
go build -o daily_summary
```

2. **（可选）使用分钟级调度**

编辑配置文件 `~/.config/daily_summary/config.yaml`：
```yaml
# 添加此行启用分钟级调度
minute_interval: 30
```

3. **重启服务**
```bash
./scripts/install.sh
```

### 配置迁移

如果你使用的是 JSON 配置文件，可以继续使用，也可以转换为 YAML：

**旧的 JSON 格式（仍然支持）：**
```json
{
  "hourly_interval": 1,
  "summary_time": "00:00"
}
```

**新的 YAML 格式（推荐）：**
```yaml
hourly_interval: 1
summary_time: "00:00"
minute_interval: 30  # 新增功能
```

### 兼容性

- v1.1.0 完全向后兼容 v1.0.0
- 如果不设置 `minute_interval`，程序行为与 v1.0.0 完全相同
- 支持 JSON 和 YAML 两种配置格式
