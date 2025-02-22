package contact

import (
	"fmt"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/notify"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 將接口重命名為 ContactService
type ContactService interface {
	// 基本 CRUD
	Create(contact *models.AlertContact) error
	Get(id int64) (*models.AlertContact, error)
	List(realm string) ([]models.AlertContact, error)
	Update(contact *models.AlertContact) error
	Delete(id int64) error

	// 檢查操作
	CheckName(contact models.AlertContact) (bool, string)
}

// Service 通知管道服務
type Service struct {
	db            *gorm.DB
	logger        *zap.Logger
	notifyService *notify.Service
}

// NewService 創建通知管道服務
func NewService(db *gorm.DB, logger *zap.Logger, notifyService *notify.Service) *Service {
	return &Service{
		db:            db,
		logger:        logger,
		notifyService: notifyService,
	}
}

// Create 創建通知管道
func (s *Service) Create(contact *models.AlertContact) error {
	return s.db.Create(contact).Error
}

// Get 獲取通知管道
func (s *Service) Get(id int64) (*models.AlertContact, error) {
	var contact models.AlertContact
	err := s.db.First(&contact, id).Error
	return &contact, err
}

// List 獲取通知管道列表
func (s *Service) List(realm string) ([]models.AlertContact, error) {
	var contacts []models.AlertContact
	err := s.db.Where("realm_name = ?", realm).Find(&contacts).Error
	return contacts, err
}

// Update 更新通知管道
func (s *Service) Update(contact *models.AlertContact) error {
	return s.db.Save(contact).Error
}

// Delete 刪除通知管道
func (s *Service) Delete(id int64, userID string) models.Response {
	// 檢查權限
	var contact models.AlertContact
	if err := s.db.First(&contact, id).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     "通知管道不存在",
		}
	}

	// 執行刪除
	if err := s.db.Delete(&contact).Error; err != nil {
		return models.Response{
			Success: false,
			Msg:     err.Error(),
		}
	}

	return models.Response{
		Success: true,
		Msg:     "刪除成功",
	}
}

// NotifyTest 測試通知
func (s *Service) NotifyTest(contact models.AlertContact) models.Response {
	notify := common.NotifyConfig{
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

// IsUsedByRules 檢查通知管道是否被規則使用
func (s *Service) IsUsedByRules(id int) (bool, error) {
	var count int64
	err := s.db.Model(&models.AlertRule{}).
		Where("contact_id = ?", id).
		Count(&count).Error

	return count > 0, err
}

// CheckName 檢查名稱是否重複
func (s *Service) CheckName(contact models.AlertContact) (models.Response, error) {
	var existingContact models.AlertContact
	result := s.db.Where(&models.AlertContact{
		RealmName: contact.RealmName,
		Name:      contact.Name,
	}).First(&existingContact)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return models.Response{}, result.Error
	}

	if result.RowsAffected > 0 && existingContact.ID != contact.ID {
		return models.Response{
			Success: false,
			Msg:     "名稱已存在",
		}, nil
	}

	return models.Response{
		Success: true,
		Msg:     "",
	}, nil
}
