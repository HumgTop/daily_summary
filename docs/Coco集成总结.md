# Coco AI 提供商集成总结

## ✅ 完成的工作

已成功为 Daily Summary 项目添加 **Coco** 作为第三个 AI 提供商支持，与现有的 Codex 和 Claude Code 并列。

## 📝 修改清单

### 1. 新增文件

#### `internal/summary/coco.go`
- 实现 `CocoClient` 结构体
- 实现 `AIClient` 接口的 `GenerateSummary` 方法
- 使用命令格式：`coco -p "prompt"`
- 包含回退机制（当 coco 不可用时）
- 在项目目录下执行命令

### 2. 修改文件

#### `internal/models/models.go`
- 在 `Config` 结构体中添加 `CocoPath` 字段
- 更新 `AIProvider` 注释，添加 `"coco"` 作为可选值

#### `main.go` 
修改两处 AI 客户端初始化逻辑：

**位置 1**: `runServeWithConfig` 函数（第 139-157 行）
```go
} else if cfg.AIProvider == "coco" {
    cocoPath := cfg.CocoPath
    if cocoPath == "" {
        cocoPath = "coco"
    }
    aiClient, err = summary.NewCocoClient(cocoPath, cfg.WorkDir)
    if err != nil {
        log.Fatalf("Failed to create Coco client: %v", err)
    }
    log.Println("Using Coco for summary generation")
}
```

**位置 2**: `runSummaryWithConfig` 函数（第 301-315 行）
- 相同的初始化逻辑，确保 `summary` 命令也支持 coco

#### `config.yaml`
- 更新注释：`ai_provider` 可选值添加 `coco`
- 添加 `coco_path` 配置项，默认值为 `coco`

#### `config.example.yaml`
- 添加 coco 配置说明和示例
- 与 config.yaml 保持一致

### 3. 新增文档

#### `docs/COCO_USAGE.md`
完整的使用指南，包含：
- 配置方法（配置文件和环境变量）
- 验证步骤（测试命令、集成测试）
- 工作原理（调用流程、回退机制）
- 与其他 AI 提供商对比
- 常见问题解答
- 完整配置示例
- 实现细节

## 🎯 技术实现

### 命令格式
```bash
coco -p "工作记录 prompt"
```

### 核心代码
```go
// CocoClient 实现 AIClient 接口
type CocoClient struct {
    cocoPath string
    workDir  string
}

func (c *CocoClient) GenerateSummary(prompt string) (string, error) {
    cmd := exec.Command(cocoPath, "-p", prompt)
    cmd.Dir = c.workDir
    // ... 执行并捕获输出
}
```

### 配置示例
```yaml
ai_provider: coco
coco_path: /opt/homebrew/bin/coco  # 或 "coco" 如果在 PATH 中
```

## ✅ 验证结果

### 编译测试
```bash
go build -o daily_summary
```
✅ 编译成功，无错误

### 代码结构
- ✅ 符合现有架构模式
- ✅ 实现 `AIClient` 接口
- ✅ 与 Codex/Claude 保持一致的错误处理
- ✅ 包含回退机制
- ✅ 添加详细日志输出

## 🔄 使用方法

### 启用 Coco

修改 `config.yaml`：
```yaml
ai_provider: coco
coco_path: coco  # 或完整路径
```

### 测试总结生成

```bash
# 添加工作记录
./daily_summary add "完成 Coco 集成"

# 生成总结
./daily_summary summary

# 查看日志
tail -f ./run/logs/app.log
```

预期日志输出：
```
Using Coco for summary generation
调用 Coco: coco -p
工作目录: /Users/bytedance/go/src/humg.top/daily_summary
等待 Coco 响应...
✓ Coco 响应成功，长度: XXX 字符
```

## 📊 与其他 AI 提供商对比

| 特性 | Codex | Claude Code | Coco |
|------|-------|-------------|------|
| 命令格式 | `codex exec "{prompt}"` | `claude-code --prompt "{prompt}"` | `coco -p "{prompt}"` |
| 工作目录 | 项目目录 | 临时目录 | 项目目录 |
| 配置项 | `codex_path` | `claude_code_path` | `coco_path` |
| 回退机制 | ✅ | ✅ | ✅ |
| 状态 | 默认 | 可选 | 可选 |

## 🔧 切换 AI 提供商

只需修改配置文件的一行：

```yaml
# 使用 Codex（默认）
ai_provider: codex

# 使用 Claude Code
ai_provider: claude

# 使用 Coco（新增）
ai_provider: coco
```

重启服务即可生效。

## 📌 注意事项

1. **PATH 设置**: 如果使用 launchd 后台服务，建议使用 coco 的绝对路径
2. **回退机制**: 如果 coco 不可用，会自动生成简单模板总结
3. **兼容性**: 完全兼容现有功能，不影响 codex 和 claude 的使用
4. **测试**: 建议先手动测试 `coco -p "test"` 确认命令可用

## 📦 文件清单

新增文件：
- `internal/summary/coco.go` (97 行)
- `docs/COCO_USAGE.md` (完整使用指南)

修改文件：
- `internal/models/models.go` (+1 字段)
- `main.go` (+22 行，两处修改)
- `config.yaml` (+6 行)
- `config.example.yaml` (+6 行)

总代码变更：约 **120 行**

---

**完成时间**: 2026-01-22 14:50
**编译状态**: ✅ 通过
**文档状态**: ✅ 完整
