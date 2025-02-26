//go:build wireinject

package alert

import (
	"github.com/detect-viz/shared-lib/auth/keycloak"
	"github.com/detect-viz/shared-lib/contacts"
	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/infra/scheduler"
	"github.com/detect-viz/shared-lib/labels"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/mutes"
	"github.com/detect-viz/shared-lib/notifier"
	"github.com/detect-viz/shared-lib/rules"

	"github.com/detect-viz/shared-lib/storage/mysql"
	"github.com/detect-viz/shared-lib/templates"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// ProvideZapLogger 提供 zap.Logger
func ProvideZapLogger(log logger.Logger) *zap.Logger {
	return log.GetLogger()
}

// AlertSet 提供所有依賴的 wire Set
var AlertSet = wire.NewSet(
	// 提供 zap.Logger
	ProvideZapLogger,

	// 各模組的 wire set
	rules.RuleSet,
	notifier.NotifySet,
	scheduler.SchedulerSet,
	mutes.MuteSet,
	templates.TemplateSet,
	contacts.ContactSet,
	labels.LabelSet,
	// Alert 服務
	ProvideAlertService,
)

// InitializeAlertService 初始化 AlertService
func InitializeAlertService(
	config models.AlertConfig,
	global models.GlobalConfig,
	mysqlClient *mysql.Client,
	log logger.Logger,
	keycloak *keycloak.Client,
) (*Service, error) {
	wire.Build(AlertSet)
	return nil, nil
}

// ProvideAlertService 提供 AlertService 實例
func ProvideAlertService(
	config models.AlertConfig,
	global models.GlobalConfig,
	logSvc logger.Logger,
	mysqlClient *mysql.Client,

	rule rules.Service,
	mute mutes.Service,
	keycloak *keycloak.Client,

	notify notifier.Service,
	contact contacts.Service,
	scheduler *scheduler.Service,
	template templates.Service,

	label labels.Service,
) *Service {
	return NewService(
		config,
		global,
		mysqlClient,
		logSvc,

		rule,
		mute,
		keycloak,

		notify,
		contact,
		scheduler,
		template,

		label,
	)
}
