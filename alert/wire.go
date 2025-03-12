//go:build wireinject
// +build wireinject

package alert

import (
	"github.com/detect-viz/shared-lib/auth/keycloak"
	"github.com/detect-viz/shared-lib/contacts"
	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/infra/scheduler"
	"github.com/detect-viz/shared-lib/labels"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/notifier"
	"github.com/detect-viz/shared-lib/rules"

	"github.com/detect-viz/shared-lib/storage/mysql"
	"github.com/detect-viz/shared-lib/templates"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// 提供 zap.Logger
func ProvideZapLogger(log logger.Logger) *zap.Logger {
	if log == nil {
		panic("❌ log 是 nil")
	}
	return log.GetLogger()
}

// 提供 GlobalConfig
func ProvideGlobalConfig(global models.GlobalConfig) *models.GlobalConfig {
	return &global
}

// 提供 notifier.Service
func ProvideNotifierService(keycloakClient *keycloak.Client) notifier.Service {
	return notifier.NewService(keycloakClient)
}

// 提供 scheduler.Service
func ProvideSchedulerService(log logger.Logger) scheduler.Service {
	return scheduler.NewService(log)
}

// 提供 templates.Service
func ProvideTemplateService(log logger.Logger) templates.Service {
	return templates.NewService(log)
}

// 提供 labels.Service
func ProvideLabelService(mysqlClient *mysql.Client) labels.Service {
	return labels.NewService(mysqlClient)
}

// 提供 contacts.Service
func ProvideContactService(mysqlClient *mysql.Client, log logger.Logger, notifierService notifier.Service, keycloakClient *keycloak.Client) contacts.Service {
	return contacts.NewService(mysqlClient, log, notifierService, keycloakClient)
}

// AlertSet 提供所有依賴的 wire Set
var AlertSet = wire.NewSet(
	// 提供 zap.Logger
	ProvideZapLogger,
	// 提供 GlobalConfig
	ProvideGlobalConfig,
	// 提供 notifier.Service
	ProvideNotifierService,
	// Alert 服務
	ProvideAlertService,
	// 各模組的 wire set
	rules.RuleSet,
	scheduler.SchedulerSet,
	templates.TemplateSet,
	contacts.ContactSet,
)

// InitializeAlertService 初始化 AlertService
func InitializeAlertService(
	config models.AlertConfig,
	global models.GlobalConfig,
	mysqlClient *mysql.Client,
	log logger.Logger,
	keycloakClient *keycloak.Client,
) (*Service, error) {
	wire.Build(AlertSet)
	return &Service{}, nil
}

// ProvideAlertService 提供 AlertService 實例
func ProvideAlertService(
	config models.AlertConfig,
	global models.GlobalConfig,
	mysqlClient *mysql.Client,
	logSvc logger.Logger,
	rule rules.Service,
	notify notifier.Service,
	contact contacts.Service,
	scheduler scheduler.Service,
	template templates.Service,
) *Service {
	logSvc.Debug("檢查 ProvideAlertService 參數", zap.Any("config", config), zap.Any("mysqlClient", mysqlClient))

	if mysqlClient == nil {
		panic("❌ mysqlClient 是 nil")
	}
	if logSvc == nil {
		panic("❌ logSvc 是 nil")
	}
	if notify == nil {
		panic("❌ notify 是 nil")
	}
	if contact == nil {
		panic("❌ contact 是 nil")
	}
	if scheduler == nil {
		panic("❌ scheduler 是 nil")
	}
	if template == nil {
		panic("❌ template 是 nil")
	}

	return NewService(
		config,
		global,
		mysqlClient,
		logSvc,
		rule,
		notify,
		contact,
		scheduler,
		template,
	)
}
