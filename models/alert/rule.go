package alert

import (
	"shared-lib/models/common"

	"gorm.io/gorm"
)

// AlertRule 告警規則
type AlertRule struct {
	common.Common
	ID               int64             `json:"id" gorm:"primaryKey;autoIncrement"`
	RealmName        string            `json:"realm_name" gorm:"default:master"`
	Name             string            `json:"name"`
	Enabled          bool              `json:"enabled" gorm:"default:1"`
	IsJoint          bool              `json:"is_joint" gorm:"default:false"` // 是否為聯合規則
	ResourceGroupID  int64             `json:"resource_group_id"`
	MetricRuleID     int64             `json:"metric_rule_id"`
	InfoThreshold    *float64          `json:"info_threshold"`
	WarnThreshold    *float64          `json:"warn_threshold"`
	CritThreshold    *float64          `json:"crit_threshold"`
	Duration         *int              `json:"duration"`
	SilencePeriod    string            `json:"silence_period" gorm:"default:'1h'"`
	Times            int               `json:"times" gorm:"default:3"` // 最大告警次數，超過則靜音
	MetricRule       MetricRule        `json:"metric_rule" gorm:"-"`
	AlertRuleDetails []AlertRuleDetail `json:"alert_rule_details" gorm:"-"`
	Contacts         []AlertContact    `json:"contacts" gorm:"-"`   // 關聯的通知聯繫人
	MuteRules        []AlertMute       `json:"mute_rules" gorm:"-"` // 關聯的抑制規則
	DeletedAt        gorm.DeletedAt    `json:"deleted_at" gorm:"index"`
}

// AlertRuleDetail 告警規則詳細
type AlertRuleDetail struct {
	common.Common
	ID                  int64   `json:"id" gorm:"primaryKey"`
	RealmName           string  `json:"realm_name" gorm:"default:master"`
	AlertRuleID         int64   `json:"alert_rule_id"`
	ResourceName        string  `json:"resource_name"`
	PartitionName       *string `json:"partition_name"`
	SilenceTimeRange    *string `json:"silence_time_range"` // 格式存 "startTime-endTime"
	LastStateChangeTime *int64  `json:"last_state_change_time"`
	LastAlertTime       *int64  `json:"last_alert_time"`
	LastAlertSeverity   *string `json:"last_alert_severity"`
	EventDuration       *int    `json:"event_duration"`
}
