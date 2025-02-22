package config

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	Timeout      int    `mapstructure:"timeout"`
	NotifyPeriod int    `mapstructure:"notify_period"`
	MigratePath  string `mapstructure:"migrate_path"`
	WorkPath     string `mapstructure:"work_path"` // 工作路徑
	Rotate       rotate `mapstructure:"rotate"`
}

type rotate struct {
	Enabled   bool  `mapstructure:"enabled"`
	MaxAge    int64 `mapstructure:"max_age"`
	MaxSizeMB int64 `mapstructure:"max_size_mb"`
}
