package contacts

import (
	"fmt"

	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/notifier"
	"github.com/detect-viz/shared-lib/storage/mysql"
	"github.com/google/wire"
)

var ContactSet = wire.NewSet(
	NewService,
	wire.Bind(new(Service), new(*serviceImpl)),
)

// Service 通知管道服務
type serviceImpl struct {
	mysql         *mysql.Client
	logger        logger.Logger
	notifyService notifier.Service
}

// NewService 創建通知管道服務
func NewService(mysql *mysql.Client, logger logger.Logger, notifyService notifier.Service) *serviceImpl {
	return &serviceImpl{
		mysql:         mysql,
		logger:        logger,
		notifyService: notifyService,
	}
}

// Create 創建通知管道
func (s *serviceImpl) Create(contact *models.Contact) error {
	return s.mysql.CreateContact(contact)
}

// Get 獲取通知管道
func (s *serviceImpl) Get(id int64) (*models.Contact, error) {
	return s.mysql.GetContact(id)
}

// List 獲取通知管道列表
func (s *serviceImpl) List(realm string) ([]models.Contact, error) {
	return s.mysql.ListContacts()
}

// Update 更新通知管道
func (s *serviceImpl) Update(contact *models.Contact) error {
	return s.mysql.UpdateContact(contact)
}

// Delete 刪除通知管道
func (s *serviceImpl) Delete(id int64) error {
	return s.mysql.DeleteContact(id)
}

// NotifyTest 測試通知
func (s *serviceImpl) NotifyTest(contact models.Contact) models.Response {
	notify := common.NotifySetting{
		Type:   contact.Type,
		Config: contact.Details,
	}

	if err := s.notifyService.Send(notify); err != nil {
		return models.Response{
			Success: false,
			Msg:     fmt.Sprintf("測試失敗: %v", err),
		}
	}

	return models.Response{
		Success: true,
		Msg:     "測試成功",
	}
}

// CheckName 檢查名稱是否重複
func (s *serviceImpl) CheckName(contact models.Contact) (bool, string) {
	return s.mysql.CheckContactName(contact)
}

// IsUsedByRules 檢查通知管道是否被規則使用
func (s *serviceImpl) IsUsedByRules(id int64) (bool, error) {
	return s.mysql.IsUsedByRules(id)
}

// GetContactsByRuleID 獲取規則的通知管道
func (s *serviceImpl) GetContactsByRuleID(ruleID int64) ([]models.Contact, error) {
	return s.mysql.GetContactsByRuleID(ruleID)
}
