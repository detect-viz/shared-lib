package config

import (
	"time"
)

// Config Parser 服務配置
type ParserConfig struct {
	UploadPath  string        // 上傳檔案路徑
	BackupPath  string        // 備份檔案路徑
	ErrorPath   string        // 錯誤檔案路徑
	CheckPeriod time.Duration // 檢查週期
	MappingTag  MappingTag    // 映射標籤
	MetricSpec  MetricSpec    // 指標規格
	MaxDataAge  int64         // 最大數據年齡
	AwrConfig   AwrConfig     // AWR 配置
}

type AwrSetting struct {
	DefaultTables        AwrTablesSetting `json:"default_tables"`
	Information          AwrTablesSetting `json:"information"`
	InstanceEfficiency   AwrTablesSetting `json:"instance_efficiency"`
	CacheSizes           AwrTablesSetting `json:"cache_sizes"`
	Iostat               AwrTablesSetting `json:"iostat"`
	SharedPoolStatistics AwrTablesSetting `json:"shared_pool_statistics"`
	MemoryStatistics     AwrTablesSetting `json:"memory_statistics"`
	OperatingSystem      AwrTablesSetting `json:"operating_system"`
	UndoSegment          AwrTablesSetting `json:"undo_segment"`
	InstanceActivity     AwrTablesSetting `json:"instance_activity"`
	WaitEventHistogram   AwrTablesSetting `json:"wait_event_histogram"`
	TablespaceIoStats    AwrTablesSetting `json:"tablespace_io_stats"`
	BufferPoolStatistics AwrTablesSetting `json:"buffer_pool_statistics"`
	SegmentStatistics    AwrTablesSetting `json:"segment_statistics"`
	SqlStatistics        AwrTablesSetting `json:"sql_statistics"`
	//* 新版本 Information
	DatabaseSummary AwrTablesSetting `json:"database_summary"`
}

type AwrConfig struct {
	EnabledTables AwrSetting `yaml:"enabled_tables" json:"enabled_tables"`
	SQLText       struct {
		Enable        bool `yaml:"enable" json:"enable"`
		RetentionDays int  `yaml:"retention_days" json:"retention_days"`
	}
}

type AwrTablesSetting struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	RawName string `yaml:"raw_name" json:"raw_name"`
	NewName string `yaml:"new_name" json:"new_name"`
}

// Metric 指標定義
type MetricSpec struct {
	Category     string `yaml:"category"`
	MetricName   string `yaml:"metric_name"`
	AliasName    string `yaml:"alias_name"`
	RawUnit      string `yaml:"raw_unit"`
	DisplayUnit  string `yaml:"display_unit"`
	PartitionTag string `yaml:"partition_tag"`
}

type MetricSpecConfig struct {
	CPU struct {
		Usage  MetricSpec `yaml:"usage"`
		Idle   MetricSpec `yaml:"idle"`
		System MetricSpec `yaml:"system"`
		User   MetricSpec `yaml:"user"`
		IOWait MetricSpec `yaml:"iowait"`
		Nice   MetricSpec `yaml:"nice"`
		Steal  MetricSpec `yaml:"steal"`
	} `yaml:"cpu"`

	Memory struct {
		TotalBytes  MetricSpec `yaml:"total_bytes"`
		UsedBytes   MetricSpec `yaml:"used_bytes"`
		BufferBytes MetricSpec `yaml:"buffer_bytes"`
		CacheBytes  MetricSpec `yaml:"cache_bytes"`
		BufferUsage MetricSpec `yaml:"buffer_usage"`
		CacheUsage  MetricSpec `yaml:"cache_usage"`
		Usage       MetricSpec `yaml:"usage"`
		FreeUsage   MetricSpec `yaml:"free_usage"`
	} `yaml:"memory"`

	Swap struct {
		TotalBytes MetricSpec `yaml:"total_bytes"`
		UsedBytes  MetricSpec `yaml:"used_bytes"`
		Usage      MetricSpec `yaml:"usage"`
	} `yaml:"swap"`

	Network struct {
		SentBytes   MetricSpec `yaml:"sent_bytes"`
		RecvBytes   MetricSpec `yaml:"recv_bytes"`
		SentPackets MetricSpec `yaml:"sent_packets"`
		RecvPackets MetricSpec `yaml:"recv_packets"`
		SentErrors  MetricSpec `yaml:"sent_errs"`
		RecvErrors  MetricSpec `yaml:"recv_errs"`
	} `yaml:"network"`

	Disk struct {
		Busy        MetricSpec `yaml:"busy"`
		ReadBytes   MetricSpec `yaml:"read_bytes"`
		WriteBytes  MetricSpec `yaml:"write_bytes"`
		Reads       MetricSpec `yaml:"reads"`
		Writes      MetricSpec `yaml:"writes"`
		Xfers       MetricSpec `yaml:"xfers"`
		Rqueue      MetricSpec `yaml:"rqueue"`
		Wqueue      MetricSpec `yaml:"wqueue"`
		QueueLength MetricSpec `yaml:"queue_length"`
		IOTime      MetricSpec `yaml:"io_time"`
		ReadTime    MetricSpec `yaml:"read_time"`
		WriteTime   MetricSpec `yaml:"write_time"`
	} `yaml:"disk"`

	Filesystem struct {
		TotalBytes MetricSpec `yaml:"total_bytes"`
		FreeBytes  MetricSpec `yaml:"free_bytes"`
		UsedBytes  MetricSpec `yaml:"used_bytes"`
		Usage      MetricSpec `yaml:"usage"`
		FreeUsage  MetricSpec `yaml:"free_usage"`
	} `yaml:"filesystem"`

	Process struct {
		PID        MetricSpec `yaml:"pid"`
		PsCPUUsage MetricSpec `yaml:"ps_cpu_usage"`
		PsMemUsage MetricSpec `yaml:"ps_mem_usage"`
	} `yaml:"process"`

	System struct {
		Uptime MetricSpec `yaml:"uptime"`
	} `yaml:"system"`

	Connection struct {
		Refused MetricSpec `yaml:"refused"`
		Status  MetricSpec `yaml:"status"`
	} `yaml:"connection"`

	Tablespace struct {
		Autoextensible MetricSpec `yaml:"autoextensible"`
		FilesCount     MetricSpec `yaml:"files_count"`
		TotalSpaceGB   MetricSpec `yaml:"total_space_gb"`
		UsedSpaceGB    MetricSpec `yaml:"used_space_gb"`
		FreeSpaceGB    MetricSpec `yaml:"free_space_gb"`
		UsedPct        MetricSpec `yaml:"used_pct"`
		FreePct        MetricSpec `yaml:"free_pct"`
		MaxSizeGB      MetricSpec `yaml:"max_size_gb"`
		MaxFreeGB      MetricSpec `yaml:"max_free_gb"`
		AutoUsedPct    MetricSpec `yaml:"auto_used_pct"`
		AutoFreePct    MetricSpec `yaml:"auto_free_pct"`
	} `yaml:"tablespace"`
}

type MappingTag struct {
	Base struct {
		Host      string `json:"host"`
		Database  string `json:"database"`
		Source    string `json:"source"`
		Os        string `json:"os"`
		IP        string `json:"ip"`
		Timestamp string `json:"timestamp"`
		Value     string `json:"value"`
		Metric    string `json:"metric"`
	} `json:"base"`
	Os struct {
		Aix     string `json:"aix"`
		Linux   string `json:"linux"`
		Windows string `json:"windows"`
	} `json:"os"`
	Source struct {
		Nmon             string `json:"nmon"`
		Njmon            string `json:"njmon"`
		Logman           string `json:"logman"`
		Sysstat          string `json:"sysstat"`
		OracleAwrrpt     string `json:"oracle_awrrpt"`
		OracleTablespace string `json:"oracle_tablespace"`
		OracleConnection string `json:"oracle_connection"`
	} `json:"source"`
	Partition struct {
		TablespaceName string `json:"tablespace_name"`
		CPUName        string `json:"cpu_name"`
		DiskName       string `json:"disk_name"`
		FilesystemName string `json:"filesystem_name"`
		SwapName       string `json:"swap_name"`
		ProcessName    string `json:"process_name"`
		NetworkName    string `json:"network_name"`
	} `json:"partition"`
	Category struct {
		CPU        string `json:"cpu"`
		Memory     string `json:"memory"`
		Swap       string `json:"swap"`
		Network    string `json:"network"`
		Disk       string `json:"disk"`
		Filesystem string `json:"filesystem"`
		Process    string `json:"process"`
		System     string `json:"system"`
		Connection string `json:"connection"`
		Tablespace string `json:"tablespace"`
	} `json:"category"`
	Sql struct {
		ID     string `json:"id"`
		Module string `json:"module"`
		Text   string `json:"text"`
	} `json:"sql"`
}

// IsEmpty 檢查 AWR 配置是否為空
func (c AwrConfig) IsEmpty() bool {
	return c == AwrConfig{}
}
