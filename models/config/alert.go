package config

// AlertConfig 告警配置
type AlertConfig struct {
	Enabled      bool `mapstructure:"enabled"`
	Timeout      int  `mapstructure:"timeout"`
	NotifyPeriod int  `mapstructure:"notify_period"`
	WorkPath     struct {
		Sent       string // 待轉換的觸發日誌目錄
		Notify     string // 通知目錄
		Silence    string // 靜音目錄
		Unresolved string // 未解決目錄
		Resolved   string // 已解決目錄
	}
	AlertCodes struct {
		Level struct {
			Crit Code `yaml:"crit"`
			Warn Code `yaml:"warn"`
			Info Code `yaml:"info"`
		} `yaml:"level"`
		Contact struct {
			Type struct {
				Discord Code `yaml:"discord"`
				Line    Code `yaml:"line"`
				Mail    Code `yaml:"mail"`
				Slack   Code `yaml:"slack"`
				Teams   Code `yaml:"teams"`
				Webex   Code `yaml:"webex"`
				Webhook Code `yaml:"webhook"`
			} `yaml:"type"`
			State struct {
				Muting  Code `yaml:"muting"`
				Normal  Code `yaml:"normal"`
				Pending Code `yaml:"pending"`
			} `yaml:"state"`
		} `yaml:"contact"`
		Health struct {
			Disconnected Code `yaml:"disconnected"`
			Error        Code `yaml:"error"`
			Healthy      Code `yaml:"healthy"`
			Warning      Code `yaml:"warning"`
		} `yaml:"health"`
		Mute struct {
			Active    Code `yaml:"active"`
			Disable   Code `yaml:"disable"`
			Ended     Code `yaml:"ended"`
			Scheduled Code `yaml:"scheduled"`
		} `yaml:"mute"`
		Rule struct {
			Alerting Code `yaml:"alerting"`
			Disable  Code `yaml:"disable"`
			Normal   Code `yaml:"normal"`
			Resolved Code `yaml:"resolved"`
		} `yaml:"rule"`
	}
}
type Code struct {
	Name  string `yaml:"name"`
	Alias string `yaml:"alias"`
	Desc  string `yaml:"desc"`
}
