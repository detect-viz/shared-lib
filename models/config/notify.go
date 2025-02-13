package config

import (
	"shared-lib/models/common"
)

// NotifyConfig 通知配置
type NotifyConfig struct {
	MaxRetry      int   `mapstructure:"max_retry"`
	RetryInterval int64 `mapstructure:"retry_interval"`
	WorkDir       struct {
		NotifyPath string // 通知路徑
		BackupPath string // 備份路徑
		FailedPath string // 失敗路徑
	}
	Rotate common.RotateTask
}

// ChannelConfig 通知渠道配置
type ChannelConfig struct {
	ID      int64             `json:"id"`
	Type    string            `json:"type"`
	Name    string            `json:"name"`
	Enabled bool              `json:"enabled"`
	Config  map[string]string `json:"config"`
}
