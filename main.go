package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"humg.top/daily_summary/config"
	"humg.top/daily_summary/internal/cli"
	"humg.top/daily_summary/internal/dialog"
	"humg.top/daily_summary/internal/scheduler"
	"humg.top/daily_summary/internal/storage"
	"humg.top/daily_summary/internal/summary"
)

func main() {
	// 解析全局 flag
	globalFlags := flag.NewFlagSet("global", flag.ContinueOnError)
	configPath := globalFlags.String("config", getDefaultConfigPath(), "配置文件路径")
	globalFlags.SetOutput(os.Stderr)

	// 查找第一个非 flag 参数作为子命令，同时解析全局 flags
	subcommand := ""
	subcommandIndex := -1
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		// 跳过 flag 参数
		if strings.HasPrefix(arg, "-") {
			// 如果是带值的 flag（如 --config path），跳过下一个参数
			if i+1 < len(os.Args) && !strings.Contains(arg, "=") &&
			   (arg == "--config" || arg == "-config") {
				i++
			}
			continue
		}
		// 找到第一个非 flag 参数
		subcommand = arg
		subcommandIndex = i
		break
	}

	// 解析全局 flags（在子命令之前的所有参数）
	if subcommandIndex > 0 {
		globalFlags.Parse(os.Args[1:subcommandIndex])
	} else {
		globalFlags.Parse(os.Args[1:])
	}

	// 如果没有找到子命令，默认启动服务
	if subcommand == "" {
		runServeWithConfig(*configPath)
		return
	}

	// 处理子命令
	switch subcommand {
	case "serve":
		runServeWithConfig(*configPath)
	case "add":
		runAddWithConfig(*configPath, os.Args[subcommandIndex+1:])
	case "list":
		runListWithConfig(*configPath)
	case "summary":
		runSummaryWithConfig(*configPath, os.Args[subcommandIndex+1:])
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("未知命令: %s\n\n", subcommand)
		printHelp()
		os.Exit(1)
	}
}

// runServeWithConfig 启动后台服务
func runServeWithConfig(configPath string) {

	// 检查并获取进程锁
	if err := cli.CheckAndAcquireLock(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 确保退出时释放锁
	defer cli.ReleaseLock()

	// 加载配置
	var err error
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 确保目录存在
	if err := config.EnsureDirectories(cfg); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// 设置日志
	if cfg.EnableLogging {
		// 使用项目目录下的 run/logs
		logFile := filepath.Join("run", "logs", "app.log")
		// 确保日志目录存在
		os.MkdirAll(filepath.Dir(logFile), 0755)
		setupLogging(logFile)
	}

	log.Println("Daily Summary Tool starting...")
	log.Printf("Data directory: %s", cfg.DataDir)
	log.Printf("Summary directory: %s", cfg.SummaryDir)

	// 初始化组件
	dialogTimeout := time.Duration(cfg.DialogTimeout) * time.Second
	dlg := dialog.NewOSAScriptDialog(dialogTimeout)

	store := storage.NewJSONStorage(cfg.DataDir, cfg.SummaryDir)

	// 根据配置创建 AI 客户端
	var aiClient summary.AIClient

	// 默认使用 codex
	if cfg.AIProvider == "" || cfg.AIProvider == "codex" {
		codexPath := cfg.CodexPath
		if codexPath == "" {
			codexPath = "codex"
		}
		aiClient, err = summary.NewCodexClient(codexPath)
		if err != nil {
			log.Fatalf("Failed to create Codex client: %v", err)
		}
		log.Println("Using Codex for summary generation")
	} else if cfg.AIProvider == "claude" {
		var claudeClient *summary.ClaudeClient
		claudeClient, err = summary.NewClaudeClient(cfg.ClaudeCodePath)
		if err != nil {
			log.Fatalf("Failed to create Claude client: %v", err)
		}
		aiClient = claudeClient
		log.Println("Using Claude for summary generation")
	} else {
		log.Fatalf("Unknown AI provider: %s (supported: codex, claude)", cfg.AIProvider)
	}

	gen := summary.NewGenerator(store, aiClient)

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

// runAddWithConfig 添加工作记录
func runAddWithConfig(configPath string, args []string) {
	// 检查工作内容参数
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: 请提供工作内容")
		fmt.Fprintln(os.Stderr, "\n用法: daily_summary add \"工作内容\"")
		fmt.Fprintln(os.Stderr, "示例: daily_summary add \"完成需求文档审查\"")
		os.Exit(1)
	}

	content := args[0]

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 确保目录存在
	if err := config.EnsureDirectories(cfg); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// 初始化存储
	store := storage.NewJSONStorage(cfg.DataDir, cfg.SummaryDir)

	// 执行添加
	if err := cli.RunAdd(store, content); err != nil {
		log.Fatalf("Failed to add entry: %v", err)
	}
}

// runListWithConfig 查看今日记录
func runListWithConfig(configPath string) {
	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化存储
	store := storage.NewJSONStorage(cfg.DataDir, cfg.SummaryDir)

	// 执行列表
	if err := cli.RunList(store); err != nil {
		log.Fatalf("Failed to list entries: %v", err)
	}
}

// runSummaryWithConfig 生成工作总结
func runSummaryWithConfig(configPath string, args []string) {
	// 解析参数
	summaryCmd := flag.NewFlagSet("summary", flag.ExitOnError)
	dateStr := summaryCmd.String("date", "", "指定日期 (格式: 2006-01-02，默认今天)")
	summaryCmd.Parse(args)

	// 加载配置
	var err error
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 确定日期
	var targetDate time.Time
	if *dateStr == "" {
		targetDate = time.Now()
	} else {
		targetDate, err = time.Parse("2006-01-02", *dateStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: 无效的日期格式，应为 YYYY-MM-DD\n")
			os.Exit(1)
		}
	}

	// 初始化存储
	store := storage.NewJSONStorage(cfg.DataDir, cfg.SummaryDir)

	// 创建 AI 客户端
	var aiClient summary.AIClient
	if cfg.AIProvider == "" || cfg.AIProvider == "codex" {
		codexPath := cfg.CodexPath
		if codexPath == "" {
			codexPath = "codex"
		}
		aiClient, err = summary.NewCodexClient(codexPath)
		if err != nil {
			log.Fatalf("Failed to create Codex client: %v", err)
		}
	} else if cfg.AIProvider == "claude" {
		var claudeClient *summary.ClaudeClient
		claudeClient, err = summary.NewClaudeClient(cfg.ClaudeCodePath)
		if err != nil {
			log.Fatalf("Failed to create Claude client: %v", err)
		}
		aiClient = claudeClient
	} else {
		log.Fatalf("Unknown AI provider: %s", cfg.AIProvider)
	}

	// 创建生成器
	gen := summary.NewGenerator(store, aiClient)

	// 生成总结
	fmt.Printf("正在生成 %s 的工作总结...\n", targetDate.Format("2006-01-02"))
	if err := gen.GenerateDailySummary(targetDate); err != nil {
		fmt.Fprintf(os.Stderr, "Error: 生成总结失败: %v\n", err)
		os.Exit(1)
	}

	// 标记总结已生成
	if err := store.MarkSummaryGenerated(targetDate); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: 标记总结状态失败: %v\n", err)
		// 不终止程序，因为总结已经成功生成
	}

	// 构建总结文件路径
	summaryPath := filepath.Join(cfg.SummaryDir, targetDate.Format("2006-01-02")+".md")
	fmt.Printf("✓ 总结已生成并保存到: %s\n", summaryPath)
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println(`Daily Summary Tool - 工作记录助手

用法:
  daily_summary [命令] [选项]

命令:
  serve            启动后台服务（默认）
  add <content>    手动添加工作记录
  list             查看今日记录
  summary [--date] 生成工作总结
  help             显示此帮助信息

全局选项:
  --config PATH  配置文件路径 (默认: ~/.config/daily_summary/config.yaml)

示例:
  daily_summary                                    # 启动后台服务
  daily_summary serve                              # 启动后台服务
  daily_summary add "完成需求文档审查"              # 添加工作记录
  daily_summary list                               # 查看今日记录
  daily_summary summary                            # 生成今日总结
  daily_summary summary --date 2026-01-19          # 生成指定日期的总结
  daily_summary --config ~/my-config.yaml          # 使用自定义配置启动服务

说明:
  - 后台服务通过 install.sh 安装后会自动启动
  - 手动添加的记录会立即保存，并在下次定时弹窗中显示
  - 如果后台服务已在运行，执行 serve 命令会提示并退出`)
}

// getDefaultConfigPath 获取默认配置文件路径
func getDefaultConfigPath() string {
	// 使用项目目录下的 config.yaml
	return "config.yaml"
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
