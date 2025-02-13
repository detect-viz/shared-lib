package scheduler

import (
	"fmt"
	"sync"
	"time"

	"shared-lib/interfaces"
	"shared-lib/logger"
	"shared-lib/models"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// scheduler 排程管理器實現
type scheduler struct {
	cron   *cron.Cron
	jobs   map[string]jobInfo
	logger *logger.Logger
	config *models.SchedulerConfig
	mu     sync.RWMutex
}

// jobInfo 任務資訊
type jobInfo struct {
	id       cron.EntryID
	schedule time.Duration
	status   string
	lastErr  error
}

// NewScheduler 創建排程管理器
func NewScheduler(config *models.SchedulerConfig, log *logger.Logger) interfaces.Scheduler {
	return &scheduler{
		cron:   cron.New(cron.WithSeconds()),
		jobs:   make(map[string]jobInfo),
		logger: log,
		config: config,
	}
}

// RegisterTask 註冊任務
func (s *scheduler) RegisterTask(name string, schedule time.Duration, task interfaces.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 檢查是否已存在
	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("task %s already exists", name)
	}

	// 創建 cron 表達式
	spec := fmt.Sprintf("@every %s", schedule.String())

	// 包裝任務以記錄狀態
	wrappedJob := func() {
		info := s.jobs[name]
		info.lastErr = task.Execute()
		s.jobs[name] = info

		if info.lastErr != nil {
			s.logger.Error("任務執行失敗",
				zap.String("name", name),
				zap.Error(info.lastErr))
		}
	}

	// 添加任務
	id, err := s.cron.AddFunc(spec, wrappedJob)
	if err != nil {
		return fmt.Errorf("add task failed: %w", err)
	}

	s.jobs[name] = jobInfo{
		id:       id,
		schedule: schedule,
		status:   "registered",
	}

	s.logger.Info("任務註冊成功",
		zap.String("name", name),
		zap.Duration("schedule", schedule))

	return nil
}

// Start 啟動排程器
func (s *scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cron.Start()

	// 更新所有任務狀態
	for name, info := range s.jobs {
		info.status = "running"
		s.jobs[name] = info
	}

	s.logger.Info("排程器已啟動")
	return nil
}

// Stop 停止排程器
func (s *scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cron.Stop()

	// 更新所有任務狀態
	for name, info := range s.jobs {
		info.status = "stopped"
		s.jobs[name] = info
	}

	s.logger.Info("排程器已停止")
}

// GetTaskInfo 獲取任務資訊
func (s *scheduler) GetTaskInfo(name string) (interfaces.TaskInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.jobs[name]
	if !exists {
		return interfaces.TaskInfo{}, false
	}

	entry := s.cron.Entry(info.id)
	return interfaces.TaskInfo{
		Name:     name,
		Schedule: info.schedule,
		LastRun:  entry.Prev,
		NextRun:  entry.Next,
		Status:   info.status,
		Error:    info.lastErr,
	}, true
}

// ListTasks 列出所有任務
func (s *scheduler) ListTasks() map[string]interfaces.TaskInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make(map[string]interfaces.TaskInfo)
	for name, info := range s.jobs {
		entry := s.cron.Entry(info.id)
		tasks[name] = interfaces.TaskInfo{
			Name:     name,
			Schedule: info.schedule,
			LastRun:  entry.Prev,
			NextRun:  entry.Next,
			Status:   info.status,
			Error:    info.lastErr,
		}
	}
	return tasks
}

// PauseTask 暫停任務
func (s *scheduler) PauseTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	s.cron.Remove(info.id)
	info.status = "paused"
	s.jobs[name] = info

	return nil
}

// ResumeTask 恢復任務
func (s *scheduler) ResumeTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	if info.status != "paused" {
		return fmt.Errorf("task %s is not paused", name)
	}

	spec := fmt.Sprintf("@every %s", info.schedule.String())
	id, err := s.cron.AddFunc(spec, func() {})
	if err != nil {
		return fmt.Errorf("resume task failed: %w", err)
	}

	info.id = id
	info.status = "running"
	s.jobs[name] = info

	return nil
}

// RunTaskNow 立即執行任務
func (s *scheduler) RunTaskNow(name string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	entry := s.cron.Entry(info.id)
	entry.Job.Run()

	return nil
}
