package contacts

import (
	"context"

	"github.com/detect-viz/shared-lib/auth/keycloak"
	"github.com/detect-viz/shared-lib/models"
)

// ContactChangeCallback 聯絡人變更回調函數類型
type ContactChangeCallback func(contact models.Contact, operation string)

// Service 通知管道服務接口
type Service interface {
	// SetContactChangeCallback 設置聯絡人變更回調函數
	SetContactChangeCallback(callback ContactChangeCallback)

	// Create 創建聯絡人
	Create(realm string, contactResp *models.ContactResponse) (*models.ContactResponse, error)

	// Get 獲取聯絡人
	Get(id string) (*models.ContactResponse, error)

	// List 獲取聯絡人列表
	List(realm string, cursor int64, limit int) ([]models.ContactResponse, int64, error)

	// Update 更新聯絡人
	Update(realm string, contactResp *models.ContactResponse) (*models.ContactResponse, error)

	// Delete 刪除聯絡人
	Delete(id string) error

	// NotifyTest 測試聯絡人通知
	NotifyTest(realm string, contactResp *models.ContactResponse) error

	// GetContactsByRuleID 獲取規則的聯絡人列表
	GetContactsByRuleID(ruleID string) ([]models.ContactResponse, error)

	// IsUsedByRules 檢查聯絡人是否被規則使用
	IsUsedByRules(id string) (bool, error)

	// GetConfig 獲取通知配置
	GetConfig(ctx context.Context, notifyType string) (map[string]string, error)

	// GetAllConfigs 獲取所有通知配置
	GetAllConfigs(ctx context.Context) (map[string]map[string]string, error)

	// GetNotifyMethods 獲取所有通知方法
	GetNotifyMethods() []string

	// GetNotifyOptions 獲取通知選項
	GetNotifyOptions(ctx context.Context, realm string) (map[string]map[string][]string, error)

	// SetKeycloakClient 設置 Keycloak 客戶端
	SetKeycloakClient(client keycloak.KeycloakClient) error

	// ToResponse 將 Contact 轉換為 ContactResponse
	ToResponse(contact models.Contact) models.ContactResponse

	// FromResponse 將 ContactResponse 轉換為 Contact
	FromResponse(contactResp models.ContactResponse, realm string) models.Contact
}
