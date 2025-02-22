package config

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/detect-viz/shared-lib/models"

	"github.com/spf13/viper"
)

// Manager 配置管理器
type Manager struct {
	sync.RWMutex
	viper   *viper.Viper
	config  *models.Config
	mapping *models.MappingConfig
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
func (m *Manager) loadOtherConfigs(configFiles []string) error {
	var mapping models.MappingConfig
	// 讀取所有配置文件
	if err := m.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("沒有發現配置文件，改抓取環境變數")
			m.viper.AutomaticEnv()
			m.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		} else {
			return err
		}
	}
	for _, name := range configFiles {
		m.viper.SetConfigName(name)
		m.viper.SetConfigType("yaml")
		m.viper.AddConfigPath("conf") // 設置配置文件路徑
		m.viper.AddConfigPath("/etc/bimap-ipoc")
		if err := m.viper.MergeInConfig(); err != nil {
			return fmt.Errorf("合併配置文件失敗 %s: %w", name, err)
		}
	}

	if err := m.viper.Unmarshal(&mapping); err != nil {
		return fmt.Errorf("解析配置失敗: %w", err)
	}
	m.mapping = &mapping
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

// GetRawConfig 獲取原始 viper 實例
func (m *Manager) GetRawConfig() *viper.Viper {
	m.RLock()
	defer m.RUnlock()
	return m.viper
}

// 獲取配置值的方法
func (m *Manager) GetCodes() models.CodeConfig {
	m.RLock()
	defer m.RUnlock()
	return m.mapping.Code
}

func (m *Manager) GetMetrics() models.MetricConfig {
	m.RLock()
	defer m.RUnlock()
	return m.mapping.Metric
}

func (m *Manager) GetTags() models.TagConfig {
	m.RLock()
	defer m.RUnlock()
	return m.mapping.Tag
}
