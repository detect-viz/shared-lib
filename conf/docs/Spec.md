# Share 共用模組

此目錄包含了可以被其他服務共用的核心模組，專門用於**監控與告警解決服務**。整個系統由多個子模組組成，分別負責不同的功能，通過協同工作實現高效的監控、告警及通知。

## 目錄結構

```
shared-lib/
├── alert/         # 告警核心邏輯
├── api/           # API 處理器
├── auth/          # 認證模組
├── config/        # 配置管理
├── databases/     # 資料庫操作封裝
├── label/         # 標籤管理
├── logger/        # 日誌處理
├── models/        # 資料模型
├── notify/        # 通知服務
├── parser/        # 解析器
└── template/      # 模板管理
```

## 模組說明

### 1. Parser 解析器模組

**職責：**
- 負責從各種數據源獲取並解析監控數據。
- 支援多種數據格式，確保數據的準確性和一致性。
- 提供統一的解析接口，便於擴展新的數據源解析器。

**設定：**
- 配置文件 `configs/metrics.yaml` 定義了需要監控的指標及其相關信息。
- 環境變數或配置文件用於設置數據源的詳細參數，如路徑、文件格式等。

**資料流：**
1. 從指定的數據源（如 NMON、LOGMAN）讀取原始數據。
2. 解析數據並轉換為結構化格式存入資料庫。
3. 提供給告警模組進行異常檢測。

### 2. Alert 告警模組

**職責：**
- 實時檢測監控數據中的異常情況。
- 支援兩種檢查類型：
  - **絕對值（Absolute）**：基於固定閾值判斷異常。
  - **變化幅度（Amplitude）**：基於變化的幅度判斷異常。
- 提供告警狀態管理和靜默期設置，避免重複告警。
- 支援多租戶（realm）環境，確保數據隔離與安全。

**設定：**
- 配置文件 `configs/alert_rules.yaml` 定義了各種告警規則及其參數。
- 環境變數用於設置告警相關的全局參數，如告警級別、重試策略等。

**資料流：**
1. 從 **Parser** 模組獲取最新的監控數據。
2. 根據配置的告警規則進行異常檢測。
3. 異常情況下，生成告警並更新告警狀態。

### 3. Notify 通知模組

**職責：**
- 根據告警模組生成的告警，通過不同的通知管道（如 Email、Teams、Webhook 等）發送通知。
- 支援多種通知類型，提供靈活的通知配置。
- 提供通知重試機制，保證通知的可靠性。
- 統一的通知配置管理，方便維護和擴展。

**設定：**
- 配置文件 `configs/notify_config.yaml` 定義了各種通知類型及其參數。
- 環境變數用於設置通知相關的全局參數，如 SMTP 服務器配置、Webhook URL 等。

**資料流：**
1. 接收 **Alert** 模組生成的告警信息。
2. 根據告警信息選擇相應的通知管道。
3. 使用 **Template** 模組生成符合格式的通知內容。
4. 發送通知並記錄通知狀態。

### 4. Template 模板管理模組

**職責：**
- 提供統一的模板系統，用於生成各種通知的內容。
- 支援多種模板格式（如 HTML、Markdown），滿足不同通知需求。
- 管理和維護通知模板，支持動態更新和擴展。

**設定：**
- 模板文件存儲在 `configs/templates.yaml` 目錄下，可根據需求新增或修改模板。
- 環境變數配置模板引擎的相關參數，如模板緩存策略等。

**資料流：**
1. 接收來自 **Notify** 模組的告警信息。
2. 根據通知類型選擇相應的模板。
3. 渲染模板生成最終的通知內容。
4. 將生成的內容傳遞給 **Notify** 模組進行發送。

### 5. Label 標籤管理模組

**職責：**
- 管理系統中的各類標籤，支持標籤的 CRUD 操作。
- 支援多租戶（realm）環境，確保標籤的隔離與安全。
- 標籤值支援一對多關係，方便靈活的標籤分類。

**設定：**
- 配置文件 `configs/labels.yaml` 定義了標籤的結構和屬性。
- 使用環境變數設置標籤管理的相關參數，如標籤的命名規則等。

**資料流：**
1. 通過 API 接口新增、更新或刪除標籤。
2. 標籤用於分類和過濾告警，提高告警的精確性。
3. 提供給 **Alert** 和 **Notify** 模組進行數據篩選和通知路由。

## 整體資料流

1. **Parser** 模組從各個數據源收集原始監控數據並解析成結構化格式，存入資料庫。
2. **Alert** 模組根據配置的告警規則對數據進行異常檢測，生成告警並更新狀態。
3. **Label** 模組提供標籤管理功能，對告警進行分類和過濾，提升告警的精確性。
4. **Notify** 模組接收告警信息，利用 **Template** 模組生成通知內容，通過配置的通知管道發送通知。
5. 所有模組均使用 **Logger** 模組進行日誌記錄，便於追蹤和排查問題。

## 使用說明

1. **依賴注入：**
   - 所有模組應遵循依賴注入原則，確保模組之間的解耦和靈活擴展。
   - 使用接口和構造函數進行依賴注入，避免直接依賴具體實現。

2. **錯誤處理：**
   - 錯誤處理應使用統一的錯誤類型，便於集中管理和處理。
   - 使用自定義錯誤類型（如 `NotifyError`）來提供更豐富的錯誤信息。

3. **配置管理：**
   - 配置應通過環境變數或配置文件注入，保持靈活性和可配置性。
   - 使用統一的配置管理模組，避免配置散亂和重複。

4. **日誌記錄：**
   - 日誌記錄應使用統一的 **Logger** 模組，確保日誌格式和級別的一致性。
   - 支援日誌輪轉和清理，避免日誌佔用過多磁碟空間。

## 開發規範

1. **函數註釋：**
   - 所有公開的函數都應該有詳細的註釋說明，便於理解和使用。

2. **錯誤訊息：**
   - 錯誤訊息應該使用中文，確保開發和運維人員能夠快速理解問題。

3. **單元測試：**
   - 所有模組應維護完善的單元測試，確保功能的正確性和穩定性。

4. **Go 標準項目結構：**
   - 遵循 Go 標準項目結構，保持代碼的整齊和可維護性。
   - 使用模組分層，確保代碼的可重用性和擴展性。

## 各模組詳細介紹

### Parser 解析器模組

- **功能：**
  - 解析來自不同數據源（如 NMON、LOGMAN）的監控數據。
  - 支援數據的清洗、轉換和存儲。
  
- **配置：**
  - `configs/metrics.yaml` 定義需要監控的指標。
  - 每個數據源有專屬的配置文件，指定數據源的位置和格式。

- **數據流：**
  1. 從數據源讀取原始數據文件。
  2. 解析數據並轉換為結構化數據。
  3. 存入資料庫，供後續告警模組使用。

### Alert 告警模組

- **功能：**
  - 根據解析後的監控數據，應用告警規則檢測異常。
  - 支援基於絕對值和變化幅度的告警檢測。
  - 管理告警狀態，避免重複告警，並支援靜默期設置。
  
- **配置：**
  - `configs/alert_rules.yaml` 定義各種告警規則，包括檢測類型、閾值、持續時間等。
  - 靜默期和重試策略等全局告警參數通過環境變數配置。

- **數據流：**
  1. 從資料庫讀取最新的監控數據。
  2. 應用告警規則進行異常檢測。
  3. 若檢測到異常，生成告警並更新狀態。

### Notify 通知模組

- **功能：**
  - 接收告警信息，選擇合適的通知管道（Email、Teams、Webhook 等）進行通知。
  - 使用模板生成通知內容，確保通知的一致性和可讀性。
  - 提供通知重試機制，提升通知可靠性。
  
- **配置：**
  - `configs/notify_config.yaml` 定義各種通知類型及其相關參數。
  - 通知管道的具體參數（如 SMTP 配置、Webhook URL 等）通過配置文件設置。

- **數據流：**
  1. 接收 **Alert** 模組生成的告警信息。
  2. 根據告警信息選擇合適的通知管道。
  3. 使用 **Template** 模組生成通知內容。
  4. 發送通知並記錄通知狀態。

### Template 模板管理模組

- **功能：**
  - 提供統一的模板系統，用於生成不同通知管道的通知內容。
  - 支援多種模板格式（如 HTML、Markdown），滿足不同通知需求。
  - 管理和維護通知模板，支持動態更新和擴展。

- **配置：**
  - 模板文件存儲在 `share/template/` 目錄下。
  - 可以根據不同通知類型和需求新增或修改模板文件。

- **數據流：**
  1. 接收 **Notify** 模組的告警信息。
  2. 根據通知類型選擇相應的模板文件。
  3. 渲染模板生成最終的通知內容。
  4. 將生成的內容傳遞給 **Notify** 模組發送。

### Label 標籤管理模組

- **功能：**
  - 管理系統中的各類標籤，支援標籤的 CRUD 操作。
  - 支援多租戶（realm）環境，確保標籤的隔離與安全。
  - 標籤值支援一對多關係，方便靈活的標籤分類。
  
- **配置：**
  - `configs/labels.yaml` 定義了標籤的結構和屬性。
  - 使用環境變數設置標籤管理的相關參數，如標籤的命名規則等。

- **數據流：**
  1. 通過 API 接口新增、更新或刪除標籤。
  2. 標籤用於分類和過濾告警，提高告警的精確性。
  3. 提供給 **Alert** 和 **Notify** 模組進行數據篩選和通知路由。

## 整體資料流

1. **Parser** 模組從各個數據源收集原始監控數據並解析成結構化格式，存入資料庫。
2. **Alert** 模組根據配置的告警規則對數據進行異常檢測，生成告警並更新狀態。
3. **Label** 模組提供標籤管理功能，對告警進行分類和過濾，提升告警的精確性。
4. **Notify** 模組接收告警信息，利用 **Template** 模組生成通知內容，通過配置的通知管道發送通知。
5. 所有模組均使用 **Logger** 模組進行日誌記錄，便於追蹤和排查問題。

## 詳細模組互動

### Parser -> Alert

1. **Parser** 收集並解析監控數據，存入資料庫中的各個監控指標表。
2. **Alert** 定期或實時從資料庫中讀取最新的監控數據。
3. **Alert** 根據告警規則對數據進行檢測，確定是否有異常狀況。

### Alert -> Label

1. **Alert** 生成告警時，根據預設或動態標籤將告警歸類。
2. **Label** 提供標籤服務，允許用戶對告警進行細粒度的分類和管理。

### Alert -> Notify

1. **Alert** 發現異常後，將告警信息傳遞給 **Notify** 模組。
2. **Notify** 根據告警信息中的標籤和配置選擇合適的通知管道。

### Notify -> Template

1. **Notify** 模組將告警信息傳遞給 **Template** 模組。
2. **Template** 模組根據通知類型和模板配置，生成格式化的通知內容。

### Notify -> Label

1. **Notify** 根據告警的標籤信息進行通知路由，確保通知發送到正確的聯絡人或頻道。

## API 功能總覽

### Parser 模組

- **RunParser**：啟動解析器，定時從數據源讀取並解析監控數據。
- **LogRotate**：日誌輪轉與清理，確保日誌不佔用過多磁碟空間。

### Alert 模組

- **CreateAlertRule**：創建新的告警規則。
- **UpdateAlertRule**：更新現有的告警規則。
- **DeleteAlertRule**：刪除告警規則。
- **GetAllAlertRules**：獲取所有告警規則。
- **GetAlertRuleByID**：根據 ID 獲取具體的告警規則。

### Notify 模組

- **ProcessNotifications**：處理待發送的通知。
- **SendNotification**：通過配置的通知管道發送通知。
- **RetryNotification**：重試發送失敗的通知。

### Template 模組

- **RenderTemplate**：根據模板渲染通知內容。
- **ManageTemplates**：新增、更新、刪除通知模板。

### Label 模組

- **CreateLabel**：創建新的標籤。
- **UpdateLabel**：更新現有的標籤。
- **DeleteLabel**：刪除標籤。
- **GetAllLabels**：獲取所有標籤。
- **GetLabelsByKey**：根據鍵值獲取標籤。

## 設定詳細說明

### 配置文件

- **metrics.yaml**：定義需監控的指標及其屬性。
- **alert_rules.yaml**：定義各種告警規則，包括檢測類型、閾值、持續時間等。
- **notify_config.yaml**：定義通知管道及其配置參數。
- **labels.yaml**：定義系統內使用的標籤結構和屬性。

### 環境變數

- **DATABASE_URL**：資料庫連接字符串。
- **INFLUXDB_URL**：InfluxDB 連接地址。
- **SMTP_HOST**、**SMTP_PORT**、**SMTP_USER**、**SMTP_PASS**：SMTP 服務器配置，用於 Email 通知。
- **WEBHOOK_URL**：Webhook URL，用於 Teams、Slack 等通知。
- **LOG_LEVEL**：日誌級別配置。

## 部署與運行

1. **配置環境**：
   - 根據需要修改 `configs/*.yaml` 配置文件，設置監控指標、告警規則、通知管道等。
   - 設置必要的環境變數，如資料庫連接、通知服務配置等。

2. **啟動服務**：
   - 使用以下命令啟動後台服務：
     ```bash
     go run main.go &
     ```
   - 這將啟動監控數據解析、告警檢測與通知發送的流程。

3. **持續運行與監控**：
   - 確保服務持續運行，並使用 **Logger** 模組監控日誌，以便及時發現和解決問題。

## 擴展與維護

1. **新增數據源解析器**：
   - 在 `share/parser/` 目錄下新增對應的解析器模組。
   - 更新 `parser.go` 以註冊新的解析器。

2. **新增通知管道**：
   - 在 `share/notify/channels/` 目錄下新增通知器模組（如新的 Webhook 類型）。
   - 更新 `factory.go` 和 `types.go` 以支持新的通知類型。

3. **新增告警類型**：
   - 在 `share/alert/` 模組中新增新的檢查器和告警規則。
   - 更新 `alert_rules.yaml` 配置文件以定義新的告警規則。

4. **優化模板系統**：
   - 在 `share/template/` 目錄下新增或修改模板文件。
   - 根據需要調整模板渲染邏輯，確保通知內容符合需求。

## 常見問題

1. **通知無法發送**：
   - 檢查 `notify_config.yaml` 中的通知管道配置是否正確，如 SMTP 服務器信息、Webhook URL 是否有效。
   - 查看 **Logger** 日誌，確認發送過程中是否有錯誤信息。

2. **告警規則未觸發**：
   - 確認 `alert_rules.yaml` 中的告警規則是否配置正確，閾值是否合適。
   - 檢查監控數據是否已正確存入資料庫，並符合告警條件。

3. **解析器無法讀取數據源**：
   - 確認 `metrics.yaml` 中的數據源配置是否正確，路徑和文件格式是否正確。
   - 查看 **Logger** 日誌，確認解析過程中是否有錯誤信息。


---

## 模組詳細介紹

### 1. Parser 解析器模組

**目錄：** `share/parser/`

**功能：**
- 從各種數據源（如 NMON、LOGMAN、Oracle 等）收集原始監控數據。
- 解析數據並轉換為結構化格式，存入資料庫。
- 提供統一的解析接口，方便擴展新的數據源解析器。

**主要文件：**
- `parser.go`：統一的解析接口和主要的解析流程。
- 各數據源的解析器（如 `nmon/parser.go`, `logman/parser.go` 等）。
- `common/constants.go`：定義解析相關的常量。

**使用說明：**
1. **配置數據源**：在 `configs/metrics.yaml` 中定義需要監控的指標及其數據源配置。
2. **啟動解析器**：運行 `RunParser` 函數，開始定時從數據源讀取並解析數據。

### 2. Alert 告警模組

**目錄：** `share/alert/`

**功能：**
- 基於解析後的監控數據，應用配置的告警規則進行異常檢測。
- 支援絕對值和變化幅度兩種告警檢測類型。
- 管理告警狀態，支援靜默期設置和重試機制。
- 提供多租戶支持，確保數據隔離和安全。

**主要文件：**
- `check.go`：告警檢測邏輯實現。
- `alert_rule.go`：告警規則的 CRUD 操作。
- `mute/mute.go`：靜默期間的管理和執行。
- `factory.go`：檢查器工廠，用於創建不同類型的檢查器。

**使用說明：**
1. **配置告警規則**：在 `configs/alert_rules.yaml` 中定義各種告警規則。
2. **啟動告警服務**：確保告警模組已經啟動，並且與資料庫連接正常。

### 3. Notify 通知模組

**目錄：** `share/notify/`

**功能：**
- 根據告警模組生成的告警信息，選擇合適的通知管道（如 Email、Teams、Webhook 等）發送通知。
- 使用模板生成格式化的通知內容，確保通知的一致性和可讀性。
- 提供通知重試機制，提升通知的可靠性。
- 統一管理通知配置，方便維護和擴展新的通知類型。

**主要文件：**
- `service.go`：主要的通知服務實現。
- `sender.go`：根據通知類型選擇並創建對應的通知器。
- `channels/`：各種通知器的實現（如 `email.go`, `teams.go`, `webhook.go` 等）。
- `factory/`：通知器工廠，用於註冊和創建通知器。
- `validate/`：通知配置的驗證邏輯。
- `config/`：通知模組的配置處理。

**使用說明：**
1. **配置通知管道**：在 `configs/notify_config.yaml` 中定義各種通知類型及其參數。
2. **啟動通知服務**：通過 API 或自動化流程觸發通知發送。

### 4. Template 模板管理模組

**目錄：** `share/template/`

**主要功能：**
- 提供統一的模板系統，用於生成不同通知管道的通知內容。
- 支援多種模板格式（如 HTML、Markdown），滿足不同通知需求。
- 管理和維護通知模板，支持動態更新和擴展。

**主要文件：**
- `template.go`：模板渲染的主要邏輯。
- 各種模板文件（如 `email.html`, `teams.md` 等）。

**使用說明：**
1. **創建模板**：在 `share/template/` 目錄下新增或修改模板文件。
2. **渲染模板**：使用 **Notify** 模組調用模板渲染函數，生成最終的通知內容。

### 5. Label 標籤管理模組

**目錄：** `share/label/`

**主要功能：**
- 管理系統中的各類標籤，支援標籤的 CRUD 操作。
- 支援多租戶（realm）環境，確保標籤的隔離與安全。
- 標籤值支援一對多關係，方便靈活的標籤分類。
- 提供標籤值的檢索和分類功能，提升告警的精確性。

**主要文件：**
- `service.go`：標籤管理的主要服務實現。
- `router.go`：標籤相關的 API 路由定義。
- `README.md`：標籤模組的使用說明和功能介紹。

**使用說明：**
1. **管理標籤**：通過 API 接口新增、更新、刪除標籤。
2. **應用標籤**：在告警配置中使用標籤進行告警分類和過濾。

## 配置文件說明

### 1. metrics.yaml

```yaml
metrics:
  cpu:
    usage:
      category: "cpu"
      metric_name: "cpu_usage"
      alias_name: "cpu_usage"
      unit: "%"
      partition_tag: "cpu_name"
    # ... 其他指標配置
```

**說明：**
- 定義需要監控的指標及其屬性。
- 每個指標包含類別、名稱、別名、單位和分區標籤。

### 2. alert_rules.yaml

```yaml
alert_rules:
  - id: 1
    name: "High CPU Usage"
    metric: "cpu_usage"
    check_type: "absolute"
    operator: ">"
    threshold: 80
    duration: 300
    severity: "warn"
    labels:
      - key: "environment"
        value: "production"
  # ... 其他告警規則
```

**說明：**
- 定義各種告警規則，包括檢測指標、閾值、持續時間、嚴重程度等。
- 可附加多個標籤，用於告警分類和過濾。

### 3. notify_config.yaml

```yaml
notify:
  email:
    host: "smtp.example.com"
    port: "587"
    username: "user@example.com"
    password: "password"
    from: "alert@example.com"
    to:
      - "admin@example.com"
    cc: []
    bcc: []
    reply_to: "noreply@example.com"
    use_tls: true
    use_auth: true
  teams:
    webhook_url: "https://outlook.office.com/webhook/..."
  webhook:
    url: "https://customwebhook.example.com/notify"
    method: "POST"
    token: "your_token_here"
  # ... 其他通知類型配置
```

**說明：**
- 定義各種通知類型及其具體參數。
- 支援多種通知管道，便於靈活配置和擴展。

### 4. labels.yaml

```yaml
labels:
  - id: 1
    realm_name: "production"
    key: "environment"
    value: "production"
  - id: 2
    realm_name: "production"
    key: "application"
    value: "frontend"
  # ... 其他標籤
```

**說明：**
- 定義系統內使用的標籤，包括標籤的鍵和值。
- 支援多租戶，確保標籤的隔離與安全。

## 日誌管理

### Logger 日誌模組

**目錄：** `share/logger/`

**功能：**
- 提供統一的日誌記錄接口，支援多種日誌級別（如 Info、Warn、Error）。
- 支援日誌輪轉和清理，確保日誌不佔用過多磁碟空間。
- 可配置的日誌級別和格式，滿足不同環境的需求。

**主要文件：**
- `manager.go`：日誌管理的主要實現。
- `cleaner.go`：日誌清理和輪轉的實現。
- `README.md`：日誌模組的使用說明和功能介紹。

**使用說明：**
1. **初始化日誌**：
   ```go
   import "shared-lib/logger"

   func main() {
       logger.InitLogger()
       defer logger.Sync()
       // 其他初始化代碼
   }
   ```
2. **記錄日誌**：
   ```go
   import "go.uber.org/zap"

   var logger = zap.L()

   func someFunction() {
       logger.Info("This is an info message")
       logger.Error("This is an error message")
   }
   ```

### 日誌輪轉與清理

**功能：**
- 定時壓縮和清理舊的日誌文件，避免磁碟空間被佔滿。
- 配置文件 `configs/logger_config.yaml` 定義壓縮和清理的策略。

**主要文件：**
- `cleaner.go`：實現日誌輪轉和清理的邏輯。
- `scheduler.go`：設定定時任務來執行日誌圧縮和清理。

**使用說明：**
1. **配置日誌輪轉策略**：在 `configs/logger_config.yaml` 中設置壓縮和清理的條件，如天數和大小。
2. **啟動日誌清理任務**：
   ```go
   import "shared-lib/logger"

   func main() {
       logger.StartLogScheduler()
       // 其他初始化代碼
   }
   ```

## 資料庫操作模組

### Databases 資料庫模組

**目錄：** `share/databases/`

**功能：**
- 封裝對資料庫的操作，提供通用的 CRUD 接口。
- 支援多種資料庫類型（如 MySQL、InfluxDB）。
- 提供事務處理，保證操作的原子性和一致性。

**主要文件：**
- `mysql.go`：MySQL 資料庫操作的實現。
- `database.go`：定義統一的資料庫接口。
- `migrations/`：資料庫遷移腳本。

**使用說明：**
1. **初始化資料庫**：
   ```go
   import "shared-lib/databases"

   func main() {
       db := databases.NewMySQL(yourGormDBInstance)
       // 其他初始化代碼
   }
   ```
2. **執行 CRUD 操作**：
   ```go
   metricRule, err := db.GetMetricRule(1)
   if err != nil {
       // 處理錯誤
   }
   ```

### MySQL 資料庫操作

**主要文件：** `share/databases/mysql.go`

**功能：**
- 提供 MySQL 資料庫的具體操作實現，如獲取指標規則、告警詳情、聯絡人等。
- 支援事務操作，確保數據的一致性。

**使用說明：**
1. **獲取自定義標籤**：
   ```go
   labels, err := db.GetCustomLabels(ruleID)
   if err != nil {
       // 處理錯誤
   }
   ```
2. **獲取告警聯絡人**：
   ```go
   contacts, err := db.GetAlertContacts(ruleID)
   if err != nil {
       // 處理錯誤
   }
   ```

## 模型定義

### Models 資料模型模組

**目錄：** `share/models/`

**功能：**
- 定義所有資料結構，包括資料庫模型和 API 請求/響應結構。
- 提供資料驗證方法，確保數據的有效性和一致性。

**主要文件：**
- `contact.go`：聯絡人相關的模型定義。
- `event_log.go`：事件日誌相關的模型定義。
- `alert_state.go`：告警狀態相關的模型定義。
- `check_rule.go`：檢查規則相關的模型定義。
- `alert_codes.go`：告警代碼相關的模型定義。

**使用說明：**
1. **定義資料模型**：
   ```go
   type AlertContact struct {
       ID        int64  `json:"id" gorm:"primaryKey"`
       Name      string `json:"name"`
       Type      string `json:"type"`
       Target    string `json:"target"`
       Severity  string `json:"severity"`
       Status    string `json:"status"`
       CreatedAt time.Time `json:"created_at"`
       UpdatedAt time.Time `json:"updated_at"`
   }
   ```
2. **進行資料操作**：
   ```go
   contact := models.AlertContact{
       Name:     "Admin",
       Type:     "email",
       Target:   "admin@example.com",
       Severity: "crit",
       Status:   "enabled",
   }
   err := db.Create(&contact).Error
   if err != nil {
       // 處理錯誤
   }
   ```

## 常用操作示例

### 添加新的告警規則

1. **定義告警規則**：
   在 `configs/alert_rules.yaml` 中新增新的告警規則。
   ```yaml
   - id: 2
     name: "Disk Space Low"
     metric: "disk_free_space"
     check_type: "absolute"
     operator: "<"
     threshold: 20
     duration: 600
     severity: "crit"
     labels:
       - key: "environment"
         value: "production"
   ```
2. **重啟服務**：
   使新的告警規則生效。
   ```bash
   go run main.go &
   ```

### 發送測試通知

1. **使用 API 發送測試通知**：
   發送一條測試通知，檢查通知管道是否配置正確。
   ```bash
   curl -X POST http://localhost:8700/api/v1/notify/process
   ```
2. **檢查通知狀態**：
   查看 **Logger** 模組的日誌，確認通知是否成功發送。

### 管理標籤

1. **創建新的標籤**：
   透過 API 創建新的標籤。
   ```bash
   curl -X POST http://localhost:8700/api/v1/alert/alert-label/create \
        -H "Content-Type: application/json" \
        -d '[{"key":"priority","value":"high"}]'
   ```
2. **更新標籤**：
   更新已有標籤的值。
   ```bash
   curl -X PUT http://localhost:8700/api/v1/alert/alert-label/update/priority \
        -H "Content-Type: application/json" \
        -d '[{"value":"medium"},{"value":"low"}]'
   ```
3. **刪除標籤**：
   刪除指定的標籤。
   ```bash
   curl -X DELETE http://localhost:8700/api/v1/alert/alert-label/delete/priority
   ```

## 測試與驗證

1. **單元測試**：
   在各個模組中撰寫單元測試，確保功能的正確性。
   ```bash
   go test ./share/alert/...
   go test ./share/notify/...
   ```
2. **集成測試**：
   測試各模組之間的協同工作，確保數據流和功能的完整性。

## 部署注意事項

1. **環境配置**：
   確保所有配置文件和環境變數正確設置，包括資料庫連接、通知管道配置等。

2. **資料庫遷移**：
   在首次部署或更新版本時，運行資料庫遷移腳本。
   ```bash
   go run share/databases/migrate.go
   ```

3. **日誌管理**：
   配置 **Logger** 模組的日誌存儲路徑和輪轉策略，確保日誌的持久性和可用性。

4. **安全設置**：
   確保所有 API 端點都有適當的認證和授權機制，保護系統不受未授權訪問。

## 參考資源

- [Gin 官方文檔](https://gin-gonic.com/docs/)
- [GORM 官方文檔](https://gorm.io/docs/)
- [Zap 日誌庫](https://pkg.go.dev/go.uber.org/zap)
- [Cron 調度庫](https://github.com/robfig/cron)

---

感謝您使用**iPOC Golang API**！如果有任何疑問或建議，請隨時聯繫我們的技術支持團隊。

---

## 初始化流程

### 1. 配置初始化
首先必須通過 config 模組初始化全局配置：

```go
import (
    "shared-lib/config"
    "shared-lib/logger"
    "shared-lib/notify"
)

func main() {
    // 1. 初始化配置
    if err := config.InitConfig("/etc/your-service/config.yml"); err != nil {
        panic(err)
    }

    // 2. 初始化日誌
    logConfig := config.GetLoggerConfig()
    log, err := logger.NewLogger(&logConfig)
    if err != nil {
        panic(err)
    }

    // 3. 初始化通知服務
    notifyConfig := config.GetNotifyConfig()
    notifyService := notify.NewService(db, &logConfig)
    
    // 4. 初始化通知路徑
    notify.InitPaths("/var/log/your-service")
    if err := notify.InitNotifyDirs(log); err != nil {
        log.Fatal("初始化通知目錄失敗", zap.Error(err))
    }

    // 5. 其他服務初始化...
}
```

### 2. 配置文件結構
配置文件 (`config.yml`) 應包含所有需要的模組配置：

```yaml
# 日誌配置
logger:
  level: info
  file:
    path: /var/log/your-service
    max_size: 100
    max_age: 7
    max_backups: 5

# 通知配置
notify:
  max_retry: 3
  retry_interval: 300
  channels:
    email:
      host: smtp.example.com
      port: 587
    teams:
      webhook_url: https://...

# 資料庫配置
database:
  mysql:
    host: localhost
    port: 3306
    
# 服務自己的配置
your_service:
  custom_setting: value
```

### 3. 初始化順序
正確的初始化順序很重要：

1. **Config 模組**：
   - 最先初始化，因為其他模組都依賴配置
   - 讀取並解析配置文件
   - 提供配置訪問接口

2. **Logger 模組**：
   - 第二個初始化，因為後續模組都需要日誌功能
   - 使用配置模組提供的日誌配置
   - 初始化日誌實例

3. **Database 模組**：
   - 在日誌之後初始化
   - 使用配置模組提供的資料庫配置
   - 建立資料庫連接

4. **其他核心模組**：
   - Notify、Alert 等模組
   - 使用已初始化的日誌和資料庫實例
   - 根據配置初始化各自的功能

### 4. 錯誤處理
- 初始化過程中的錯誤應立即處理
- 關鍵模組初始化失敗應該中止程序
- 使用日誌記錄詳細的錯誤信息

### 5. 優雅關閉
確保在程序結束時正確關閉各個模組：

```go
func main() {
    // ... 初始化程式 ...

    // 設置優雅關閉
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    // 等待關閉信號
    <-c

    // 按照依賴關係的相反順序關閉服務
    notifyService.Close()
    db.Close()
    log.Sync()
}
```
