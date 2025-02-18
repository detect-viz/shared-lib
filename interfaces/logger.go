package interfaces

import (
	"shared-lib/models/common"
	"time"

	"go.uber.org/zap"
)

// Logger 定義日誌接口
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Close() error
}

// LogManager 日誌管理介面
type LogManager interface {
	// 啟動日誌管理服務
	Start() error
	// 停止日誌管理服務
	Stop()
	// 獲取 Logger 實例
	GetLogger() Logger
	// 註冊輪轉任務
	RegisterRotateTask(task common.RotateTask) error
}

// LogCleanerOptions 清理配置
type LogCleanerOptions struct {
	BasePath    string        // 基礎路徑
	Pattern     string        // 檔案匹配模式
	MaxAge      time.Duration // 最大保留時間
	MaxSize     int64         // 最大檔案大小
	Compress    bool          // 是否壓縮
	CompressAge time.Duration // 壓縮時間閾值
}
