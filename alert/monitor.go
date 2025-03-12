package alert

import (
	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/models"
)

func (s *Service) ListRuleState(realm string) ([]models.RuleState, error) {
	//TODO: 實作 Alert State 查詢
	return nil, nil
}

func (s *Service) ListAlertHistory(realm string) ([]models.TriggeredLog, error) {
	//TODO: 實作 Triggered Log 查詢
	return nil, nil
}

func (s *Service) GetMetricRuleCategoryOptions() ([]models.OptionResponse, error) {
	result := []models.OptionResponse{}
	categoryMap := make(map[string]bool)

	// 遍歷所有規則，收集不重複的類別
	for _, rule := range s.global.MetricRules {
		if !categoryMap[rule.Category] {
			categoryMap[rule.Category] = true
			result = append(result, models.OptionResponse{
				Text:  rule.Category,
				Value: rule.Category,
			})
		}
	}

	return result, nil
}

// 獲取指標規則選項
func (s *Service) GetMetricRuleOptions(category string) ([]models.OptionResponse, error) {
	result := []models.OptionResponse{}

	// 如果沒有指定類別，返回所有規則
	if category == "" {
		for uid, rule := range s.global.MetricRules {
			result = append(result, models.OptionResponse{
				Text:  rule.Name,
				Value: uid,
			})
		}
		return result, nil
	}

	// 如果指定了類別，返回該類別下的規則
	for uid, rule := range s.global.MetricRules {
		if rule.Category == category {
			result = append(result, models.OptionResponse{
				Text:  rule.Name,
				Value: uid,
			})
		}
	}

	if len(result) == 0 {
		return nil, apierrors.ErrNotFound
	}

	return result, nil
}

func (s *Service) GetMetricRule(uid string) (*models.MetricRule, error) {
	// 直接從 global.MetricRules 中獲取
	rule, exists := s.global.MetricRules[uid]
	if exists {
		return &rule, nil
	}

	return nil, apierrors.ErrNotFound
}
