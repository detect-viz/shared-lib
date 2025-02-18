package alert

import (
	"fmt"
	"shared-lib/models"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// checkAbsolute 絕對值檢查
// func (c *Service) CheckAbsolute(rule models.CheckRule, file models.FileInfo, metrics []map[string]interface{}) bool {
// 	if len(metrics) == 0 {
// 		return false
// 	}

// 	// 取最新的一筆數據
// 	latest := metrics[len(metrics)-1]
// 	value, ok := latest["value"].(string)
// 	if !ok {
// 		c.logger.Error("無法解析指標值")
// 		return false
// 	}
// 	timestampStr, ok := latest["timestamp"].(string)
// 	if !ok {
// 		c.logger.Error("無法解析時間戳")
// 		return false
// 	}
// 	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
// 	if err != nil {
// 		c.logger.Error("轉換時間戳失敗", zap.Error(err))
// 		return false
// 	}
// 	// 轉換為 float64
// 	floatValue, err := strconv.ParseFloat(value, 64)
// 	if err != nil {
// 		c.logger.Error("轉換指標值失敗", zap.Error(err))
// 		return false
// 	}

// 	// 使用狀態管理器處理
// 	_, exceeded, err := c.stateManager.GetAndUpdateState(rule, floatValue, timestamp)
// 	if err != nil {
// 		c.logger.Error("檢查告警狀態失敗", zap.Error(err))
// 		return false
// 	}

// 	return exceeded
// }

func (c *Service) CheckAbsolute(rule models.CheckRule, file models.FileInfo, metrics []map[string]interface{}) bool {
	if len(metrics) == 0 {
		return false
	}

	// **查詢 `alert_states`**
	state, err := c.db.GetAlertState(rule.RuleID, file.Host, rule.MetricName)
	if err != nil {
		c.logger.Error("查詢 Alert 狀態失敗", zap.Error(err))
		return false
	}

	// **初始化異常累積時間**
	if state.StartTime == 0 {
		state.StackDuration = 0
	}

	// **檢查最近 N 筆數據**
	const minRequiredCount = 5
	count := 0
	var floatValue float64
	for _, data := range metrics[max(0, len(metrics)-minRequiredCount):] {

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

		// **判斷是否超過閾值**
		if CheckThreshold(rule, rule.Operator, floatValue) {
			count++
		} else {
			count = 0 // 若中間有數據未超過閾值，則重置
		}
	}

	// **觸發告警條件**
	if count >= minRequiredCount {
		fmt.Println("🎯 連續超過閾值，觸發告警！")

		// **更新 Severity 狀態**
		rule.Severity = GetSeverity(rule, floatValue)

		// **更新 Alert 狀態**
		if state.StartTime == 0 {
			state.StartTime = time.Now().Unix()
		}
		state.LastTime = time.Now().Unix()
		state.StackDuration += int(time.Now().Unix() - state.LastTime)
		state.LastValue = floatValue

		// **存入 `alert_states`**
		c.db.SaveAlertState(state)
		return true
	}

	// **如果恢復正常，重置 `alert_states`**
	state.StartTime = 0
	state.StackDuration = 0
	c.db.SaveAlertState(state)
	return false
}
