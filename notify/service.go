package notify

import (
	"encoding/json"
	"fmt"

	"os"
	"path/filepath"
	"time"

	"shared-lib/models"
	"shared-lib/models/common"

	"go.uber.org/zap"

	"shared-lib/interfaces"
)

const (
	NotifyPath = "notify"
	BackupPath = "backup"
	FailedPath = "failed"
)

// Service 通知服務
type Service struct {
	config models.NotifyConfig // 只使用自己的配置
	logger interfaces.Logger
	logMgr interfaces.LogManager
	db     interfaces.Database
}

// NewService 創建通知服務
func NewService(config models.NotifyConfig, logSvc interfaces.Logger, db interfaces.Database) *Service {
	logger := logSvc.With(zap.String("module", "notify"))

	return &Service{
		config: config,
		logger: logger,
		db:     db,
	}
}

// Init 初始化服務
func (s *Service) Init() error {
	// 初始化目錄
	workDir := filepath.Join(s.config.WorkPath, BackupPath) // 備份路徑
	if err := os.MkdirAll(workDir, 0755); err != nil {
		s.logger.Error("創建通知目錄失敗",
			zap.String("path", workDir),
			zap.Error(err))
		return err
	}

	// 註冊輪轉任務
	if s.config.Rotate.Enabled {
		task := common.RotateTask{
			JobID:      "notify_rotate_" + workDir,
			SourcePath: workDir,
			DestPath:   workDir,
			RotateSetting: common.RotateSetting{
				Schedule:            "0 0 1 * * *",
				MaxAge:              time.Duration(s.config.Rotate.MaxAge),
				MaxSizeMB:           s.config.Rotate.MaxSizeMB,
				CompressEnabled:     true,
				CompressMatchRegex:  "*${YYYYMMDD}*.log",
				CompressOffsetHours: 2,
				CompressSaveRegex:   "${YYYYMMDD}.tar.gz",
				MinDiskFreeMB:       300,
			},
		}

		if err := s.logMgr.RegisterRotateTask(task); err != nil {
			return fmt.Errorf("註冊輪轉任務失敗: %w", err)
		}
		s.logger.Info("已註冊通知日誌輪轉任務",
			zap.String("source", task.SourcePath),
			zap.String("dest", task.DestPath))
	}

	s.logger.Info("通知服務初始化完成")
	return nil
}

// ProcessNotifications 處理通知日誌
func (s *Service) ProcessNotifications() error {
	notifications, err := s.readNotificationLogs()
	if err != nil {
		s.logger.Error("讀取通知日誌失敗",
			zap.Error(err),
			zap.String("path", s.config.WorkPath))
		return fmt.Errorf("讀取通知日誌失敗: %w", err)
	}

	if len(notifications) == 0 {
		s.logger.Debug("沒有需要處理的通知")
		return nil
	}

	for _, notification := range notifications {
		if err := s.processNotification(&notification); err != nil {
			s.logger.Error("處理通知失敗",
				zap.Error(err),
				zap.String("uuid", notification.UUID))
			continue
		}
	}

	return nil
}

// processNotification 處理單個通知
func (s *Service) processNotification(notification *models.NotificationLog) error {
	// 檢查重試
	if s.shouldStopRetry(*notification) {
		return s.handleFailedNotification(notification, "超過重試限制或截止時間")
	}

	// 發送通知
	if err := s.sendNotification(*notification); err != nil {
		notification.NotifyRetry++
		return s.handleFailedNotification(notification, err.Error())
	}

	// 處理成功
	return s.handleSuccessNotification(notification)
}

// handleFailedNotification 處理失敗的通知
func (s *Service) handleFailedNotification(notification *models.NotificationLog, errMsg string) error {
	notification.Status = "failed"
	notification.Error = &errMsg

	if err := s.db.WriteNotificationLog(*notification); err != nil {
		s.logger.Error("更新通知狀態失敗",
			zap.Error(err),
			zap.String("uuid", notification.UUID))
		return err
	}

	return nil
}

// handleSuccessNotification 處理成功的通知
func (s *Service) handleSuccessNotification(notification *models.NotificationLog) error {
	now := time.Now().Unix()
	notification.Status = "sent"
	notification.SentAt = &now

	// 更新狀態
	if err := s.db.WriteNotificationLog(*notification); err != nil {
		s.logger.Error("更新通知狀態失敗",
			zap.Error(err),
			zap.String("uuid", notification.UUID))
		return err
	}

	// 歸檔通知
	if err := s.archiveNotification(*notification); err != nil {
		s.logger.Error("歸檔通知失敗",
			zap.Error(err),
			zap.String("uuid", notification.UUID))
		return err
	}

	return nil
}

// shouldStopRetry 檢查是否應該停止重試
func (s *Service) shouldStopRetry(notification models.NotificationLog) bool {
	return notification.NotifyRetry >= s.config.MaxRetry
}

// sendNotification 發送通知
func (s *Service) sendNotification(notification models.NotificationLog) error {
	// 創建發送器
	channel, err := NewSender(notification)
	if err != nil {
		return err
	}

	// 轉換並發送消息
	message := toAlertMessage(notification)
	return channel.Send(message)
}

// archiveNotification 歸檔通知日誌
func (s *Service) archiveNotification(notification models.NotificationLog) error {
	if notification.FilePath == nil {
		return fmt.Errorf("notification file path is nil")
	}

	oldPath := *notification.FilePath
	fileName := filepath.Base(oldPath)
	newPath := filepath.Join(s.config.WorkPath, BackupPath, fileName)

	// 確保目標目錄存在
	if err := os.MkdirAll(filepath.Join(s.config.WorkPath, BackupPath), 0755); err != nil {
		return fmt.Errorf("create backup directory failed: %w", err)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("move file failed: %w", err)
	}

	return nil
}

// readNotificationLogs 讀取通知日誌
func (s *Service) readNotificationLogs() ([]models.NotificationLog, error) {
	var notifications []models.NotificationLog

	// 讀取目錄下所有檔案
	files, err := os.ReadDir(filepath.Join(s.config.WorkPath, NotifyPath))
	if err != nil {
		return nil, fmt.Errorf("讀取目錄失敗: %v", err)
	}

	// 依序讀取每個檔案
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(s.config.WorkPath, NotifyPath, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			s.logger.Error("讀取檔案失敗",
				zap.String("file", path),
				zap.Error(err))
			continue
		}

		var notification models.NotificationLog
		if err := json.Unmarshal(data, &notification); err != nil {
			s.logger.Error("解析通知日誌失敗",
				zap.String("file", path),
				zap.Error(err))
			continue
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// SendNotification 發送通知
func (s *Service) SendNotification(notification *models.NotificationLog) error {
	return s.sendNotification(*notification)
}
