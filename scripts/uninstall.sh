#!/bin/bash

set -e

PLIST_DEST="$HOME/Library/LaunchAgents/com.humg.daily_summary.plist"

echo "====================================="
echo "Daily Summary Tool - Uninstallation"
echo "====================================="
echo ""

# 卸载服务
echo "Stopping and unloading service..."
launchctl unload "$PLIST_DEST" 2>/dev/null || true
rm -f "$PLIST_DEST"
echo "✓ Service unloaded"
echo ""

echo "====================================="
echo "Uninstallation complete!"
echo "====================================="
echo ""
echo "Note: Data files are preserved in ~/daily_summary/"
echo "To remove all data, run: rm -rf ~/daily_summary"
echo ""
