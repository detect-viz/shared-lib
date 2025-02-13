package common

// Metric 定義所有指標名稱常量
const (
	// CPU 指標
	CPUUsage    = "cpu_usage"    // CPU 總使用率
	IdleUsage   = "idle_usage"   // 空閒率
	SystemUsage = "system_usage" // 系統使用率(win:privileged_time)
	UserUsage   = "user_usage"   // 用戶使用率(win:user_time)
	IOWaitUsage = "iowait_usage" // IO 等待率
	NiceUsage   = "nice_usage"   // Nice 值
	StealUsage  = "steal_usage"  // Steal 值(win:interrupt_time)
	// PrivilegedTime = "privileged_time" // Windows 特權時間
	// UserTime       = "user_time"       // Windows 用戶時間
	// InterruptTime  = "interrupt_time"  // 中斷時間

	// 記憶體指標
	MemTotalBytes = "mem_total_bytes" // 總記憶體
	MemUsedBytes  = "mem_used_bytes"  // 已用記憶體
	CacheBytes    = "cache_bytes"     // 快取大小
	BufferBytes   = "buffer_bytes"    // 緩衝區大小
	MemUsage      = "mem_usage"       // 記憶體使用率
	CacheUsage    = "cache_usage"     // 快取使用率
	BufferUsage   = "buffer_usage"    // 緩衝區使用率

	// 交換空間指標
	SwapTotalBytes = "swap_total_bytes" // 交換空間總量
	SwapUsedBytes  = "swap_used_bytes"  // 已用交換空間
	SwapUsage      = "swap_usage"       // 交換空間使用率
	PagingUsage    = "paging_usage"     // 分頁使用率
	PgIn           = "pgin"             // 頁面調入
	PgOut          = "pgout"            // 頁面調出

	// 網絡指標
	SentBytes   = "sent_bytes"   // 發送字節數
	RecvBytes   = "recv_bytes"   // 接收字節數
	SentPackets = "sent_packets" // 發送包數
	RecvPackets = "recv_packets" // 接收包數
	SentErrs    = "sent_errs"    // 發送錯誤數
	RecvErrs    = "recv_errs"    // 接收錯誤數

	// 磁盤指標
	Busy        = "busy"         // 磁盤忙碌度
	ReadBytes   = "read_bytes"   // 讀取字節數
	WriteBytes  = "write_bytes"  // 寫入字節數
	Reads       = "reads"        // 讀取次數
	Writes      = "writes"       // 寫入次數
	Xfers       = "xfers"        // 傳輸次數
	Rqueue      = "rqueue"       // 讀取隊列
	Wqueue      = "wqueue"       // 寫入隊列
	QueueLength = "queue_length" // 隊列長度
	IOTime      = "io_time"      // IO 時間
	WriteTime   = "write_time"   // 寫入時間
	ReadTime    = "read_time"    // 讀取時間

	// 文件系統指標
	FSTotalBytes = "fs_total_bytes" // 總空間
	FSFreeBytes  = "fs_free_bytes"  // 剩餘空間
	FSUsedBytes  = "fs_used_bytes"  // 已用空間
	FSUsage      = "fs_usage"       // 使用率

	// 進程指標
	PID        = "pid"          // 進程 ID
	PSCPUUsage = "ps_cpu_usage" // 進程 CPU 使用率
	PSMemUsage = "ps_mem_usage" // 進程記憶體使用率

	// 系統指標
	Uptime = "uptime" // 系統運行時間
)

// Label 定義指標標籤常量
const (
	Hostname    = "host"            // 主機名
	Instance    = "instance"        // 實例名稱
	CPUName     = "cpu_name"        // CPU 名稱
	DiskName    = "disk_name"       // 磁盤名稱
	Filesystem  = "filesystem_name" // 文件系統名稱
	ProcessName = "process_name"    // 進程名稱
	NetworkName = "network_name"    // 網絡接口名稱
	Timestamp   = "timestamp"       // 時間戳
	Value       = "value"           // 值
)
