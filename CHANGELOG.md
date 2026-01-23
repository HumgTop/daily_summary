# 更新日志

## [v1.6.0] - 2026-01-23

### 架构优化 🏗️

#### 2段式调度判断架构
- **问题**：daily-summary 任务 next_run 为零值，架构不一致
- **改进**：实现2段式判断提升性能和可维护性
  - 第一段（Scheduler）：基于 NextRun 粗粒度时间过滤，未到期任务直接跳过
  - 第二段（Task）：业务逻辑细粒度判断（检查数据、生成状态等）
- **修复**：
  - 初始化时正确设置 daily-summary 的 NextRun
  - OnExecuted() 更新 NextRun 为明天总结时间
  - 移除文件信号机制，CLI 直接修改 tasks.json
- **效果**：所有任务统一维护 NextRun，调度更高效，配置更直观

---

## [v1.5.1] - 2026-01-22

### 重要修复 🔧

#### 心跳唤醒检测失效问题
- **问题描述**：系统从睡眠状态唤醒后，心跳监控无法检测到唤醒事件
  - 日志中缺少 "=== System wake-up detected! ===" 记录
  - 定时任务无法重置，导致时间错乱
  - 只能通过延迟跳过机制被动处理，无法主动重置

- **根本原因**：Go 的 `time.Ticker` 在系统睡眠时会暂停
  - 原代码使用 `case now := <-ticker.C` 获取 ticker 的时间戳
  - ticker 在睡眠时暂停，唤醒后继续，时间戳不反映墙上时钟跳变
  - 从 ticker 角度看，两次 tick 之间只有正常的 10 秒间隔
  - 无法超过 20 秒阈值，导致唤醒检测失效

- **修复方案**：使用 `time.Now()` 获取墙上时钟时间
  - 改为 `case <-ticker.C` 只接收信号，不使用 ticker 的时间
  - 手动调用 `now := time.Now()` 获取当前墙上时钟
  - 墙上时钟能正确反映睡眠期间经过的真实时间
  - 现在可以正确检测到 12+ 小时的睡眠并触发重置

### 技术细节

**时间机制对比**：

| 方法 | 时间来源 | 睡眠时行为 | 唤醒后时间跳变 |
|------|---------|-----------|--------------|
| `case now := <-ticker.C` (旧) | Ticker 时间戳 | 暂停 | ❌ 不反映 |
| `time.Now()` (新) | 墙上时钟 | 继续 | ✅ 正确反映 |

**修复前的时间线**：
```
21:50:00 - 最后一次心跳
[系统睡眠 12+ 小时]
11:12:00 - 唤醒后第一次 tick
11:12:10 - 下一次 tick
elapsed = 11:12:10 - 11:12:00 = 10s ✗ 小于 20s 阈值
```

**修复后的时间线**：
```
21:50:00 - 最后一次心跳 (lastHeartbeat)
[系统睡眠 12+ 小时]
11:12:00 - 唤醒后获取 time.Now()
elapsed = 11:12:00 - 21:50:00 = 12h22m ✓ 触发唤醒检测！
```

### 相关文件

- [internal/scheduler/scheduler.go:357-360](internal/scheduler/scheduler.go#L357-L360) - 修复心跳时间获取逻辑
- [internal/scheduler/heartbeat_test.go](internal/scheduler/heartbeat_test.go) - 新增测试验证修复
- [docs/BUGFIX_HEARTBEAT_WAKEUP.md](docs/BUGFIX_HEARTBEAT_WAKEUP.md) - 详细技术分析文档

### 测试验证

```bash
# 运行心跳测试
go test -v ./internal/scheduler/ -run TestHeartbeat

# 真实睡眠测试
# 1. 重新构建并安装
go build -o daily_summary && ./scripts/install.sh

# 2. 让电脑进入睡眠 10+ 分钟
# 3. 唤醒后查看日志
tail -50 run/logs/app.log | grep "wake-up"
```

### 预期日志输出

修复后唤醒应该看到：
```
2026/01/22 11:12:xx === System wake-up detected! ===
2026/01/22 11:12:xx Last heartbeat: 12h22m ago (threshold: 20s)
2026/01/22 11:12:xx Handling system wake-up (sleep duration: 12h22m)...
2026/01/22 11:12:xx ✓ Hourly task reset signal sent
2026/01/22 11:12:xx ✓ Summary task reset signal sent
2026/01/22 11:12:xx === Wake-up handling completed ===
```

### 升级说明

从 v1.5.0 升级到 v1.5.1：

```bash
# 1. 更新代码
git pull
go build -o daily_summary

# 2. 重启服务
./scripts/uninstall.sh
./scripts/install.sh

# 3. 验证修复（可选）
# 让电脑睡眠后唤醒，检查日志确认唤醒检测正常工作
```

### 兼容性

- ✅ **完全向后兼容** v1.5.0
- ✅ **无需修改配置**
- ✅ **无需迁移数据**
- ✅ **运行时行为改进**：唤醒检测更可靠

### 最佳实践参考

这次修复揭示了一个重要的 Go 时间处理原则：

**检测时间跳变时，应该使用墙上时钟而非 Timer/Ticker：**

```go
// ❌ 错误：无法检测睡眠导致的时间跳变
case now := <-ticker.C:
    elapsed := now.Sub(lastTime)

// ✅ 正确：使用墙上时钟检测真实时间变化
case <-ticker.C:
    now := time.Now()
    elapsed := now.Sub(lastTime)
```

相关资源：
- Go time package: https://pkg.go.dev/time
- Go Blog - Timers: https://go.dev/blog/timers

---

## [v1.5.0] - 2026-01-21

### 重要变更

#### 心跳唤醒机制 💓
- **自动检测 Mac 睡眠唤醒**：彻底解决 Mac 睡眠导致的定时器失效问题
  - 每 10 秒心跳检测，唤醒阈值 20 秒
  - 检测到系统睡眠唤醒后自动重置所有定时器
  - 纯 Go 实现，无需外部依赖，跨平台支持
- **双重保护机制**：
  - 延迟检测：跳过延迟过长的调度（原有功能）
  - 唤醒检测：主动检测睡眠并重置定时器（新增功能）
- **全面重置**：唤醒后同时重置两个定时任务
  - Hourly Task（工作记录提醒）
  - Summary Task（每日总结生成）

### 技术实现

#### Scheduler 增强
- **新增字段**：
  - `summaryResetCh chan struct{}` - 总结任务重置通道
  - `lastHeartbeat time.Time` - 上次心跳时间
  - `heartbeatMu sync.Mutex` - 心跳时间锁
- **新增方法**：
  - `monitorHeartbeat()` - 心跳监控主循环（10秒间隔）
  - `handleWakeUp(duration)` - 唤醒事件处理
- **改进方法**：
  - `runDailySummaryTask()` - 添加 `summaryResetCh` 监听
  - `Start()` - 启动心跳监控 goroutine

#### 工作原理
```
正常情况：
10:00:00 心跳 → 10:00:10 心跳 → 10:00:20 心跳（间隔 10s，正常）

睡眠场景：
10:00:00 心跳 → [Mac 睡眠 2 分钟] → 10:02:10 心跳
                                      ↑
                                间隔 130s > 20s 阈值
                                检测到唤醒！触发重置
```

### 测试验证

#### 单元测试
- ✅ `TestHeartbeatMonitor` - 心跳监控和唤醒检测
- ✅ `TestWakeUpHandler` - 唤醒处理逻辑
- ✅ `TestSummaryResetChannel` - 总结任务重置通道

#### 真实睡眠测试
参考 [docs/HEARTBEAT_WAKEUP_TESTING.md](docs/HEARTBEAT_WAKEUP_TESTING.md) 进行验证

### 方案对比

| 特性 | serve (旧) | tick (launchd) | 心跳唤醒 (新) |
|-----|-----------|----------------|--------------|
| 睡眠兼容性 | ❌ timer 失效 | ✅ 系统调度 | ✅ 主动检测 |
| 智能重置 | ✅ 支持 | ❌ 不支持 | ✅ 支持 |
| 资源占用 | 中 | 低 | 中+ |
| 实现复杂度 | 低 | 中 | 低 |
| 跨平台 | ✅ | ❌ macOS | ✅ |

### 移除功能

#### 移除 launchd 定时调度模式
- 移除 `tick` 命令（不再需要）
- 删除 `launchd/com.humg.daily_summary.scheduled.plist`
- 删除 `scripts/install_scheduled.sh`
- 删除 `docs/LAUNCHD_SCHEDULING.md`

**原因**：心跳唤醒机制提供了更好的解决方案
- 保留智能重置功能（手动 add 后顺延）
- 跨平台支持
- 实现更简单，无需维护 plist 配置
- 程序持续运行，响应更快

### 新增文档

- `docs/HEARTBEAT_WAKEUP_TESTING.md` - 心跳唤醒测试指南
  - 真实睡眠测试步骤
  - 预期日志输出
  - 常见问题排查

### 使用示例

**启动服务后的日志**：
```
2026/01/21 11:31:34 Scheduler started
2026/01/21 11:31:34 Heartbeat monitor started (interval: 10s, threshold: 20s)
2026/01/21 11:31:34 Using minute-based scheduling: every 45 minute(s)
2026/01/21 11:31:34 Next reminder scheduled at 12:16:00
```

**Mac 睡眠唤醒后的日志**：
```
2026/01/21 11:37:15 === System wake-up detected! ===
2026/01/21 11:37:15 Last heartbeat: 2m15s ago (threshold: 20s)
2026/01/21 11:37:15 Handling system wake-up (sleep duration: 2m15s)...
2026/01/21 11:37:15 ✓ Hourly task reset signal sent
2026/01/21 11:37:15 ✓ Summary task reset signal sent
2026/01/21 11:37:15 === Wake-up handling completed ===
2026/01/21 11:37:15 Received reset signal, rescheduling next reminder
2026/01/21 11:37:15 Next reminder scheduled at 12:22:00  ← 重新计算！
2026/01/21 11:37:15 Summary task received wake-up signal, recalculating...
```

### 相关文件

- [internal/scheduler/scheduler.go](internal/scheduler/scheduler.go) - 心跳监控实现（+80 行）
- [internal/scheduler/scheduler_test.go](internal/scheduler/scheduler_test.go) - 单元测试（+75 行）
- [main.go](main.go) - 移除 tick 命令（-160 行）
- [docs/HEARTBEAT_WAKEUP_TESTING.md](docs/HEARTBEAT_WAKEUP_TESTING.md) - 测试文档

### 兼容性

- ✅ **完全向后兼容** v1.4.0
- ✅ **配置文件不变**：无需修改配置
- ✅ **数据格式不变**：与现有记录完全兼容
- ✅ **自动启用**：无需配置，开箱即用

### 升级说明

从 v1.4.0 升级到 v1.5.0：

```bash
# 1. 更新代码
git pull
go build -o daily_summary

# 2. 重启服务
./scripts/uninstall.sh
./scripts/install.sh

# 3. 验证心跳监控已启动
tail -f run/logs/app.log
# 应该看到：Heartbeat monitor started (interval: 10s, threshold: 20s)

# 4. 测试睡眠唤醒（可选）
# - 让 Mac 进入睡眠 1-2 分钟
# - 唤醒后查看日志确认检测到唤醒并重置了定时器
```

### 技术细节

**心跳间隔和阈值**：
- 心跳间隔：10 秒（可在代码中调整）
- 唤醒阈值：20 秒（2倍心跳间隔）
- 平衡了检测精度和系统开销

**非阻塞设计**：
- 使用 `select` + `default` 避免通道阻塞
- 带缓冲通道（容量为1）防止信号丢失

**Mutex 保护**：
- `heartbeatMu` 保护 `lastHeartbeat` 时间戳
- 避免并发读写竞态条件

---

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
