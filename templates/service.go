package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// RenderMessage 渲染通知模板
func (s *serviceImpl) RenderMessage(t models.Template, data map[string]interface{}) (string, error) {
	var messageTemplate string

	// 根據 FormatType 選擇對應的模板
	switch t.FormatType {
	case "html":
		messageTemplate = t.Message
	case "markdown":
		messageTemplate = t.Message
	case "text":
		messageTemplate = t.Message
	case "json":
		// JSON 需要額外解析成標準格式
		return renderJSONTemplate(t.Message, data)
	default:
		return "", fmt.Errorf("未知的模板格式: %s", t.FormatType)
	}

	// 渲染模板
	tmpl, err := template.New("message-" + t.FormatType).Parse(messageTemplate)
	if err != nil {
		return "", fmt.Errorf("解析模板失敗: %w", err)
	}

	var messageBuf bytes.Buffer
	if err := tmpl.Execute(&messageBuf, data); err != nil {
		return "", fmt.Errorf("渲染模板失敗: %w", err)
	}

	return messageBuf.String(), nil
}

func renderJSONTemplate(jsonTemplate string, data map[string]interface{}) (string, error) {
	// 渲染 JSON 模板
	tmpl, err := template.New("json").Parse(jsonTemplate)
	if err != nil {
		return "", fmt.Errorf("解析 JSON 模板失敗: %w", err)
	}

	var jsonBuf bytes.Buffer
	if err := tmpl.Execute(&jsonBuf, data); err != nil {
		return "", fmt.Errorf("渲染 JSON 內容失敗: %w", err)
	}

	// 解析渲染後的 JSON 來確保格式正確
	var formattedJSON map[string]interface{}
	if err := json.Unmarshal(jsonBuf.Bytes(), &formattedJSON); err != nil {
		return "", fmt.Errorf("渲染後的 JSON 格式錯誤: %w", err)
	}

	// 美化輸出的 JSON
	formattedOutput, err := json.MarshalIndent(formattedJSON, "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化 JSON 失敗: %w", err)
	}

	return string(formattedOutput), nil
}
