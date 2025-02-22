package config

// LoggerConfig 日誌配置
type LoggerConfig struct {
	Level string `mapstructure:"level"` // 日誌級別
	Path  string `mapstructure:"path"`  // 日誌路徑
}

// SchedulerConfig 排程配置
type SchedulerConfig struct {
	Timezone   string         `mapstructure:"timezone"`
	Jobs       []SchedulerJob `mapstructure:"jobs"`
	MaxRetries int            `mapstructure:"max_retries"`
	RetryDelay int            `mapstructure:"retry_delay"`
}

// SchedulerJob 排程任務配置
type SchedulerJob struct {
	Name    string `mapstructure:"name"`
	Spec    string `mapstructure:"spec"`
	Type    string `mapstructure:"type"`
	Enabled bool   `mapstructure:"enabled"`
	Func    func() `mapstructure:"func"`
}
