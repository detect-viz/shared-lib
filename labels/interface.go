package labels

import (
	"mime/multipart"

	"github.com/detect-viz/shared-lib/models"
)

type Service interface {
	// 基本 CRUD
	Create(realm string, label *models.LabelDTO) (*models.LabelDTO, error)
	Get(id int64) (*models.LabelDTO, error)
	List(realm string, cursor int64, limit int) ([]models.LabelDTO, int64, error)
	Update(realm string, label *models.LabelDTO) (*models.LabelDTO, error)
	UpdateKeyName(realm string, oldKey, newKey string) (*models.LabelDTO, error)
	Delete(id int64) error

	// 進階功能
	GetKeyOptions(realm string) ([]models.OptionResponse, error)
	ExportCSV(realm string) ([][]string, error)
	ImportCSV(realm string, file *multipart.FileHeader) error
}
