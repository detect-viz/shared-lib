package alert

type MetricRule struct {
	UID                  string    `yaml:"uid" json:"uid"`
	Name                 string    `yaml:"name" json:"name"`
	Category             string    `yaml:"category" json:"category"`
	MatchDatasourceNames []string  `yaml:"match_datasource_names" json:"match_datasource_names"`
	MatchTargetPattern   string    `yaml:"match_target_pattern" json:"match_target_pattern"`
	DetectionType        string    `yaml:"detection_type" json:"detection_type"`
	MetricRawName        string    `yaml:"metric_raw_name" json:"metric_raw_name"`
	MetricDisplayName    string    `yaml:"metric_display_name" json:"metric_display_name"`
	RawUnit              string    `yaml:"raw_unit" json:"raw_unit"`
	DisplayUnit          string    `yaml:"display_unit" json:"display_unit"`
	Scale                float64   `yaml:"scale" json:"scale"`
	Duration             string    `yaml:"duration" json:"duration"`
	Operator             string    `yaml:"operator" json:"operator"`
	Thresholds           Threshold `yaml:"thresholds" json:"thresholds"`
}

type Threshold struct {
	Info *float64 `yaml:"info" json:"info"`
	Warn *float64 `yaml:"warn" json:"warn"`
	Crit float64  `yaml:"crit" json:"crit"`
}
