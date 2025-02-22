package errors

const (
	ErrInvalidTime  = "invalid timestamp: %v"
	ErrInvalidValue = "invalid value: %v"
	ErrExpiredData  = "data too old: %v"
	ErrNilContent   = "nil content"
	ErrNoValidData  = "no valid data"
	ErrDivideByZero = "divide by zero"
	ErrScannerError = "scanner error: %v"
)

// 添加錯誤類型常量
const (
	ErrParseHTML          = "解析 HTML 失敗: %v"
	ErrInvalidDBInfo      = "無效的數據庫信息: %v"
	ErrInvalidTimestamp   = "無效的時間戳: %v"
	ErrInvalidMetric      = "無效的指標值: %v"
	ErrSaveSQLText        = "保存 SQL 文本失敗: %v"
	ErrTableNotFound      = "找不到指定的表格: %s"
	ErrInvalidTableFormat = "表格格式無效: %s"
	ErrEmptyContent       = "empty content"
	ErrInvalidContent     = "invalid content"
	ErrInvalidConfig      = "invalid configuration: %v"
	ErrParseTime          = "failed to parse time: %v"
)
