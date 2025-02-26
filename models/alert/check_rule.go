package alert

// CheckRule - 異常檢測規則
type CheckRule struct {
	// 動態檢查
	Timestamp    int64    `json:"timestamp"`           // 異常檢測時間
	TriggerValue float64  `json:"trigger_value"`       // 異常數值
	Threshold    *float64 `json:"threshold,omitempty"` // 異常對應閾值
	Severity     string   `json:"severity"`            // 異常對應嚴重性
	RuleState    string   `json:"rule_state"`          // alerting / resolved / normal / disable
	ContactState string   `json:"contact_state"`       // normal / muting / silence

	// 靜態資訊
	RealmName         string            `json:"realm_name"`               // 告警規則所在的 realm
	ResourceGroupID   int64             `json:"resource_group_id"`        // 資源群組 ID
	ResourceGroupName string            `json:"resource_group_name"`      // 資源群組
	ResourceName      string            `json:"resource_name"`            // 監控的主機/設備
	PartitionName     *string           `json:"partition_name,omitempty"` // 分區名稱 (可選)
	MetricName        string            `json:"metric_name"`              // 監控指標
	CheckType         string            `json:"check_type"`               // absolute / amplitude
	Operator          string            `json:"operator"`                 // > / < / = / !=
	InfoThreshold     *float64          `json:"info_threshold,omitempty"` // 資訊閾值
	WarnThreshold     *float64          `json:"warn_threshold,omitempty"` // 警告閾值
	CritThreshold     *float64          `json:"crit_threshold,omitempty"` // 嚴重閾值
	RawUnit           string            `json:"raw_unit"`                 // 原始單位 (% / MB / ms)
	DisplayUnit       string            `json:"display_unit"`             // 顯示單位 (% / MB / ms)
	Scale             float64           `json:"scale"`                    // 比例
	Tags              map[string]string `json:"tags"`                     // 標籤
	Duration          int               `json:"duration"`                 // 異常持續時間
	RuleID            int64             `json:"rule_id"`                  // 關聯的告警規則 ID
	RuleDetailID      int64             `json:"rule_detail_id"`           // 關聯的告警規則詳細 ID
	RuleName          string            `json:"rule_name"`                // 規則名稱
	SilenceStart      *int64            `json:"silence_start,omitempty"`  // 靜音開始時間
	SilenceEnd        *int64            `json:"silence_end,omitempty"`    // 靜音結束時間
	MuteStart         *int64            `json:"mute_start,omitempty"`     // 抑制開始時間
	MuteEnd           *int64            `json:"mute_end,omitempty"`       // 抑制結束時間
	Labels            JSONMap           `json:"labels"`                   // 其他標籤
	Contacts          []Contact         `json:"contacts,omitempty"`       // 通知對象
}
