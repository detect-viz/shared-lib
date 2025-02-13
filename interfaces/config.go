package interfaces

import (
	"shared-lib/models"

	"github.com/spf13/viper"
)

// Config 配置管理介面
type Config interface {
	// 加載配置
	Load(configPath string) error

	// 重新加載配置
	Reload() error

	// 獲取配置
	GetConfig() *models.Config

	// 模組配置訪問
	GetLoggerConfig() models.LoggerConfig
	GetNotifyConfig() models.NotifyConfig
	GetAlertConfig() models.AlertConfig
	GetDatabaseConfig() models.DatabaseConfig
	GetMetricSpecConfig() models.MetricSpecConfig

	// 原始配置訪問
	GetRawConfig() *viper.Viper
}
