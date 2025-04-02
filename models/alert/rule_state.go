package alert

import "github.com/detect-viz/shared-lib/models/common"

type RuleState struct {
	RuleID                []byte   `json:"rule_id" gorm:"foreignKey:RuleID;index"`
	State                 string   `json:"state"`
	ContactState          string   `json:"contact_state"`
	ContactCounter        int      `json:"contact_counter"`
	SilenceStartAt        *int64   `json:"silence_start_at"`
	SilenceEndAt          *int64   `json:"silence_end_at"`
	LastCheckValue        float64  `json:"last_check_value"`
	LastTriggeredValue    *float64 `json:"last_triggered_value"`
	LastTriggeredSeverity *string  `json:"last_triggered_severity"`
	FirstTriggeredAt      *int64   `json:"first_triggered_at"`
	LastTriggeredAt       *int64   `json:"last_triggered_at"`
	LastTriggeredLogID    *[]byte  `json:"last_triggered_log_id" gorm:"type:binary(16)"`
	common.AuditTimeModel
}
