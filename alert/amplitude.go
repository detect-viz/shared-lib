package alert

import (
	"math"
	"shared-lib/models"
	"strconv"

	"go.uber.org/zap"
)

func (c *Service) CheckAmplitude(rule models.CheckRule, file models.FileInfo, metrics []map[string]interface{}) bool {
	if len(metrics) < 2 { // 振幅檢查需要至少兩個點
		return false
	}

	// 取最新的兩筆數據
	latest := metrics[len(metrics)-1]
	previous := metrics[len(metrics)-2]

	latestValue, ok := latest["value"].(string)
	if !ok {
		c.logger.Error("無法解析最新指標值")
		return false
	}
	latestTimestampStr, ok := latest["timestamp"].(string)
	if !ok {
		c.logger.Error("無法解析最新時間戳")
		return false
	}
	latestTimestamp, err := strconv.ParseInt(latestTimestampStr, 10, 64)
	if err != nil {
		c.logger.Error("轉換時間戳失敗", zap.Error(err))
		return false
	}

	previousValue, ok := previous["value"].(string)
	if !ok {
		c.logger.Error("無法解析前一個指標值")
		return false
	}

	// 轉換為 float64
	latestFloat, err := strconv.ParseFloat(latestValue, 64)
	if err != nil {
		c.logger.Error("轉換最新指標值失敗", zap.Error(err))
		return false
	}
	previousFloat, err := strconv.ParseFloat(previousValue, 64)
	if err != nil {
		c.logger.Error("轉換前一個指標值失敗", zap.Error(err))
		return false
	}

	// 計算振幅
	amplitude := math.Abs(latestFloat - previousFloat)

	// 使用狀態管理器處理
	_, exceeded, err := c.stateManager.GetAndUpdateState(rule, amplitude, latestTimestamp)
	if err != nil {
		c.logger.Error("檢查告警狀態失敗", zap.Error(err))
		return false
	}

	return exceeded
}
