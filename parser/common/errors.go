package common

// Errors 錯誤訊息常量
const (
	// 一般錯誤
	ErrNilContent     = "content buffer is nil"
	ErrNoMetricGroups = "failed to initialize metric groups"
	ErrNoValidData    = "no valid metrics data found in content"
	ErrScannerError   = "scanner error: %v"

	// 檔案相關錯誤
	ErrUnknownSource = "unknown source type for file: %s"
	ErrNoHostname    = "cannot detect hostname from file: %s"
	ErrNoTimestamp   = "cannot parse timestamp from file: %s"

	// 數據相關錯誤
	ErrInvalidValue = "invalid value: %v"
	ErrDivideByZero = "divide by zero error"
	ErrInvalidTime  = "invalid timestamp: %v"
	ErrExpiredData  = "data is too old: %v"
)
