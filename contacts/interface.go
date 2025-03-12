package contacts

import (
	"context"

	"github.com/detect-viz/shared-lib/auth/keycloak"
	"github.com/detect-viz/shared-lib/models"
)

type Service interface {
	SetContactChangeCallback(callback ContactChangeCallback)
	Create(contact *models.Contact) (*models.Contact, error)
	Get(id []byte) (*models.Contact, error)
	List(realm string, cursor int64, limit int) ([]models.Contact, int64, error)
	Update(contact *models.Contact) (*models.Contact, error)
	Delete(id []byte) error
	IsUsedByRules(id []byte) (bool, error)
	GetContactsByRuleID(ruleID []byte) ([]models.Contact, error)
	NotifyTest(contact models.Contact) error
	GetConfig(ctx context.Context, notifyType string) (map[string]string, error)
	GetAllConfigs(ctx context.Context) (map[string]map[string]string, error)
	GetNotifyMethods() []string
	GetNotifyOptions(ctx context.Context, realm string) (map[string]map[string][]string, error)
	SetKeycloakClient(client keycloak.KeycloakClient) error
	ToResponse(contact models.Contact) models.ContactResponse
	FromResponse(resp models.ContactResponse) models.Contact
}
