package alert

// RuleConfig 代表 YAML 中的告警規則配置
type RuleConfig struct {
	Rules []RuleDefinition `yaml:"rules"`
}

// RuleDefinition 代表單個告警規則的定義
type RuleDefinition struct {
	Name              string            `yaml:"name"`
	ResourceGroupName string            `yaml:"resource_group_name"`
	MetricRuleID      string            `yaml:"metric_rule_id"`
	Thresholds        ThresholdConfig   `yaml:"thresholds"`
	Duration          int               `yaml:"duration"`
	Labels            map[string]string `yaml:"labels"`
}

// ThresholdConfig 代表不同級別的閾值配置
type ThresholdConfig struct {
	Info *float64 `yaml:"info,omitempty"`
	Warn *float64 `yaml:"warn,omitempty"`
	Crit *float64 `yaml:"crit,omitempty"`
}

// ToCheckRule 將 RuleDefinition 轉換為 CheckRule
func (rd *RuleDefinition) ToCheckRule() *CheckRule {
	return &CheckRule{
		RuleName:          rd.Name,
		ResourceGroupName: rd.ResourceGroupName, // name 轉 id
		MetricName:        rd.MetricRuleID,      // id 找 metric name
		Duration:          rd.Duration,
		Labels:            rd.Labels,
		InfoThreshold:     rd.Thresholds.Info,
		WarnThreshold:     rd.Thresholds.Warn,
		CritThreshold:     rd.Thresholds.Crit,
	}
}

// ToAlertRule 將 RuleDefinition 轉換為 AlertRule
func (rd *RuleDefinition) ToAlertRule() *AlertRule {
	return &AlertRule{
		Name: rd.Name,

		Duration:      &rd.Duration,
		InfoThreshold: rd.Thresholds.Info,
		WarnThreshold: rd.Thresholds.Warn,
		CritThreshold: rd.Thresholds.Crit,
	}
	//TODO: ResourceGroup: rd.ResourceGroupName, // name 轉 id
	//TODO: Labels:          rd.Labels, // 綁定 labels
	//TODO:MetricRule:      rd.MetricRuleID, // id 找 metric
}
