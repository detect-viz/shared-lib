package mysql

import (
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/label"
)

// 資料庫介面
type Database interface {

	// 告警檢查規則相關
	GetActiveRules(realm, resourceName string) ([]models.Rule, error)
	GetAllActiveRules() ([]models.Rule, error)

	// 告警規則設定相關
	CreateRule(rule *models.Rule) (*models.Rule, error)
	CreateRules(rules []models.Rule) error
	ListRules(realm string, cursor int64, limit int) ([]models.Rule, int64, error)
	GetRule(id []byte) (*models.Rule, error)
	UpdateRule(rule *models.Rule) (*models.Rule, error)
	DeleteRule(id []byte) error

	// 獲取自動匹配的規則
	GetAutoApplyRulesByMetricRuleUIDs(realm string, metricRuleUIDs []string) ([]models.Rule, error)

	// 聯絡人相關
	CreateContact(contact *models.Contact) (*models.Contact, error)
	ListContacts(realm string, cursor int64, limit int) ([]models.Contact, int64, error)
	GetContact(id []byte) (*models.Contact, error)
	UpdateContact(contact *models.Contact) (*models.Contact, error)
	DeleteContact(id []byte) error
	GetContactsByRuleID(ruleID []byte) ([]models.Contact, error)
	IsUsedByRules(contactID []byte) (bool, error)

	// 載入告警遷移
	LoadAlertMigrate(path string) error

	// 告警狀態相關
	GetRuleStateAndLock(ruleID []byte) (*models.RuleState, error)
	UpdateRuleStateWithUpdates(oldState, newState models.RuleState) error

	// 資源群組相關
	GetResourceGroupName(id []byte) (string, error)

	// 監控對象相關
	CheckTargetExists(realm, dataSource, resourceName, partitionName string) (bool, error)
	CreateTarget(target *models.Target) (*models.Target, error)

	// TriggeredLog 相關
	GetActiveTriggeredLog(ruleID []byte, resourceName, metricName string) (*models.TriggeredLog, error)
	CreateTriggeredLog(triggered models.TriggeredLog) error
	UpdateTriggeredLog(triggered models.TriggeredLog) error
	UpdateTriggeredLogNotifyState(id []byte, state string) error
	UpdateTriggeredLogResolvedNotifyState(id []byte, state string) error
	GetTriggeredLogsForAlertNotify(timestamp int64) ([]models.TriggeredLog, error)
	GetTriggeredLogsForResolvedNotify(timestamp int64) ([]models.TriggeredLog, error)

	// NotifyLog 相關
	CreateNotifyLog(notify models.NotifyLog) error
	UpdateNotifyLog(notify models.NotifyLog) error

	// CheckTriggeredLogExists 相關
	CheckTriggeredLogExists(ruleID []byte, resourceName string, metricName string, firstTriggeredTime int64) (bool, error)
	UpdateTriggeredLogResolved(ruleID []byte, resourceName, metricName string, resolvedTime int64) error

	// Template 相關
	GetTemplate(realm string, ruleState string, format string) (models.Template, error)

	// 抑制規則相關
	CreateMute(mute *models.Mute) error
	ListMutes(realm string) ([]models.Mute, error)
	GetMute(id string) (*models.Mute, error)
	UpdateMute(mute *models.Mute) error
	DeleteMute(id string) error

	// 標籤相關
	CreateLabel(label *label.LabelKey, values []string) (*label.LabelKey, error)
	GetLabel(id int64) (*label.LabelKey, error)
	ListLabels(realm string, cursor int64, limit int) ([]label.LabelKey, int64, error)
	UpdateLabel(id int64, values []string) (*label.LabelKey, error)
	DeleteLabel(id int64) error
	UpdateLabelKeyName(realm, oldKey, newKey string) (*label.LabelKey, error)
	BulkCreateOrUpdateLabel(realm string, labels []models.LabelDTO) error
	GetRuleLabelByRuleID(ruleID []byte) (map[string]string, error)

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
