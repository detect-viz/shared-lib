package alert

import "github.com/detect-viz/shared-lib/models/common"

// 通知模板
type Template struct {
	RealmName  string `yaml:"realm_name" json:"realm_name"`
	ID         []byte `yaml:"id" json:"id" gorm:"primary_key"`
	Name       string `yaml:"name" json:"name"`
	CreateType string `yaml:"create_type" json:"create_type" gorm:"type:enum('system','user');not null"`
	FormatType string `yaml:"format_type" json:"format_type" gorm:"type:enum('html','text','markdown','json');not null"`
	RuleState  string `yaml:"rule_state" json:"rule_state" gorm:"type:enum('alerting','resolved');not null"`
	Title      string `yaml:"title" json:"title" gorm:"type:text;not null"`
	Message    string `yaml:"message" json:"message" gorm:"type:text;not null"`
	common.AuditUserModel
	common.AuditTimeModel
}
