package constants

// 單位轉換常數
const (
	BytesPerKB = 1024
	BytesPerMB = 1024 * 1024
	BytesPerGB = 1024 * 1024 * 1024
)

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

// 日期格式常數
const (
	FormatYYYYMMDDHHMM = "200601021504" // 202403251430
	FormatYYYYMMDDHH   = "2006010215"   // 2024032514
	FormatYYYYMMDD     = "20060102"     // 20240325
)

// 時間轉換常數
const (
	SecondsPerMinute = 60
	SecondsPerHour   = 3600
	SecondsPerDay    = 86400
)
