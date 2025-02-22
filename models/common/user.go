package common

type SSOUser struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Realm       string   `json:"realm"`
	RealmGroups []string `json:"relam_groups"`
	Roles       []string `json:"roles"`
	IsAdmin     bool     `json:"is_admin"`
	AccessHosts []string `json:"access_hosts"`
	OrgID       string   `json:"org_id"`
	OrgName     string   `json:"org_name"`
}
