package config

// Config 全局配置結構
type Config struct {
	Config    []string        `mapstructure:"config"`
	Mapping   MappingConfig   `mapstructure:"mapping"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Alert     AlertConfig     `mapstructure:"alert"`
	Server    ServerConfig    `mapstructure:"server"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
}
