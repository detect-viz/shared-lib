package interfaces

import (
	"github.com/detect-viz/shared-lib/models"
)

// 標籤服務介面
type LabelService interface {
	CreateLabel(realm string, labels []models.Label) error
	GetAllLabels(realm string) []models.Label
	GetLabelsByKey(realm, key string) ([]models.OptionResponse, error)
	UpdateLabel(realm, key string, labels []models.Label) ([]models.Label, error)
	DeleteLabelByKey(realm, key string) models.Response
	CheckLabelKey(label models.Label) models.Response
}
