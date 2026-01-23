package scheduler

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Scheduler 通用调度器（基于短周期检查）
type Scheduler struct {
	registry      *Registry       // 任务注册表
	tasks         map[string]Task // 任务实例映射
	runningTasks  map[string]bool // 正在执行的任务标记
	runningMu     sync.Mutex      // 保护 runningTasks 的互斥锁
	checkLogger   *log.Logger     // 调度检查专用日志记录器
	stopCh        chan struct{}   // 停止信号
	checkInterval time.Duration   // 检查间隔
	runDir        string          // 运行目录
}

// NewScheduler 创建调度器
// maxLogSizeMB: 调度器检查日志文件最大大小（MB），0 表示不限制
func NewScheduler(runDir string, maxLogSizeMB int) *Scheduler {
	// 创建日志目录
	logDir := filepath.Join(runDir, "logs")
	os.MkdirAll(logDir, 0755)

	// 创建调度检查日志文件
	checkLogPath := filepath.Join(logDir, "scheduler_check.log")

	// 检查日志文件大小并执行轮转
	if maxLogSizeMB > 0 {
		if err := rotateSchedulerLogIfNeeded(checkLogPath, maxLogSizeMB); err != nil {
			log.Printf("Warning: failed to rotate scheduler check log: %v", err)
		}
	}

	checkLogFile, err := os.OpenFile(checkLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: failed to open scheduler check log file: %v", err)
		checkLogFile = nil
	}

	// 创建专用的检查日志记录器
	var checkLogger *log.Logger
	if checkLogFile != nil {
		checkLogger = log.New(checkLogFile, "", log.LstdFlags)
	} else {
		// 如果无法创建文件，使用标准日志
		checkLogger = log.Default()
	}

	return &Scheduler{
		registry:      NewRegistry(runDir),
		tasks:         make(map[string]Task),
		runningTasks:  make(map[string]bool),
		checkLogger:   checkLogger,
		stopCh:        make(chan struct{}),
		checkInterval: 1 * time.Minute, // 固定 1 分钟检查间隔
		runDir:        runDir,
	}
}

// RegisterTask 注册任务
func (s *Scheduler) RegisterTask(task Task) {
	s.tasks[task.ID()] = task
	log.Printf("Task registered: %s (%s)", task.ID(), task.Name())
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	log.Println("Scheduler started with task-based scheduling (check interval: 1 minute)")

	// 加载任务配置
	if err := s.registry.Load(); err != nil {
		log.Printf("Warning: failed to load tasks registry: %v", err)
	}

	// 打印已注册的任务
	configs := s.registry.GetAllTasks()
	if len(configs) > 0 {
		log.Printf("Loaded %d task(s) from registry:", len(configs))
		for _, config := range configs {
			status := "disabled"
			if config.Enabled {
				status = "enabled"
			}
			// 打印任务基本信息和下一次调度时间
			if config.Enabled && !config.NextRun.IsZero() {
				log.Printf("  - %s: %s (%s, next run: %s)",
					config.ID, config.Name, status, config.NextRun.Format("2006-01-02 15:04:05"))
			} else {
				log.Printf("  - %s: %s (%s)", config.ID, config.Name, status)
			}
		}
	} else {
		log.Println("No tasks found in registry, will initialize from config")
	}

	// 启动调度循环
	go s.runScheduler()

	// 等待停止信号
	<-s.stopCh
	log.Println("Scheduler stopped")
	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	close(s.stopCh)
}

// runScheduler 调度循环
func (s *Scheduler) runScheduler() {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAndRunTasks()

		case <-s.stopCh:
			log.Println("Scheduler main loop exiting")
			return
		}
	}
}

// checkAndRunTasks 检查并执行所有到期的任务
func (s *Scheduler) checkAndRunTasks() {
	now := time.Now()
	configs := s.registry.GetAllTasks()

	// 记录检查周期开始
	s.checkLogger.Printf("[CHECK] Starting task check at %s", now.Format("2006-01-02 15:04:05"))

	for _, config := range configs {
		// 记录每个任务的检查状态
		if !config.Enabled {
			s.checkLogger.Printf("[SKIP] Task %s (%s): disabled", config.ID, config.Name)
			continue
		}

		// 第一段判断：基于 NextRun 的粗粒度时间检查
		if !config.NextRun.IsZero() && now.Before(config.NextRun) {
			s.checkLogger.Printf("[SKIP] Task %s (%s): not yet time (NextRun: %s)",
				config.ID, config.Name, config.NextRun.Format("2006-01-02 15:04:05"))
			continue
		}

		// 获取任务实例
		task, exists := s.tasks[config.ID]
		if !exists {
			s.checkLogger.Printf("[ERROR] Task %s not registered", config.ID)
			log.Printf("Warning: task %s not registered", config.ID)
			continue
		}

		// 第二段判断：任务的细粒度业务逻辑检查
		if !task.ShouldRun(now, config) {
			s.checkLogger.Printf("[SKIP] Task %s (%s): ShouldRun() returned false", config.ID, config.Name)
			continue
		}

		// 检查任务是否正在执行中（防止重复触发）
		s.runningMu.Lock()
		if s.runningTasks[config.ID] {
			s.runningMu.Unlock()
			s.checkLogger.Printf("[SKIP] Task %s (%s): already running", config.ID, config.Name)
			log.Printf("Task %s is already running, skipping this cycle", config.ID)
			continue
		}
		// 标记任务为执行中
		s.runningTasks[config.ID] = true
		s.runningMu.Unlock()

		// 执行任务
		s.checkLogger.Printf("[EXECUTE] Task %s (%s): starting execution", config.ID, config.Name)
		log.Printf("Executing task: %s (%s)", task.ID(), task.Name())
		err := task.Execute()

		// 清除执行状态标记
		s.runningMu.Lock()
		delete(s.runningTasks, config.ID)
		s.runningMu.Unlock()

		if err != nil {
			s.checkLogger.Printf("[EXECUTE] Task %s (%s): execution failed - %v", config.ID, config.Name, err)
		} else {
			s.checkLogger.Printf("[EXECUTE] Task %s (%s): execution completed successfully", config.ID, config.Name)
		}

		// 回调处理（更新配置）
		task.OnExecuted(now, config, err)

		// 保存更新后的配置
		if err := s.registry.UpdateTask(config); err != nil {
			log.Printf("Failed to update task config: %v", err)
		}

		// 保存注册表到文件
		if err := s.registry.Save(); err != nil {
			log.Printf("Failed to save task registry: %v", err)
		}
	}

	// 记录检查周期结束
	s.checkLogger.Printf("[CHECK] Task check completed at %s\n", now.Format("2006-01-02 15:04:05"))
}

// GetRegistry 获取任务注册表（用于外部访问）
func (s *Scheduler) GetRegistry() *Registry {
	return s.registry
}

// rotateSchedulerLogIfNeeded 检查调度器日志文件大小，如果超过限制则进行轮转
func rotateSchedulerLogIfNeeded(logFile string, maxSizeMB int) error {
	// 检查文件是否存在
	info, err := os.Stat(logFile)
	if os.IsNotExist(err) {
		// 文件不存在，无需轮转
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat scheduler log file: %w", err)
	}

	// 计算文件大小（字节转 MB）
	fileSizeMB := float64(info.Size()) / (1024 * 1024)

	// 如果文件大小未超过限制，无需轮转
	if fileSizeMB <= float64(maxSizeMB) {
		return nil
	}

	// 执行日志轮转：重命名为 .old
	oldLogFile := logFile + ".old"

	// 如果 .old 文件已存在，先删除
	if _, err := os.Stat(oldLogFile); err == nil {
		if err := os.Remove(oldLogFile); err != nil {
			return fmt.Errorf("remove old backup: %w", err)
		}
	}

	// 重命名当前日志文件为 .old
	if err := os.Rename(logFile, oldLogFile); err != nil {
		return fmt.Errorf("rename scheduler log file: %w", err)
	}

	log.Printf("Scheduler check log file rotated: %s (%.2f MB) -> %s", logFile, fileSizeMB, oldLogFile)
	return nil
}
