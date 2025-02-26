package archiver

import (
	"fmt"
	"os"

	"github.com/detect-viz/shared-lib/models/common"
	"go.uber.org/zap"
)

// RotateTask 管理日誌輪轉
type RotateTask struct {
	service *Service
}

// NewRotateTask 建立 RotateTask
func NewRotateTask(service *Service) *RotateTask {
	return &RotateTask{service: service}
}

// 註冊輪轉任務
func (r *RotateTask) RegisterRotateTask(archiverTask common.RotateTask) error {
	task := archiverTask.Task
	taskSetting := common.Task{
		Enabled:     true,
		Name:        task.Name,
		Timezone:    task.Timezone,
		Description: task.Description,
		RetryCount:  task.RetryCount,
		RetryDelay:  task.RetryDelay,
		Duration:    task.Duration,
		Spec:        task.Spec,
		Type:        task.Type,
		ExecFunc: func() error {
			return r.ExecuteRotateTask(archiverTask)
		},
	}
	err := r.service.scheduler.RegisterTask(taskSetting)
	if err != nil {
		return fmt.Errorf("註冊輪轉任務失敗: %w", err)
	}

	return nil
}

// ExecuteRotateTask 實現輪轉任務
func (r *RotateTask) ExecuteRotateTask(task common.RotateTask) error {
	// 1. 檢查並創建目標目錄
	if err := os.MkdirAll(task.RotateSetting.DestPath, 0755); err != nil {
		return fmt.Errorf("創建目標目錄失敗: %w", err)
	}

	// 2. 檢查磁碟空間
	if r.service.isLowDiskSpace(task.RotateSetting.SourcePath, task.RotateSetting.MinDiskFreeMB) {
		r.service.logger.Warn("磁碟空間不足，執行緊急清理",
			zap.String("path", task.RotateSetting.SourcePath),
			zap.Int64("min_free_mb", task.RotateSetting.MinDiskFreeMB))
		if err := r.service.cleanOldFiles(task.RotateSetting.SourcePath, task.RotateSetting.MaxAge); err != nil {
			return fmt.Errorf("緊急清理失敗: %w", err)
		}
	}

	// 3. 檢查目錄大小並清理
	if r.service.isExceedMaxSize(task.RotateSetting.SourcePath, task.RotateSetting.MaxSizeMB) {
		if err := r.service.cleanOldFiles(task.RotateSetting.SourcePath, task.RotateSetting.MaxAge); err != nil {
			return fmt.Errorf("清理超大目錄失敗: %w", err)
		}
	}

	// 4. 壓縮舊日誌
	if task.RotateSetting.CompressEnabled {
		if err := r.service.compressOldLogs(task.RotateSetting.SourcePath, task.RotateSetting.DestPath, task.RotateSetting.CompressMatchRegex, task.RotateSetting.CompressSaveRegex, task.RotateSetting.CompressOffsetHours); err != nil {
			return fmt.Errorf("壓縮舊日誌失敗: %w", err)
		}
	}

	r.service.logger.Info("輪轉任務執行完成",
		zap.String("source", task.RotateSetting.SourcePath),
		zap.String("dest", task.RotateSetting.DestPath))

	return nil
}
