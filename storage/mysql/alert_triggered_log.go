package mysql

import (
	"fmt"
	"time"

	"github.com/detect-viz/shared-lib/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 檢查 TriggeredLog 是否已存在
func (c *Client) CheckTriggeredLogExists(ruleID string, resourceName string, metricName string, firstTriggeredTime int64) (bool, error) {
	var count int64
	err := c.db.Model(&models.TriggeredLog{}).
		Where("rule_id = ? AND resource_name = ? AND metric_name = ? AND first_triggered_time = ?",
			ruleID, resourceName, metricName, firstTriggeredTime).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 寫入觸發日誌
func (c *Client) CreateTriggeredLog(triggered models.TriggeredLog) error {
	c.logger.Debug("寫入觸發日誌", zap.Any("triggered", triggered))
	return c.db.Create(&triggered).Error
}

// 更新觸發日誌
func (c *Client) UpdateTriggeredLog(triggered models.TriggeredLog) error {
	return c.db.Save(&triggered).Error
}

func (c *Client) UpdateTriggeredLogResolved(ruleID string, resourceName, metricName string, resolvedTime int64) error {
	// 查詢 TriggeredLog 是否存在
	var triggered models.TriggeredLog
	err := c.db.
		Where("rule_id = ? AND resource_name = ? AND metric_name = ?", ruleID, resourceName, metricName).
		First(&triggered).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// **找不到 TriggeredLog，可能已經 resolved，無需處理**
			return nil
		}
		return fmt.Errorf("查詢 TriggeredLog 失敗: %w", err)
	}

	// **更新 TriggeredLog 狀態**

	triggered.ResolvedAt = &resolvedTime

	// **寫入資料庫**
	err = c.db.Save(&triggered).Error
	if err != nil {
		return fmt.Errorf("更新 TriggeredLog 為 resolved 失敗: %w", err)
	}

	return nil
}

// 獲取活動的觸發日誌
func (c *Client) GetActiveTriggeredLog(ruleID []byte, resourceName, metricName string) (*models.TriggeredLog, error) {
	var triggered models.TriggeredLog
	err := c.db.
		Where(`rule_id = ? AND resource_name = ? AND metric_name = ? 
			   AND resolved_time IS NULL 
			   AND (notify_state IS NULL OR notify_state = 'failed')`,
			ruleID, resourceName, metricName).
		First(&triggered).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查詢 TriggeredLog 失敗: %w", err)
	}
	return &triggered, nil
}

// 獲取需要發送告警通知的觸發日誌
func (c *Client) GetTriggeredLogsForAlertNotify(timestamp int64) ([]models.TriggeredLog, error) {
	var triggereds []models.TriggeredLog
	err := c.db.
		Where(`timestamp < ? 
			   AND contact_state != 'mute' 
			   AND (silence_end IS NULL OR silence_end < ?) 
			   AND (notify_state IS NULL OR notify_state = 'failed')
			   AND resolved_time IS NULL`,
			timestamp, time.Now().Unix()).
		Find(&triggereds).Error

	if err != nil {
		return nil, fmt.Errorf("查詢待通知的 TriggeredLog 失敗: %w", err)
	}
	return triggereds, nil
}

// 需要發送恢復通知的觸發日誌
func (c *Client) GetTriggeredLogsForResolvedNotify(timestamp int64) ([]models.TriggeredLog, error) {
	var triggereds []models.TriggeredLog

	err := c.db.
		Where("timestamp < ? AND contact_state != ? AND (silence_end IS NULL OR silence_end < ?) AND resolved_time != 0 AND resolved_notify_state IS NULL",
			timestamp, "mute", time.Now().Unix()).
		Find(&triggereds).Error

	if err != nil {
		return nil, fmt.Errorf("查詢待發送恢復通知的 `TriggeredLog` 失敗: %w", err)
	}

	return triggereds, nil
}

// 當異常發生後，發送 TriggeredLog 通知後，應該更新 notify_state
func (c *Client) UpdateTriggeredLogNotifyState(triggeredID []byte, notifyState string) error {
	return c.db.
		Model(&models.TriggeredLog{}).
		Where("id = ?", triggeredID).
		Update("notify_state", notifyState).
		Error
}

// 當異常恢復 (resolved)，發送恢復通知後，應該更新 resolved_notify_state
func (c *Client) UpdateTriggeredLogResolvedNotifyState(triggeredID []byte, notifyState string) error {
	return c.db.
		Model(&models.TriggeredLog{}).
		Where("triggered_id = ?", triggeredID).
		Update("resolved_notify_state", notifyState).
		Error
}
