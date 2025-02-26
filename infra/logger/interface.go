package logger

import (
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
	IsDebugMode() bool
	Clone() Logger
	Sync() error
	Close() error
	GetLogger() *zap.Logger
}
