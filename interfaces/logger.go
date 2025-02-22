package interfaces

import (
	"github.com/detect-viz/shared-lib/models/common"

	"go.uber.org/zap"
)

// Logger 定義日誌接口
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Named(name string) Logger
	Clone() Logger
	Sync() error
	Close() error
}

// LogRotator 日誌輪轉器接口
type LogRotator interface {
	// 啟動日誌管理服務
	Start() error
	// 停止日誌管理服務
	Stop()
	// 獲取 Logger 實例
	GetLogger() Logger
	// // 註冊輪轉任務
	// RegisterRotateTask(task common.RotateTask) error
	// 執行輪轉任務
	ExecuteRotateTask(task common.RotateTask) error
}
