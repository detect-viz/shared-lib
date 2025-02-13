package label

// Label 自定義標籤
type Label struct {
	ID        int64  `json:"id" gorm:"primaryKey"`
	RealmName string `json:"realm_name"`
	Key       string `json:"key"`   // 標籤名稱
	Value     string `json:"value"` // 標籤值
}
