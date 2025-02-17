## SpecStory

## 使用 YAML 設定告警規則 (2025/2/14 下午4:13:37)

_**User**_

我想靠yaml設定告警規則可以嗎

---

_**Assistant**_

是的，可以使用 YAML 來設定告警規則。我建議在 `settings/metrics.yaml` 中定義告警規則的配置。以下是一個建議的實現方式：

首先，讓我們創建一個新的 YAML 配置文件：

```yaml:settings/alert_rules.yaml
rules:
  - name: "CPU 使用率過高"
    resource_group: "servers"
    metric: "cpu_usage"
    check_type: "absolute"
    operator: ">"
    thresholds:
      info: 70
      warn: 80
      crit: 90
    unit: "%"
    duration: 300  # 5分鐘
    labels:
      team: "ops"
      environment: "prod"

  - name: "記憶體使用量突增"
    resource_group: "servers"
    metric: "memory_usage"
    check_type: "amplitude"
    operator: ">"
    thresholds:
      warn: 30  # 30% 增幅
      crit: 50  # 50% 增幅
    unit: "%"
    duration: 180  # 3分鐘
    labels:
      team: "ops"
      environment: "prod"
```

然後，我們需要創建一個結構體來解析這個 YAML 配置：

```go:models/alert/rule_config.go
package alert

// RuleConfig 代表 YAML 中的告警規則配置
type RuleConfig struct {
	Rules []RuleDefinition `yaml:"rules"`
}

// RuleDefinition 代表單個告警規則的定義
type RuleDefinition struct {
	Name          string             `yaml:"name"`
	ResourceGroup string             `yaml:"resource_group"`
	Metric        string             `yaml:"metric"`
	CheckType     string             `yaml:"check_type"`
	Operator      string             `yaml:"operator"`
	Thresholds    ThresholdConfig    `yaml:"thresholds"`
	Unit          string             `yaml:"unit"`
	Duration      int                `yaml:"duration"`
	Labels        map[string]string  `yaml:"labels"`
}

// ThresholdConfig 代表不同級別的閾值配置
type ThresholdConfig struct {
	Info  *float64 `yaml:"info,omitempty"`
	Warn  *float64 `yaml:"warn,omitempty"`
	Crit  *float64 `yaml:"crit,omitempty"`
}

// ToCheckRule 將 RuleDefinition 轉換為 CheckRule
func (rd *RuleDefinition) ToCheckRule() *CheckRule {
	return &CheckRule{
		RuleName:      rd.Name,
		ResourceGroup: rd.ResourceGroup,
		Metric:        rd.Metric,
		CheckType:     rd.CheckType,
		Operator:      rd.Operator,
		Unit:          rd.Unit,
		Duration:      rd.Duration,
		Labels:        rd.Labels,
		InfoThreshold: rd.Thresholds.Info,
		WarnThreshold: rd.Thresholds.Warn,
		CritThreshold: rd.Thresholds.Crit,
	}
}
```

最後，我們需要一個函數來加載和解析這個配置：

```go:alert/service.go
package alert

import (
	"gopkg.in/yaml.v2"
	"os"
)

// LoadRuleConfig 從 YAML 文件加載告警規則配置
func LoadRuleConfig(filepath string) (*RuleConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var config RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
```

使用這個配置系統的主要優點是：

1. 配置更容易維護和修改，不需要修改代碼
2. 可以輕鬆地為不同環境（測試、生產等）維護不同的規則集
3. 規則的變更可以通過版本控制系統追蹤
4. 可以動態重新加載規則，無需重啟應用

要使用這個配置，你可以這樣做：

```go
config, err := LoadRuleConfig("settings/alert_rules.yaml")
if err != nil {
    log.Fatalf("無法加載告警規則配置: %v", err)
}

// 將配置轉換為 CheckRule
for _, ruleDef := range config.Rules {
    checkRule := ruleDef.ToCheckRule()
    // 使用 checkRule 進行告警檢查
}
```

這樣的設計讓告警規則的管理變得更加靈活和可維護。你可以根據需要輕鬆地添加或修改規則，而不需要修改代碼。

---

_**User**_

@contact.go 通知管道也可以透過yaml讀取設定嗎

---

_**Assistant**_

是的，我們可以通過 YAML 來設定通知管道。以下是建議的實現方式：

首先，創建一個通知管道的 YAML 配置文件：

```yaml:settings/contacts.yaml
contacts:
  - name: "運維團隊"
    type: "telegram"
    enabled: true
    severities: ["warn", "crit"]
    details:
      chat_id: "-100123456789"
      bot_token: "your_bot_token"

  - name: "開發團隊"
    type: "slack"
    enabled: true
    severities: ["crit"]
    details:
      webhook_url: "https://hooks.slack.com/services/xxx/yyy/zzz"
      channel: "#alerts"

  - name: "系統管理員"
    type: "email"
    enabled: true
    severities: ["info", "warn", "crit"]
    details:
      address: "admin@example.com"
```

然後，在 `models/alert/contact.go` 中添加相關的配置結構：

```go:models/alert/contact.go
// ContactConfig 代表 YAML 中的通知管道配置
type ContactConfig struct {
	Contacts []ContactDefinition `yaml:"contacts"`
}

// ContactDefinition 代表單個通知管道的定義
type ContactDefinition struct {
	Name       string                 `yaml:"name"`
	Type       string                 `yaml:"type"`
	Enabled    bool                   `yaml:"enabled"`
	Severities []string              `yaml:"severities"`
	Details    map[string]interface{} `yaml:"details"`
}

// ToAlertContact 將 ContactDefinition 轉換為 AlertContact
func (cd *ContactDefinition) ToAlertContact() *AlertContact {
	// 將 severities 字符串轉換為 AlertContactSeverity
	var severities []AlertContactSeverity
	for _, s := range cd.Severities {
		severities = append(severities, AlertContactSeverity{
			Severity: "level_" + s,
		})
	}

	return &AlertContact{
		Name:       cd.Name,
		Type:       cd.Type,
		Enabled:    cd.Enabled,
		Details:    cd.Details,
		Severities: severities,
	}
}

// LoadContactConfig 從 YAML 文件加載通知管道配置
func LoadContactConfig(filepath string) (*ContactConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var config ContactConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
```

在 `alert/service.go` 中添加加載通知管道配置的方法：

```go:alert/service.go
// InitContacts 初始化通知管道
func (s *Service) InitContacts(configPath string) error {
	// 加載配置
	config, err := models.LoadContactConfig(configPath)
	if err != nil {
		return fmt.Errorf("加載通知管道配置失敗: %w", err)
	}

	// 轉換並保存到資料庫
	for _, contactDef := range config.Contacts {
		contact := contactDef.ToAlertContact()
		
		// 設置默認值
		contact.RealmName = "master"
		
		// 保存到資料庫
		if err := s.db.CreateOrUpdateAlertContact(contact); err != nil {
			s.logger.Error("保存通知管道失敗",
				zap.String("name", contact.Name),
				zap.Error(err))
			continue
		}
	}

	s.logger.Info("通知管道初始化完成",
		zap.Int("contact_count", len(config.Contacts)))

	return nil
}
```

使用方式：

1. 在 `main.go` 或服務初始化時加載通知管道配置：

```go
if err := alertService.InitContacts("settings/contacts.yaml"); err != nil {
    log.Fatalf("初始化通知管道失敗: %v", err)
}
```

這樣設計的優點是：

1. 配置更容易維護和修改
2. 可以為不同環境維護不同的通知管道配置
3. 敏感資訊（如 API token）可以從代碼中分離
4. 可以動態更新通知管道，無需重啟應用
5. 配置變更可以通過版本控制系統追蹤

注意事項：

1. 確保敏感資訊（如 API token）的安全性
2. 可以考慮使用環境變數或其他安全的方式來存儲敏感資訊
3. 建議在加載配置時驗證必要的字段
4. 可以添加配置熱重載功能

這樣的設計讓通知管道的管理變得更加靈活和安全。

---

_**User**_

For the code present, we get this error:
```
cannot convert cd.Details (variable of type interface{}) to type models.JSONMap: need type assertion
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_

我們需要將 `interface{}` 類型的 `Details` 正確轉換為 `models.JSONMap`。以下是修復方案：

```go:alert/setting.go
// ToAlertContact 將 ContactDefinition 轉換為 AlertContact
func (cd *ContactDefinition) ToAlertContact() *models.AlertContact {
	// ... 其他代碼保持不變 ...

	return &models.AlertContact{
		Name:       cd.Name,
		Type:       cd.Type,
		Enabled:    cd.Enabled,
		Details:    models.JSONMap(cd.Details.(map[string]interface{})),  // 添加類型斷言
		Severities: severities,
	}
}
```

這個修改將 `interface{}` 先斷言為 `map[string]interface{}`，然後再轉換為 `models.JSONMap`。

---

_**User**_

For the code present, we get this error:
```
cannot convert cd.Details (variable of type map[string]interface{}) to type models.JSONMap
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_

我們需要修改 `models.JSONMap` 的類型定義。這個錯誤表明 `JSONMap` 可能沒有正確定義為 `map[string]interface{}`。

請檢查 `models/models.go` 文件，確保有以下定義：

```go:models/models.go
type JSONMap map[string]interface{}
```

如果已經有這個定義，那麼我們需要修改 `ToAlertContact` 中的轉換方式：

```go:alert/setting.go
Details: cd.Details, // 直接賦值，不需要類型轉換
```

因為如果 `JSONMap` 就是 `map[string]interface{}`，那麼它們應該是可以直接賦值的。

---

_**User**_

For the code present, we get this error:
```
Property rules is not allowed.
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_



我們需要修改 `RuleConfig` 結構體的 yaml tag。讓我們更新 `models/alert/rule_config.go`：

```go:models/alert/rule_config.go
// RuleConfig 代表 YAML 中的告警規則配置
type RuleConfig struct {
	Rules []RuleDefinition `yaml:"alert_rules"` // 改為 alert_rules
}
```

然後相應地修改 YAML 配置文件：

```yaml:conf/alert_rules.yaml
alert_rules:  # 改為 alert_rules
  - name: "CPU 使用率過高"
    resource_group: "servers"
    metric: "cpu_usage"
    # ... 其他配置保持不變 ...
```

這樣修改後，YAML 的結構會與 Go 結構體定義匹配。

---

_**User**_

如何設定不要關聯 prometheus.rules.json

---

_**Assistant**_



我們需要修改 YAML 配置文件的名稱和結構，讓它完全獨立於 Prometheus 的配置。建議創建一個新的配置文件：

```yaml:conf/monitoring_rules.yaml
rules:
  - name: "CPU 使用率過高"
    resource_group: "servers"
    metric: "cpu_usage"
    check_type: "absolute"
    operator: ">"
    thresholds:
      info: 70
      warn: 80
      crit: 90
    unit: "%"
    duration: 300
    contacts:
      - name: "Messaging API 運維團隊"
    labels:
      team: "ops"
      environment: "prod"
```

然後修改 `RuleConfig` 結構體：

```go:models/alert/rule_config.go
// RuleConfig 代表 YAML 中的告警規則配置
type RuleConfig struct {
	Rules []RuleDefinition `yaml:"rules"`  // 改回 rules
}
```

這樣就可以完全獨立於 Prometheus 的配置文件。在使用時，需要修改配置文件的路徑：

```go:alert/service.go
config, err := LoadRuleConfig("conf/monitoring_rules.yaml")  // 使用新的配置文件
```

這樣的設計更清晰地表明這是我們自己的監控規則配置，而不是 Prometheus 的規則。

---

_**User**_

 @run.go 創建 Service

---

_**Assistant**_



我們需要修改 `alert/cmd/run.go` 來正確初始化 `Service`。以下是修改建議：

```go:alert/cmd/run.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"shared-lib/alert"
	"shared-lib/databases"
	"shared-lib/logger"
	"shared-lib/models"
	"shared-lib/models/config"
)

// RunResourceAlert 會將 metrics 轉為 JSON 並存入檔案
func RunResourceAlert(realm, source, filename, hostname string, metrics map[string][]map[string]interface{}) {
	file := models.FileInfo{
		Realm:    realm,
		Source:   source,
		FileName: filename,
		Host:     hostname,
	}
	outputMetrics(file, metrics)

	// 初始化配置
	cfg := config.AlertConfig{
		WorkPath: config.WorkPath{
			Sent:       "data/alert/sent",
			Notify:     "data/alert/notify",
			Silence:    "data/alert/silence",
			Unresolved: "data/alert/unresolved",
			Resolved:   "data/alert/resolved",
		},
	}

	// 初始化 logger
	logSvc := logger.NewLogger()

	// 初始化資料庫連接
	db := databases.NewMySQLDB()

	// 創建 Service 實例並處理告警
	alertSvc := alert.NewService(cfg, logSvc, db)
	if err := alertSvc.ProcessFile(file, metrics); err != nil {
		fmt.Println("處理告警失敗:", err)
	}
}
```

這個修改：
1. 引入必要的依賴包
2. 創建基本的 AlertConfig 配置
3. 初始化 logger 和資料庫服務
4. 使用這些組件創建 AlertService

注意：你可能需要根據實際情況調整配置路徑和其他參數。

---

_**User**_

For the code present, we get this error:
```
undefined: config.InitConfig
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_



我們需要在 `config/config.go` 中添加 `InitConfig` 函數：

```go:config/config.go
package config

import (
	"os"
	"shared-lib/models/config"

	"gopkg.in/yaml.v3"
)

// InitConfig 初始化配置
func InitConfig(configPath string) config.Config {
	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return cfg
}
```

這個函數會讀取 YAML 配置文件並解析為 `config.Config` 結構體。

---

_**User**_

哪個方式正確

---

_**Assistant**_



根據代碼結構，應該使用 `Manager` 的方式來管理配置。以下是正確的修改方式：

1. 首先修改 `main.go`：

```go:main.go
func main() {
	// 1️⃣ 讀取設定檔
	configManager := config.New()
	if err := configManager.Load("/etc/bimap-ipoc/config.yml"); err != nil {
		panic(err)
	}
	cfg := configManager.GetConfig()

	// 2️⃣ 初始化各個 Service
	alertService := alert.NewService(cfg.Alert, logger, db)
	// ...其他代碼
}
```

2. 修改 `alert/cmd/run.go`：

```go:alert/cmd/run.go
func RunResourceAlert(realm, source, filename, hostname string, metrics map[string][]map[string]interface{}) {
	// ...其他代碼

	// 初始化配置
	configManager := config.New()
	if err := configManager.Load("/etc/bimap-ipoc/config.yml"); err != nil {
		fmt.Println("加載配置失敗:", err)
		return
	}
	cfg := configManager.GetConfig()

	// ...其他代碼
}
```

這樣使用 `Manager` 的方式有以下優點：
1. 統一的配置管理
2. 支持配置熱重載
3. 線程安全
4. 更好的錯誤處理

建議移除之前添加的 `InitConfig` 函數，統一使用 `Manager` 的方式。

---

_**User**_

初始化各個 Service 所以 logger 跟 db 要先初始化 請根據我的代碼結構幫我完成

---

_**Assistant**_



根據代碼結構，我們需要在 `main.go` 中初始化 logger 和 database。以下是修改建議：

```go:main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"shared-lib/alert"
	"shared-lib/config"
	"shared-lib/databases"
	"shared-lib/logger"
	"shared-lib/models"
	"shared-lib/notify"

	"go.uber.org/zap"
)

func main() {
	// 1️⃣ 讀取設定檔
	configManager := config.New()
	if err := configManager.Load("/etc/bimap-ipoc/config.yml"); err != nil {
		panic(err)
	}
	cfg := configManager.GetConfig()

	// 2️⃣ 初始化 logger
	logSvc := logger.NewLogger(cfg.Logger, logger.WithLevel(zap.InfoLevel))

	// 3️⃣ 初始化資料庫
	db := databases.NewMySQLDB(cfg.Database)

	// 4️⃣ 初始化各個 Service
	alertService := alert.NewService(cfg.Alert, logSvc, db)

	// 5️⃣ 啟動模組
	if err := alertService.Init(); err != nil {
		logSvc.Error("初始化告警服務失敗", zap.Error(err))
		return
	}

	file := models.FileInfo{
		Realm:    "master",
		Source:   "test",
		FileName: "test",
		Host:     "test",
	}

	metrics := readJsonFile("output/AL2SUB_08310400.csv.json")

	if err := alertService.ProcessFile(file, metrics); err != nil {
		fmt.Println("處理告警失敗:", err)
	}
}

// ... readJsonFile 函數保持不變 ...
```

主要修改：
1. 添加必要的 import
2. 按順序初始化各個組件
3. 移除未定義的 `CheckAlert` 和 `SendNotification` 調用
4. 使用 `alertService.Init()` 進行初始化
5. 添加錯誤處理

這樣可以確保各個組件按正確的順序初始化和啟動。

---

_**User**_

For the code present, we get this error:
```
cannot use cfg.Logger (variable of type "shared-lib/models/config".LoggerConfig) as *models.LoggerConfig value in argument to logger.NewLogger
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_



需要修改 `logger.NewLogger` 的調用方式，因為它期望一個指針類型的參數：

```go:main.go
// 2️⃣ 初始化 logger
logSvc, err := logger.NewLogger(&cfg.Logger, logger.WithLevel(cfg.Logger.Level))
if err != nil {
	panic(err)
}
```

主要變更是在 `cfg.Logger` 前加上 `&` 運算符，將其轉換為指針類型。

---

_**User**_

For the code present, we get this error:
```
undefined: logger.WithLevel
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_





我們需要修改 logger 的初始化方式。由於 `WithLevel` 未定義，我們可以直接使用配置：

```go:main.go
// 2️⃣ 初始化 logger
logSvc, err := logger.NewLogger(&cfg.Logger)
if err != nil {
    panic(err)
}
```

移除了 `logger.WithLevel` 選項，因為日誌級別應該已經在 `cfg.Logger` 配置中定義。

---

_**User**_

For the code present, we get this error:
```
cannot use db (variable of type *databases.MySQL) as interfaces.Database value in argument to alert.NewService: *databases.MySQL does not implement interfaces.Database (missing method Close)
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_



我們需要在 `databases/mysql.go` 中為 `MySQL` 結構體實現 `Close` 方法：

```go:databases/mysql.go
// Close 關閉資料庫連接
func (db *MySQL) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
```

這樣 `MySQL` 結構體就完整實現了 `interfaces.Database` 接口。

---

_**User**_

For the code present, we get this error:
```
cannot use db (variable of type *databases.MySQL) as interfaces.Database value in argument to alert.NewService: *databases.MySQL does not implement interfaces.Database (missing method CreateAlertRule)
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_





我們需要在 `databases/mysql.go` 中添加 `CreateAlertRule` 方法：

```go:databases/mysql.go
// CreateAlertRule 創建告警規則
func (m *MySQL) CreateAlertRule(rule models.AlertRule) error {
	return m.db.Create(&rule).Error
}

// CreateOrUpdateAlertRule 創建或更新告警規則
func (m *MySQL) CreateOrUpdateAlertRule(rule models.AlertRule) error {
	return m.db.Save(&rule).Error
}

// CreateOrUpdateAlertContact 創建或更新通知管道
func (m *MySQL) CreateOrUpdateAlertContact(contact models.AlertContact) error {
	return m.db.Save(&contact).Error
}
```


這樣就實現了 `interfaces.Database` 接口所需的所有方法。

---

_**User**_

For the code present, we get this error:
```
cannot use db (variable of type *databases.MySQL) as interfaces.Database value in argument to alert.NewService: *databases.MySQL does not implement interfaces.Database (wrong type for method CreateAlertRule)
		have CreateAlertRule(models.AlertRule) error
		want CreateAlertRule(*models.AlertRule) error
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_





修改 `databases/mysql.go` 中的 `CreateAlertRule` 方法，使用指針參數：

```go:databases/mysql.go
// CreateAlertRule 創建告警規則
func (m *MySQL) CreateAlertRule(rule *models.AlertRule) error {
	return m.db.Create(rule).Error
}

// CreateOrUpdateAlertRule 創建或更新告警規則
func (m *MySQL) CreateOrUpdateAlertRule(rule *models.AlertRule) error {
	return m.db.Save(rule).Error
}

// CreateOrUpdateAlertContact 創建或更新通知管道
func (m *MySQL) CreateOrUpdateAlertContact(contact *models.AlertContact) error {
	return m.db.Save(contact).Error
}
```


所有方法都改為使用指針參數，以匹配接口定義。

---

_**User**_

For the code present, we get this error:
```
contactDef.ToAlertContact undefined (type "shared-lib/models/alert".ContactDefinition has no field or method ToAlertContact)
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_





我們需要在 `models/alert/contact_config.go` 中定義 `ToAlertContact` 方法：

```go:models/alert/contact_config.go
package alert

// ContactConfig 代表 YAML 中的通知管道配置
type ContactConfig struct {
	Contacts []ContactDefinition `yaml:"contacts"`
}

// ContactDefinition 代表單個通知管道的定義
type ContactDefinition struct {
	Name       string                 `yaml:"name"`
	Type       string                 `yaml:"type"`
	Enabled    bool                   `yaml:"enabled"`
	Severities []string               `yaml:"severities"`
	Details    map[string]interface{} `yaml:"details"`
}

// ToAlertContact 將 ContactDefinition 轉換為 AlertContact
func (cd *ContactDefinition) ToAlertContact() *AlertContact {
	var severities []AlertContactSeverity
	for _, s := range cd.Severities {
		severities = append(severities, AlertContactSeverity{
			Severity: "level_" + s,
		})
	}

	j := make(JSONMap)
	for k, v := range cd.Details {
		j[k] = v.(string)
	}

	return &AlertContact{
		Name:       cd.Name,
		Type:       cd.Type,
		Enabled:    cd.Enabled,
		Details:    j,
		Severities: severities,
	}
}
```

然後移除 `alert/setting.go` 中的 `ToAlertContact` 方法。

---

_**User**_

For the code present, we get this error:
```
ruleDef.ToAlertRule undefined (type "shared-lib/models/alert".RuleDefinition has no field or method ToAlertRule)
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_



---



---

