package alert

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// TriggerLog 告警觸發記錄
type TriggerLog struct {
	RealmName         string  `json:"realm_name"`                               // 告警規則所在的 realm
	UUID              string  `json:"uuid" gorm:"type:varchar(36);uniqueIndex"` // 唯一識別碼
	Timestamp         int64   `json:"timestamp"`                                // 觸發時間
	RuleID            int64   `json:"rule_id"`                                  // 規則 ID
	RuleName          string  `json:"rule_name"`                                // 規則名稱
	ResourceGroupName string  `json:"resource_group_name"`                      // 資源群組名稱
	ResourceName      string  `json:"resource_name"`                            // 資源名稱
	PartitionName     *string `json:"partition_name,omitempty"`                 // 分區名稱
	MetricName        string  `json:"metric_name"`                              // 監控指標
	Unit              string  `json:"unit"`                                     // 單位
	TriggerValue      float64 `json:"trigger_value"`                            // 觸發值
	Threshold         float64 `json:"threshold"`                                // 閾值
	Severity          string  `json:"severity"`                                 // 嚴重程度(info/warn/crit)
	FirstTriggerTime  int64   `json:"first_trigger_time"`                       // 首次觸發時間
	SilenceStart      *int64  `json:"silence_start,omitempty"`                  // 靜音開始時間
	SilenceEnd        *int64  `json:"silence_end,omitempty"`                    // 靜音結束時間
	MuteStart         *int64  `json:"mute_start,omitempty"`                     // 抑制開始時間
	MuteEnd           *int64  `json:"mute_end,omitempty"`                       // 抑制結束時間
	ContactState      string  `json:"contact_state"`                            // 通知狀態(normal/muting/silence)

	//* 以下為後續更新欄位
	NotifyState         string `json:"notify_state"`            // 通知狀態(solved/failed)
	ResolvedNotifyState string `json:"resolved_notify_state"`   // 恢復通知狀態(solved/failed)
	ResolvedTime        *int64 `json:"resolved_time,omitempty"` // 恢復時間

	//* 關聯
	Labels   JSONMap        `json:"labels" gorm:"-"`   // 標籤
	Contacts []AlertContact `json:"contacts" gorm:"-"` // 聯絡人列表

	CreatedAt time.Time `json:"-" form:"created_at"`
	UpdatedAt time.Time `json:"-" form:"updated_at"`
}

// NotifyLog 通知日誌
type NotifyLog struct {
	RealmName         string `json:"realm_name"` // 告警規則所在的 realm
	UUID              string `json:"uuid" gorm:"type:varchar(36);uniqueIndex"`
	Timestamp         int64  `json:"timestamp"`    // 通知時間
	ContactID         int64  `json:"contact_id"`   // 聯絡人 ID
	ContactName       string `json:"contact_name"` // 聯絡人名稱
	ContactType       string `json:"contact_type"` // 通知類型(email/teams/webhook)
	ContactMaxRetry   int    `json:"max_retry" gorm:"default:3"`
	ContactRetryDelay int    `json:"retry_delay" gorm:"default:300"`

	Title          string `json:"title"`                    // 通知標題
	Message        string `json:"message" gorm:"type:text"` // 通知內容
	RuleState      string `json:"rule_state"`               // 規則狀態(alerting/resolved)
	NotifyState    string `json:"notify_state"`             // 通知狀態(solved/failed)
	LastRetryTime  int64  `json:"last_retry_time"`          // 最後重試時間
	LastFailedTime int64  `json:"last_failed_time"`         // 最後失敗時間
	SentAt         *int64 `json:"sent_at,omitempty"`        // 發送時間

	Error        string `json:"error,omitempty"`                // 錯誤訊息(failed)
	RetryCounter int    `json:"retry_counter" gorm:"default:0"` // 重試次數

	CreatedAt time.Time `json:"-" form:"created_at"`
	UpdatedAt time.Time `json:"-" form:"updated_at"`

	//* 關聯
	ContactConfig JSONMap       `json:"contact_config" gorm:"-"`                           // 通知設定
	TriggerLogs   []*TriggerLog `json:"trigger_logs" gorm:"many2many:notify_trigger_logs"` // 關聯的觸發記錄
}

// NotifyTriggerLog 關聯表 (多對多)
type NotifyTriggerLog struct {
	NotifyLogUUID  string `json:"notify_log_uuid" gorm:"primaryKey"`
	TriggerLogUUID string `json:"trigger_log_uuid" gorm:"primaryKey"`
}

// JSONMap 讓 GORM 正確處理 JSON 欄位
type JSONMap map[string]string

func (j *JSONMap) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}
