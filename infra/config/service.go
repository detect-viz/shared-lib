package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/alert"
	"github.com/fsnotify/fsnotify"
	"github.com/google/wire"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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
	v.AddConfigPath("./config") // 設定檔目錄

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
		configDir = "./config/conf.d"
	}
	globalCfg, err := loadGlobalConfigs(v, configDir, m.config.Logger.Level)
	if err != nil {
		panic(fmt.Errorf("載入額外設定檔失敗: %w", err))
	}
	m.global = globalCfg

	// 5️⃣ 監聽 `conf.d/` 設定變更
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("偵測到設定變更:", e.Name)
		globalCfg, err := loadGlobalConfigs(v, configDir, m.config.Logger.Level)
		if err != nil {
			panic(fmt.Errorf("載入額外設定檔失敗: %w", err))
		}
		m.global = globalCfg

	})
	if strings.ToLower(m.config.Logger.Level) == "debug" {
		// **確認環境變數是否生效**
		fmt.Printf("確認%v:%+v\n", configDir, m.config.Logger)
	}
}

// 讀取 `conf.d/*.yaml` 設定
func loadGlobalConfigs(v *viper.Viper, configDir string, level string) (*models.GlobalConfig, error) {
	files, err := filepath.Glob(filepath.Join(configDir, "*.yaml"))
	if err != nil {
		fmt.Println("❌ 無法讀取 `conf.d/` 目錄:", err)
		return nil, err
	}

	if len(files) == 0 {
		fmt.Println("⚠️ `conf.d/` 目錄內沒有 `.yaml` 設定檔")
		return nil, nil
	}

	// 初始化 GlobalConfig
	conf := &models.GlobalConfig{
		MetricRules: make(map[string]models.MetricRule),
		Templates:   []models.Template{},
	}

	// 遍歷所有 YAML 文件
	for _, file := range files {
		fmt.Println("✅ 處理設定檔:", file)

		// 讀取文件內容
		yamlFile, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("❌ 讀取 YAML 文件失敗:", file, err)
			continue
		}

		// 如果文件名包含 metric_rule，則解析 metric_rules
		if strings.Contains(file, "metric_rule") {
			fmt.Println("✅ 解析 metric_rules 設定檔:", file)

			// 解析 YAML 文件
			var data map[string]interface{}
			if err := yaml.Unmarshal(yamlFile, &data); err != nil {
				fmt.Println("❌ 解析 YAML 文件失敗:", file, err)
				continue
			}

			// 檢查是否包含 metric_rules
			if rulesRaw, ok := data["metric_rules"].([]interface{}); ok {
				fmt.Println("✅ 從 YAML 文件中讀取到", len(rulesRaw), "條規則")

				// 遍歷每個規則
				for _, ruleRaw := range rulesRaw {
					if ruleMap, ok := ruleRaw.(map[string]interface{}); ok {
						uid, _ := ruleMap["uid"].(string)
						if uid == "" {
							fmt.Println("⚠️ 跳過沒有 UID 的規則")
							continue
						}

						// 創建 MetricRule 結構體
						rule := models.MetricRule{
							UID:                uid,
							Name:               getString(ruleMap, "name"),
							Category:           getString(ruleMap, "category"),
							MatchTargetPattern: getString(ruleMap, "match_target_pattern"),
							DetectionType:      getString(ruleMap, "detection_type"),
							MetricRawName:      getString(ruleMap, "metric_raw_name"),
							MetricDisplayName:  getString(ruleMap, "metric_display_name"),
							RawUnit:            getString(ruleMap, "raw_unit"),
							DisplayUnit:        getString(ruleMap, "display_unit"),
							Scale:              getFloat64(ruleMap, "scale"),
							Duration:           getString(ruleMap, "duration"),
							Operator:           getString(ruleMap, "operator"),
							Thresholds:         alert.Threshold{},
						}

						// 處理 MatchDatasourceNames
						if datasources, ok := ruleMap["match_datasource_names"].([]interface{}); ok {
							for _, ds := range datasources {
								if dsStr, ok := ds.(string); ok {
									rule.MatchDatasourceNames = append(rule.MatchDatasourceNames, dsStr)
								}
							}
						}

						// 處理 Threshold
						if threshold, ok := ruleMap["threshold"].(map[string]interface{}); ok {
							// 處理 info 閾值
							if infoVal, ok := threshold["info"]; ok {
								// 嘗試不同類型的轉換
								if infoFloat, ok := infoVal.(float64); ok {
									infoPtr := infoFloat
									rule.Thresholds.Info = &infoPtr
									fmt.Printf("設置 %s 的 info 閾值為 %v (float64)\n", uid, infoFloat)
								} else if infoInt, ok := infoVal.(int); ok {
									infoFloat := float64(infoInt)
									rule.Thresholds.Info = &infoFloat
									fmt.Printf("設置 %s 的 info 閾值為 %v (int)\n", uid, infoInt)
								} else if infoStr, ok := infoVal.(string); ok {
									if infoFloat, err := strconv.ParseFloat(infoStr, 64); err == nil {
										rule.Thresholds.Info = &infoFloat
										fmt.Printf("設置 %s 的 info 閾值為 %v (string)\n", uid, infoStr)
									}
								} else {
									fmt.Printf("無法解析 %s 的 info 閾值: %v (類型: %T)\n", uid, infoVal, infoVal)
								}
							}

							// 處理 warn 閾值
							if warnVal, ok := threshold["warn"]; ok {
								// 嘗試不同類型的轉換
								if warnFloat, ok := warnVal.(float64); ok {
									warnPtr := warnFloat
									rule.Thresholds.Warn = &warnPtr
									fmt.Printf("設置 %s 的 warn 閾值為 %v (float64)\n", uid, warnFloat)
								} else if warnInt, ok := warnVal.(int); ok {
									warnFloat := float64(warnInt)
									rule.Thresholds.Warn = &warnFloat
									fmt.Printf("設置 %s 的 warn 閾值為 %v (int)\n", uid, warnInt)
								} else if warnStr, ok := warnVal.(string); ok {
									if warnFloat, err := strconv.ParseFloat(warnStr, 64); err == nil {
										rule.Thresholds.Warn = &warnFloat
										fmt.Printf("設置 %s 的 warn 閾值為 %v (string)\n", uid, warnStr)
									}
								} else {
									fmt.Printf("無法解析 %s 的 warn 閾值: %v (類型: %T)\n", uid, warnVal, warnVal)
								}
							}

							// 處理 crit 閾值
							if critVal, ok := threshold["crit"]; ok {
								// 嘗試不同類型的轉換
								if critFloat, ok := critVal.(float64); ok {
									rule.Thresholds.Crit = critFloat
									fmt.Printf("設置 %s 的 crit 閾值為 %v (float64)\n", uid, critFloat)
								} else if critInt, ok := critVal.(int); ok {
									rule.Thresholds.Crit = float64(critInt)
									fmt.Printf("設置 %s 的 crit 閾值為 %v (int)\n", uid, critInt)
								} else if critStr, ok := critVal.(string); ok {
									if critFloat, err := strconv.ParseFloat(critStr, 64); err == nil {
										rule.Thresholds.Crit = critFloat
										fmt.Printf("設置 %s 的 crit 閾值為 %v (string)\n", uid, critStr)
									}
								} else {
									fmt.Printf("無法解析 %s 的 crit 閾值: %v (類型: %T)\n", uid, critVal, critVal)
								}
							}

							// 輸出閾值的最終結果
							fmt.Printf("%s 的閾值: Info=%v, Warn=%v, Crit=%v\n",
								uid,
								rule.Thresholds.Info,
								rule.Thresholds.Warn,
								rule.Thresholds.Crit)
						}

						// 調試輸出
						fmt.Printf("載入規則 %s: MetricRawName=%s, MatchDatasourceNames=%v, Thresholds=%+v\n",
							rule.UID, rule.MetricRawName, rule.MatchDatasourceNames, rule.Thresholds)

						// 添加到 map
						conf.MetricRules[uid] = rule
					}
				}
			}
		} else if strings.Contains(file, "template") {
			// 如果文件名包含 template，則解析 templates
			fmt.Println("✅ 解析 templates 設定檔:", file)

			// 解析 YAML 文件
			var data map[string]interface{}
			if err := yaml.Unmarshal(yamlFile, &data); err != nil {
				fmt.Println("❌ 解析 YAML 文件失敗:", file, err)
				continue
			}

			// 檢查是否包含 templates
			if templatesRaw, ok := data["templates"].([]interface{}); ok {
				fmt.Println("✅ 從 YAML 文件中讀取到", len(templatesRaw), "個模板")

				// 遍歷每個模板
				for _, templateRaw := range templatesRaw {
					if templateMap, ok := templateRaw.(map[string]interface{}); ok {
						name, _ := templateMap["name"].(string)
						if name == "" {
							fmt.Println("⚠️ 跳過沒有名稱的模板")
							continue
						}

						// 創建 Template 結構體
						template := models.Template{
							Name:       name,
							FormatType: getString(templateMap, "format_type"),
							RuleState:  getString(templateMap, "rule_state"),
							Title:      getString(templateMap, "title"),
							Message:    getString(templateMap, "message"),
						}

						// 調試輸出
						fmt.Printf("載入模板 %s: FormatType=%s, RuleState=%s\n",
							template.Name, template.FormatType, template.RuleState)

						// 添加到數組
						conf.Templates = append(conf.Templates, template)
					}
				}
			}
		} else {
			// 其他設定檔，使用 viper 解析
			subViper := viper.New()
			subViper.SetConfigType("yaml")
			if err := subViper.ReadConfig(strings.NewReader(string(yamlFile))); err != nil {
				fmt.Println("❌ 解析設定檔失敗:", file, err)
				continue
			}

			// 合併到主 viper
			v.MergeConfigMap(subViper.AllSettings())
		}
	}

	// 解析其他設定
	if err := v.Unmarshal(conf); err != nil {
		fmt.Println("❌ 解析 GlobalConfig 失敗:", err)
		return nil, err
	}

	fmt.Println("✅ 載入 GlobalConfig 成功，MetricRules 數量:", len(conf.MetricRules))
	fmt.Println("✅ 載入 GlobalConfig 成功，Templates 數量:", len(conf.Templates))

	if strings.ToLower(level) == "debug" {
		// 輸出所有規則的詳細信息
		for uid, rule := range conf.MetricRules {
			fmt.Printf("規則 %s: MetricRawName=%s, MatchDatasourceNames=%v, Category=%s\n",
				uid, rule.MetricRawName, rule.MatchDatasourceNames, rule.Category)
		}

		// 輸出所有模板的詳細信息
		for i, template := range conf.Templates {
			fmt.Printf("模板 %d: Name=%s, FormatType=%s, RuleState=%s\n",
				i, template.Name, template.FormatType, template.RuleState)
		}
	}

	// json, _ := json.MarshalIndent(conf.MetricRules, "", "\t")
	// fmt.Println("✅ MetricRules:", string(json))

	return conf, nil
}

// 輔助函數：從 map 中獲取字符串
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// 輔助函數：從 map 中獲取 float64
func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}

// 輔助函數：從 map 中獲取 float64 指針
func getFloat64Ptr(m map[string]interface{}, key string) *float64 {
	if val, ok := m[key].(float64); ok {
		return &val
	}
	return nil
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
