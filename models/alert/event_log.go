package alert

import (
	"database/sql/driver"
	"encoding/json"

	"shared-lib/models/common"
)

// TriggerLog 告警觸發記錄
type TriggerLog struct {
	common.Common
	ID                uint           `json:"id" gorm:"primaryKey"`
	UUID              string         `json:"uuid"`                     // 唯一識別碼
	Timestamp         int64          `json:"timestamp"`                // 觸發時間
	FirstTriggerTime  int64          `json:"first_trigger_time"`       // 首次觸發時間
	RuleID            int64          `json:"rule_id"`                  // 規則 ID
	RuleName          string         `json:"rule_name"`                // 規則名稱
	ResourceGroupName string         `json:"resource_group_name"`      // 資源群組名稱
	ResourceName      string         `json:"resource_name"`            // 資源名稱
	PartitionName     *string        `json:"partition_name,omitempty"` // 分區名稱
	MetricName        string         `json:"metric_name"`              // 監控指標
	Value             float64        `json:"value"`                    // 當前值
	Threshold         float64        `json:"threshold"`                // 閾值
	Unit              string         `json:"unit"`                     // 單位
	Severity          string         `json:"severity"`                 // 嚴重程度(info/warn/crit)
	Duration          int            `json:"duration"`                 // 持續時間(秒)
	Status            string         `json:"status"`                   // 狀態(alerting/muting/pending)
	NotifyStatus      string         `json:"notify_status"`            // 通知狀態(alerting/muting/pending)
	SilenceStart      *int64         `json:"silence_start,omitempty"`  // 靜音開始時間
	SilenceEnd        *int64         `json:"silence_end,omitempty"`    // 靜音結束時間
	PendingEnd        *int64         `json:"pending_end,omitempty"`    // 抑制結束時間
	ResolvedTime      *int64         `json:"resolved_time,omitempty"`  // 恢復時間
	Labels            JSONMap        `json:"labels"`                   // 標籤
	Contacts          []AlertContact `json:"contacts" gorm:"-"`        // 聯絡人列表
}

// NotificationLog 通知日誌
type NotificationLog struct {
	common.Common
	ID            int64         `json:"id" gorm:"primaryKey"`
	UUID          string        `json:"uuid" gorm:"type:varchar(36);uniqueIndex"`
	Timestamp     int64         `json:"timestamp"`                                               // 通知時間
	ContactID     int64         `json:"contact_id"`                                              // 聯絡人 ID
	ContactName   string        `json:"contact_name"`                                            // 聯絡人名稱
	ChannelType   string        `json:"contact_type"`                                            // 通知類型(email/teams/webhook)
	ContactConfig JSONMap       `json:"contact_config" gorm:"-"`                                 // 通知設定(不存DB)
	Target        string        `json:"target"`                                                  // 通知目標(email/webhook url)
	Severity      string        `json:"severity"`                                                // 嚴重程度(info/warn/crit)
	Subject       string        `json:"subject"`                                                 // 通知標題
	Body          string        `json:"body" gorm:"-"`                                           // 通知內容(不存DB)
	FilePath      *string       `json:"file_path,omitempty"`                                     // 通知內容檔案路徑
	Status        string        `json:"status" gorm:"type:enum('sent','failed','pending')"`      // 通知狀態
	SentAt        *int64        `json:"sent_at,omitempty"`                                       // 發送時間
	Error         *string       `json:"error,omitempty"`                                         // 錯誤訊息
	NotifyRetry   int           `json:"notify_retry" gorm:"default:0"`                           // 重試次數
	RetryDeadline int64         `json:"retry_deadline"`                                          // 重試截止時間
	TriggerLogs   []*TriggerLog `json:"trigger_logs" gorm:"many2many:notification_trigger_logs"` // 關聯的觸發記錄
}

// NotificationTriggerLog 關聯表 (多對多)
type NotificationTriggerLog struct {
	NotificationID int64 `gorm:"primaryKey"`
	TriggerLogID   int64 `gorm:"primaryKey"`
}

// JSONMap 讓 GORM 正確處理 JSON 欄位
type JSONMap map[string]string

func (j *JSONMap) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}
