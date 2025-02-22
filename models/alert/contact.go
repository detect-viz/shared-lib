package alert

import (
	"time"

	"gorm.io/gorm"
)

// AlertContact 通知聯繫人
type AlertContact struct {
	ID           int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	RealmName    string `json:"realm_name" gorm:"default:master"`
	Name         string `json:"name"` // 聯繫人名稱
	Type         string `json:"type"` // 通知類型
	Enabled      bool   `json:"enabled" gorm:"default:1"`
	SentResolved bool   `json:"sent_resolved" gorm:"default:1"`
	MaxRetry     int    `json:"max_retry" gorm:"default:3"`
	RetryDelay   int    `json:"retry_delay" gorm:"default:300"`

	Details    JSONMap                `json:"details" gorm:"type:json"`
	Severities []AlertContactSeverity `json:"severities"  gorm:"many2many:alert_contact_severities"`
	CreatedAt  time.Time              `json:"-" form:"created_at"`
	UpdatedAt  time.Time              `json:"-" form:"updated_at"`
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
