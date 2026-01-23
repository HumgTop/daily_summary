package scheduler

import (
	"time"
)

// Task 任务接口
type Task interface {
	// ID 返回任务唯一标识
	ID() string

	// Name 返回任务名称
	Name() string

	// ShouldRun 判断是否应该执行（基于当前时间和任务配置）
	// 返回值：
	//   - shouldRun: 是否应该执行任务
	//   - newConfig: 如果需要更新配置（如延迟检测重新计算 NextRun），返回新的配置；否则返回 nil
	ShouldRun(now time.Time, config *TaskConfig) (shouldRun bool, newConfig *TaskConfig)

	// Execute 执行任务
	Execute() error

	// OnExecuted 任务执行后的回调（用于更新下次执行时间等）
	OnExecuted(now time.Time, config *TaskConfig, err error)
}

// TaskType 任务类型
type TaskType string

const (
	TaskTypeInterval TaskType = "interval" // 间隔型任务（如每 N 分钟）
	TaskTypeDaily    TaskType = "daily"    // 每日定时任务（如每天 11:00）
	TaskTypeOnce     TaskType = "once"     // 一次性任务
)

// TaskConfig 任务配置（存储在 JSON 文件中）
type TaskConfig struct {
	ID              string    `json:"id"`                         // 任务 ID
	Name            string    `json:"name"`                       // 任务名称
	Type            TaskType  `json:"type"`                       // 任务类型
	Enabled         bool      `json:"enabled"`                    // 是否启用
	IntervalMinutes int       `json:"interval_minutes,omitempty"` // 间隔分钟数（interval 类型）
	Time            string    `json:"time,omitempty"`             // 执行时间 HH:MM（daily 类型）
	NextRun         time.Time `json:"next_run,omitempty"`         // 下次执行时间（interval/once 类型）
	LastRun         time.Time `json:"last_run,omitempty"`         // 上次执行时间
	LastSuccess     time.Time `json:"last_success,omitempty"`     // 上次成功时间
	LastError       string    `json:"last_error,omitempty"`       // 上次错误信息

	// 业务特定数据（可选）
	Data map[string]interface{} `json:"data,omitempty"`
}

// TaskRegistry 任务注册表
type TaskRegistry struct {
	Tasks []*TaskConfig `json:"tasks"`
}
