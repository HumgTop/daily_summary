package scheduler

import (
	"log"
	"time"
)

// Scheduler 通用调度器（基于短周期检查）
type Scheduler struct {
	registry      *Registry           // 任务注册表
	tasks         map[string]Task     // 任务实例映射
	stopCh        chan struct{}       // 停止信号
	checkInterval time.Duration       // 检查间隔
	runDir        string              // 运行目录
}

// NewScheduler 创建调度器
func NewScheduler(runDir string) *Scheduler {
	return &Scheduler{
		registry:      NewRegistry(runDir),
		tasks:         make(map[string]Task),
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
			log.Printf("  - %s: %s (%s)", config.ID, config.Name, status)
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

	for _, config := range configs {
		if !config.Enabled {
			continue
		}

		// 第一段判断：基于 NextRun 的粗粒度时间检查
		if !config.NextRun.IsZero() && now.Before(config.NextRun) {
			continue
		}

		// 获取任务实例
		task, exists := s.tasks[config.ID]
		if !exists {
			log.Printf("Warning: task %s not registered", config.ID)
			continue
		}

		// 第二段判断：任务的细粒度业务逻辑检查
		if !task.ShouldRun(now, config) {
			continue
		}

		// 执行任务
		log.Printf("Executing task: %s (%s)", task.ID(), task.Name())
		err := task.Execute()

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
}

// GetRegistry 获取任务注册表（用于外部访问）
func (s *Scheduler) GetRegistry() *Registry {
	return s.registry
}
