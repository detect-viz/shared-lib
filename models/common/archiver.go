package common

import "time"

// 檔案輪轉配置 [各模組使用]
// 排程優先度: MinDiskFreeMB > MaxSizeMB > MaxAge (3個條件滿足1個即可)

// RotateTask 檔案輪轉任務
type RotateTask struct {
	Task          Task          `json:"task"`
	RotateSetting RotateSetting `json:"archiver_setting"`
}

// RotateSetting 日誌輪轉配置
type RotateSetting struct {
	SourcePath          string        `json:"source_path"` // 來源目錄 (data/master/nmon)
	DestPath            string        `json:"dest_path"`   // 目標目錄 (backup/master/nmon)
	CompressEnabled     bool          `mapstructure:"compress_enabled"`
	CompressMatchRegex  string        `mapstructure:"compress_match_regex"` // *_${YYYYMMDD}.nmon
	CompressOffsetHours int           `mapstructure:"compress_offset_hours"`
	CompressSaveRegex   string        `mapstructure:"compress_save_regex"` // NMON-${YYYYMMDD}.tar.gz
	MaxAge              time.Duration `mapstructure:"max_age"`
	MaxSizeMB           int64         `mapstructure:"max_size_mb"`
	MinDiskFreeMB       int64         `mapstructure:"min_disk_free_mb"`
}

type BackupTask struct {
	Task          Task          `json:"task"`
	BackupSetting BackupSetting `json:"backup_setting"`
}

type BackupSetting struct {
	SourcePath    string        `json:"source_path"`
	DestPath      string        `json:"dest_path"`
	BackupType    string        `json:"backup_type"`
	MaxAge        time.Duration `mapstructure:"max_age"`
	MinDiskFreeMB int64         `mapstructure:"min_disk_free_mb"`
}
