package config

// ServerConfig 服務器配置
type ServerConfig struct {
	Port            int    `mapstructure:"port"`
	Mode            string `mapstructure:"mode"`
	ConfigDirectory string `mapstructure:"config_directory"`
}
