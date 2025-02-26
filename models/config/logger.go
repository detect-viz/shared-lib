package config

// LoggerConfig 日誌配置
type LoggerConfig struct {
	Level string `mapstructure:"level"` // 日誌級別
	Path  string `mapstructure:"path"`  // 日誌路徑
}
