#!/bin/bash

# 快速测试脚本 - 创建测试数据并验证程序功能

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "====================================="
echo "Daily Summary Tool - Quick Test"
echo "====================================="
echo ""

# 1. 检查编译
echo "1. Checking if binary exists..."
if [ ! -f "$PROJECT_DIR/daily_summary" ]; then
    echo "   Binary not found. Building..."
    cd "$PROJECT_DIR"
    go build -o daily_summary
    echo "   ✓ Build complete"
else
    echo "   ✓ Binary exists"
fi
echo ""

# 2. 创建测试目录
echo "2. Creating test directories..."
TEST_DIR="$HOME/daily_summary_test"
mkdir -p "$TEST_DIR/data"
mkdir -p "$TEST_DIR/summaries"
mkdir -p "$TEST_DIR/logs"
echo "   ✓ Test directories created at $TEST_DIR"
echo ""

# 3. 创建测试数据
echo "3. Creating test work entry data..."
TODAY=$(date +%Y-%m-%d)
YESTERDAY=$(date -v-1d +%Y-%m-%d)

cat > "$TEST_DIR/data/$YESTERDAY.json" <<EOF
{
  "date": "$YESTERDAY",
  "entries": [
    {
      "timestamp": "${YESTERDAY}T09:00:00+08:00",
      "content": "完成了用户认证模块的代码审查，修复了2个安全漏洞"
    },
    {
      "timestamp": "${YESTERDAY}T10:00:00+08:00",
      "content": "修复了前端登录页面的3个 UI bug"
    },
    {
      "timestamp": "${YESTERDAY}T11:00:00+08:00",
      "content": "参加团队站会，讨论本周的开发计划"
    },
    {
      "timestamp": "${YESTERDAY}T14:00:00+08:00",
      "content": "开发新的数据导出功能，完成基础框架"
    },
    {
      "timestamp": "${YESTERDAY}T18:00:00+08:00",
      "content": "编写了数据导出功能的单元测试，覆盖率达到85%"
    }
  ]
}
EOF
echo "   ✓ Test data created for $YESTERDAY"
echo ""

# 4. 测试 JSON 读取
echo "4. Testing JSON data reading..."
if [ -f "$TEST_DIR/data/$YESTERDAY.json" ]; then
    ENTRY_COUNT=$(cat "$TEST_DIR/data/$YESTERDAY.json" | grep -c "timestamp")
    echo "   ✓ Found $ENTRY_COUNT work entries in test data"
else
    echo "   ✗ Test data file not found"
    exit 1
fi
echo ""

echo "====================================="
echo "Test Setup Complete!"
echo "====================================="
echo ""
echo "Test data location: $TEST_DIR"
echo ""
echo "Next steps:"
echo "  1. View test data: cat $TEST_DIR/data/$YESTERDAY.json"
echo "  2. To test the program with test data, create a config file pointing to test directories"
echo "  3. Run: ./daily_summary --config /path/to/test/config.json"
echo ""
echo "To clean up test data: rm -rf $TEST_DIR"
echo ""
