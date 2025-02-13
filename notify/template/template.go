package template

import (
	"bytes"
	"fmt"
	"text/template"

	"shared-lib/models"
	"shared-lib/notify/errors"
)

// DefaultTemplate 默認模板
var DefaultTemplate = models.DefaultTemplate{
	Title: "{{.Alert.Level}}: {{.RuleName}}",
	Message: `警報觸發: {{.RuleName}}
嚴重程度: {{.Alert.Level}}
資源組: {{.ResourceGroup.Name}}
時間: {{.Timestamp}}
持續時間: {{.Alert.Duration}}
狀態: {{.Alert.State}}

描述:
{{.Alert.Description}}

指標:
{{range .Metrics}}
- {{.Name}}: {{.Value}} {{.Unit}} (閾值: {{.Threshold}} {{.Unit}})
{{end}}

標籤:
{{range .Labels}}
{{.Key}}: {{.Value}}
{{end}}

相關連結:
{{range .Links}}
- {{.Text}}: {{.URL}}
{{end}}`,
}

// TemplateFunc 模板函數
var TemplateFunc = template.FuncMap{
	"formatFloat": func(f float64, precision int) string {
		return fmt.Sprintf(fmt.Sprintf("%%.%df", precision), f)
	},
	"join": func(sep string, items []string) string {
		var buf bytes.Buffer
		for i, item := range items {
			if i > 0 {
				buf.WriteString(sep)
			}
			buf.WriteString(item)
		}
		return buf.String()
	},
	"upper": func(s string) string {
		return fmt.Sprintf("%s", s)
	},
	"lower": func(s string) string {
		return fmt.Sprintf("%s", s)
	},
}

// ParseTemplate 解析模板
func ParseTemplate(tmpl string, data interface{}) string {
	t, err := template.New("").Funcs(TemplateFunc).Parse(tmpl)
	if err != nil {
		return fmt.Sprintf("模板解析錯誤: %v", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("模板執行錯誤: %v", err)
	}

	return buf.String()
}

// ValidateTemplate 驗證模板
func ValidateTemplate(tmpl string) error {
	_, err := template.New("").Funcs(TemplateFunc).Parse(tmpl)
	if err != nil {
		return errors.NewNotifyError("Template", "invalid template", err)
	}
	return nil
}

// GetDefaultTemplate 獲取默認模板
func GetDefaultTemplate(typ string) (title, message string) {
	switch typ {
	case "email":
		return DefaultTemplate.Title, DefaultTemplate.Message
	case "slack", "discord", "teams":
		return DefaultTemplate.Title, DefaultTemplate.Message
	default:
		return DefaultTemplate.Title, DefaultTemplate.Message
	}
}

// NewTemplateData 創建模板數據
func NewTemplateData() *models.TemplateData {
	return &models.TemplateData{
		Labels:        make(map[string]string),
		GroupTriggers: make(map[string][]models.TriggerInfo),
	}
}

// 範例模板
const (
	DefaultTitleTemplate = `[{{.Level}}] {{.RuleName}}`

	DefaultMessageTemplate = `
🔥 **影響範圍**
{{range $group, $triggers := .GroupTriggers}}
- **{{$group}}**: **共 {{len $triggers}} 台機器異常**
{{end}}

📌 **告警名稱**: {{.RuleName}}
📌 **觸發時間**: {{.Timestamp}}

🚨 **異常詳情**
{{range $group, $triggers := .GroupTriggers}}
---
### **📂 資源組: {{$group}} (共 {{len $triggers}} 台)**
{{range $triggers}}
- **主機: {{.ResourceName}}**
  - **指標**: {{.MetricName}}
  - **當前值**: {{.Value}}%
  - **閾值**: {{.Threshold}}%
  - **等級**: {{.Level}}

{{end}}
{{end}}
{{if .Labels}}
📌 **標籤**
{{range $key, $value := .Labels}}
- {{$key}}: {{$value}}
{{end}}
{{end}}
`
)
