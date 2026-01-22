#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PLIST_SRC="$PROJECT_DIR/launchd/com.humg.daily_summary.plist"
PLIST_DEST="$HOME/Library/LaunchAgents/com.humg.daily_summary.plist"

echo "====================================="
echo "Daily Summary Tool - Installation"
echo "====================================="
echo ""

# 编译程序
echo "Building daily_summary..."
cd "$PROJECT_DIR"
go build -o daily_summary
echo "✓ Build complete"
echo ""

# 创建必要的目录
echo "Creating directories..."
mkdir -p "$HOME/.config/daily_summary"
mkdir -p "$PROJECT_DIR/run/data"
mkdir -p "$PROJECT_DIR/run/summaries"
mkdir -p "$PROJECT_DIR/run/logs"
echo "✓ Directories created"
echo ""

# 复制 plist 文件
echo "Installing launchd service..."
cp "$PLIST_SRC" "$PLIST_DEST"
echo "✓ Plist copied to $PLIST_DEST"
echo ""

# 卸载旧服务（如果存在）
echo "Reloading service..."
launchctl unload "$PLIST_DEST" 2>/dev/null || true

# 等待旧进程完全退出（最多等待 5 秒）
LOCK_FILE="$PROJECT_DIR/run/daily_summary.lock"
if [ -f "$LOCK_FILE" ]; then
    echo "Waiting for old process to exit..."
    for i in {1..10}; do
        if [ ! -f "$LOCK_FILE" ]; then
            echo "✓ Old process exited"
            break
        fi
        sleep 0.5
    done

    # 如果锁文件仍存在，强制删除（可能是僵死进程）
    if [ -f "$LOCK_FILE" ]; then
        echo "⚠ Old process did not exit cleanly, removing stale lock file"
        rm -f "$LOCK_FILE"
    fi
fi

launchctl load "$PLIST_DEST"
echo "✓ Service loaded"
echo ""

echo "====================================="
echo "Installation complete!"
echo "====================================="
echo ""
echo "The service is now running in the background."
echo "It will automatically start on system boot."
echo ""
echo "Useful commands:"
echo "  - Check status: launchctl list | grep daily_summary"
echo "  - View logs: tail -f $PROJECT_DIR/run/logs/stdout.log"
echo "  - View app log: tail -f $PROJECT_DIR/run/logs/app.log"
echo "  - Uninstall: ./scripts/uninstall.sh"
echo ""
