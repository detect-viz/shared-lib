package config

// DatabaseConfig 資料庫配置
type DatabaseConfig struct {
	InfluxDB InfluxDBConfig `mapstructure:"influxdb"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
}

// InfluxDB 配置
type InfluxDBConfig struct {
	Version string `mapstructure:"version"`
	URL     string `mapstructure:"url"`
	Token   string `mapstructure:"token"`
	Org     string `mapstructure:"org"`
	Bucket  string `mapstructure:"bucket"`
}

// MySQL 配置
type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	Level    string `mapstructure:"level"`
	MaxIdle  int    `mapstructure:"max_idle"`
	MaxOpen  int    `mapstructure:"max_open"`
	MaxLife  string `mapstructure:"max_life"`
}
