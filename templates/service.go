package templates

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/models"
	"github.com/google/wire"
)

var TemplateSet = wire.NewSet(NewService, wire.Bind(new(Service), new(*serviceImpl)))

// Service 處理格式轉換
type serviceImpl struct {
	logger logger.Logger
}

// NewService 創建模板服務
func NewService(logger logger.Logger) *serviceImpl {
	return &serviceImpl{
		logger: logger,
	}
}

// 動態渲染模板
func (s *serviceImpl) RenderMessage(t models.Template, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("alert").Parse(t.Message)
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	err = tmpl.Execute(&result, data)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// RenderJSON 格式化為 JSON
func (s *serviceImpl) RenderJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

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
