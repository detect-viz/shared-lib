package alert

import (
	"shared-lib/models"
	"strconv"

	"go.uber.org/zap"
)

// checkAbsolute 絕對值檢查
func (c *Service) CheckAbsolute(rule models.CheckRule, file models.FileInfo, metrics []map[string]interface{}) bool {
	if len(metrics) == 0 {
		return false
	}

	// 取最新的一筆數據
	latest := metrics[len(metrics)-1]
	value, ok := latest["value"].(string)
	if !ok {
		c.logger.Error("無法解析指標值")
		return false
	}
	timestampStr, ok := latest["timestamp"].(string)
	if !ok {
		c.logger.Error("無法解析時間戳")
		return false
	}
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		c.logger.Error("轉換時間戳失敗", zap.Error(err))
		return false
	}
	// 轉換為 float64
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		c.logger.Error("轉換指標值失敗", zap.Error(err))
		return false
	}

	// 使用狀態管理器處理
	_, exceeded, err := c.stateManager.GetAndUpdateState(rule, floatValue, timestamp)
	if err != nil {
		c.logger.Error("檢查告警狀態失敗", zap.Error(err))
		return false
	}

	return exceeded
}
