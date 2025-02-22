package interfaces

import "github.com/detect-viz/shared-lib/models"

type TemplateService interface {
	RenderMessage(alertTemplate models.AlertTemplate, data map[string]interface{}) (string, error)
	RenderJSON(data interface{}) (string, error)
}
