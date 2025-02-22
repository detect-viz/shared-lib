package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/detect-viz/shared-lib/interfaces"
	"github.com/detect-viz/shared-lib/models"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 日誌記錄器
type Logger struct {
	zap    *zap.Logger
	config *models.LoggerConfig
	files  []*os.File
}

// LoggerOption 日誌選項
type LoggerOption func(*Logger)

// WithCallerSkip 設置調用者跳過層數
func WithCallerSkip(skip int) LoggerOption {
	return func(l *Logger) {
		l.zap = l.zap.WithOptions(zap.AddCallerSkip(skip))
	}
}

// WithFields 添加固定字段
func WithFields(fields ...zap.Field) LoggerOption {
	return func(l *Logger) {
		l.zap = l.zap.With(fields...)
	}
}

// validateConfig 驗證配置
func validateConfig(config *models.LoggerConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能為空")
	}
	if config.Path == "" {
		return fmt.Errorf("日誌路徑不能為空")
	}
	return nil
}

// NewLogger 創建日誌記錄器
func NewLogger(config *models.LoggerConfig, opts ...LoggerOption) (*Logger, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("配置驗證失敗: %w", err)
	}

	// 確保日誌目錄存在
	if err := os.MkdirAll(filepath.Dir(config.Path), 0755); err != nil {
		return nil, fmt.Errorf("創建日誌目錄失敗: %w", err)
	}

	// 創建 encoder 配置
	encoderConfig := newEncoderConfig()

	// 設置日誌級別
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, err
	}

	// 創建輸出
	outputs, files, err := createOutputs(config.Path)
	if err != nil {
		return nil, err
	}

	// 創建 core
	// core := zapcore.NewCore(
	// 	zapcore.NewJSONEncoder(encoderConfig),
	// 	zapcore.NewMultiWriteSyncer(outputs...),
	// 	level,
	// )

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(outputs...),
		level,
	)

	// 創建 logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	logger := &Logger{
		zap:    zapLogger,
		config: config,
		files:  files,
	}

	// 應用選項
	for _, opt := range opts {
		opt(logger)
	}

	return logger, nil
}

// newEncoderConfig 創建 encoder 配置
func newEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    "func",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// parseLogLevel 解析日誌級別
func parseLogLevel(levelStr string) (zapcore.Level, error) {
	level := zap.InfoLevel
	if levelStr != "" {
		if err := level.UnmarshalText([]byte(levelStr)); err != nil {
			return level, fmt.Errorf("解析日誌級別失敗: %w", err)
		}
	}
	return level, nil
}

func createOutputs(path string) ([]zapcore.WriteSyncer, []*os.File, error) {
	var outputs []zapcore.WriteSyncer
	var files = []*os.File{} // 初始化為空 slice

	// 添加文件輸出
	if path != "" {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// 清理已打開的文件
			for _, f := range files {
				f.Close()
			}
			return nil, nil, fmt.Errorf("打開日誌文件失敗: %w", err)
		}
		files = append(files, file)
		outputs = append(outputs, zapcore.AddSync(file))
	}

	// 添加控制台輸出
	outputs = append(outputs, zapcore.AddSync(os.Stdout))

	return outputs, files, nil
}

// Close 關閉日誌記錄器
func (l *Logger) Close() error {
	// 同步緩衝區
	if err := l.zap.Sync(); err != nil {
		return fmt.Errorf("同步日誌緩衝區失敗: %w", err)
	}

	// 關閉所有文件
	for _, file := range l.files {
		if err := file.Close(); err != nil {
			return fmt.Errorf("關閉日誌文件失敗: %w", err)
		}
	}

	return nil
}

// 實現 interfaces.Logger 介面
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// With 創建帶有額外字段的 Logger
func (l *Logger) With(fields ...zap.Field) interfaces.Logger {
	newLogger := *l
	newLogger.zap = l.zap.With(fields...)
	return &newLogger
}

// Sync 同步緩衝區
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Named 創建命名日誌記錄器
func (l *Logger) Named(name string) interfaces.Logger {
	newLogger := *l
	newLogger.zap = l.zap.Named(name)
	return &newLogger
}

// Sugar 獲取 SugaredLogger
func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.zap.Sugar()
}

// Clone 克隆日誌記錄器
func (l *Logger) Clone() interfaces.Logger {
	newLogger := *l
	newLogger.zap = l.zap
	newLogger.files = make([]*os.File, len(l.files))
	copy(newLogger.files, l.files)
	return &newLogger
}

// GetLogger 獲取原始日誌實例
func (l *Logger) GetLogger() *zap.Logger {
	return l.zap
}
