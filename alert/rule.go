package alert

import (
	"sync"

	"github.com/detect-viz/shared-lib/models"
)

var (
	// AlertRules 告警規則映射表 realm -> rules
	AlertRules = make(map[int64][]models.AlertRule)
	// MetricRules 指標規則映射表 id -> rule
	MetricRules = make(map[int64]models.MetricRule)
	ruleMutex   sync.RWMutex
)

// GetAlertRules 獲取指定域的告警規則
func (s *Service) GetAlertRules(realm string) []models.AlertRule {
	ruleMutex.RLock()
	defer ruleMutex.RUnlock()
	rules, err := s.db.GetAlertRulesByRealm(realm)
	if err != nil {
		return []models.AlertRule{}
	}
	return rules
}

// GetAlertRuleByID 根據 ID 獲取告警規則
func (s *Service) GetAlertRuleByID(id int64) (models.AlertRule, bool) {
	ruleMutex.RLock()
	defer ruleMutex.RUnlock()
	rule, err := s.db.GetAlertRuleByID(id)
	if err != nil {
		return models.AlertRule{}, false
	}
	return rule, true
}

// GetMetricRule 根據 ID 獲取指標規則
// GetMetricRule 根據 ID 獲取指標規則
func (s *Service) GetMetricRule(id int64) (models.MetricRule, bool) {
	ruleMutex.RLock()
	defer ruleMutex.RUnlock()
	rule, exists := MetricRules[id]
	return rule, exists
}

// SetMetricRules 設置所有指標規則
func (s *Service) SetMetricRules(rules []models.MetricRule) {
	ruleMutex.Lock()
	defer ruleMutex.Unlock()
	MetricRules = make(map[int64]models.MetricRule)
	for _, rule := range rules {
		MetricRules[rule.ID] = rule
	}
}

// GetAlertRule 根據 ID 獲取告警規則
func (s *Service) GetAlertRule(id int64) (models.AlertRule, bool) {
	ruleMutex.RLock()
	defer ruleMutex.RUnlock()
	rule, err := s.db.GetAlertRuleByID(id)
	if err != nil {
		return models.AlertRule{}, false
	}
	return rule, true
}

// CreateAlertRule 創建告警規則
func (s *Service) Create(rule models.AlertRule) {
	ruleMutex.Lock()
	defer ruleMutex.Unlock()
	s.db.CreateAlertRule(&rule)
}

// DeleteAlertRule 刪除告警規則
func (s *Service) Delete(id int64) {
	ruleMutex.Lock()
	defer ruleMutex.Unlock()
	s.db.DeleteAlertRule(id)
}

// UpdateAlertRule 更新告警規則
func (s *Service) Update(rule models.AlertRule) {
	ruleMutex.Lock()
	defer ruleMutex.Unlock()
	s.db.UpdateAlertRule(&rule)
}
