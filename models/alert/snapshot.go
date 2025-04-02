package alert

import "github.com/detect-viz/shared-lib/models/common"

type RuleSnapshot struct {
	RealmName     string    `json:"realm_name"`
	ID            []byte    `json:"id" gorm:"primaryKey"`
	TargetID      []byte    `json:"target_id"`
	MetricRuleUID string    `json:"metric_rule_uid"`
	CreateType    string    `json:"create_type"`
	AutoApply     bool      `json:"auto_apply" gorm:"default:false"`
	Enabled       bool      `json:"enabled" gorm:"default:1"`
	InfoThreshold *float64  `json:"info_threshold"`
	WarnThreshold *float64  `json:"warn_threshold"`
	CritThreshold float64   `json:"crit_threshold"`
	Times         int       `json:"times" gorm:"default:3"`
	Duration      string    `json:"duration" gorm:"default:'5m'"`
	SilencePeriod string    `json:"silence_period" gorm:"default:'1h'"`
	Contacts      []Contact `json:"contacts" gorm:"many2many:rule_contacts"`
	Target        Target    `json:"target" gorm:"foreignKey:TargetID"`
	common.AuditUserModel
	common.AuditTimeModel
}

type ContactSnapshot struct {
	ID           []byte         `json:"id" gorm:"primaryKey"`
	RealmName    string         `json:"realm_name" gorm:"default:master"`
	Name         string         `json:"name"`
	ChannelType  string         `json:"channel_type"`
	Enabled      bool           `json:"enabled" gorm:"default:1"`
	SendResolved bool           `json:"send_resolved" gorm:"default:1"`
	MaxRetry     int            `json:"max_retry" gorm:"default:3"`
	RetryDelay   string         `json:"retry_delay" gorm:"default:5m"`
	Config       common.JSONMap `json:"config" gorm:"type:json"`
	Severities   SeveritySet    `json:"severities" gorm:"type:set('info','warn','crit');default:'crit'"`
	common.AuditUserModel
	common.AuditTimeModel
}

type RuleStateSnapshot struct {
	RuleID                []byte   `json:"rule_id" gorm:"foreignKey:RuleID"`
	State                 string   `json:"state"`
	ContactState          string   `json:"contact_state"`
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
