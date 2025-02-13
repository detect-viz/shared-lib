package common

// TimeConstants 時間相關常量
const (
	// 時間轉換
	SecondsPerMinute = 60
	SecondsPerHour   = 3600
	SecondsPerDay    = 86400

	// 數據保留時間 (30天)
	MaxDataAge = 30 * SecondsPerDay

	// 單位轉換
	BytesPerKB = 1024
	BytesPerMB = 1024 * 1024
	BytesPerGB = 1024 * 1024 * 1024
	MilliToSec = 1000.0

	// 百分比轉換
	PercentMultiplier = 100.0
)

// MetricConstants 指標相關常量
const (
	// 通用字段
	Timestamp = "timestamp"
	Value     = "value"

	// CPU 相關
	CPUName     = "cpu_name"
	CPUUsage    = "cpu_usage"
	IOWaitUsage = "iowait_usage"
	IdleUsage   = "idle_usage"

	// 記憶體相關
	MemTotalBytes = "mem_total_bytes"
	MemUsedBytes  = "mem_used_bytes"
	CacheBytes    = "cache_bytes"
	VirtualBytes  = "virtual_bytes"

	// 磁盤相關
	DiskName  = "disk_name"
	DiskBusy  = "disk_busy"
	DiskRead  = "disk_read"
	DiskWrite = "disk_write"

	// 網絡相關
	NetInterface = "interface"
	NetRead      = "net_read_bytes"
	NetWrite     = "net_write_bytes"
	NetPacketIn  = "net_packet_in"
	NetPacketOut = "net_packet_out"
	NetErrorIn   = "net_error_in"
	NetErrorOut  = "net_error_out"

	// 進程相關
	ProcessID      = "process_id"
	ProcessName    = "process_name"
	ProcessCPU     = "process_cpu"
	ProcessMemory  = "process_memory"
	ProcessVirtual = "process_virtual"

	// 使用率相關
	MemUsagePercent = "mem_usage_percent"
	DiskUtilization = "disk_utilization"
)

// ErrorMessages 錯誤信息
const (
	ErrInvalidTime  = "invalid timestamp: %v"
	ErrInvalidValue = "invalid value: %v"
	ErrExpiredData  = "data too old: %v"
	ErrNilContent   = "nil content"
	ErrNoValidData  = "no valid data"
	ErrDivideByZero = "divide by zero"
	ErrScannerError = "scanner error: %v"
)
