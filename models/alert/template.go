package alert

import "shared-lib/models/common"

// AlertTemplate 通知模板
type AlertTemplate struct {
	common.Common
	RealmName       string `json:"realm_name" gorm:"default:master"`
	ID              int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	IsDefault       bool   `json:"is_default" gorm:"default:false;index"`
	Name            string `json:"name"`
	Format          string `json:"format" gorm:"type:enum('html','text','markdown','json')"`
	Status          string `json:"status" gorm:"type:enum('rule_alerting','rule_resolved')"`
	TitleTemplate   string `json:"title_template" gorm:"type:text"`
	MessageTemplate string `json:"message_template" gorm:"type:text"`
}
