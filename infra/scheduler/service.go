package scheduler

import (
	"fmt"
	"sync"

	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/google/wire"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// SchedulerSet 提供 Scheduler
var SchedulerSet = wire.NewSet(
	NewService,
	wire.Bind(new(Service), new(*serviceImpl)),
)

// Scheduler 排程管理器
type serviceImpl struct {
	cron   *cron.Cron
	tasks  map[string]models.TaskInfo
	logger logger.Logger
	mu     sync.RWMutex
}

// NewScheduler 創建排程管理器
func NewService(logger logger.Logger) *serviceImpl {
	return &serviceImpl{
		cron:   cron.New(cron.WithSeconds()),
		tasks:  make(map[string]models.TaskInfo),
		logger: logger.With(zap.String("module", "scheduler")), // 添加模組標識
	}
}

// 註冊任務
func (s *serviceImpl) RegisterTask(task models.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validateJob(task.Name, task.Spec); err != nil {
		return err
	}

	id, err := s.addTask(task.Spec, task.ExecFunc)
	if err != nil {
		return err
	}

	s.tasks[task.Name] = common.TaskInfo{
		ID:     id,
		Spec:   task.Spec,
		Status: "registered",
	}

	s.logger.Info("任務註冊成功",
		zap.String("name", task.Name),
		zap.String("spec", task.Spec))

	return nil
}

// validateJob 驗證任務
func (s *serviceImpl) validateJob(name, spec string) error {
	if _, exists := s.tasks[name]; exists {
		return fmt.Errorf("task %s already exists", name)
	}

	// 驗證 spec 格式
	if _, err := cron.ParseStandard(spec); err != nil {
		return fmt.Errorf("invalid schedule spec: %w", err)
	}

	return nil
}

// 添加任務
func (s *serviceImpl) addTask(spec string, fn func() error) (cron.EntryID, error) {
	id, err := s.cron.AddFunc(spec, func() {
		err := fn()
		if err != nil {
			s.logger.Error("任務執行錯誤", zap.Error(err))
		}
	})
	if err != nil {
		return 0, fmt.Errorf("add task failed: %w", err)
	}

	return id, nil
}

// Start 啟動排程器
func (s *serviceImpl) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cron.Start()

	// 更新所有任務狀態
	for name, info := range s.tasks {
		info.Status = "running"
		s.tasks[name] = info
	}

	s.logger.Info("排程器已啟動")
	return nil
}

// Stop 停止排程器
func (s *serviceImpl) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cron.Stop()

	// 更新所有任務狀態
	for name, info := range s.tasks {
		info.Status = "stopped"
		s.tasks[name] = info
	}

	s.logger.Info("排程器已停止")
}

// GetTaskInfo 獲取任務資訊
func (s *serviceImpl) GetTaskInfo(name string) (common.TaskInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.tasks[name]
	if !exists {
		return common.TaskInfo{}, false
	}

	entry := s.cron.Entry(info.ID)
	return common.TaskInfo{
		Name:    name,
		Spec:    info.Spec,
		LastRun: entry.Prev,
		NextRun: entry.Next,
		Status:  info.Status,
		Error:   info.Error,
	}, true
}

// ListTasks 列出所有任務
func (s *serviceImpl) ListTasks() map[string]common.TaskInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make(map[string]common.TaskInfo)
	for name, info := range s.tasks {
		entry := s.cron.Entry(info.ID)
		tasks[name] = common.TaskInfo{
			Name:    name,
			Spec:    info.Spec,
			LastRun: entry.Prev,
			NextRun: entry.Next,
			Status:  info.Status,
			Error:   info.Error,
		}
	}
	return tasks
}

// PauseTask 暫停任務
func (s *serviceImpl) PauseTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	s.cron.Remove(info.ID)
	info.Status = "paused"
	s.tasks[name] = info

	return nil
}

// ResumeTask 恢復任務
func (s *serviceImpl) ResumeTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	if info.Status != "paused" {
		return fmt.Errorf("task %s is not paused", name)
	}

	id, err := s.cron.AddFunc(info.Spec, func() {})
	if err != nil {
		return fmt.Errorf("resume task failed: %w", err)
	}

	info.ID = id
	info.Status = "running"
	s.tasks[name] = info

	return nil
}

// RunTaskNow 立即執行任務
func (s *serviceImpl) RunTaskNow(name string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	entry := s.cron.Entry(info.ID)
	entry.Job.Run()

	return nil
}
