package label

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Label 自定義標籤
type Label struct {
	ID        int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	RealmName string         `json:"realm_name" gorm:"index"`                   // 多租戶支持
	KeyName   string         `json:"key_name" gorm:"uniqueIndex:idx_realm_key"` // 標籤名稱
	Value     datatypes.JSON `json:"value" gorm:"type:json"`                    // 標籤值，存 JSON
	CreatedAt int64          `json:"created_at"`
	UpdatedAt int64          `json:"updated_at"`
}

// BeforeCreate - 確保 JSON 格式正確
func (l *Label) BeforeCreate(tx *gorm.DB) (err error) {
	if l.Value == nil {
		l.Value = datatypes.JSON(`{}`)
	}
	return nil
}
