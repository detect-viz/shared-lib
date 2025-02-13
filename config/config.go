package config

import (
	"fmt"
	"sync"

	"shared-lib/models"

	"github.com/spf13/viper"
)

// Manager 配置管理器
type Manager struct {
	sync.RWMutex
	viper  *viper.Viper
	config *models.Config
}

// New 創建配置管理器
func New() *Manager {
	return &Manager{
		viper: viper.New(),
	}
}

// Load 加載配置
func (m *Manager) Load(configPath string) error {
	m.Lock()
	defer m.Unlock()

	m.viper.SetConfigFile(configPath)
	if err := m.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("讀取配置文件失敗: %w", err)
	}

	return m.parseConfig()
}

// Reload 重新加載配置
func (m *Manager) Reload() error {
	m.Lock()
	defer m.Unlock()

	if err := m.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("重新加載配置失敗: %w", err)
	}

	return m.parseConfig()
}

// parseConfig 解析配置
func (m *Manager) parseConfig() error {
	var conf models.Config
	if err := m.viper.Unmarshal(&conf); err != nil {
		return fmt.Errorf("解析配置失敗: %w", err)
	}
	m.config = &conf
	return nil
}

// GetConfig 獲取完整配置
func (m *Manager) GetConfig() *models.Config {
	m.RLock()
	defer m.RUnlock()
	return m.config
}

// GetLoggerConfig 獲取日誌配置
func (m *Manager) GetLoggerConfig() models.LoggerConfig {
	m.RLock()
	defer m.RUnlock()
	return m.config.Logger
}

// GetNotifyConfig 獲取通知配置
func (m *Manager) GetNotifyConfig() models.NotifyConfig {
	m.RLock()
	defer m.RUnlock()
	return m.config.Notify
}

// GetAlertConfig 獲取告警配置
func (m *Manager) GetAlertConfig() models.AlertConfig {
	m.RLock()
	defer m.RUnlock()
	return m.config.Alert
}

// GetDatabaseConfig 獲取資料庫配置
func (m *Manager) GetDatabaseConfig() models.DatabaseConfig {
	m.RLock()
	defer m.RUnlock()
	return m.config.Database
}

// GetMetricSpecConfig 獲取指標配置
func (m *Manager) GetMetricSpecConfig() models.MetricSpecConfig {
	m.RLock()
	defer m.RUnlock()
	return m.config.Metric
}

// GetRawConfig 獲取原始 viper 實例
func (m *Manager) GetRawConfig() *viper.Viper {
	m.RLock()
	defer m.RUnlock()
	return m.viper
}
