package alert

// CheckRule - 異常檢測規則
type CheckRule struct {
	UUID          string   `json:"uuid"`                     // 唯一識別碼
	Timestamp     int64    `json:"timestamp"`                // 異常檢測時間
	ResourceGroup string   `json:"resource_group"`           // 資源群組
	ResourceName  string   `json:"resource_name"`            // 監控的主機/設備
	PartitionName *string  `json:"partition_name,omitempty"` // 分區名稱 (可選)
	Metric        string   `json:"metric"`                   // 監控指標
	CheckType     string   `json:"check_type"`               // absolute / amplitude
	Operator      string   `json:"operator"`                 // > / < / = / !=
	Value         float64  `json:"value"`                    // 異常數值
	Threshold     float64  `json:"threshold"`                // 閾值
	Unit          string   `json:"unit"`                     // 單位 (% / MB / ms)
	Duration      int      `json:"duration"`                 // 異常持續時間
	Severity      string   `json:"severity"`                 // info / warn / crit
	Status        string   `json:"status"`                   // alerting / normal / silenced
	NotifyStatus  string   `json:"notify_status"`            // pending / sent / failed
	RuleID        int64    `json:"rule_id"`                  // 關聯的告警規則 ID
	RuleName      string   `json:"rule_name"`                // 規則名稱
	SilenceStart  *int64   `json:"silence_start,omitempty"`  // 靜音開始時間 (Unix Timestamp)
	SilenceEnd    *int64   `json:"silence_end,omitempty"`    // 靜音結束時間 (Unix Timestamp)
	MuteStart     *int64   `json:"mute_start,omitempty"`     // 抑制開始時間 (Unix Timestamp)
	MuteEnd       *int64   `json:"mute_end,omitempty"`       // 抑制結束時間 (Unix Timestamp)
	Labels        JSONMap  `json:"labels"`                   // 其他標籤
	InfoThreshold *float64 `json:"info_threshold,omitempty"` // 資訊閾值
	WarnThreshold *float64 `json:"warn_threshold,omitempty"` // 警告閾值
	CritThreshold *float64 `json:"crit_threshold,omitempty"` // 嚴重閾值
	Contacts      []string `json:"contacts,omitempty"`       // 通知對象
}
