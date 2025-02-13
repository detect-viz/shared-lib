package databases

import (
	"fmt"

	"shared-lib/models"

	"gorm.io/gorm"
)

type MySQL struct {
	db *gorm.DB
}

func NewMySQL(db *gorm.DB) *MySQL {
	return &MySQL{db: db}
}

// GetMetricRule 獲取指標規則定義
func (m *MySQL) GetMetricRule(id int64) (models.MetricRule, error) {
	var rule models.MetricRule
	err := m.db.First(&rule, id).Error
	return rule, err
}

// GetAlertRuleDetails 獲取告警規則詳情
func (m *MySQL) GetAlertRuleDetails(ruleID int64) ([]models.AlertRuleDetail, error) {
	var details []models.AlertRuleDetail
	err := m.db.Where("alert_rule_id = ?", ruleID).Find(&details).Error
	return details, err
}

// GetCustomLabels 獲取規則相關的標籤
func (m *MySQL) GetCustomLabels(ruleID int64) (map[string]string, error) {
	var labels []models.Label
	err := m.db.Where("resource_id = ? AND type = ?", ruleID, "rule").Find(&labels).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, label := range labels {
		result[label.Key] = label.Value
	}
	return result, nil
}

// GetAlertState 獲取告警狀態
func (m *MySQL) GetAlertState(ruleID int64, resourceName string, metricName string) (models.AlertState, error) {
	var state models.AlertState
	err := m.db.Where(
		"rule_id = ? AND resource_name = ? AND metric_name = ?",
		ruleID, resourceName, metricName,
	).First(&state).Error

	if err == gorm.ErrRecordNotFound {
		// 如果不存在，返回新的狀態
		return models.AlertState{
			RuleID:       ruleID,
			ResourceName: resourceName,
			MetricName:   metricName,
		}, nil
	}
	return state, err
}

// SaveAlertState 保存告警狀態
func (m *MySQL) SaveAlertState(state models.AlertState) error {
	// 使用 Upsert 操作
	return m.db.Save(&state).Error
}

// GetAlertContacts 獲取規則的聯絡人列表
func (m *MySQL) GetAlertContacts(ruleID string) ([]models.AlertContact, error) {
	var contacts []models.AlertContact

	// 1. 查詢規則關聯的聯絡人
	err := m.db.Table("alert_rule_contacts").
		Select("alert_contacts.*").
		Joins("LEFT JOIN alert_contacts ON alert_rule_contacts.contact_id = alert_contacts.id").
		Where("alert_rule_contacts.rule_id = ? AND alert_contacts.status = ?", ruleID, "enabled").
		Find(&contacts).Error

	if err != nil {
		return nil, fmt.Errorf("查詢聯絡人失敗: %v", err)
	}

	return contacts, nil
}

// WriteTriggeredLog 寫入觸發日誌
func (m *MySQL) WriteTriggeredLog(trigger models.TriggerLog) error {
	return m.db.Create(&trigger).Error
}

// WriteNotificationLog 寫入通知日誌
func (m *MySQL) WriteNotificationLog(notification models.NotificationLog) error {
	// 開啟交易
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 1. 寫入通知日誌
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 寫入關聯表
	for _, trigger := range notification.TriggerLogs {
		relation := models.NotificationTriggerLog{
			NotificationID: notification.ID,
			TriggerLogID:   int64(trigger.ID),
		}
		if err := tx.Create(&relation).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交交易
	return tx.Commit().Error
}
