package contacts

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/auth/keycloak"
	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/notifier"
	"github.com/detect-viz/shared-lib/storage/mysql"
	"github.com/google/uuid"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var ContactSet = wire.NewSet(
	NewService,
	wire.Bind(new(Service), new(*serviceImpl)),
)

// Service 通知管道服務
type serviceImpl struct {
	mysql               *mysql.Client
	logger              logger.Logger
	notifyService       notifier.Service
	keycloakClient      *keycloak.Client
	onContactChangeFunc ContactChangeCallback
}

// 創建通知管道服務
func NewService(mysql *mysql.Client, logger logger.Logger, notifyService notifier.Service, keycloakClient *keycloak.Client) *serviceImpl {
	return &serviceImpl{
		mysql:          mysql,
		logger:         logger,
		notifyService:  notifyService,
		keycloakClient: keycloakClient,
	}
}

// 設置聯絡人變更回調函數
func (s *serviceImpl) SetContactChangeCallback(callback ContactChangeCallback) {
	s.onContactChangeFunc = callback
}

// 創建通知管道
func (s *serviceImpl) Create(realm string, contactResp *models.ContactResponse) (*models.ContactResponse, error) {
	contact := s.FromResponse(*contactResp, realm)
	createdContact, err := s.mysql.CreateContact(&contact)
	if err != nil {
		return nil, err
	}

	// 創建成功後，觸發回調函數
	if s.onContactChangeFunc != nil {
		s.onContactChangeFunc(*createdContact, "create")
	}

	response := s.ToResponse(*createdContact)
	return &response, nil
}

// 獲取通知管道
func (s *serviceImpl) Get(id string) (*models.ContactResponse, error) {
	// 將 ID 從 string 轉換為 []byte
	idBytes, err := uuid.Parse(id)
	if err != nil {
		return nil, apierrors.ErrInvalidID
	}

	contact, err := s.mysql.GetContact(idBytes[:])
	if err != nil {
		return nil, err
	}

	if contact.ChannelType == "email" || contact.ChannelType == "line" {
		cfg, err := s.GetConfig(context.Background(), contact.ChannelType)
		if err != nil {
			return nil, err
		}
		fmt.Printf("cfg: %v\n", cfg)
		contact.Config = cfg
	}

	response := s.ToResponse(*contact)
	return &response, nil
}

// 獲取通知管道列表
func (s *serviceImpl) List(realm string, cursor int64, limit int) ([]models.ContactResponse, int64, error) {
	contacts, nextCursor, err := s.mysql.ListContacts(realm, cursor, limit)
	if err != nil {
		return nil, 0, err
	}

	// 將 Contact 轉換為 ContactResponse
	contactResponses := make([]models.ContactResponse, len(contacts))
	for i, contact := range contacts {
		contactResponses[i] = s.ToResponse(contact)
	}

	return contactResponses, nextCursor, nil
}

// 更新通知管道
func (s *serviceImpl) Update(realm string, contactResp *models.ContactResponse) (*models.ContactResponse, error) {
	contact := s.FromResponse(*contactResp, realm)
	updatedContact, err := s.mysql.UpdateContact(&contact)
	if err != nil {
		return nil, err
	}

	// 更新成功後，觸發回調函數
	if s.onContactChangeFunc != nil {
		s.onContactChangeFunc(*updatedContact, "update")
	}

	response := s.ToResponse(*updatedContact)
	return &response, nil
}

// 刪除通知管道
func (s *serviceImpl) Delete(id string) error {
	// 將 ID 從 string 轉換為 []byte
	idBytes, err := uuid.Parse(id)
	if err != nil {
		return apierrors.ErrInvalidID
	}

	// 獲取聯絡人，以便在刪除後觸發回調
	contact, err := s.mysql.GetContact(idBytes[:])
	if err != nil {
		return err
	}

	//* 檢查 contact_id 是否仍被使用
	used, err := s.mysql.IsUsedByRules(idBytes[:])
	if err != nil {
		return apierrors.ErrInternalError
	}
	if used {
		return apierrors.ErrUsedByRules
	}

	// 刪除聯絡人
	if err := s.mysql.DeleteContact(idBytes[:]); err != nil {
		return err
	}

	// 刪除成功後，觸發回調函數
	if s.onContactChangeFunc != nil && contact != nil {
		s.onContactChangeFunc(*contact, "delete")
	}

	return nil
}

// 測試通知
func (s *serviceImpl) NotifyTest(realm string, contactResp *models.ContactResponse) error {
	contact := s.FromResponse(*contactResp, realm)
	notify := common.NotifySetting{
		Type:   contact.ChannelType,
		Config: contact.Config,
	}

	notify.Config["title"] = "測試通知"
	notify.Config["message"] = "這是一個測試通知"

	if err := s.notifyService.Send(notify); err != nil {
		return err
	}

	return nil
}

// 檢查通知管道是否被規則使用
func (s *serviceImpl) IsUsedByRules(id string) (bool, error) {
	// 將 ID 從 string 轉換為 []byte
	idBytes, err := uuid.Parse(id)
	if err != nil {
		return false, apierrors.ErrInvalidID
	}

	return s.mysql.IsUsedByRules(idBytes[:])
}

// 獲取規則的通知管道
func (s *serviceImpl) GetContactsByRuleID(ruleID string) ([]models.ContactResponse, error) {
	// 將 ID 從 string 轉換為 []byte
	idBytes, err := uuid.Parse(ruleID)
	if err != nil {
		return nil, apierrors.ErrInvalidID
	}

	contacts, err := s.mysql.GetContactsByRuleID(idBytes[:])
	if err != nil {
		return nil, err
	}

	// 將 Contact 轉換為 ContactResponse
	contactResponses := make([]models.ContactResponse, len(contacts))
	for i, contact := range contacts {
		contactResponses[i] = s.ToResponse(contact)
	}

	return contactResponses, nil
}

// 獲取所有通知管道方法
func (s *serviceImpl) GetNotifyMethods() []string {
	return []string{"email", "slack", "discord", "teams", "webex", "webhook", "line"}
}

// 根據通知類型獲取對應的配置
func (s *serviceImpl) GetConfig(ctx context.Context, notifyType string) (map[string]string, error) {
	// 檢查 keycloakClient 是否為 nil
	if s.keycloakClient == nil {
		s.logger.Warn("keycloak client is nil, returning empty config")
		return make(map[string]string), nil
	}

	configKey := fmt.Sprintf("%s_config", notifyType)
	s.logger.Info("GetConfig", zap.String("configKey", configKey))
	configJSON, err := s.keycloakClient.GetRealmAttribute(ctx, configKey)
	if err != nil {
		// 如果是找不到屬性的錯誤，返回空配置而不是錯誤
		if strings.Contains(err.Error(), "no attributes found in realm") {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to get %s config: %w", notifyType, err)
	}

	// 如果配置為空，返回空映射
	if configJSON == "" {
		return make(map[string]string), nil
	}

	// 解析 JSON
	var config map[string]string
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to parse %s config: %w", notifyType, err)
	}

	return config, nil
}

// 獲取所有通知配置
func (s *serviceImpl) GetAllConfigs(ctx context.Context) (map[string]map[string]string, error) {
	notifyTypes := s.GetNotifyMethods()
	configs := make(map[string]map[string]string)

	for _, notifyType := range notifyTypes {
		config, err := s.GetConfig(ctx, notifyType)
		if err == nil && len(config) > 0 {
			configs[notifyType] = config
		}
	}

	return configs, nil
}

// 獲取通知選項
func (s *serviceImpl) GetNotifyOptions(ctx context.Context, realm string) (map[string]map[string][]string, error) {
	// 檢查 keycloakClient 是否為 nil
	if s.keycloakClient == nil {
		s.logger.Warn("keycloak client is nil, using default notify options")
		return getNotifyOptions(nil), nil
	}

	// 從 Keycloak 獲取租戶的通知配置
	tenantConfigs, err := s.GetAllConfigs(ctx)
	if err != nil {
		s.logger.Warn("failed to get tenant configs, using default notify options",
			zap.String("error", err.Error()))
		return getNotifyOptions(nil), nil
	}

	return getNotifyOptions(tenantConfigs), nil
}

// getNotifyOptions 根據租戶配置生成通知選項
func getNotifyOptions(tenantConfigs map[string]map[string]string) map[string]map[string][]string {
	options := map[string]map[string][]string{
		"email": {
			"required": {"to"},
			"optional": {"cc", "bcc", "from"},
		},
		"line": {
			"required": {"to"},
			"optional": {},
		},
		"slack": {
			"required": {"url"},
			"optional": {},
		},
		"discord": {
			"required": {"url"},
			"optional": {},
		},
		"teams": {
			"required": {"url"},
			"optional": {},
		},
		"webex": {
			"required": {"url"},
			"optional": {},
		},
		"webhook": {
			"required": {"url"},
			"optional": {},
		},
	}

	if tenantConfigs == nil {
		return options
	}

	// 處理 SMTP 配置
	smtpConfig, hasSmtp := tenantConfigs["email"]
	if !hasSmtp || smtpConfig["host"] == "" {
		options["email"]["required"] = append(options["email"]["required"], "host", "port", "user", "password")
	}

	// 處理 LINE 配置
	lineConfig, hasLine := tenantConfigs["line"]
	if hasLine && lineConfig["channel_token"] != "" {
		// 如果已經配置了 channel_token，則不需要用戶再次輸入
		// 不添加到任何列表中
	} else {
		// 否則將其添加到 required
		options["line"]["required"] = append(options["line"]["required"], "channel_token")
	}

	// 處理其他通知類型的配置
	for notifyType, config := range tenantConfigs {
		switch notifyType {
		case "slack", "discord", "teams", "webex", "webhook":
			if config["url"] != "" {
				// 如果已經配置了 URL，則從必填字段中移除
				options[notifyType]["optional"] = append(options[notifyType]["optional"], "url")
				options[notifyType]["required"] = []string{}
			}
		}
	}

	return options
}

// 設置 Keycloak 客戶端
func (s *serviceImpl) SetKeycloakClient(client keycloak.KeycloakClient) error {
	if client == nil {
		return fmt.Errorf("keycloak client is nil")
	}

	if kc, ok := client.(*keycloak.Client); ok {
		s.keycloakClient = kc
		return nil
	}

	return fmt.Errorf("invalid keycloak client type")
}

func (s *serviceImpl) ToResponse(contact models.Contact) models.ContactResponse {
	// 將 JSONMap 轉換為 map[string]interface{}
	configMap := make(map[string]interface{})
	for k, v := range contact.Config {
		configMap[k] = v
	}

	// 將 SeveritySet 轉換為 []string
	severities := []string{}
	for _, sev := range contact.Severities {
		if sev == "info" || sev == "warn" || sev == "crit" {
			severities = append(severities, sev)
		}
	}

	return models.ContactResponse{
		ID:           hex.EncodeToString(contact.ID),
		Name:         contact.Name,
		ChannelType:  contact.ChannelType,
		Enabled:      contact.Enabled,
		SendResolved: contact.SendResolved,
		AutoApply:    contact.AutoApply,
		MaxRetry:     contact.MaxRetry,
		RetryDelay:   contact.RetryDelay,
		Config:       configMap,
		Severities:   severities,
	}
}

// 將 ContactResponse 轉換為 Contact
func (s *serviceImpl) FromResponse(resp models.ContactResponse, realm string) models.Contact {
	// 將 ID 從 string 轉換為 []byte
	var id []byte
	if resp.ID != "" {
		parsedID, err := uuid.Parse(resp.ID)
		if err == nil {
			id = parsedID[:]
		}
	}

	// 將 map[string]interface{} 轉換為 common.JSONMap
	config := make(models.JSONMap)
	for k, v := range resp.Config {
		if strVal, ok := v.(string); ok {
			config[k] = strVal
		}
	}

	// 創建 Contact 對象
	contact := models.Contact{
		ID:           id,
		RealmName:    realm,
		Name:         resp.Name,
		ChannelType:  resp.ChannelType,
		Enabled:      resp.Enabled,
		SendResolved: resp.SendResolved,
		AutoApply:    resp.AutoApply,
		MaxRetry:     resp.MaxRetry,
		RetryDelay:   resp.RetryDelay,
		Config:       config,
	}

	// 設置 Severities
	for _, sev := range resp.Severities {
		contact.Severities = append(contact.Severities, sev)
	}

	return contact
}
