package templates

import "github.com/detect-viz/shared-lib/models"

type Service interface {
	RenderMessage(template models.Template, data map[string]interface{}) (string, error)
}
