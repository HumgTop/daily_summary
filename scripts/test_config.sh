#!/bin/bash

# 测试 YAML 配置文件加载

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "====================================="
echo "Testing YAML Configuration"
echo "====================================="
echo ""

# 创建测试配置目录
TEST_CONFIG_DIR="/tmp/daily_summary_config_test"
mkdir -p "$TEST_CONFIG_DIR"

# 1. 测试 YAML 配置
echo "1. Testing YAML config..."
cat > "$TEST_CONFIG_DIR/test.yaml" <<EOF
data_dir: /tmp/test_data
summary_dir: /tmp/test_summaries
hourly_interval: 2
summary_time: "23:00"
claude_code_path: /usr/local/bin/claude-code
dialog_timeout: 600
enable_logging: false
EOF

echo "   Created YAML config:"
cat "$TEST_CONFIG_DIR/test.yaml"
echo ""

# 2. 测试 JSON 配置
echo "2. Testing JSON config..."
cat > "$TEST_CONFIG_DIR/test.json" <<EOF
{
  "data_dir": "/tmp/test_data_json",
  "summary_dir": "/tmp/test_summaries_json",
  "hourly_interval": 3,
  "summary_time": "22:00",
  "claude_code_path": "/opt/claude",
  "dialog_timeout": 120,
  "enable_logging": true
}
EOF

echo "   Created JSON config:"
cat "$TEST_CONFIG_DIR/test.json"
echo ""

echo "====================================="
echo "Configuration files created!"
echo "====================================="
echo ""
echo "Test files location: $TEST_CONFIG_DIR"
echo ""
echo "To test loading:"
echo "  YAML: ./daily_summary --config $TEST_CONFIG_DIR/test.yaml"
echo "  JSON: ./daily_summary --config $TEST_CONFIG_DIR/test.json"
echo ""
echo "To clean up: rm -rf $TEST_CONFIG_DIR"
echo ""
