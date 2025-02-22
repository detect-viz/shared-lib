package label

import (
	"fmt"

	"github.com/detect-viz/shared-lib/models"

	"gorm.io/gorm"
)

// Service 標籤服務
type Service struct {
	db *gorm.DB
}

// NewService 創建標籤服務
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateLabel 創建標籤
func (s *Service) CreateLabel(realm string, labels []models.Label) error {
	for i := range labels {
		labels[i].RealmName = realm
	}
	return s.db.Create(&labels).Error
}

// GetAllLabels 獲取所有標籤
func (s *Service) GetAllLabels(realm string) []models.Label {
	var labels []models.Label
	s.db.Where("realm_name = ?", realm).Find(&labels)
	return labels
}

// GetLabelsByKey 根據 key 獲取標籤值
func (s *Service) GetLabelsByKey(realm, key string) ([]models.OptionResponse, error) {
	var values []string
	err := s.db.Model(&models.Label{}).
		Where("realm_name = ? AND key = ?", realm, key).
		Pluck("value", &values).Error

	OptionResponse := make([]models.OptionResponse, len(values))
	for i, value := range values {
		OptionResponse[i] = models.OptionResponse{
			Text:  value,
			Value: value,
		}
	}
	return OptionResponse, err
}

// CheckLabelKey 檢查標籤 key 是否存在
func (s *Service) CheckLabelKey(label models.Label) models.Response {
	var count int64
	s.db.Model(&models.Label{}).
		Where("realm_name = ? AND key = ?", label.RealmName, label.Key).
		Count(&count)

	if count > 0 {
		return models.Response{
			Success: false,
			Msg:     fmt.Sprintf("標籤 [%s] 已存在", label.Key),
		}
	}
	return models.Response{Success: true}
}

// UpdateLabel 更新標籤
func (s *Service) UpdateLabel(realm, key string, labels []models.Label) ([]models.Label, error) {
	// 刪除舊標籤
	if err := s.db.Where("realm_name = ? AND key = ?", realm, key).
		Delete(&models.Label{}).Error; err != nil {
		return nil, err
	}

	// 創建新標籤
	for i := range labels {
		labels[i].RealmName = realm
		labels[i].Key = key
	}

	if err := s.db.Create(&labels).Error; err != nil {
		return nil, err
	}

	return labels, nil
}

// DeleteLabelByKey 根據 key 刪除標籤
func (s *Service) DeleteLabelByKey(realm, key string) models.Response {
	result := s.db.Where("realm_name = ? AND key = ?", realm, key).
		Delete(&models.Label{})

	if result.Error != nil {
		return models.Response{
			Success: false,
			Msg:     result.Error.Error(),
		}
	}

	return models.Response{
		Success: true,
		Msg:     fmt.Sprintf("標籤 [%s] 刪除成功", key),
	}
}
