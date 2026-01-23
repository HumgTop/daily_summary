package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Registry 任务注册表管理器
type Registry struct {
	filePath string
	registry *TaskRegistry
	mu       sync.RWMutex
}

// NewRegistry 创建任务注册表
func NewRegistry(runDir string) *Registry {
	return &Registry{
		filePath: filepath.Join(runDir, "tasks.json"),
		registry: &TaskRegistry{
			Tasks: make([]*TaskConfig, 0),
		},
	}
}

// Load 从文件加载任务配置
func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		// 文件不存在，使用空注册表
		r.registry = &TaskRegistry{
			Tasks: make([]*TaskConfig, 0),
		}
		return nil
	}

	// 读取文件
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return fmt.Errorf("failed to read tasks file: %w", err)
	}

	// 解析 JSON
	var registry TaskRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("failed to parse tasks file: %w", err)
	}

	r.registry = &registry
	return nil
}

// Save 保存任务配置到文件
func (r *Registry) Save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 序列化为 JSON
	data, err := json.MarshalIndent(r.registry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}

// GetTask 获取指定 ID 的任务配置
func (r *Registry) GetTask(id string) *TaskConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, task := range r.registry.Tasks {
		if task.ID == id {
			return task
		}
	}
	return nil
}

// GetAllTasks 获取所有任务配置
func (r *Registry) GetAllTasks() []*TaskConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 返回副本，避免并发问题
	tasks := make([]*TaskConfig, len(r.registry.Tasks))
	copy(tasks, r.registry.Tasks)
	return tasks
}

// UpdateTask 更新任务配置
func (r *Registry) UpdateTask(config *TaskConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, task := range r.registry.Tasks {
		if task.ID == config.ID {
			r.registry.Tasks[i] = config
			return nil
		}
	}

	return fmt.Errorf("task not found: %s", config.ID)
}

// AddTask 添加新任务
func (r *Registry) AddTask(config *TaskConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查是否已存在
	for _, task := range r.registry.Tasks {
		if task.ID == config.ID {
			return fmt.Errorf("task already exists: %s", config.ID)
		}
	}

	r.registry.Tasks = append(r.registry.Tasks, config)
	return nil
}

// RemoveTask 移除任务
func (r *Registry) RemoveTask(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, task := range r.registry.Tasks {
		if task.ID == id {
			r.registry.Tasks = append(r.registry.Tasks[:i], r.registry.Tasks[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("task not found: %s", id)
}
