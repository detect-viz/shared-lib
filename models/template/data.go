package template

// TemplateData 模板數據結構
type TemplateData struct {
	RuleName        string                     // 規則名稱
	Timestamp       string                     // 觸發時間
	Labels          map[string]string          // 標籤
	GroupTriggereds map[string][]TriggeredInfo // 按資源組分類的觸發詳情
}

// ResourceGroupInfo 資源組信息
type ResourceGroupInfo struct {
	Name string // 資源組名稱
	ID   string // 資源組ID
}

// TriggeredInfo 觸發詳情
type TriggeredInfo struct {
	ResourceName string  // 資源名稱 (主機名)
	MetricName   string  // 指標名稱
	Value        float64 // 當前值
	Threshold    float64 // 閾值
	Level        string  // 告警等級
}

// DefaultTemplate 默認模板內容
type DefaultTemplate struct {
	Title   string
	Message string
}
