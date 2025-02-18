package config

// NotifyConfig 通知配置
type NotifyConfig struct {
	MaxRetry      int    `mapstructure:"max_retry"`
	RetryInterval int64  `mapstructure:"retry_interval"`
	WorkPath      string `mapstructure:"work_path"` // 工作路徑
	Rotate        rotate `mapstructure:"rotate"`
}

type rotate struct {
	Enabled   bool  `mapstructure:"enabled"`
	MaxAge    int64 `mapstructure:"max_age"`
	MaxSizeMB int64 `mapstructure:"max_size_mb"`
}

// ChannelConfig 通知渠道配置
type ChannelConfig struct {
	ID      int64             `json:"id"`
	Type    string            `json:"type"`
	Name    string            `json:"name"`
	Enabled bool              `json:"enabled"`
	Config  map[string]string `json:"config"`
}
