package alert

type MetricRule struct {
	ID               int64   `json:"id" gorm:"primaryKey"` // CPU001
	Name             string  `json:"name"`
	MetricName       string  `json:"metric_name"`
	PartitionTag     string  `json:"partition_tag"`
	Operator         string  `json:"operator"`
	Type             string  `json:"type"`       // cpu, memory, disk, network, etc.
	CheckType        string  `json:"check_type"` // absolute, amplitude
	DefaultThreshold float64 `json:"default_threshold"`
	DefaultDuration  *int    `json:"default_duration"`
	Unit             string  `json:"unit"` // 添加單位字段
}
