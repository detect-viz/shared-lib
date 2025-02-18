package interfaces

import (
	"shared-lib/models"
)

// 資料庫介面
type Database interface {
	// 載入告警遷移
	LoadAlertMigrate(path string) error
	// 告警規則相關
	GetAlertRules() (map[string][]models.AlertRule, error)
	GetAlertRulesByRealm(realm string) ([]models.AlertRule, error)
	GetAlertRuleByID(id int64) (models.AlertRule, error)
	CreateAlertRule(rule *models.AlertRule) error
	UpdateAlertRule(rule *models.AlertRule) error
	DeleteAlertRule(id int64) error

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
	GetContactByID(id int64) (models.AlertContact, error)
	CreateOrUpdateAlertContact(contact *models.AlertContact) error
	CreateOrUpdateAlertRule(rule *models.AlertRule) error

	// 資源群組相關
	GetResourceGroupName(id int64) (string, error)

	// 日誌相關
	WriteTriggeredLog(trigger models.TriggerLog) error
	WriteNotificationLog(notification models.NotificationLog) error

	Close() error
}
