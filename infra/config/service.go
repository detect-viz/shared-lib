package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/detect-viz/shared-lib/models"
	"github.com/fsnotify/fsnotify"
	"github.com/google/wire"

	"github.com/spf13/viper"
)

var ConfigSet = wire.NewSet(NewConfigManager)

// 配置管理器
type ConfigManager struct {
	sync.RWMutex
	viper  *viper.Viper
	config *models.Config
	global *models.GlobalConfig
}

// 創建配置管理器
func NewConfigManager() *ConfigManager {
	cm := &ConfigManager{
		viper: viper.New(),
	}
	cm.loadConfig()
	return cm
}

// LoadConfig 載入設定
func (m *ConfigManager) loadConfig() {
	m.Lock()
	defer m.Unlock()

	v := viper.New()
	configExists := true

	// 1️⃣ **先嘗試讀取 `config.yaml`**
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./conf") // 設定檔目錄

	if err := v.ReadInConfig(); err != nil {
		fmt.Println("⚠️ 找不到 `config.yaml`，將使用環境變數與 `.env`。")
		configExists = false
	}

	// 2️⃣ **只有當 `config.yaml` 不存在時，才讀取環境變數**
	if !configExists {
		v.AutomaticEnv() // 讓系統環境變數可用
	}

	// 4️⃣ **最後讀取 `custom.yaml`，讓 `custom.yaml` 覆蓋設定**
	envViper := viper.New()
	envViper.SetConfigName("custom")
	envViper.SetConfigType("yaml")
	envViper.AddConfigPath(".") // `custom.yaml` 在專案根目錄

	if err := envViper.ReadInConfig(); err == nil {
		// **合併 `custom.yaml` 設定**
		_ = v.MergeConfigMap(envViper.AllSettings())
	}

	// 6️⃣ 解析配置
	conf, err := parseConfig(v)
	if err != nil {
		panic(fmt.Errorf("解析配置失敗: %w", err))
	}
	m.config = conf

	// 4️⃣ 讀取 `conf.d/` 內的 YAML 檔案
	configDir := v.GetString("server.config_directory")
	if configDir == "" {
		configDir = "./conf/conf.d"
	}
	globalCfg, err := loadGlobalConfigs(v, configDir)
	if err != nil {
		panic(fmt.Errorf("載入額外設定檔失敗: %w", err))
	}
	m.global = globalCfg

	// 5️⃣ 監聽 `conf.d/` 設定變更
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("偵測到設定變更:", e.Name)
		globalCfg, err := loadGlobalConfigs(v, configDir)
		if err != nil {
			panic(fmt.Errorf("載入額外設定檔失敗: %w", err))
		}
		m.global = globalCfg
	})
	if strings.ToLower(m.config.Logger.Level) == "debug" {
		// **確認環境變數是否生效**
		b, _ := json.MarshalIndent(m.config.Logger, "", "\t")
		os.Stdout.Write(b)
	}
}

// 讀取 `conf.d/*.yaml` 設定
func loadGlobalConfigs(v *viper.Viper, configDir string) (*models.GlobalConfig, error) {
	files, err := filepath.Glob(filepath.Join(configDir, "*.yaml"))
	if err != nil {
		fmt.Println("無法讀取 `conf.d/` 目錄:", err)
		return nil, err
	}

	if len(files) == 0 {
		return nil, nil
	}

	for _, file := range files {
		subViper := viper.New()
		subViper.SetConfigFile(file)

		if err := subViper.ReadInConfig(); err != nil {
			fmt.Println("讀取設定檔失敗:", file, err)
			continue
		}

		// **合併 `conf.d/*.yaml` 設定**
		v.MergeConfigMap(subViper.AllSettings())

		fmt.Println("載入設定檔:", file)
	}
	var conf models.GlobalConfig
	if err := v.Unmarshal(&conf); err != nil {
		fmt.Println("解析設定檔失敗:", err)
	}

	return &conf, nil
}

// 解析配置
func parseConfig(v *viper.Viper) (*models.Config, error) {
	var conf models.Config
	if err := v.Unmarshal(&conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

// 獲取完整配置
func (m *ConfigManager) GetConfig() *models.Config {
	m.RLock()
	defer m.RUnlock()
	return m.config
}

// 獲取日誌配置
func (m *ConfigManager) GetLoggerConfig() models.LoggerConfig {
	m.RLock()
	defer m.RUnlock()
	return m.config.Logger
}

// 獲取告警配置
func (m *ConfigManager) GetAlertConfig() models.AlertConfig {
	m.RLock()
	defer m.RUnlock()
	return m.config.Alert
}

// 獲取資料庫配置
func (m *ConfigManager) GetDatabaseConfig() models.DatabaseConfig {
	m.RLock()
	defer m.RUnlock()
	return m.config.Database
}

// 獲取原始 viper 實例
func (m *ConfigManager) GetRawConfig() *viper.Viper {
	m.RLock()
	defer m.RUnlock()
	return m.viper
}

// 獲取 Keycloak 配置
func (m *ConfigManager) GetKeycloakConfig() models.KeycloakConfig {
	m.RLock()
	defer m.RUnlock()
	return m.config.Keycloak
}

// 獲取 Global 配置
func (m *ConfigManager) GetGlobalConfig() models.GlobalConfig {
	m.RLock()
	defer m.RUnlock()
	return *m.global
}
