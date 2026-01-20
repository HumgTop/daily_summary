#!/bin/bash

# 测试分钟级调度功能

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "========================================"
echo "Testing Minute-based Scheduling"
echo "========================================"
echo ""

# 检查编译
if [ ! -f "$PROJECT_DIR/daily_summary" ]; then
    echo "Binary not found. Building..."
    cd "$PROJECT_DIR"
    go build -o daily_summary
    echo "✓ Build complete"
    echo ""
fi

# 创建测试配置
TEST_CONFIG="/tmp/daily_summary_minute_test.yaml"
cat > "$TEST_CONFIG" <<EOF
data_dir: /tmp/daily_summary_test/data
summary_dir: /tmp/daily_summary_test/summaries
minute_interval: 5
summary_time: "23:59"
claude_code_path: claude-code
dialog_timeout: 60
enable_logging: false
EOF

echo "Created test configuration:"
echo "---"
cat "$TEST_CONFIG"
echo "---"
echo ""

# 创建测试目录
mkdir -p /tmp/daily_summary_test/data
mkdir -p /tmp/daily_summary_test/summaries

echo "========================================"
echo "Test Setup Complete!"
echo "========================================"
echo ""
echo "Configuration file: $TEST_CONFIG"
echo "Test data directory: /tmp/daily_summary_test"
echo ""
echo "To test the program:"
echo "  ./daily_summary --config $TEST_CONFIG"
echo ""
echo "Expected behavior:"
echo "  - Dialog will appear every 5 minutes"
echo "  - First dialog at next minute boundary (e.g., XX:05, XX:10, etc.)"
echo "  - Check logs to confirm minute-based scheduling is active"
echo ""
echo "To monitor:"
echo "  - Watch logs in real-time (if logging enabled)"
echo "  - Dialog should appear at XX:05, XX:10, XX:15, etc."
echo ""
echo "To clean up:"
echo "  rm -rf /tmp/daily_summary_test"
echo "  rm $TEST_CONFIG"
echo ""
