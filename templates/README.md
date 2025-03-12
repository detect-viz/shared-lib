# Template Service 模板服務

此模組負責處理告警通知的模板管理和內容渲染。

## 模組架構

```
template/
├── service.go         # 模板服務實現
├── interfaces.go      # 模板介面定義
└── models/           # 模板相關模型
    └── template.go   # 模板資料結構
```

## 核心功能

### 1. 模板管理

- 支援多種格式模板：
  - HTML (Email)
  - Markdown (Slack/Teams/Discord)
  - JSON (Webhook)
  - Text (預設)
- 模板來源：
  - 內建預設模板 (YAML)
  - 資料庫儲存模板
  - 動態生成模板

### 2. 格式轉換

- 自動匹配通知類型：
  | 通知類型 | 格式 | 使用場景 |
  |---------|------|---------|
  | email | HTML | 郵件通知 |
  | slack/teams/discord | Markdown | 即時通訊 |
  | webhook | JSON | API 回調 |
  | 其他 | Text | 預設格式 |

### 3. 模板渲染

- 使用 Go template 引擎
- 支援的變數：
  ```go
  data := map[string]interface{}{
    "timestamp":     "2024-03-21 15:04:05",
    "resource_name": "server-01",
    "rule_name":     "CPU Usage High",
    "severity":      "critical",
    "current_value": 95.5,
    "threshold":     90,
    "unit":         "%",
    "labels":       map[string]string{...},
  }
  ```

## 使用方式

### 1. 預設模板配置

```yaml
templates:
  - name: "email-alert"
    format: "html"
    rule_state: "alerting"
    title: "系統告警: {{.resource_name}}"
    message: |
      <h2>告警通知</h2>
      <p>時間: {{.timestamp}}</p>
      <p>資源: {{.resource_name}}</p>
      <p>規則: {{.rule_name}}</p>
      <p>當前值: {{.current_value}}{{.unit}}</p>

  - name: "slack-alert"
    format: "markdown"
    rule_state: "alerting"
    title: "🚨 系統告警"
    message: |
      **告警通知**
      - 時間: {{.timestamp}}
      - 資源: {{.resource_name}}
      - 規則: {{.rule_name}}
      - 當前值: {{.current_value}}{{.unit}}
```

### 2. 程式調用

```go
// 初始化模板服務
templateSvc := template.NewService(logger)

// 獲取適用的模板
tmpl, err := templateSvc.GetTemplate(realm, "alerting", "markdown")

// 渲染模板內容
message, err := templateSvc.RenderMessage(tmpl, data)
```

## 注意事項

1. 模板更新

   - 系統啟動時載入預設模板
   - 資料庫模板優先於預設模板
   - 避免重複載入相同模板

2. 格式轉換

   - 確保模板格式符合通知類型
   - 處理特殊字符轉義
   - 驗證模板語法正確性

3. 效能考量
   - 模板快取機制
   - 避免重複解析模板
   - 大量通知時的效能優化

## 未來規劃

1. UI 管理介面

   - 模板列表查看
   - 在線編輯模板
   - 模板測試功能

2. 進階功能

   - 模板版本控制
   - 模板權限管理
   - 模板審核流程

3. 效能優化
   - 模板預編譯
   - 分散式快取
   - 批次處理機制

## **通知模板管理方式**

1. **手動設定通知模板**，存放在 `default-templates.yaml`，Alert Service 啟動時會自動載入。

📌 這個 YAML 存放在 `/etc/${APP_NAME}/provisioning/notifiers/default-templates.yaml`。  
📌 在 **重啟時** 讀取這些模板。

```yaml
template_path: /etc/${APP_NAME}/provisioning/notifiers/
```

✅ **適合預設模板，但不支援 UI 編輯**。  
✅ **適合靜態設定，不適合經常變更的通知格式**。

## **1️⃣ 在 UI / API 管理 (儲存在 DB)**

在 **Grafana UI** 設定 `Contact Points` 時，可以直接修改 `Message`：

- 支援 `Go Template`
- 可用變數 `{{ .alertName }}`, `{{ .labels }}`, `{{ .values }}` 來動態插入數據。

### **📌 Contact Point 設定範例**

```json
{
  "name": "Slack Notification",
  "type": "slack",
  "settings": {
    "url": "https://hooks.slack.com/services/XXXX",
    "message": "🚨 **Alert:** {{ .alertName }}\n🔹 *Status:* {{ .status }}\n🔹 *Value:* {{ .values }}"
  }
}
```

```bash
curl -X PATCH http://localhost:3000/api/v1/provisioning/contact-points/1 \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Slack Notification",
    "type": "slack",
    "settings": {
      "message": "🚨 Alert: {{ .alertName }} - Status: {{ .status }}"
    }
  }'
```

### **📌 3. API 介面 (類似 Grafana `PATCH /contact-points`)**

**`PATCH /templates/{id}`**

```json
{
  "name": "Custom Slack Template",
  "format": "markdown",
  "message": "**[告警]** {{ .alertName }}\n🔹 *狀態:* {{ .status }}\n🔹 *數值:* {{ .values }}"
}
```

📌 這與 **Grafana `PATCH /contact-points`** 相同，可讓 UI 修改模板。

### **1️⃣ 啟動時讀取 `default-templates.yaml`**

**步驟**

1.  準備 `default-templates.yaml` 內建靜態模板
2.  `iPOC` 啟動時讀取 YAML，並載入 DB
3.  若 DB 已有相同名稱的模板，則不覆蓋（避免重複載入）

### **2️⃣ Template 模組應獨立負責格式轉換**

📌 **Template 模組應負責：**

1.  **解析 AlertTemplate**（title & message）
2.  **根據 contactType 選擇 format**
3.  **動態渲染 TriggerLog 為不同格式**
4.  **支持 HTML, Markdown, JSON, Text 等格式輸出**

### **4️⃣ AlertTemplate 自動適配 format**

## **🚀 Template 模組如何動態適配不同格式**

| **通知類型 (type)**                        | **匹配格式 (format)** | **範本引擎**    |
| ------------------------------------------ | --------------------- | --------------- |
| **email**                                  | html                  | **HTML 模板**   |
| **slack / discord / teams / webex / line** | markdown              | **Markdown**    |
| **webhook**                                | json                  | **JSON 格式化** |
| **其他**                                   | text                  | **純文字**      |

## **🚀 總結**

✔ **讓 TemplateService 負責通知內容格式轉換，保持 alert 乾淨**  
✔ **使用 text/template 來靈活渲染 AlertTemplate**  
✔ **不同 type 可自動匹配 format，減少硬編碼邏輯**  
✔ **通知格式變更時，只需修改 AlertTemplate，不需改動 Go 代碼**

✅ **這樣 `iPOC` 啟動時，會先確認 DB 是否有內建模板，若無則載入 YAML**。

---

欄位
// 準備通知內容
data := map[string]interface{}{
"realm_name": triggers[0].RealmName,
"timestamp": time.Unix(now, 0).Format("2006-01-02 15:04:05"),
"notify_uuid": triggers[0].UUID,
"rule_id": triggers[0].RuleID,
"rule_name": triggers[0].RuleName,
"resource_group": triggers[0].ResourceGroupName,
"resource_name": triggers[0].ResourceName,
"partition_name": triggers[0].PartitionName,
"metric_name": triggers[0].MetricName,
"unit": triggers[0].Unit,
"trigger_value": triggers[0].TriggerValue,
"threshold": triggers[0].Threshold,
"severity": triggers[0].Severity,
"first_trigger_time": triggers[0].FirstTriggerTime,
"silence_start": triggers[0].SilenceStart,
"silence_end": triggers[0].SilenceEnd,
"mute_start": triggers[0].MuteStart,
"mute_end": triggers[0].MuteEnd,
"contact_state": triggers[0].ContactState,
"notify_state": triggers[0].NotifyState,
"resolved_notify_state": triggers[0].ResolvedNotifyState,
"resolved_time": triggers[0].ResolvedTime,
"labels": triggers[0].Labels,
"contacts": triggers[0].Contacts,
}

// 複製 TriggerLog
for i := range triggers {
notify.TriggerLogs[i] = &triggers[i]
}

## **4️⃣ 未來開放 UI 讓用戶管理模板**

當 YAML 內建模板載入 DB 後，下一步可以開放 UI 讓用戶管理模板。

✅ **這樣 `NotifyLog` 會根據 `contactType` 自動匹配 `format`，確保通知內容符合適用格式。**

---

### **4️⃣ `AlertTemplate` 自動適配 `format`**

如果 `AlertTemplate` 需要基於 `type` 設定 `format`，可以在查詢時自動匹配：

```go
// GetAlertTemplate 根據 `contactType` 獲取適用的模板
func (s *Service) GetAlertTemplate(contactType string) models.AlertTemplate {
	format := GetFormatByType(contactType) // 🔹 獲取適用格式
	var template models.AlertTemplate

	err := s.db.Where("format = ?", format).
		Where("is_default = ?", true).
		First(&template).Error

	if err != nil {
		s.logger.Warn("找不到適用的通知模板，將使用 text 格式",
			zap.String("contact_type", contactType),
			zap.String("format", format))
		template.Format = "text"
		template.Title = "【告警通知】"
		template.Message = "未找到對應的模板，請聯繫管理員"
	}

	return template
}
```

✅ **確保 `AlertTemplate` 會根據 `type` 自動匹配 `format`，不會出現不適用的模板格式。**

---

### **🔍 總結**

| **通知類型 (type)**                                        | **匹配格式 (format)** | **適用場景**   |
| ---------------------------------------------------------- | --------------------- | -------------- |
| **email**                                                  | `html`                | 郵件通知       |
| **slack** / **discord** / **teams** / **webex** / **line** | `markdown`            | IM 聊天工具    |
| **webhook**                                                | `json`                | API 回調       |
| **其他**                                                   | `text`                | 預設純文字格式 |

✅ **所有 `type` 都可以自動匹配 `format`，不需要手動設定**  
✅ **`NotifyLog` 與 `AlertTemplate` 都能根據 `contactType` 自動適配格式**  
✅ **iPOC `alert` 可以兼容不同的通知管道，確保發送格式正確！🚀**

## Response:

格式轉換有更靈活套件處理嗎？ 功能應該獨立在 Template 模組

## Prompt:

### **🔍 更靈活的 `format` 轉換與 `Template` 模組處理**

---

### **1️⃣ 使用 `text/template` 或 `goTemplate` 進行格式轉換**

目前 `format` 的內容生成是手寫 `string.Builder`，可以改用 **`text/template`** 或 **`goTemplate`** 來動態渲染不同格式的通知內容。

---

### **2️⃣ `Template` 模組應獨立負責格式轉換**

📌 **`Template` 模組應負責：**

1.  **解析 `AlertTemplate`**（`title` & `message`）
2.  **根據 `contactType` 選擇 `format`**
3.  **動態渲染 `TriggerLog` 為不同格式**
4.  **支持 `HTML`, `Markdown`, `JSON`, `Text` 等格式輸出**

---

## **🚀 方案：用 `text/template` 讓格式轉換更靈活**

### **📌 3️⃣ `template_service.go`**

```go
package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"shared-lib/models"
	"go.uber.org/zap"
)

// TemplateService 負責處理 `format` 轉換
type TemplateService struct {
	logger *zap.Logger
}

// NewTemplateService 創建模板服務
func NewTemplateService(logger *zap.Logger) *TemplateService {
	return &TemplateService{logger: logger}
}

// GetFormatByType 根據通知類型獲取 `format`
func GetFormatByType(contactType string) string {
	switch contactType {
	case "email":
		return "html"
	case "slack", "discord", "teams", "webex", "line":
		return "markdown"
	case "webhook":
		return "json"
	default:
		return "text"
	}
}

// RenderMessage 根據 `format` 渲染模板
func (s *TemplateService) RenderMessage(template models.AlertTemplate, data map[string]interface{}) (string, error) {
	// `message` 是存儲在 `AlertTemplate` 內的模板
	tmpl, err := template.New("alert").Parse(template.Message)
	if err != nil {
		s.logger.Error("模板解析失敗", zap.Error(err))
		return "", err
	}

	var result bytes.Buffer
	err = tmpl.Execute(&result, data)
	if err != nil {
		s.logger.Error("模板執行失敗", zap.Error(err))
		return "", err
	}

	return result.String(), nil
}

// RenderJSON 格式化為 JSON
func (s *TemplateService) RenderJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON 轉換失敗: %v", err)
	}
	return string(jsonBytes), nil
}
```

✅ **使用 `text/template` 解析 `AlertTemplate`**  
✅ **動態選擇 `format`，支援 `HTML`, `Markdown`, `Text`, `JSON`**  
✅ **若 `format = "json"`，可直接轉 JSON**

---

### **📌 4️⃣ `generateNotifyLog()` 更新，使用 `TemplateService`**

```go
func (s *Service) generateNotifyLog(key string, triggers []models.TriggerLog) models.NotifyLog {
	now := time.Now().Unix()

	// 解析聯絡人資訊
	parts := strings.Split(key, "_")
	contactID, _ := strconv.ParseInt(parts[0], 10, 64)
	contactType := parts[1]
	notifyFormat := GetFormatByType(contactType) // 🔹 自動匹配 format

	// 取得模板
	alertTemplate := s.templateService.GetAlertTemplate(contactType)

	// 準備通知內容
	data := map[string]interface{}{
		"timestamp":      time.Unix(now, 0).Format("2006-01-02 15:04:05"),
		"resource_name":  triggers[0].ResourceName,
		"rule_name":      triggers[0].RuleName,
		"severity":       triggers[0].Severity,
		"current_value":  triggers[0].TriggerValue,
		"threshold":      triggers[0].Threshold,
		"unit":           triggers[0].Unit,
		"labels":         triggers[0].Labels,
	}

	// 渲染通知內容
	message, err := s.templateService.RenderMessage(alertTemplate, data)
	if err != nil {
		s.logger.Error("渲染通知內容失敗", zap.Error(err))
		message = "告警通知發生錯誤，請聯繫管理員"
	}

	notify := models.NotifyLog{
		UUID:         uuid.New().String(),
		Timestamp:    now,
		ContactID:    contactID,
		ContactName:  triggers[0].Contacts[0].Name,
		ContactType:  contactType,
		Format:       notifyFormat,
		Title:        alertTemplate.Title,
		Message:      message,
		RetryCounter: 0,
		TriggerLogs:  make([]*models.TriggerLog, len(triggers)),
	}

	// 複製 TriggerLog 指針
	for i := range triggers {
		notify.TriggerLogs[i] = &triggers[i]
	}

	return notify
}
```

✅ **`generateNotifyLog()` 透過 `TemplateService` 來渲染不同 `format`**  
✅ **通知內容會根據 `AlertTemplate` 自動轉換格式**

---

## **🚀 `Template` 模組如何動態適配不同格式**

| **通知類型 (type)**                        | **匹配格式 (format)** | **範本引擎**    |
| ------------------------------------------ | --------------------- | --------------- |
| **email**                                  | `html`                | **HTML 模板**   |
| **slack / discord / teams / webex / line** | `markdown`            | **Markdown**    |
| **webhook**                                | `json`                | **JSON 格式化** |
| **其他**                                   | `text`                | **純文字**      |

### **📌 5️⃣ `AlertTemplate` 設定範例**

```yaml
templates:
  - name: "Email 通知"
    format: "html"
    rule_state: "alerting"
    title: "【系統告警】{{.resource_name}}"
    message: |
      <html>
      <body>
        <h2>告警通知</h2>
        <p>觸發時間: {{.timestamp}}</p>
        <p>設備: {{.resource_name}}</p>
        <p>規則: {{.rule_name}}</p>
        <p>當前數值: {{.current_value}} {{.unit}}</p>
        <p>設定閾值: {{.threshold}} {{.unit}}</p>
      </body>
      </html>

  - name: "Slack 通知"
    format: "markdown"
    rule_state: "alerting"
    title: "**[系統告警] {{.resource_name}}**"
    message: |
      - **觸發時間**: {{.timestamp}}
      - **設備**: {{.resource_name}}
      - **規則**: {{.rule_name}}
      - **當前數值**: {{.current_value}} {{.unit}}
      - **設定閾值**: {{.threshold}} {{.unit}}

  - name: "Webhook JSON"
    format: "json"
    rule_state: "alerting"
    title: "{{.rule_name}}"
    message: |
      {
        "timestamp": "{{.timestamp}}",
        "resource_name": "{{.resource_name}}",
        "rule_name": "{{.rule_name}}",
        "severity": "{{.severity}}",
        "current_value": "{{.current_value}}",
        "threshold": "{{.threshold}}",
        "unit": "{{.unit}}"
      }
```

✅ **不同 `format` 使用不同模板，不用在 `go` 代碼內手動拼接通知內容**  
✅ **支援 `HTML`, `Markdown`, `JSON`, `Text`，無需手寫 `if-else` 判斷格式**

---

## **🚀 總結**

✔ **讓 `TemplateService` 負責通知內容格式轉換，保持 `alert` 乾淨**  
✔ **使用 `text/template` 來靈活渲染 `AlertTemplate`**  
✔ **不同 `type` 可自動匹配 `format`，減少硬編碼邏輯**  
✔ **通知格式變更時，只需修改 `AlertTemplate`，不需改動 `Go` 代碼**

🎯 **這樣的架構讓 `alert` 模組更靈活、更容易擴展，確保 `iPOC` 能支援多種通知格式！🚀**

是的，以上規劃是正確且符合最佳實踐的，這樣的設計：

保持 Template struct 不變 → 讓資料庫與 API 結構維持一致，避免多個欄位存相同的數據
GetFormatByType(contactType string) 放在 notify.go → 由通知邏輯決定使用哪種格式
RenderMessage() 在 templates 模組中負責渲染 → templates 模組只需管理模板渲染邏輯
提供 template-default-alerting.yaml 和 template-default-resolved.yaml → 讓 alerting 和 resolved 各自有 4 種格式（HTML、Markdown、Text、JSON）
