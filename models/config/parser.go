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

// IsEmpty 檢查 AWR 配置是否為空
func (c AwrConfig) IsEmpty() bool {
	return c == AwrConfig{}
}
