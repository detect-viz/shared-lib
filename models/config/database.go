package config

// DatabaseConfig 資料庫配置
type DatabaseConfig struct {
	MySQL struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
	} `mapstructure:"mysql"`
	InfluxDB struct {
		URL    string `mapstructure:"url"`
		Token  string `mapstructure:"token"`
		Org    string `mapstructure:"org"`
		Bucket string `mapstructure:"bucket"`
	} `mapstructure:"influxdb"`
}
