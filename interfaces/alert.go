package interfaces

import (
	"github.com/detect-viz/shared-lib/models"
)

// 告警服務介面
type AlertService interface {
	// 告警檢查
	ProcessFile(file models.FileInfo) error
	ProcessTriggers() error
	// 監控頁面
	GetHistoryAlert(user models.SSOUser) ([]models.HistoryAlert, error)
	GetCurrentAlert() ([]models.CurrentAlert, error)
	GetHistoryAlertMetric(user models.SSOUser, body models.HistoryAlert) ([]models.MetricResponse, error)
	// 告警規則管理基本 CRUD
	Create(contact *models.AlertRule) error
	Get(id int64) (*models.AlertRule, error)
	List(realm string) ([]models.AlertRule, error)
	Update(contact *models.AlertRule) error
	Delete(id int64) error

	// 檢查操作
	CheckName(contact models.AlertRule) (bool, string)

	// 指標規則管理
	GetMetricRule(id int64) (models.MetricRule, bool)
}

// 告警檢查器介面
type AlertChecker interface {
	Check(rule models.AlertRule, file models.FileInfo, metrics map[string]interface{}) bool
}
