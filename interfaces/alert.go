package interfaces

import (
	"shared-lib/models"
)

// 告警服務介面
type AlertService interface {
	// 告警檢查
	ProcessFile(file models.FileInfo) error
	ProcessTriggers() error

	// 告警規則管理
	GetAlertRules(realm string) []models.AlertRule
	SetAlertRules(realm string, rules []models.AlertRule)
	GetMetricRule(id int64) (models.MetricRule, bool)
	SetMetricRules(rules []models.MetricRule)
}

// 告警檢查器介面
type AlertChecker interface {
	Check(rule models.AlertRule, file models.FileInfo, metrics map[string]interface{}) bool
}
