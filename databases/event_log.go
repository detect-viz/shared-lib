package databases

import (
	"fmt"
	"time"

	"github.com/detect-viz/shared-lib/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 檢查 TriggerLog 是否已存在
func (m *MySQL) CheckTriggerLogExists(ruleID int64, resourceName string, metric string, firstTriggerTime int64) (bool, error) {
	var count int64
	err := m.db.Model(&models.TriggerLog{}).
		Where("rule_id = ? AND resource_name = ? AND metric_name = ? AND first_trigger_time = ?",
			ruleID, resourceName, metric, firstTriggerTime).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 寫入觸發日誌
func (m *MySQL) CreateTriggerLog(trigger models.TriggerLog) error {
	m.logger.Debug("寫入觸發日誌", zap.Any("trigger", trigger))
	return m.db.Create(&trigger).Error
}

// 更新觸發日誌
func (m *MySQL) UpdateTriggerLog(trigger models.TriggerLog) error {
	return m.db.Save(&trigger).Error
}

// 寫入通知日誌
func (m *MySQL) CreateNotifyLog(notify models.NotifyLog) error {
	// 開啟交易
	tx := m.db.Begin()
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
func (m *MySQL) GetAlertState(RuleDetailID int64) (models.AlertState, error) {
	var state models.AlertState
	err := m.db.Where(
		"rule_detail_id = ?",
		RuleDetailID,
	).First(&state).Error

	if err == gorm.ErrRecordNotFound {
		// 如果不存在，返回新的狀態
		return models.AlertState{
			RuleDetailID: RuleDetailID,
		}, nil
	}
	return state, err
}

// SaveAlertState 保存告警狀態
func (m *MySQL) SaveAlertState(state models.AlertState) error {
	// 使用 Upsert 操作
	return m.db.Save(&state).Error
}

func (m *MySQL) UpdateTriggerLogResolved(ruleID int64, resourceName, metricName string, resolvedTime int64) error {
	// 查詢 TriggerLog 是否存在
	var trigger models.TriggerLog
	err := m.db.
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
	err = m.db.Save(&trigger).Error
	if err != nil {
		return fmt.Errorf("更新 TriggerLog 為 resolved 失敗: %w", err)
	}

	return nil
}

// GetActiveTriggerLog 獲取活動的觸發日誌
func (m *MySQL) GetActiveTriggerLog(ruleID int64, resourceName, metricName string) (*models.TriggerLog, error) {
	var trigger models.TriggerLog
	err := m.db.
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
func (m *MySQL) GetTriggerLogsForAlertNotify(timestamp int64) ([]models.TriggerLog, error) {
	var triggers []models.TriggerLog
	err := m.db.
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
func (m *MySQL) GetTriggerLogsForResolvedNotify(timestamp int64) ([]models.TriggerLog, error) {
	var triggers []models.TriggerLog

	err := m.db.
		Where("timestamp < ? AND contact_state != ? AND (silence_end IS NULL OR silence_end < ?) AND resolved_time != 0 AND resolved_notify_state IS NULL",
			timestamp, "mute", time.Now().Unix()).
		Find(&triggers).Error

	if err != nil {
		return nil, fmt.Errorf("查詢待發送恢復通知的 `TriggerLog` 失敗: %w", err)
	}

	return triggers, nil
}

// 當異常發生後，發送 TriggerLog 通知後，應該更新 notify_state
func (m *MySQL) UpdateTriggerLogNotifyState(triggerUUID string, notifyState string) error {
	return m.db.
		Model(&models.TriggerLog{}).
		Where("uuid = ?", triggerUUID).
		Update("notify_state", notifyState).
		Error
}

// 當異常恢復 (resolved)，發送恢復通知後，應該更新 resolved_notify_state
func (m *MySQL) UpdateTriggerLogResolvedNotifyState(triggerUUID string, notifyState string) error {
	return m.db.
		Model(&models.TriggerLog{}).
		Where("uuid = ?", triggerUUID).
		Update("resolved_notify_state", notifyState).
		Error
}
