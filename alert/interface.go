package alert

import (
	"github.com/detect-viz/shared-lib/contacts"
	"github.com/detect-viz/shared-lib/infra/scheduler"
	"github.com/detect-viz/shared-lib/labels"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/mutes"
	"github.com/detect-viz/shared-lib/notifier"
	"github.com/detect-viz/shared-lib/rules"
	"github.com/detect-viz/shared-lib/templates"
	"go.uber.org/zap"
)

// 告警服務介面
type AlertService interface {
	GetRuleService() rules.Service
	GetMuteService() mutes.Service
	GetNotifyService() notifier.Service
	GetSchedulerService() scheduler.Service
	GetTemplateService() templates.Service
	GetContactService() contacts.Service
	GetLabelService() labels.Service

	// 監控頁面
	ListAlertState(realm string) ([]models.AlertState, error)
	ListAlertHistory(realm string) ([]models.TriggerLog, error)

	// 檢查操作
	CheckName(name string) bool
	GetLogger() *zap.Logger
	// 指標規則管理
	GetMetricRule(id int64) (models.MetricRule, bool)
}

// 告警檢查器介面
type AlertChecker interface {
	Check(rule models.Rule, file models.FileInfo, metrics map[string]interface{}) bool
}
