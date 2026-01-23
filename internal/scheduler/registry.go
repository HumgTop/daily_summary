package scheduler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Registry 任务注册表管理器（基于文件的数据存储，确保数据一致性）
type Registry struct {
	filePath string     // JSON 文件路径
	mu       sync.Mutex // 文件操作互斥锁
}

// NewRegistry 创建任务注册表
func NewRegistry(runDir string) *Registry {
	return &Registry{
		filePath: filepath.Join(runDir, "tasks.json"),
	}
}

// load 从文件加载任务配置（内部方法，调用者需持有锁）
func (r *Registry) load() (*TaskRegistry, error) {
	// 检查文件是否存在
	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		// 文件不存在，返回空注册表
		return &TaskRegistry{
			Tasks: make([]*TaskConfig, 0),
		}, nil
	}

	// 读取文件
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}

	// 解析 JSON
	var registry TaskRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse tasks JSON: %w", err)
	}

	return &registry, nil
}

// save 保存任务配置到文件（内部方法，调用者需持有锁）
func (r *Registry) save(registry *TaskRegistry) error {
	// 序列化为 JSON
	data, err := json.MarshalIndent(registry, "", "  ")
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

// Load 从文件加载任务配置（公开方法，用于初始化验证）
func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 只是验证文件可以正常加载
	_, err := r.load()
	return err
}

// GetTask 获取指定任务配置
func (r *Registry) GetTask(id string) *TaskConfig {
	r.mu.Lock()
	defer r.mu.Unlock()

	registry, err := r.load()
	if err != nil {
		return nil
	}

	for _, task := range registry.Tasks {
		if task.ID == id {
			return task
		}
	}

	return nil
}

// GetAllTasks 获取所有任务配置
func (r *Registry) GetAllTasks() []*TaskConfig {
	r.mu.Lock()
	defer r.mu.Unlock()

	registry, err := r.load()
	if err != nil {
		return make([]*TaskConfig, 0)
	}

	// 返回副本
	tasks := make([]*TaskConfig, len(registry.Tasks))
	copy(tasks, registry.Tasks)
	return tasks
}

// PatchTask 增量更新任务配置（安全模式）
// 通过回调函数修改最新配置，避免并发覆盖问题。
func (r *Registry) PatchTask(id string, updateFunc func(task *TaskConfig)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 加载现有数据
	registry, err := r.load()
	if err != nil {
		return err
	}

	// 查找并更新
	found := false
	for _, task := range registry.Tasks {
		if task.ID == id {
			// 在最新配置上执行更新
			updateFunc(task)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task not found: %s", id)
	}

	// 保存回文件
	return r.save(registry)
}

// AddTask 添加任务配置
func (r *Registry) AddTask(config *TaskConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 加载现有数据
	registry, err := r.load()
	if err != nil {
		return err
	}

	// 检查是否已存在
	for _, task := range registry.Tasks {
		if task.ID == config.ID {
			return fmt.Errorf("task already exists: %s", config.ID)
		}
	}

	// 添加新任务
	registry.Tasks = append(registry.Tasks, config)

	// 保存回文件
	return r.save(registry)
}

// RemoveTask 移除任务配置
func (r *Registry) RemoveTask(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 加载现有数据
	registry, err := r.load()
	if err != nil {
		return err
	}

	// 查找并移除
	found := false
	newTasks := make([]*TaskConfig, 0, len(registry.Tasks))
	for _, task := range registry.Tasks {
		if task.ID != id {
			newTasks = append(newTasks, task)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("task not found: %s", id)
	}

	registry.Tasks = newTasks

	// 保存回文件
	return r.save(registry)
}
