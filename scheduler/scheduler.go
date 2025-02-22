package scheduler

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/detect-viz/shared-lib/interfaces"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler 排程管理器
type Scheduler struct {
	cron   *cron.Cron
	jobs   map[string]jobInfo
	logger interfaces.Logger // 使用接口而不是具體實現
	config *models.SchedulerConfig
	mu     sync.RWMutex
}

// NewScheduler 創建排程管理器
func NewScheduler(config *models.SchedulerConfig, logger interfaces.Logger) interfaces.Scheduler {
	return &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		jobs:   make(map[string]jobInfo),
		logger: logger.With(zap.String("module", "scheduler")), // 添加模組標識
		config: config,
	}
}

// jobInfo 任務資訊
type jobInfo struct {
	id       cron.EntryID
	schedule time.Duration
	status   string
	lastErr  error
}

// RegisterCronJob 註冊 cron 任務
func (s *Scheduler) RegisterCronJob(job models.SchedulerJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validateJob(job.Name, job.Spec); err != nil {
		return err
	}

	id, schedule, err := s.addCronJob(job.Spec, job.Func)
	if err != nil {
		return err
	}

	s.jobs[job.Name] = jobInfo{
		id:       id,
		schedule: schedule,
		status:   "registered",
	}

	s.logger.Info("任務註冊成功",
		zap.String("name", job.Name),
		zap.String("spec", job.Spec))

	return nil
}

// validateJob 驗證任務
func (s *Scheduler) validateJob(name, spec string) error {
	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	// 驗證 spec 格式
	if _, err := cron.ParseStandard(spec); err != nil {
		return fmt.Errorf("invalid schedule spec: %w", err)
	}

	return nil
}

// addCronJob 添加 cron 任務
func (s *Scheduler) addCronJob(spec string, fn func()) (cron.EntryID, time.Duration, error) {
	id, err := s.cron.AddFunc(spec, fn)
	if err != nil {
		return 0, 0, fmt.Errorf("add job failed: %w", err)
	}

	schedule, err := time.ParseDuration(strings.TrimPrefix(spec, "@every "))
	if err != nil {
		s.cron.Remove(id) // 清理已添加的任務
		return 0, 0, fmt.Errorf("parse schedule failed: %w", err)
	}

	return id, schedule, nil
}

// RegisterTask 註冊任務
func (s *Scheduler) RegisterTask(name string, schedule string, task interfaces.LogRotator) error {
	return s.RegisterCronJob(models.SchedulerJob{
		Name: name,
		Spec: schedule,
		Func: func() {
			task.ExecuteRotateTask(common.RotateTask{
				JobID:      name,
				SourcePath: name, // 或從其他地方獲取
				RotateSetting: common.RotateSetting{
					Schedule: schedule,
				},
			})
		},
	})
}

// Start 啟動排程器
func (s *Scheduler) Start() error {
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
func (s *Scheduler) Stop() {
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
func (s *Scheduler) GetTaskInfo(name string) (common.TaskInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.jobs[name]
	if !exists {
		return common.TaskInfo{}, false
	}

	entry := s.cron.Entry(info.id)
	return common.TaskInfo{
		Name:     name,
		Schedule: info.schedule,
		LastRun:  entry.Prev,
		NextRun:  entry.Next,
		Status:   info.status,
		Error:    info.lastErr,
	}, true
}

// ListTasks 列出所有任務
func (s *Scheduler) ListTasks() map[string]common.TaskInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make(map[string]common.TaskInfo)
	for name, info := range s.jobs {
		entry := s.cron.Entry(info.id)
		tasks[name] = common.TaskInfo{
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
func (s *Scheduler) PauseTask(name string) error {
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
func (s *Scheduler) ResumeTask(name string) error {
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
func (s *Scheduler) RunTaskNow(name string) error {
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
