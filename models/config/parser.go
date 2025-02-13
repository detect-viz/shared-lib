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
	} `yaml:"disk"`

	Filesystem struct {
		TotalBytes MetricSpec `yaml:"total_bytes"`
		FreeBytes  MetricSpec `yaml:"free_bytes"`
		UsedBytes  MetricSpec `yaml:"used_bytes"`
		Usage      MetricSpec `yaml:"usage"`
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
		Host   string `json:"host"`
		Source string `json:"source"`
		Os     string `json:"os"`
		IP     string `json:"ip"`
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
		InstanceName   string `json:"instance_name"`
		TablespaceName string `json:"tablespace_name"`
		CPUName        string `json:"cpu_name"`
		DiskName       string `json:"disk_name"`
		FilesystemName string `json:"filesystem_name"`
		SwapName       string `json:"swap_name"`
		ProcessName    string `json:"process_name"`
		NetworkName    string `json:"network_name"`
	} `json:"partition"`
	Sql struct {
		ID     string `json:"id"`
		Module string `json:"module"`
		Text   string `json:"text"`
	} `json:"sql"`
}
