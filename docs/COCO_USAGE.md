# Coco AI 提供商使用指南

本文档说明如何在 Daily Summary 中使用 Coco 作为 AI 总结生成器。

## 功能说明

Coco 是第三个支持的 AI 提供商（继 Codex 和 Claude Code 之后）。Coco 使用非交互模式命令格式：

```bash
coco -p "prompt content"
```

## 配置方法

### 方式 1: 修改配置文件

编辑 `config.yaml` 文件：

```yaml
# AI 总结生成配置
# 可选值：codex, claude, coco
ai_provider: coco

# Coco CLI 可执行文件路径
# 如果 coco 在 PATH 中，直接写命令名即可
coco_path: coco

# 如果 coco 不在 PATH 中，需要提供完整路径
# coco_path: /usr/local/bin/coco
# 或者（Apple Silicon）
# coco_path: /opt/homebrew/bin/coco
```

### 方式 2: 使用环境变量

确保 `coco` 命令在系统 PATH 中可用：

```bash
which coco
# 输出：/usr/local/bin/coco 或类似路径
```

**注意**: 如果使用 launchd 后台服务，请使用绝对路径，因为 launchd 的 PATH 可能不完整。

## 验证配置

### 1. 测试 Coco 命令

手动测试 coco 是否正常工作：

```bash
coco -p "Hello, this is a test"
```

应该能看到 Coco 的响应。

### 2. 测试 Daily Summary 集成

使用 `go run` 命令测试：

```bash
# 切换到项目目录
cd /Users/bytedance/go/src/humg.top/daily_summary

# 测试配置加载
go run main.go --config ./config.yaml serve

# 查看日志确认 AI 提供商
tail -f ./run/logs/app.log
# 应该看到：Using Coco for summary generation
```

### 3. 手动生成测试总结

```bash
# 添加一些测试记录
./daily_summary add "完成 Coco 集成开发"
./daily_summary add "测试 Coco 总结生成功能"

# 手动生成今日总结
./daily_summary summary

# 查看生成的总结
cat run/summaries/$(date +%Y-%m-%d).md
```

## 工作原理

### 调用流程

1. **初始化**：`main.go` 根据 `ai_provider: coco` 创建 `CocoClient` 实例
2. **生成总结**：调用 `CocoClient.GenerateSummary(prompt)` 方法
3. **执行命令**：执行 `coco -p "{工作记录 prompt}"`
4. **返回结果**：捕获 stdout 作为 AI 生成的总结内容

### 回退机制

如果 Coco CLI 不可用（未安装或路径错误），系统会：

1. 记录警告日志：`Warning: coco not found, using fallback summary`
2. 生成简单模板总结（不会失败）
3. 继续正常运行

## 与其他 AI 提供商对比

| 特性 | Codex | Claude Code | Coco |
|------|-------|-------------|------|
| 命令格式 | `codex exec "{prompt}"` | `claude-code --prompt "{prompt}"` | `coco -p "{prompt}"` |
| 工作目录 | 项目目录 | 临时目录 | 项目目录 |
| 配置项 | `codex_path` | `claude_code_path` | `coco_path` |
| 回退机制 | ✅ | ✅ | ✅ |

## 常见问题

### Q1: Coco 不在 PATH 中怎么办？

**A**: 在配置文件中指定完整路径：

```yaml
coco_path: /usr/local/bin/coco
# 或根据实际安装位置调整
```

### Q2: 如何切换回 Codex？

**A**: 修改 `config.yaml`：

```yaml
ai_provider: codex
```

重启服务即可。

### Q3: Coco 执行失败如何排查？

**A**: 查看日志获取详细信息：

```bash
tail -f ./run/logs/app.log
```

日志中会显示：
- Coco 调用命令
- 工作目录
- 错误输出（如果有）

## 完整配置示例

```yaml
# 项目工作目录
work_dir: /Users/bytedance/go/src/humg.top/daily_summary

# 数据目录
data_dir: run/data
summary_dir: run/summaries

# 提醒设置
minute_interval: 45
summary_time: "11:00"

# AI 配置 - 使用 Coco
ai_provider: coco
coco_path: /opt/homebrew/bin/coco

# 其他配置
dialog_timeout: 3600
enable_logging: true
log_file: run/logs/app.log
max_log_size_mb: 10
```

## 实现细节

文件结构：
- `internal/summary/coco.go` - CocoClient 实现
- `internal/models/models.go` - 添加 `CocoPath` 配置字段
- `main.go` - 添加 coco 提供商初始化逻辑

核心代码：

```go
// 调用 coco -p "{prompt}"
cmd := exec.Command(cocoPath, "-p", prompt)
cmd.Dir = c.workDir
```

符合 `AIClient` 接口：

```go
type AIClient interface {
    GenerateSummary(prompt string) (string, error)
}
```

---

更新时间：2026-01-22
