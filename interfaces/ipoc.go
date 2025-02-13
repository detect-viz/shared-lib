package interfaces

import "context"

// MetricData 系統指標資料結構
type MetricData struct {
	Color     string  `json:"color"`
	Metric    string  `json:"metric"`
	Timestamp int64   `json:"timestamp"`
	Unit      string  `json:"unit"`
	Value     float64 `json:"value"`
}

// IPOCClient 定義 IPOC 客戶端介面
type IPOCClient interface {
	GetSysloadMetric(ctx context.Context, realm, token string, start, end int64, host, metric string, is_long_term bool) ([]MetricData, error)
}
