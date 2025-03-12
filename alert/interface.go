package alert

import (
	"github.com/detect-viz/shared-lib/contacts"
	"github.com/detect-viz/shared-lib/infra/scheduler"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/notifier"
	"github.com/detect-viz/shared-lib/rules"
	"github.com/detect-viz/shared-lib/templates"
	"go.uber.org/zap"
)

type AlertService interface {
	GetLogger() *zap.Logger
	GetRuleService() rules.Service
	GetNotifyService() notifier.Service
	GetSchedulerService() scheduler.Service
	GetTemplateService() templates.Service
	GetContactService() contacts.Service

	//* 監控頁面
	ListRuleState(realm string) ([]models.RuleState, error)
	ListAlertHistory(realm string) ([]models.TriggeredLog, error)

	//* 指標規則管理
	GetMetricRule(uid string) (*models.MetricRule, error)
	GetMetricRuleCategoryOptions() ([]models.OptionResponse, error)
	GetMetricRuleOptions(category string) ([]models.OptionResponse, error)

	//* 告警處理 (主入口)
	ProcessAlert(payload models.AlertPayload) error

	//* 批次通知 (主入口)
	ProcessNotifyLog() error
}
