package resource

type ResourceGroup struct {
	RealmName string     `json:"realm_name"       form:"realm_name"` // from sso
	Type      string     `json:"type"             form:"type"`
	ID        int        `json:"id"               form:"id"           gorm:"primaryKey" `
	Name      string     `json:"name"             form:"name"`
	Resources []Resource `json:"resources"  form:"resources" gorm:"foreignKey:ResourceGroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Resource struct {
	Name            string        `json:"name"                form:"name"`
	ResourceGroupID int           `json:"resource_group_id"      form:"resource_group_id"`
	ResourceGroup   ResourceGroup `json:"-"         form:"resource_groups" gorm:"foreignKey:ID;references:resource_group_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
