package alert

import (
	"github.com/detect-viz/shared-lib/models/common"
)

// * 指標規則總覽
type MetricRuleOverview struct {
	Category       string         `json:"category"`         // 指標規則類型
	MetricRuleUID  string         `json:"metric_rule_uid"`  // 指標規則UID
	MetricRuleName string         `json:"metric_rule_name"` // 指標規則名稱
	RuleCount      int            `json:"rule_count"`       // 規則數量
	TargetCount    int            `json:"target_count"`     // 監控對象數量
	PartitionCount int            `json:"partition_count"`  // 監控分區數量
	ContactCount   int            `json:"contact_count"`    // 通知通道數量
	Rules          []RuleResponse `json:"rules"`            // 規則列表
}

// * 指標規則下的規則總覽
type RuleOverview struct {
	Enabled        bool     `json:"enabled"`         // 是否啟用
	AutoApply      bool     `json:"auto_apply"`      // 是否自動適用
	CreateType     string   `json:"create_type"`     // 創建方式
	Duration       string   `json:"duration"`        // 持續時間
	InfoThreshold  *string  `json:"info_threshold"`  // 加上單位
	WarnThreshold  *string  `json:"warn_threshold"`  // 加上單位
	CritThreshold  string   `json:"crit_threshold"`  // 加上單位
	RuleCount      int      `json:"rule_count"`      // 規則數量
	ResourceCount  int      `json:"resource_count"`  // 監控對象數量
	PartitionCount int      `json:"partition_count"` // 監控分區數量
	ContactCount   int      `json:"contact_count"`   // 通知通道數量
	Targets        []Target `json:"targets"`         // 監控對象
}

// * 規則設定
type RuleResponse struct {
	ID            string                `json:"id"`
	AutoApply     bool                  `json:"auto_apply"`
	Enabled       bool                  `json:"enabled"`
	InfoThreshold *float64              `json:"info_threshold"`
	WarnThreshold *float64              `json:"warn_threshold"`
	CritThreshold float64               `json:"crit_threshold"`
	Times         int                   `json:"times"`
	Duration      string                `json:"duration"`
	SilencePeriod string                `json:"silence_period"`
	MetricRule    MetricRule            `json:"metric_rule"`
	Target        Target                `json:"target"`
	Contacts      []RuleContactResponse `json:"contacts"`
}

// * 聯絡人設定
type RuleContactResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Severities []string `json:"severities"`
}

type Rule struct {
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
