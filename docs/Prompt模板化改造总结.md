# Prompt 模板化改造总结

## 改造时间

2026-01-21

## 改造背景

原有的 prompt 构建使用硬编码字符串拼接方式，导致：
- 修改 prompt 需要改代码并重新编译
- 无法快速迭代优化 prompt 质量
- 不支持多种 prompt 风格切换

## 改造目标

将 prompt 构建改为基于外部 markdown 模板文件的模式，实现：
- 代码与内容分离
- 动态内容通过变量注入
- 支持多模板切换
- 快速迭代无需重新编译

## 主要改动

### 1. 新增模板文件

```
templates/
├── summary_prompt.md         # 详细版模板
└── summary_prompt_simple.md  # 简洁版模板（默认）
```

**模板特性**：
- 使用 Go `text/template` 标准语法
- 支持变量注入: `{{.Date}}`, `{{.EntryCount}}`, `{{.Entries}}`
- 支持循环: `{{range .Entries}}...{{end}}`
- Markdown 格式，易于阅读和编辑

### 2. 代码重构 (`internal/summary/generator.go`)

#### 新增数据结构

```go
// PromptData 模板数据结构
type PromptData struct {
    Date       string        // 日期
    EntryCount int          // 记录条数
    Entries    []PromptEntry // 工作记录列表
}

// PromptEntry 单条工作记录
type PromptEntry struct {
    Time    string  // 时间
    Content string  // 内容
}
```

#### 新增字段和方法

```go
type Generator struct {
    storage      storage.Storage
    aiClient     AIClient
    notifier     Notifier
    templatePath string // 新增：模板路径配置
}

// SetTemplatePath 设置自定义模板路径
func (g *Generator) SetTemplatePath(path string) {
    g.templatePath = path
}
```

#### 重构 buildPrompt 方法

**原逻辑**：硬编码字符串拼接
```go
func (g *Generator) buildPrompt(dailyData *models.DailyData) string {
    var builder strings.Builder
    builder.WriteString("请为以下工作记录生成...")
    // ... 大量字符串拼接
    return builder.String()
}
```

**新逻辑**：模板加载和渲染
```go
func (g *Generator) buildPrompt(dailyData *models.DailyData) string {
    // 1. 准备模板数据
    data := PromptData{...}
    
    // 2. 读取模板文件
    templateContent, err := os.ReadFile(templatePath)
    
    // 3. 解析模板
    tmpl, err := template.New("prompt").Parse(string(templateContent))
    
    // 4. 执行模板渲染
    var buf bytes.Buffer
    tmpl.Execute(&buf, data)
    
    return buf.String()
}
```

#### 新增降级方法

```go
// buildFallbackPrompt 降级方案：使用原有的硬编码逻辑
func (g *Generator) buildFallbackPrompt(dailyData *models.DailyData) string {
    // 保留原有的字符串拼接逻辑作为降级方案
}
```

### 3. 错误处理机制

实现了三层错误处理：
1. **模板文件读取失败** → 降级到硬编码逻辑
2. **模板解析失败** → 降级到硬编码逻辑
3. **模板执行失败** → 降级到硬编码逻辑

确保功能的鲁棒性和可用性。

### 4. 单元测试 (`internal/summary/generator_test.go`)

新增三个测试用例：
- `TestBuildPrompt`: 测试默认模板渲染
- `TestBuildPromptWithCustomTemplate`: 测试自定义模板
- `TestBuildPromptFallback`: 测试降级逻辑

### 5. 文档

创建 `docs/Prompt模板系统说明.md`，包含：
- 架构设计
- 使用方法
- 模板语法
- 最佳实践
- 实现细节
- 未来扩展方向

## 使用示例

### 默认使用（简洁版模板）

```go
generator := summary.NewGenerator(storage, aiClient, notifier)
// 自动使用 templates/summary_prompt_simple.md
```

### 使用详细版模板

```go
generator := summary.NewGenerator(storage, aiClient, notifier)
generator.SetTemplatePath("templates/summary_prompt.md")
```

### 自定义模板

```go
generator := summary.NewGenerator(storage, aiClient, notifier)
generator.SetTemplatePath("templates/my_custom_prompt.md")
```

## 技术亮点

### 1. 代码与内容分离

✅ **前**：修改 prompt 需要改 Go 代码 → 编译 → 测试  
✅ **后**：直接编辑 markdown 文件 → 测试（无需编译）

### 2. 模板复用

使用 Go 标准库的 `text/template`，成熟稳定，支持：
- 变量插值
- 条件判断
- 循环遍历
- 函数调用

### 3. 健壮的降级机制

任何模板相关错误都会自动降级到硬编码逻辑，确保功能始终可用。

### 4. 易于扩展

未来可以轻松支持：
- 配置文件指定默认模板
- 模板市场和版本管理
- 更多上下文变量注入
- 多语言模板

## 文件变更清单

### 新增文件

- ✅ `templates/summary_prompt.md` - 详细版模板
- ✅ `templates/summary_prompt_simple.md` - 简洁版模板
- ✅ `internal/summary/generator_test.go` - 单元测试
- ✅ `docs/Prompt模板系统说明.md` - 系统文档

### 修改文件

- ✅ `internal/summary/generator.go` - 重构 prompt 构建逻辑

## Git 提交建议

根据用户规则，建议使用以下 commit message：

```bash
feat:重构Prompt构建为模板化架构

主要改动:
- 新增 markdown 模板文件支持,实现代码与内容分离
- 重构 buildPrompt 方法,使用 text/template 渲染
- 添加降级机制,确保模板加载失败时的鲁棒性
- 新增单元测试覆盖模板功能
- 完善文档说明模板系统设计和使用方法

优势:
- 修改 prompt 无需重新编译代码
- 支持多模板切换和自定义
- 提高 prompt 迭代优化效率
```

## 待办事项（可选）

以下是可以进一步优化的方向：

1. **配置化模板选择**
   ```yaml
   # config.yaml
   summary:
     template:
       type: "simple"  # simple | detailed | custom
       custom_path: ""
   ```

2. **模板校验命令**
   ```bash
   ds template validate templates/my_template.md
   ```

3. **更多预设模板**
   - 周报模板
   - 月报模板
   - 项目总结模板

4. **扩展模板变量**
   - 统计数据（总耗时、任务数等）
   - 环境信息（Git 分支、项目名等）
   - 历史对比（与昨日、上周对比）

## 兼容性说明

- ✅ **向后兼容**：保留了原有的硬编码逻辑作为降级方案
- ✅ **无破坏性变更**：对外接口未改变，现有调用代码无需修改
- ✅ **渐进式增强**：默认行为与原来一致，可选择启用模板功能

## 测试建议

运行以下命令测试改造效果：

```bash
# 1. 运行单元测试
cd internal/summary
go test -v -run TestBuildPrompt

# 2. 生成实际总结测试
cd ../..
./daily_summary summary

# 3. 尝试自定义模板
# 编辑 templates/summary_prompt.md
# 在代码中设置 generator.SetTemplatePath("templates/summary_prompt.md")
# 重新运行 ./daily_summary summary
```

## 总结

本次改造成功实现了 prompt 构建的模板化，主要成果：

1. ✅ **提高可维护性**：模板独立管理，修改无需改代码
2. ✅ **增强灵活性**：支持多模板切换和自定义
3. ✅ **保障稳定性**：降级机制确保功能可用
4. ✅ **完善测试**：单元测试覆盖核心功能
5. ✅ **文档齐全**：详细的使用和设计文档

为未来的 prompt 优化和功能扩展奠定了良好基础。
