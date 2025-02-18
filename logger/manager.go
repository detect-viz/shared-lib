package logger

import (
	"fmt"
	"sync"

	"shared-lib/interfaces"
	"shared-lib/models"
	"shared-lib/models/common"
)

// LogManager 日誌管理器
type LogManager struct {
	sync.RWMutex
	config    *models.LoggerConfig
	logger    *Logger
	rotator   *RotateManager
	stopChan  chan struct{}
	waitGroup sync.WaitGroup
}

// NewLogManager 創建日誌管理器
func NewLogManager(config *models.LoggerConfig) (*LogManager, error) {
	logger, err := NewLogger(config, WithCallerSkip(1))
	if err != nil {
		return nil, fmt.Errorf("創建日誌管理器失敗: %w", err)
	}

	return &LogManager{
		config:   config,
		logger:   logger,
		rotator:  NewRotateManager(logger),
		stopChan: make(chan struct{}),
	}, nil
}

// RegisterRotateTask 註冊輪轉任務
func (m *LogManager) RegisterRotateTask(task common.RotateTask) error {
	return m.rotator.RegisterTask(task)
}

// Start 啟動日誌管理服務
func (m *LogManager) Start() error {
	m.Lock()
	defer m.Unlock()

	// 啟動輪轉管理器
	if err := m.rotator.Start(); err != nil {
		return fmt.Errorf("啟動輪轉管理器失敗: %w", err)
	}

	m.logger.Info("日誌管理服務已啟動")
	return nil
}

// Stop 停止日誌管理服務
func (m *LogManager) Stop() {
	close(m.stopChan)
	m.rotator.Stop()
	m.waitGroup.Wait()
}

// GetLogger 獲取日誌實例
func (m *LogManager) GetLogger() interfaces.Logger {
	return m.logger
}
