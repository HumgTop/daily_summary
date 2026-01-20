# 更新日志

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
