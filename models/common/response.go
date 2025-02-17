package common

// 下拉選單響應
type OptionResponse struct {
	Text  string `json:"text"  from:"text"`
	Value string `json:"value" from:"value"`
}

type Response struct {
	Msg     string
	Success bool
}

type MetricResponse struct {
	Metric string  `json:"metric"`
	Unit   string  `json:"unit,omitempty"`
	Color  string  `json:"color,omitempty"`
	Time   int64   `json:"timestamp,omitempty"`
	Value  float64 `json:"value"`
	Sort   int64   `json:"-"`
}
