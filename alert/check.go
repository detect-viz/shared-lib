package alert

import (
	"fmt"
	"shared-lib/models"
)

// CheckThreshold 檢查閾值
func CheckThreshold(rule models.CheckRule, operator string, value float64) bool {
	switch operator {
	case ">", ">=":
		if rule.CritThreshold != nil && value > *rule.CritThreshold {
			return true
		}
		if rule.WarnThreshold != nil && value > *rule.WarnThreshold {
			return true
		}
		if rule.InfoThreshold != nil && value > *rule.InfoThreshold {
			return true
		}
	case "<", "<=":
		if rule.CritThreshold != nil && value < *rule.CritThreshold {
			return true
		}
		if rule.WarnThreshold != nil && value < *rule.WarnThreshold {
			return true
		}
		if rule.InfoThreshold != nil && value < *rule.InfoThreshold {
			return true
		}
	}
	return false
}

// getCurrentValue 從指標數據中獲取當前值和時間戳
func (c *Service) GetCurrentValue(metricName string, metrics map[string]interface{}) (float64, int64, error) {
	data, ok := metrics[metricName].([]map[string]interface{})
	if !ok || len(data) == 0 {
		return 0, 0, fmt.Errorf("no data for metric: %s", metricName)
	}

	latestData := data[len(data)-1]
	value, ok := latestData["value"].(float64)
	if !ok {
		return 0, 0, fmt.Errorf("invalid value type for metric: %s", metricName)
	}

	timestamp, ok := latestData["timestamp"].(int64)
	if !ok {
		return 0, 0, fmt.Errorf("invalid timestamp type for metric: %s", metricName)
	}

	return value, timestamp, nil
}
