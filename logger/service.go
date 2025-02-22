package logger

import (
	"fmt"

	"github.com/detect-viz/shared-lib/interfaces"

	"go.uber.org/zap"
)

// Service 日誌服務
type Service struct {
	config    interfaces.Config
	scheduler interfaces.Scheduler
	logger    interfaces.Logger
	tasks     interfaces.LogRotator
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
		tasks:     NewLogRotator(logger),
	}

	return svc, nil
}

// GetLogger 獲取 Logger
func (s *Service) GetLogger() interfaces.Logger {
	// 不需要重新包裝，直接返回
	return s.logger
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
