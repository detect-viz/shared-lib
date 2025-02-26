package mysql

import (
	"fmt"
	"time"

	"github.com/detect-viz/shared-lib/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UpdateNotifyLog 更新通知日誌
func (c *Client) UpdateNotifyLog(notify models.NotifyLog) error {
	return c.db.Save(&notify).Error
}

// 檢查 TriggerLog 是否已存在
func (c *Client) CheckTriggerLogExists(ruleID int64, resourceName string, metric string, firstTriggerTime int64) (bool, error) {
	var count int64
	err := c.db.Model(&models.TriggerLog{}).
		Where("rule_id = ? AND resource_name = ? AND metric_name = ? AND first_trigger_time = ?",
			ruleID, resourceName, metric, firstTriggerTime).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 寫入觸發日誌
func (c *Client) CreateTriggerLog(trigger models.TriggerLog) error {
	c.logger.Debug("寫入觸發日誌", zap.Any("trigger", trigger))
	return c.db.Create(&trigger).Error
}

// 更新觸發日誌
func (c *Client) UpdateTriggerLog(trigger models.TriggerLog) error {
	return c.db.Save(&trigger).Error
}

// 寫入通知日誌
func (c *Client) CreateNotifyLog(notify models.NotifyLog) error {
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

	// 2. 寫入關聯表
	for _, trigger := range notify.TriggerLogs {
		relation := models.NotifyTriggerLog{
			NotifyLogUUID:  notify.UUID,
			TriggerLogUUID: trigger.UUID,
		}
		if err := tx.Create(&relation).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交交易
	return tx.Commit().Error
}

// 獲取告警狀態
func (c *Client) GetAlertState(RuleDetailID int64) (models.AlertState, error) {
	state := models.AlertState{}
	err := c.db.Where("alert_rule_detail_id = ?", RuleDetailID).First(&state).Error

	if err == gorm.ErrRecordNotFound {
		c.logger.Warn("找不到告警狀態", zap.Int64("rule_detail_id", RuleDetailID))
		return state, nil
	}
	return state, err
}

// SaveAlertState 保存告警狀態
func (c *Client) SaveAlertState(state models.AlertState) error {
	// 使用 Upsert 操作
	return c.db.Save(&state).Error
}

func (c *Client) UpdateTriggerLogResolved(ruleID int64, resourceName, metricName string, resolvedTime int64) error {
	// 查詢 TriggerLog 是否存在
	var trigger models.TriggerLog
	err := c.db.
		Where("rule_id = ? AND resource_name = ? AND metric_name = ?", ruleID, resourceName, metricName).
		First(&trigger).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// **找不到 TriggerLog，可能已經 resolved，無需處理**
			return nil
		}
		return fmt.Errorf("查詢 TriggerLog 失敗: %w", err)
	}

	// **更新 TriggerLog 狀態**

	trigger.ResolvedTime = &resolvedTime

	// **寫入資料庫**
	err = c.db.Save(&trigger).Error
	if err != nil {
		return fmt.Errorf("更新 TriggerLog 為 resolved 失敗: %w", err)
	}

	return nil
}

// GetActiveTriggerLog 獲取活動的觸發日誌
func (c *Client) GetActiveTriggerLog(ruleID int64, resourceName, metricName string) (*models.TriggerLog, error) {
	var trigger models.TriggerLog
	err := c.db.
		Where(`rule_id = ? AND resource_name = ? AND metric_name = ? 
			   AND resolved_time IS NULL 
			   AND (notify_state IS NULL OR notify_state = 'failed')`,
			ruleID, resourceName, metricName).
		First(&trigger).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查詢 TriggerLog 失敗: %w", err)
	}
	return &trigger, nil
}

// GetTriggerLogsForAlertNotify 獲取需要發送告警通知的觸發日誌
func (c *Client) GetTriggerLogsForAlertNotify(timestamp int64) ([]models.TriggerLog, error) {
	var triggers []models.TriggerLog
	err := c.db.
		Where(`timestamp < ? 
			   AND contact_state != 'mute' 
			   AND (silence_end IS NULL OR silence_end < ?) 
			   AND (notify_state IS NULL OR notify_state = 'failed')
			   AND resolved_time IS NULL`,
			timestamp, time.Now().Unix()).
		Find(&triggers).Error

	if err != nil {
		return nil, fmt.Errorf("查詢待通知的 TriggerLog 失敗: %w", err)
	}
	return triggers, nil
}

// 需要發送恢復通知的 TriggerLog
func (c *Client) GetTriggerLogsForResolvedNotify(timestamp int64) ([]models.TriggerLog, error) {
	var triggers []models.TriggerLog

	err := c.db.
		Where("timestamp < ? AND contact_state != ? AND (silence_end IS NULL OR silence_end < ?) AND resolved_time != 0 AND resolved_notify_state IS NULL",
			timestamp, "mute", time.Now().Unix()).
		Find(&triggers).Error

	if err != nil {
		return nil, fmt.Errorf("查詢待發送恢復通知的 `TriggerLog` 失敗: %w", err)
	}

	return triggers, nil
}

// 當異常發生後，發送 TriggerLog 通知後，應該更新 notify_state
func (c *Client) UpdateTriggerLogNotifyState(triggerUUID string, notifyState string) error {
	return c.db.
		Model(&models.TriggerLog{}).
		Where("uuid = ?", triggerUUID).
		Update("notify_state", notifyState).
		Error
}

// 當異常恢復 (resolved)，發送恢復通知後，應該更新 resolved_notify_state
func (c *Client) UpdateTriggerLogResolvedNotifyState(triggerUUID string, notifyState string) error {
	return c.db.
		Model(&models.TriggerLog{}).
		Where("uuid = ?", triggerUUID).
		Update("resolved_notify_state", notifyState).
		Error
}
