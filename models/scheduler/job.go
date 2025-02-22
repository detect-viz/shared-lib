package scheduler

// JobStatus 任務狀態
type JobStatus struct {
	Name       string
	LastRun    int64
	NextRun    int64
	Status     string
	Error      string
	RetryCount int
}
