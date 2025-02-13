package alert

import (
	"shared-lib/models"

	"go.uber.org/zap"
)

func (c *Service) CheckAmplitude(rule models.CheckRule, file models.FileInfo, metrics map[string]interface{}) bool {
	// 1. 獲取當前值和時間戳
	value, timestamp, err := c.GetCurrentValue(rule.Metric, metrics)
	if err != nil {
		return false
	}

	// 2. 使用狀態管理器處理
	_, exceeded, err := c.stateManager.GetAndUpdateState(rule, value, timestamp)
	if err != nil {
		c.logger.Error("檢查告警狀態失敗", zap.Error(err))
		return false
	}

	return exceeded
}
