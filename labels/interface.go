package labels

import (
	"github.com/detect-viz/shared-lib/models"
)

type Service interface {
	// 基本 CRUD
	Create(realm string, label *models.Label) (*models.Label, error)
	Get(realm, key string) (*models.Label, error)
	List(realm string, limit, offset int) ([]models.Label, error)
	Update(realm, key string, updates map[string]interface{}) error
	UpdateKey(realm, oldKey, newKey string) error
	Delete(realm, key string) error
	Exists(realm, key string) (bool, error)

	// 進階功能
	GetKeyOptions(realm, key string) ([]models.OptionResponse, error)
	BulkCreateOrUpdate(realm string, labels []models.Label) ([]models.Label, error)
}
