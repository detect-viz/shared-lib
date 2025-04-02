package alert

type AlertPayload struct {
	Metadata Metadata                 `json:"metadata"` // 告警元數據
	Data     map[string][]MetricValue `json:"data"`     // 監控數據
}

type Metadata struct {
	RealmName      string `json:"realm_name"`
	DatasourceName string `json:"datasource_name"`
	ResourceName   string `json:"resource_name"`
	Timestamp      int64  `json:"timestamp"` // 告警發送時間
}

// 監控指標數據
type MetricValue struct {
	Timestamp int64   `json:"timestamp"` // 數據發生時間
	Value     float64 `json:"value"`
}
