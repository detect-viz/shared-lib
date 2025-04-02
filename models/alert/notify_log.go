package alert

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/detect-viz/shared-lib/models/common"
)

// TriggeredLogIDsMap 讓 GORM 正確處理 TriggeredLogIDs 欄位
type TriggeredLogIDsMap []map[string]interface{}

// Scan 從資料庫讀取
func (t *TriggeredLogIDsMap) Scan(value interface{}) error {
	if value == nil {
		*t = TriggeredLogIDsMap{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan TriggeredLogIDsMap: unsupported type")
	}

	return json.Unmarshal(bytes, t)
}

// Value 存入資料庫
func (t TriggeredLogIDsMap) Value() (driver.Value, error) {
	if len(t) == 0 {
		return "[]", nil
	}
	return json.Marshal(t)
}

type NotifyLog struct {
	RealmName       string             `json:"realm_name" gorm:"index"`
	ID              []byte             `json:"id" gorm:"primaryKey"`
	SentAt          *int64             `json:"sent_at,omitempty"`
	State           string             `json:"state"`
	RetryCounter    int                `json:"retry_counter" gorm:"default:0"`
	LastRetryAt     *int64             `json:"last_retry_at,omitempty"`
	ErrorMessages   *common.JSONMap    `json:"error_messages" gorm:"type:json"`
	TriggeredLogIDs TriggeredLogIDsMap `json:"triggered_log_ids" gorm:"type:json"`
	ContactID       []byte             `json:"contact_id"`
	ChannelType     string             `json:"channel_type"`
	ContactSnapshot common.JSONMap     `json:"contact_snapshot" gorm:"type:json"`
}
