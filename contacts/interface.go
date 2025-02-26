package contacts

import (
	"github.com/detect-viz/shared-lib/models"
)

// 將接口重命名為 ContactService
type Service interface {
	// 基本 CRUD
	Create(contact *models.Contact) error
	Get(id int64) (*models.Contact, error)
	List(realm string) ([]models.Contact, error)
	Update(contact *models.Contact) error
	Delete(id int64) error
	CheckName(contact models.Contact) (bool, string)
	IsUsedByRules(id int64) (bool, error)
	GetContactsByRuleID(ruleID int64) ([]models.Contact, error)
	NotifyTest(contact models.Contact) models.Response
}
