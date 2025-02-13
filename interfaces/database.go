package interfaces

import (
	"shared-lib/models"
)

// 資料庫介面
type Database interface {
	// 告警規則相關
	GetMetricRule(id int64) (models.MetricRule, error)
	GetAlertRuleDetails(ruleID int64) ([]models.AlertRuleDetail, error)
	GetCustomLabels(ruleID int64) (map[string]string, error)
	GetAlertState(ruleID int64, resourceName string, metricName string) (models.AlertState, error)
	SaveAlertState(state models.AlertState) error

	// 聯絡人相關
	GetAlertContacts(ruleID int64) ([]models.AlertContact, error)
	CreateContact(contact *models.AlertContact) error
	UpdateContact(contact *models.AlertContact) error
	DeleteContact(id int64) error

	// 日誌相關
	WriteTriggeredLog(trigger models.TriggerLog) error
	WriteNotificationLog(notification models.NotificationLog) error

	GetPendingFiles() ([]models.FileInfo, error)
	SaveMetrics(fileID string, metrics map[string]interface{}) error
	UpdateFile(file models.FileInfo) error
	Close() error
}
