# launchd PATH 环境变量问题与解决方案

## 问题描述

当 daily_summary 作为 launchd 服务运行时，无法找到 Homebrew 安装的 `codex` 命令，导致使用回退总结（质量较差）。

### 症状

```markdown
# 生成的总结质量差
## 主要完成的任务
（由于 Codex CLI 不可用，这是一个自动生成的简单总结）
根据今天的工作记录，完成了多项任务。
```

### 日志中的警告

```
2026/01/22 11:26:45 Warning: codex not found, using fallback summary
```

## 根本原因

**launchd 服务的 PATH 环境变量不包含 Homebrew 路径**

```bash
# 你的 shell PATH（包含 Homebrew）
/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:...

# launchd 默认 PATH（不包含 Homebrew）
/usr/bin:/bin:/usr/sbin:/sbin
```

因此，当程序使用相对路径 `codex` 时，`exec.LookPath("codex")` 在 launchd 环境中会失败。

## 解决方案对比

### 方案 A：修改 launchd 的 PATH 环境变量

**实现方式**：在 plist 文件中添加 `EnvironmentVariables`：

```xml
<key>EnvironmentVariables</key>
<dict>
    <key>PATH</key>
    <string>/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
</dict>
```

**优点**：
- ✅ 与命令行环境一致
- ✅ 对多个依赖命令有效

**缺点**：
- ❌ **硬编码路径**：Apple Silicon 用 `/opt/homebrew`，Intel 用 `/usr/local`
- ❌ **跨用户不兼容**：不同用户的 Homebrew 路径可能不同
- ❌ **安全风险**：扩大 PATH 范围可能引入非预期的命令
- ❌ **维护成本**：需要在 plist 中维护完整的 PATH
- ❌ **调试困难**：PATH 问题不易排查

### 方案 B：在配置文件中使用绝对路径 ✅（推荐）

**实现方式**：在 `config.yaml` 中指定绝对路径：

```yaml
codex_path: /opt/homebrew/bin/codex
```

**优点**：
- ✅ **明确清晰**：一眼看出使用哪个 codex
- ✅ **易于配置**：用户可轻松修改
- ✅ **跨环境一致**：无论在哪运行都使用同一个 codex
- ✅ **安全性高**：不修改系统 PATH，不引入其他命令
- ✅ **易于调试**：路径错误立即可见
- ✅ **支持多版本**：可以明确指定使用特定版本的 codex

**缺点**：
- ⚠️ 需要用户手动配置（但安装时可自动检测）

## 最佳实践建议

### 推荐配置

```yaml
# config.yaml
# Apple Silicon Mac
codex_path: /opt/homebrew/bin/codex

# Intel Mac
codex_path: /usr/local/bin/codex
```

### 自动检测脚本

可以在 `install.sh` 中添加自动检测：

```bash
# 自动检测 codex 路径
CODEX_PATH=$(which codex 2>/dev/null || echo "")
if [ -n "$CODEX_PATH" ]; then
    echo "检测到 codex: $CODEX_PATH"
    # 自动更新配置文件
    sed -i '' "s|^codex_path:.*|codex_path: $CODEX_PATH|" config.yaml
else
    echo "⚠ 未检测到 codex，请手动配置 config.yaml"
fi
```

## 为什么不推荐修改 launchd PATH

### 1. 跨平台兼容性问题

不同 Mac 架构的 Homebrew 路径不同：

| Mac 类型 | Homebrew 路径 |
|---------|--------------|
| Apple Silicon (M1/M2) | `/opt/homebrew` |
| Intel | `/usr/local` |

如果在 plist 中硬编码 `/opt/homebrew/bin`，在 Intel Mac 上会失败。

### 2. 安全性考虑

扩大 launchd 的 PATH 范围可能会：
- 执行非预期的同名命令
- 引入安全风险（如果 PATH 中有恶意软件）

### 3. "最小权限原则"

launchd 服务应该只能访问它明确需要的命令，而不是整个 Homebrew bin 目录。

### 4. 可维护性

配置文件比 plist 更容易修改：
- 用户友好的 YAML 格式 vs XML
- 可以添加注释说明
- 不需要 unload/load 服务

## 其他解决方案

### 方案 C：符号链接（不推荐）

```bash
sudo ln -s /opt/homebrew/bin/codex /usr/local/bin/codex
```

**问题**：
- ❌ 需要 sudo 权限
- ❌ 可能与系统命令冲突
- ❌ 难以维护

### 方案 D：Wrapper 脚本（过度设计）

创建一个包装脚本设置 PATH 再调用程序。

**问题**：
- ❌ 增加复杂性
- ❌ 额外的故障点
- ❌ 调试困难

## 总结

**推荐使用方案 B（配置文件绝对路径）**，原因：

1. **简单明了** - 配置即文档
2. **安全可靠** - 不修改系统环境
3. **易于调试** - 路径问题一目了然
4. **跨环境兼容** - 用户可根据实际情况配置
5. **最小权限** - 只访问需要的命令

如果有多个外部命令依赖，可以在配置文件中一一指定：

```yaml
codex_path: /opt/homebrew/bin/codex
claude_code_path: /usr/local/bin/claude-code
some_other_tool_path: /path/to/tool
```

这样既清晰又可控。

## 参考资料

- [launchd.plist(5) man page](https://www.manpagez.com/man/5/launchd.plist/)
- [Apple Technical Note TN2083: Daemons and Agents](https://developer.apple.com/library/archive/technotes/tn2083/)
- [Homebrew installation paths](https://docs.brew.sh/Installation)
