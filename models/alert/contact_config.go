package alert

// ContactConfig 代表 YAML 中的通知管道配置
type ContactConfig struct {
	Contacts []ContactDefinition `yaml:"contacts"`
}

// ContactDefinition 代表單個通知管道的定義
type ContactDefinition struct {
	Name       string                 `yaml:"name"`
	Type       string                 `yaml:"type"`
	Enabled    bool                   `yaml:"enabled"`
	Severities []string               `yaml:"severities"`
	Details    map[string]interface{} `yaml:"details"`
}

// ToAlertContact 將 ContactDefinition 轉換為 Contact
func (cd *ContactDefinition) ToAlertContact() *Contact {
	var severities []AlertContactSeverity
	for _, s := range cd.Severities {
		severities = append(severities, AlertContactSeverity{
			Severity: "level_" + s,
		})
	}

	j := make(JSONMap)
	for k, v := range cd.Details {
		j[k] = v.(string)
	}

	return &Contact{
		Name:       cd.Name,
		Type:       cd.Type,
		Enabled:    cd.Enabled,
		Details:    j,
		Severities: severities,
	}
}
