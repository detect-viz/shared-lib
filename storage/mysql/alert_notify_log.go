package mysql

import (
	"fmt"

	"github.com/detect-viz/shared-lib/models"
)

// 更新通知日誌
func (c *Client) UpdateNotifyLog(notify models.NotifyLog) error {
	return c.db.Save(&notify).Error
}

// 寫入通知日誌
func (c *Client) CreateNotifyLog(notify models.NotifyLog) error {
	notify.ID = GenerateUUID16()

	// 開啟交易
	tx := c.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 1. 寫入通知日誌
	if err := tx.Create(&notify).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交交易
	return tx.Commit().Error
}

// GetFailedNotifyLogs 獲取所有失敗的通知記錄
func (c *Client) GetFailedNotifyLogs() ([]models.NotifyLog, error) {
	var logs []models.NotifyLog
	err := c.db.
		Where("state = ? AND retry_counter < ?", "failed", 3).
		Find(&logs).Error
	if err != nil {
		return nil, fmt.Errorf("查詢失敗的通知記錄失敗: %w", err)
	}
	return logs, nil
}

// GetTriggeredLog 根據 ID 獲取觸發日誌
func (c *Client) GetTriggeredLog(id []byte) (*models.TriggeredLog, error) {
	var log models.TriggeredLog
	err := c.db.Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, fmt.Errorf("查詢觸發日誌失敗: %w", err)
	}
	return &log, nil
}

// GetRuleState 獲取規則狀態
func (c *Client) GetRuleState(ruleID []byte) (*models.RuleState, error) {
	var state models.RuleState
	err := c.db.Where("rule_id = ?", ruleID).First(&state).Error
	if err != nil {
		return nil, fmt.Errorf("查詢規則狀態失敗: %w", err)
	}
	return &state, nil
}
