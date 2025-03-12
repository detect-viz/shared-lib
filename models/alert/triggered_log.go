package alert

import (
	"github.com/detect-viz/shared-lib/models/common"
)

type TriggeredLog struct {
	RealmName           string         `json:"realm_name" gorm:"index"`
	ID                  []byte         `json:"id" gorm:"primaryKey"`
	TriggeredAt         int64          `json:"triggered_at"` //first
	LastTriggeredAt     int64          `json:"last_triggered_at"`
	ResolvedAt          *int64         `json:"resolved_at,omitempty"`
	NotifyLogID         *[]byte        `json:"notify_log_id"`
	ResolvedNotifyLogID *[]byte        `json:"resolved_notify_log_id"`
	NotifyState         string         `json:"notify_state"`
	ResolvedNotifyState *string        `json:"resolved_notify_state"`
	ResourceName        string         `json:"resource_name"`
	PartitionName       string         `json:"partition_name"`
	MetricRuleUID       string         `json:"metric_rule_uid"`
	RuleID              []byte         `json:"rule_id"`
	RuleSnapshot        common.JSONMap `json:"rule_snapshot" gorm:"type:json"`
	RuleStateSnapshot   common.JSONMap `json:"rule_state_snapshot" gorm:"type:json"`
	Severity            string         `json:"severity"`
	TriggeredValue      float64        `json:"triggered_value"`
	ResolvedValue       *float64       `json:"resolved_value"`
	Threshold           float64        `json:"threshold"`
	common.AuditTimeModel
}
