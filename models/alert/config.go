package alert

// AlertConfig 告警服務配置
type AlertConfig struct {
	NotifyPeriod      int `json:"notify_period"`       // 通知週期（秒）
	NotifyPendingTime int `json:"notify_pending_time"` // 通知等待時間（秒）
	RetryDelay        int `json:"retry_delay"`         // 重試延遲時間（秒）
	MaxRetry          int `json:"max_retry"`           // 最大重試次數
}
