#!/bin/bash

# 日报目录迁移脚本
# 将 summaries/ 根目录下的日报文件迁移到 summaries/daily/ 子目录

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 读取配置获取 summaries 目录路径
SUMMARIES_DIR="$PROJECT_ROOT/run/summaries"

echo "日报文件迁移工具"
echo "================"
echo ""
echo "源目录: $SUMMARIES_DIR"
echo "目标目录: $SUMMARIES_DIR/daily"
echo ""

# 检查 summaries 目录是否存在
if [ ! -d "$SUMMARIES_DIR" ]; then
    echo "错误: summaries 目录不存在: $SUMMARIES_DIR"
    exit 1
fi

# 创建 daily 子目录（如果不存在）
mkdir -p "$SUMMARIES_DIR/daily"

# 查找所有日报文件（格式：YYYY-MM-DD.md）
# 匹配模式：4位数字-2位数字-2位数字.md
FILES=$(find "$SUMMARIES_DIR" -maxdepth 1 -type f -name '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9].md' 2>/dev/null || true)

if [ -z "$FILES" ]; then
    echo "未找到需要迁移的日报文件（格式：YYYY-MM-DD.md）"
    echo "迁移完成！"
    exit 0
fi

# 统计文件数量
FILE_COUNT=$(echo "$FILES" | wc -l | tr -d ' ')

echo "找到 $FILE_COUNT 个日报文件需要迁移："
echo "$FILES" | while read -r file; do
    basename "$file"
done
echo ""

# 询问用户确认
read -p "是否继续迁移？[y/N] " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "已取消迁移"
    exit 0
fi

# 执行迁移
echo ""
echo "开始迁移..."
MOVED_COUNT=0
FAILED_COUNT=0

echo "$FILES" | while read -r file; do
    filename=$(basename "$file")
    target="$SUMMARIES_DIR/daily/$filename"

    if [ -f "$target" ]; then
        echo "⚠️  跳过（目标文件已存在）: $filename"
        FAILED_COUNT=$((FAILED_COUNT + 1))
    else
        if mv "$file" "$target"; then
            echo "✓ 已迁移: $filename"
            MOVED_COUNT=$((MOVED_COUNT + 1))
        else
            echo "✗ 迁移失败: $filename"
            FAILED_COUNT=$((FAILED_COUNT + 1))
        fi
    fi
done

echo ""
echo "迁移完成！"
echo "成功迁移: $MOVED_COUNT 个文件"
if [ $FAILED_COUNT -gt 0 ]; then
    echo "跳过/失败: $FAILED_COUNT 个文件"
fi
echo ""
echo "新的目录结构："
echo "  $SUMMARIES_DIR/"
echo "  ├── daily/       # 日报文件"
echo "  └── weekly/      # 周报文件"
