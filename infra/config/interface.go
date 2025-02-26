package config

import (
	"github.com/detect-viz/shared-lib/models"

	"github.com/spf13/viper"
)

// Config 配置管理介面
type Config interface {

	// 獲取配置
	GetConfig() *models.Config

	// 模組配置訪問
	GetLoggerConfig() models.LoggerConfig
	GetAlertConfig() models.AlertConfig
	GetDatabaseConfig() models.DatabaseConfig
	GetKeycloakConfig() models.KeycloakConfig
	GetGlobalConfig() models.GlobalConfig
	// 原始配置訪問
	GetRawConfig() *viper.Viper
}
