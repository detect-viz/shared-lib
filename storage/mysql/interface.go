package mysql

import (
	"github.com/detect-viz/shared-lib/models"
)

// 資料庫介面
type Database interface {
	// 載入告警遷移
	LoadAlertMigrate(path string) error

	GetMetricRule(id int64) (models.MetricRule, error)
	GetAlertRuleDetails(ruleID int64) ([]models.AlertRuleDetail, error)

	GetAlertState(ruleDetailID int64) (models.AlertState, error)
	SaveAlertState(state models.AlertState) error

	// 聯絡人相關

	CreateOrUpdateAlertContact(contact *models.Contact) error
	CreateOrUpdateAlertRule(rule *models.Rule) error

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
	GetTemplate(realm string, ruleState string, format string) (models.Template, error)

	// 抑制規則相關
	CreateMute(mute *models.Mute) error
	ListMutes(realm string) ([]models.Mute, error)
	GetMute(id int64) (*models.Mute, error)
	UpdateMute(mute *models.Mute) error
	DeleteMute(id int64) error

	// 告警規則相關
	CreateRule(rule *models.Rule) error
	ListRules(realm string) ([]models.Rule, error)
	GetRule(id int64) (*models.Rule, error)
	UpdateRule(rule *models.Rule) error
	DeleteRule(id int64) error

	// 告警規則相關
	CreateLabel(label *models.Label) error
	GetLabel(id int64) (*models.Label, error)
	ListLabels(realm string, limit, offset int) ([]models.Label, error)
	UpdateLabel(label *models.Label) error
	DeleteLabel(id int64) error
	UpdateKey(realm, oldKey, newKey string) error
	ExistsLabel(realm, key string) (bool, error)
	BulkCreateOrUpdateLabel(realm string, labels []models.Label) ([]models.Label, error)
	GetLabelByRuleID(ruleID int64) ([]models.Label, error)
	// 聯絡人相關
	CreateContact(contact *models.Contact) error
	ListContacts(realm string) ([]models.Contact, error)
	GetContact(id int64) (*models.Contact, error)
	UpdateContact(contact *models.Contact) error
	DeleteContact(id int64) error
	GetContactsByRuleID(ruleID int64) ([]models.Contact, error)

	// 資源群組相關
	CreateResourceGroup(resourceGroup *models.ResourceGroup) error
	ListResourceGroups(realm string) ([]models.ResourceGroup, error)
	GetResourceGroup(id int64) (*models.ResourceGroup, error)
	UpdateResourceGroup(resourceGroup *models.ResourceGroup) error
	DeleteResourceGroup(id int64) error

	// 資源相關
	CreateResource(resource *models.Resource) error
	ListResources(realm string) ([]models.Resource, error)
	GetResource(id int64) (*models.Resource, error)
	UpdateResource(resource *models.Resource) error
	DeleteResource(id int64) error

	// Close 相關
	Close() error
}
