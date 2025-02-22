package common

import "time"

// 檔案輪轉配置 [各模組使用]
// 排程優先度: MinDiskFreeMB > MaxSizeMB > MaxAge (3個條件滿足1個即可)

// RotateSetting 日誌輪轉配置
type RotateSetting struct {
	Enabled             bool          `mapstructure:"enabled"`
	Schedule            string        `mapstructure:"schedule"` // cron 表達式
	CompressEnabled     bool          `mapstructure:"compress_enabled"`
	CompressMatchRegex  string        `mapstructure:"compress_match_regex"` // *_${YYYYMMDD}.nmon
	CompressOffsetHours int           `mapstructure:"compress_offset_hours"`
	CompressSaveRegex   string        `mapstructure:"compress_save_regex"` // NMON-${YYYYMMDD}.tar.gz
	MaxAge              time.Duration `mapstructure:"max_age"`
	MaxSizeMB           int64         `mapstructure:"max_size_mb"`
	MinDiskFreeMB       int64         `mapstructure:"min_disk_free_mb"`
}

// RotateTask 檔案輪轉任務
type RotateTask struct {
	JobID         string        `json:"job_id"`      // 輪轉任務 ID (唯一標識)
	SourcePath    string        `json:"source_path"` // 來源目錄 (data/master/nmon)
	DestPath      string        `json:"dest_path"`   // 目標目錄 (backup/master/nmon)
	RotateSetting RotateSetting `json:"rotate_setting"`
}

// 路徑 [各模組多個配置所以不放在RotateSetting]
// TaskInfo 任務資訊
type TaskInfo struct {
	Name     string        // 任務名稱
	Schedule time.Duration // 執行間隔
	LastRun  time.Time     // 上次執行時間
	NextRun  time.Time     // 下次執行時間
	Status   string        // 任務狀態 (running/stopped)
	Error    error         // 最後一次執行錯誤
}
