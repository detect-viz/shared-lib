package alert

// CheckRule - 異常檢測規則
type CheckRule struct {
	// 動態檢查
	UUID             string   `json:"uuid"`                        // 唯一識別碼
	Timestamp        int64    `json:"timestamp"`                   // 異常檢測時間
	CurrentValue     float64  `json:"current_value"`               // 異常數值
	CurrentThreshold *float64 `json:"current_threshold,omitempty"` // 異常對應閾值
	Severity         string   `json:"severity"`                    // 異常對應嚴重性
	Status           string   `json:"status"`                      // alerting / normal / silenced
	NotifyStatus     string   `json:"notify_status"`               // pending / sent / failed

	// 靜態資訊
	RealmName         string         `json:"realm_name"`               // 告警規則所在的 realm
	ResourceGroupID   int64          `json:"resource_group_id"`        // 資源群組 ID
	ResourceGroupName string         `json:"resource_group_name"`      // 資源群組
	ResourceName      string         `json:"resource_name"`            // 監控的主機/設備
	PartitionName     *string        `json:"partition_name,omitempty"` // 分區名稱 (可選)
	MetricName        string         `json:"metric_name"`              // 監控指標
	CheckType         string         `json:"check_type"`               // absolute / amplitude
	Operator          string         `json:"operator"`                 // > / < / = / !=
	InfoThreshold     *float64       `json:"info_threshold,omitempty"` // 資訊閾值
	WarnThreshold     *float64       `json:"warn_threshold,omitempty"` // 警告閾值
	CritThreshold     *float64       `json:"crit_threshold,omitempty"` // 嚴重閾值
	Unit              string         `json:"unit"`                     // 單位 (% / MB / ms)
	Duration          int            `json:"duration"`                 // 異常持續時間
	RuleID            int64          `json:"rule_id"`                  // 關聯的告警規則 ID
	RuleName          string         `json:"rule_name"`                // 規則名稱
	SilenceStart      *int64         `json:"silence_start,omitempty"`  // 靜音開始時間
	SilenceEnd        *int64         `json:"silence_end,omitempty"`    // 靜音結束時間
	MuteStart         *int64         `json:"mute_start,omitempty"`     // 抑制開始時間
	MuteEnd           *int64         `json:"mute_end,omitempty"`       // 抑制結束時間
	Labels            JSONMap        `json:"labels"`                   // 其他標籤
	Contacts          []AlertContact `json:"contacts,omitempty"`       // 通知對象
}
