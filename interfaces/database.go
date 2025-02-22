package interfaces

import (
	"github.com/detect-viz/shared-lib/models"
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
	GetLabels(ruleID int64) (map[string]string, error)
	GetAlertState(ruleDetailID int64) (models.AlertState, error)
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

	// TriggerLog 相關
	GetActiveTriggerLog(ruleID int64, resourceName, metricName string) (*models.TriggerLog, error)
	CreateTriggerLog(trigger models.TriggerLog) error
	UpdateTriggerLog(trigger models.TriggerLog) error
	UpdateTriggerLogNotifyState(uuid string, state string) error
	UpdateTriggerLogResolvedNotifyState(uuid string, state string) error
	GetTriggerLogsForAlertNotify(timestamp int64) ([]models.TriggerLog, error)
	GetTriggerLogsForResolvedNotify(timestamp int64) ([]models.TriggerLog, error)

	// NotifyLog 相關
	CreateNotifyLog(notify models.NotifyLog) error
	UpdateNotifyLog(notify models.NotifyLog) error

	// CheckTriggerLogExists 相關
	CheckTriggerLogExists(ruleID int64, resourceName string, metric string, firstTriggerTime int64) (bool, error)
	UpdateTriggerLogResolved(ruleID int64, resourceName, metricName string, resolvedTime int64) error

	// Template 相關
	GetAlertTemplate(realm string, ruleState string, format string) (models.AlertTemplate, error)

	// Close 相關
	Close() error
}
