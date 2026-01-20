#!/bin/bash

# 简单测试 osascript 对话框

echo "Testing osascript dialog..."
echo ""

result=$(osascript -e 'display dialog "这是一个测试对话框。请输入一些文本：" default answer "测试内容" with title "工作记录测试" buttons {"取消", "确定"} default button "确定"' 2>&1)

if [ $? -eq 0 ]; then
    echo "✓ Dialog test successful!"
    echo "Result: $result"
else
    echo "✗ Dialog test failed or cancelled"
    echo "Error: $result"
fi
