package alert

import (
	"shared-lib/models/common"

	"gorm.io/gorm"
)

// AlertContact 通知聯繫人
type AlertContact struct {
	common.Common
	ID         int64                  `json:"id" gorm:"primaryKey;autoIncrement"`
	RealmName  string                 `json:"realm_name" gorm:"default:master"`
	Name       string                 `json:"name"` // 聯繫人名稱
	Type       string                 `json:"type"` // 通知類型
	Enabled    bool                   `json:"enabled" gorm:"default:1"`
	Details    JSONMap                `json:"details" gorm:"type:json"`
	Severities []AlertContactSeverity `json:"severities"  gorm:"many2many:alert_contact_severities"`
	DeletedAt  gorm.DeletedAt         `json:"deleted_at" gorm:"index"`
}

// AlertContactSeverity (多對多關聯)
type AlertContactSeverity struct {
	AlertContactID int64  `json:"alert_contact_id"`
	Severity       string `json:"severity" gorm:"type:enum('info','warn','crit');unique"`
}

type AlertRuleContact struct {
	AlertRuleID    int64 `json:"alert_rule_id"`
	AlertContactID int64 `json:"alert_contact_id"`
}
