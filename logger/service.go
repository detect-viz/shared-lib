package logger

import (
	"fmt"
	"shared-lib/interfaces"
	"sync"

	"go.uber.org/zap"
)

// TaskManager 任務管理器介面
type TaskManager interface {
	RegisterTask(name string, task interfaces.Task) error
	RemoveTask(name string) error
	ListTasks() []interfaces.TaskInfo
	GetTaskStatus(name string) (interfaces.TaskInfo, bool)
}

// defaultTaskManager 默認任務管理器實現
type defaultTaskManager struct {
	tasks map[string]interfaces.Task
	mu    sync.RWMutex
}

// NewTaskManager 創建任務管理器
func NewTaskManager() TaskManager {
	return &defaultTaskManager{
		tasks: make(map[string]interfaces.Task),
	}
}

// RegisterTask 註冊任務
func (m *defaultTaskManager) RegisterTask(name string, task interfaces.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[name]; exists {
		return fmt.Errorf("任務 %s 已存在", name)
	}
	m.tasks[name] = task
	return nil
}

// RemoveTask 移除任務
func (m *defaultTaskManager) RemoveTask(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[name]; !exists {
		return fmt.Errorf("任務 %s 不存在", name)
	}
	delete(m.tasks, name)
	return nil
}

// ListTasks 列出所有任務
func (m *defaultTaskManager) ListTasks() []interfaces.TaskInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tasks []interfaces.TaskInfo
	for name, task := range m.tasks {
		tasks = append(tasks, interfaces.TaskInfo{
			Name:     name,
			Schedule: task.GetSchedule(),
		})
	}
	return tasks
}

// GetTaskStatus 獲取任務狀態
func (m *defaultTaskManager) GetTaskStatus(name string) (interfaces.TaskInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[name]
	if !exists {
		return interfaces.TaskInfo{}, false
	}

	return interfaces.TaskInfo{
		Name:     name,
		Schedule: task.GetSchedule(),
	}, true
}

// Service 日誌服務
type Service struct {
	config    interfaces.Config
	scheduler interfaces.Scheduler
	logger    *Logger
	tasks     TaskManager
}

// NewService 創建日誌服務
func NewService(config interfaces.Config, scheduler interfaces.Scheduler) (*Service, error) {
	logConfig := config.GetLoggerConfig()
	// 使用新的 Logger 選項
	logger, err := NewLogger(&logConfig,
		WithCallerSkip(1),
		WithFields(
			zap.String("service", "logger"),
			zap.String("version", "1.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("創建 logger 失敗: %w", err)
	}

	svc := &Service{
		config:    config,
		scheduler: scheduler,
		logger:    logger,
		tasks:     NewTaskManager(),
	}

	return svc, nil
}

// GetLogger 獲取 Logger
func (s *Service) GetLogger() interfaces.Logger {
	// 不需要重新包裝，直接返回
	return s.logger
}

// ListTasks 列出所有日誌相關任務
func (s *Service) ListTasks() []interfaces.TaskInfo {
	tasks := s.tasks.ListTasks()
	schedTasks := s.scheduler.ListTasks()

	// 合併運行狀態
	for i, task := range tasks {
		if info, exists := schedTasks[task.Name]; exists {
			tasks[i].LastRun = info.LastRun
			tasks[i].NextRun = info.NextRun
			tasks[i].Status = info.Status
			tasks[i].Error = info.Error
		}
	}

	return tasks
}

// GetTaskStatus 獲取任務狀態
func (s *Service) GetTaskStatus(name string) (interfaces.TaskInfo, error) {
	// 先從 TaskManager 獲取任務信息
	info, exists := s.tasks.GetTaskStatus(name)
	if !exists {
		return interfaces.TaskInfo{}, fmt.Errorf("任務 %s 不存在", name)
	}

	// 再從 Scheduler 獲取運行狀態
	schedInfo, exists := s.scheduler.GetTaskInfo(name)
	if exists {
		info.LastRun = schedInfo.LastRun
		info.NextRun = schedInfo.NextRun
		info.Status = schedInfo.Status
		info.Error = schedInfo.Error
	}

	return info, nil
}

// RunCleanupNow 立即執行清理
func (s *Service) RunCleanupNow() error {
	return s.scheduler.RunTaskNow("log_cleanup")
}

// RunCompressNow 立即執行壓縮
func (s *Service) RunCompressNow() error {
	return s.scheduler.RunTaskNow("log_compress")
}

// Start 啟動服務
func (s *Service) Start() error {
	s.logger.Info("啟動日誌服務")

	// 啟動排程器
	if err := s.scheduler.Start(); err != nil {
		return fmt.Errorf("啟動排程器失敗: %w", err)
	}

	s.logger.Info("日誌服務已啟動")
	return nil
}

// Stop 停止服務
func (s *Service) Stop() {
	s.logger.Info("停止日誌服務")
	s.scheduler.Stop()
	if err := s.logger.Close(); err != nil {
		s.logger.Error("關閉日誌失敗", zap.Error(err))
	}
	s.logger.Info("日誌服務已停止")
}
