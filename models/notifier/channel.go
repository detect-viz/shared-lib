package notifier

// EmailSetting 郵件配置
type EmailSetting struct {
	Host        string   `mapstructure:"host"`
	Port        string   `mapstructure:"port"`
	Username    string   `mapstructure:"username"`
	Password    string   `mapstructure:"password"`
	From        string   `mapstructure:"from"`
	To          []string `mapstructure:"to"`
	Cc          []string `mapstructure:"cc"`
	Bcc         []string `mapstructure:"bcc"`
	ReplyTo     string   `mapstructure:"reply_to"`
	UseTLS      bool     `mapstructure:"use_tls"`
	UseSTARTTLS bool     `mapstructure:"use_starttls"`
	UseAuth     bool     `mapstructure:"use_auth"`
}

// TeamsSetting Teams配置
type TeamsSetting struct {
	WebhookURL string `mapstructure:"url"`
}

// LineSetting Line配置
type LineConfig struct {
	WebhookURL string `mapstructure:"url"`
}

// WebhookSetting Webhook配置
type WebhookConfig struct {
	URL string `mapstructure:"url"`
}
