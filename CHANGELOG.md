# 更新日志

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
