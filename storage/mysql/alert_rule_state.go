package mysql

import (
	"fmt"
	"reflect"

	"github.com/detect-viz/shared-lib/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// * 鎖定確保一致性，防止並發更新造成數據不一致
func (c *Client) GetRuleStateAndLock(ruleID []byte) (*models.RuleState, error) {
	var state models.RuleState
	err := c.db.Raw("SELECT * FROM rule_states WHERE rule_id = ? FOR UPDATE", ruleID).Scan(&state).Error

	if err == gorm.ErrRecordNotFound {
		c.logger.Warn("找不到告警狀態", zap.String("rule_id", string(ruleID)))
		return nil, nil
	}
	return &state, nil
}

// * 如果 值為 nil 或 不變動，則不會更新該欄位。
func (c *Client) UpdateRuleStateWithUpdates(oldState, newState models.RuleState) error {

	// **Step 1: 比對 oldState & newState，若無變更則跳過更新
	if reflect.DeepEqual(oldState, newState) {
		c.logger.Warn("告警狀態無變更", zap.String("rule_id", string(oldState.RuleID)))
		return nil
	}

	// **Step 2: 更新 DB
	if err := c.db.Model(&models.RuleState{}).
		Where("rule_id = ?", oldState.RuleID).
		Updates(newState).Error; err != nil {
		return fmt.Errorf("更新 RuleState 失敗: %w", err)
	}

	return nil
}
