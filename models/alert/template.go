package alert

import "time"

// Template 通知模板
type Template struct {
	RealmName string    `json:"realm_name" gorm:"default:master"`
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	IsDefault bool      `json:"is_default" gorm:"default:false;index"`
	Name      string    `json:"name"`
	Format    string    `json:"format" gorm:"type:enum('html','text','markdown','json')"`
	RuleState string    `json:"rule_state" gorm:"type:enum('alerting','resolved','normal','disabled')"`
	Title     string    `json:"title" gorm:"type:text"`
	Message   string    `json:"message" gorm:"type:text"`
	CreatedAt time.Time `json:"-" form:"created_at"`
	UpdatedAt time.Time `json:"-" form:"updated_at"`
}
