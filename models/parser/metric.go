package parser

// MetricValue 指標數值
type MetricValue struct {
	Name      string  `json:"name"`      // 指標名稱
	Value     float64 `json:"value"`     // 數值
	Timestamp int64   `json:"timestamp"` // 時間戳
	Tags      Tags    `json:"tags"`      // 標籤
}

// Tags 標籤集合
type Tags map[string]string

// MetricLibrarys 指標庫結構
type MetricLibrarys struct {
	Metrics struct {
		CPU        map[string]MetricField
		Memory     map[string]MetricField
		Disk       map[string]MetricField
		Network    map[string]MetricField
		FileSystem map[string]MetricField
		Process    map[string]MetricField
		Tablespace map[string]MetricField
	}
}

// MetricField 指標欄位結構
type MetricField struct {
	Category     string
	MetricName   string
	AliasName    string
	RawUnit      string
	DisplayUnit  string
	PartitionTag string
}

// GetMetricField 獲取指標欄位定義
func (mc *MetricLibrarys) GetMetricField(category, name string) (MetricField, bool) {
	switch category {
	case "cpu":
		field, ok := mc.Metrics.CPU[name]
		return field, ok
	case "memory":
		field, ok := mc.Metrics.Memory[name]
		return field, ok
	case "disk":
		field, ok := mc.Metrics.Disk[name]
		return field, ok
	case "network":
		field, ok := mc.Metrics.Network[name]
		return field, ok
	case "filesystem":
		field, ok := mc.Metrics.FileSystem[name]
		return field, ok
	case "process":
		field, ok := mc.Metrics.Process[name]
		return field, ok
	case "tablespace":
		field, ok := mc.Metrics.Tablespace[name]
		return field, ok
	default:
		return MetricField{}, false
	}
}

// ConvertUnit 單位轉換
func (mf *MetricField) ConvertUnit(value float64) float64 {
	// 根據 RawUnit 和 DisplayUnit 進行單位轉換
	switch {
	case mf.RawUnit == "bytes" && mf.DisplayUnit == "MB":
		return value / 1024 / 1024
	case mf.RawUnit == "bytes" && mf.DisplayUnit == "GB":
		return value / 1024 / 1024 / 1024
	// ... 其他單位轉換邏輯
	default:
		return value
	}
}
