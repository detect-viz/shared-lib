package alert

import (
	"fmt"
	"shared-lib/models"
	"strconv"
	"time"

	"go.uber.org/zap"
)

func (c *Service) CheckAmplitude(rule models.CheckRule, file models.FileInfo, metrics []map[string]interface{}) bool {
	if len(metrics) < 2 {
		return false // 需要至少兩筆數據來比較
	}

	// **查詢 `alert_states`**
	state, err := c.db.GetAlertState(rule.RuleID, file.Host, rule.MetricName)
	if err != nil {
		c.logger.Error("查詢 Alert 狀態失敗", zap.Error(err))
		return false
	}

	// **檢查最近 N 筆數據**
	const minRequiredCount = 5
	var maxValue, minValue float64

	for _, data := range metrics[max(0, len(metrics)-minRequiredCount):] {
		var floatValue float64
		switch v := data["value"].(type) {
		case string:
			var err error
			floatValue, err = strconv.ParseFloat(v, 64)
			if err != nil {
				c.logger.Error("轉換指標值失敗", zap.Error(err))
				continue
			}
		case float64:
			floatValue = v
		default:
			c.logger.Error("無法解析指標值")
			continue
		}

		// 計算最大值與最小值
		if floatValue > maxValue {
			maxValue = floatValue
		}
		if floatValue < minValue || minValue == 0 {
			minValue = floatValue
		}
	}

	// 計算變化幅度
	amplitude := ((maxValue - minValue) / minValue) * 100

	// **判斷是否超過閾值**
	if CheckThreshold(rule, rule.Operator, amplitude) {
		fmt.Println("變化幅度超過閾值，觸發告警！")

		// **更新 Severity 狀態**
		rule.Severity = GetSeverity(rule, amplitude)

		// **更新 alert_states **
		state.StartTime = time.Now().Unix()
		state.LastTime = state.StartTime
		state.Amplitude = amplitude

		// **存入 alert_states **
		c.db.SaveAlertState(state)
		return true
	}

	state.StartTime = 0
	state.StackDuration = 0
	c.db.SaveAlertState(state)
	return false
}

// func (c *Service) CheckAmplitude(rule models.CheckRule, file models.FileInfo, metrics []map[string]interface{}) bool {
// 	if len(metrics) < 2 { // 振幅檢查需要至少兩個點
// 		return false
// 	}

// 	// 取最新的兩筆數據
// 	latest := metrics[len(metrics)-1]
// 	previous := metrics[len(metrics)-2]

// 	latestValue, ok := latest["value"].(string)
// 	if !ok {
// 		c.logger.Error("無法解析最新指標值")
// 		return false
// 	}
// 	latestTimestampStr, ok := latest["timestamp"].(string)
// 	if !ok {
// 		c.logger.Error("無法解析最新時間戳")
// 		return false
// 	}
// 	latestTimestamp, err := strconv.ParseInt(latestTimestampStr, 10, 64)
// 	if err != nil {
// 		c.logger.Error("轉換時間戳失敗", zap.Error(err))
// 		return false
// 	}

// 	previousValue, ok := previous["value"].(string)
// 	if !ok {
// 		c.logger.Error("無法解析前一個指標值")
// 		return false
// 	}

// 	// 轉換為 float64
// 	latestFloat, err := strconv.ParseFloat(latestValue, 64)
// 	if err != nil {
// 		c.logger.Error("轉換最新指標值失敗", zap.Error(err))
// 		return false
// 	}
// 	previousFloat, err := strconv.ParseFloat(previousValue, 64)
// 	if err != nil {
// 		c.logger.Error("轉換前一個指標值失敗", zap.Error(err))
// 		return false
// 	}

// 	// 計算振幅
// 	amplitude := math.Abs(latestFloat - previousFloat)

// 	// 使用狀態管理器處理
// 	_, exceeded, err := c.stateManager.GetAndUpdateState(rule, amplitude, latestTimestamp)
// 	if err != nil {
// 		c.logger.Error("檢查告警狀態失敗", zap.Error(err))
// 		return false
// 	}

// 	return exceeded
// }
