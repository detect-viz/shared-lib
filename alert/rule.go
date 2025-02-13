package alert

import (
	"shared-lib/models"
	"sync"
)

var (
	// AlertRules 告警規則映射表 realm -> rules
	AlertRules = make(map[string][]models.AlertRule)
	// MetricRules 指標規則映射表 id -> rule
	MetricRules = make(map[int64]models.MetricRule)
	ruleMutex   sync.RWMutex
)

// GetAlertRules 獲取指定域的告警規則
func GetAlertRules(realm string) []models.AlertRule {
	ruleMutex.RLock()
	defer ruleMutex.RUnlock()
	return AlertRules[realm]
}

// SetAlertRules 設置指定域的告警規則
func SetAlertRules(realm string, rules []models.AlertRule) {
	ruleMutex.Lock()
	defer ruleMutex.Unlock()
	AlertRules[realm] = rules
}

// GetMetricRule 根據 ID 獲取指標規則
func GetMetricRule(id int64) (models.MetricRule, bool) {
	ruleMutex.RLock()
	defer ruleMutex.RUnlock()
	rule, exists := MetricRules[id]
	return rule, exists
}

// SetMetricRules 設置所有指標規則
func SetMetricRules(rules []models.MetricRule) {
	ruleMutex.Lock()
	defer ruleMutex.Unlock()
	// 清空現有規則
	MetricRules = make(map[int64]models.MetricRule)
	// 添加新規則
	for _, rule := range rules {
		MetricRules[rule.ID] = rule
	}
}
