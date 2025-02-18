package config

// Config 全局配置結構
type Config struct {
	Logger   LoggerConfig   `mapstructure:"logger"`
	Database DatabaseConfig `mapstructure:"database"`
	Alert    AlertConfig    `mapstructure:"alert"`
	Notify   NotifyConfig   `mapstructure:"notify"`
	Server   ServerConfig   `mapstructure:"server"`
}

// Parser   ParserConfig     `mapstructure:"parser"`
// 	Metric   MetricSpecConfig `mapstructure:"metric"`
