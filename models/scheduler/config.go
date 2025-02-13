package scheduler

// SchedulerConfig 排程配置
type SchedulerConfig struct {
	Timezone   string         `mapstructure:"timezone"`
	Jobs       []SchedulerJob `mapstructure:"jobs"`
	MaxRetries int            `mapstructure:"max_retries"`
	RetryDelay int            `mapstructure:"retry_delay"`
}

// JobStatus 任務狀態
type JobStatus struct {
	Name       string
	LastRun    int64
	NextRun    int64
	Status     string
	Error      string
	RetryCount int
}

// SchedulerJob 排程任務配置
type SchedulerJob struct {
	Name     string            `mapstructure:"name"`
	Spec     string            `mapstructure:"spec"`
	Type     string            `mapstructure:"type"`
	Enabled  bool              `mapstructure:"enabled"`
	Metadata map[string]string `mapstructure:"metadata"`
}
