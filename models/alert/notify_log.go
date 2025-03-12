package alert

import "github.com/detect-viz/shared-lib/models/common"

type NotifyLog struct {
	RealmName       string                   `json:"realm_name" gorm:"index"`
	ID              []byte                   `json:"id" gorm:"primaryKey"`
	SentAt          *int64                   `json:"sent_at,omitempty"`
	State           string                   `json:"state"`
	RetryCounter    int                      `json:"retry_counter" gorm:"default:0"`
	LastRetryAt     *int64                   `json:"last_retry_at,omitempty"`
	ErrorMessages   *common.JSONMap          `json:"error_messages" gorm:"type:json"`
	TriggeredLogIDs []map[string]interface{} `json:"triggered_log_ids" gorm:"type:json"`
	ContactID       []byte                   `json:"contact_id"`
	ChannelType     string                   `json:"channel_type"`
	ContactSnapshot common.JSONMap           `json:"contact_snapshot" gorm:"type:json"`
}
