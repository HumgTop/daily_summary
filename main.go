package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"humg.top/daily_summary/config"
	"humg.top/daily_summary/internal/dialog"
	"humg.top/daily_summary/internal/scheduler"
	"humg.top/daily_summary/internal/storage"
	"humg.top/daily_summary/internal/summary"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", getDefaultConfigPath(), "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 确保目录存在
	if err := config.EnsureDirectories(cfg); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// 设置日志
	if cfg.EnableLogging {
		homeDir, _ := os.UserHomeDir()
		logFile := filepath.Join(homeDir, "daily_summary", "logs", "app.log")
		setupLogging(logFile)
	}

	log.Println("Daily Summary Tool starting...")
	log.Printf("Data directory: %s", cfg.DataDir)
	log.Printf("Summary directory: %s", cfg.SummaryDir)

	// 初始化组件
	dialogTimeout := time.Duration(cfg.DialogTimeout) * time.Second
	dlg := dialog.NewOSAScriptDialog(dialogTimeout)

	store := storage.NewJSONStorage(cfg.DataDir, cfg.SummaryDir)

	claudeClient, err := summary.NewClaudeClient(cfg.ClaudeCodePath)
	if err != nil {
		log.Fatalf("Failed to create Claude client: %v", err)
	}

	gen := summary.NewGenerator(store, claudeClient)

	sched := scheduler.NewScheduler(cfg, dlg, store, gen)

	// 启动调度器
	go func() {
		if err := sched.Start(); err != nil {
			log.Fatalf("Scheduler error: %v", err)
		}
	}()

	log.Println("Daily Summary Tool is now running. Press Ctrl+C to stop.")

	// 等待信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	log.Println("Shutting down...")
	sched.Stop()
	time.Sleep(1 * time.Second) // 给予一点时间完成清理
	log.Println("Goodbye!")
}

// getDefaultConfigPath 获取默认配置文件路径
func getDefaultConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "daily_summary", "config.yaml")
}

// setupLogging 设置日志输出
func setupLogging(logFile string) {
	// 确保日志目录存在
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
		return
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	log.SetOutput(f)
}
