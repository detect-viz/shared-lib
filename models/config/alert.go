package config

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	AutoApplyRule bool   `mapstructure:"auto_apply_rule"`
	NotifyPeriod  int    `mapstructure:"notify_period"`
	RetryCount    int    `mapstructure:"retry_limit"`
	RetryInterval int    `mapstructure:"retry_interval"`
	MigratePath   string `mapstructure:"migrate_path"`
	TemplatePath  string `mapstructure:"template_path"`
}
