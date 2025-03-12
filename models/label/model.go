package label

// LabelKey (標籤鍵)
type LabelKey struct {
	ID        int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	RealmName string       `json:"realm_name" gorm:"index"`                   // 多租戶支持
	KeyName   string       `json:"key_name" gorm:"uniqueIndex:idx_realm_key"` // 標籤名稱（唯一）
	Values    []LabelValue `json:"values" gorm:"foreignKey:LabelKeyID"`
	CreatedAt int64        `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64        `json:"updated_at" gorm:"autoUpdateTime"`
}

type LabelValue struct {
	ID         int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	LabelKeyID int64    `json:"label_key_id" gorm:"index"`
	Value      string   `json:"value"`
	LabelKey   LabelKey `json:"label_key" gorm:"foreignKey:LabelKeyID;references:ID"`
}
