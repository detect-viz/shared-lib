package alert

import (
	"time"

	"gorm.io/gorm"
)

// AlertRule 告警規則
type AlertRule struct {
	ID              int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	RealmName       string   `json:"realm_name" gorm:"default:master"`
	Name            string   `json:"name"`
	Enabled         bool     `json:"enabled" gorm:"default:1"`
	IsJoint         bool     `json:"is_joint" gorm:"default:false"` // 是否為聯合規則(僅預留)
	ResourceGroupID int64    `json:"resource_group_id"`
	MetricRuleID    int64    `json:"metric_rule_id"`
	InfoThreshold   *float64 `json:"info_threshold"`
	WarnThreshold   *float64 `json:"warn_threshold"`
	CritThreshold   *float64 `json:"crit_threshold"`
	Duration        *int     `json:"duration"`
	SilencePeriod   string   `json:"silence_period" gorm:"default:'1h'"`
	Times           int      `json:"times" gorm:"default:3"` // 最大告警次數，超過則靜音

	//* 定義關聯關係
	MetricRule       MetricRule        `json:"metric_rule" gorm:"foreignKey:MetricRuleID"`
	AlertRuleDetails []AlertRuleDetail `json:"alert_rule_details" gorm:"foreignKey:AlertRuleID"`
	Contacts         []AlertContact    `json:"contacts" gorm:"many2many:alert_rule_contacts"`

	CreatedAt time.Time      `json:"-" form:"created_at"`
	UpdatedAt time.Time      `json:"-" form:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// AlertRuleDetail 告警規則詳細
type AlertRuleDetail struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	RealmName     string    `json:"realm_name" gorm:"default:master"`
	AlertRuleID   int64     `json:"alert_rule_id"`
	ResourceName  string    `json:"resource_name"`
	PartitionName *string   `json:"partition_name"`
	CreatedAt     time.Time `json:"-" form:"created_at"`
	UpdatedAt     time.Time `json:"-" form:"updated_at"`
}

// AlertRuleLabel 告警規則自定義標籤
type AlertRuleLabel struct {
	RealmName string  `json:"realm_name" gorm:"default:master"`
	LabelID   int64   `json:"label_id"`        // 標籤 ID
	RuleID    int64   `json:"rule_id"`         // 告警規則 ID
	Labels    JSONMap `json:"labels" gorm:"-"` // 標籤名稱
}
