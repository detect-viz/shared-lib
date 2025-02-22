package config

// DatabaseConfig 資料庫配置
type DatabaseConfig struct {
	InfluxDB struct {
		URL    string `mapstructure:"url"`
		Token  string `mapstructure:"token"`
		Org    string `mapstructure:"org"`
		Bucket string `mapstructure:"bucket"`
	} `mapstructure:"influxdb"`
	MySQL struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"db_name"`
		Level    string `mapstructure:"level"`
		MaxIdle  int    `mapstructure:"max_idle"`
		MaxOpen  int    `mapstructure:"max_open"`
		MaxLife  string `mapstructure:"max_life"`
	} `mapstructure:"mysql"`
}
