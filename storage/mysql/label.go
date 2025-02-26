package mysql

import (
	"errors"

	"github.com/detect-viz/shared-lib/models"
	"gorm.io/gorm"
)

// CreateLabel 新增標籤
func (c *Client) CreateLabel(label *models.Label) (*models.Label, error) {
	if err := c.db.Create(label).Error; err != nil {
		return nil, err
	}
	return label, nil
}

// Get 查詢單個標籤
func (c *Client) GetLabel(realm, key string) (*models.Label, error) {
	var label models.Label
	err := c.db.Where("realm_name = ? AND key_name = ?", realm, key).First(&label).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &label, err
}

// ListLabels 列出標籤（支援分頁）
func (c *Client) ListLabels(realm string, limit, offset int) ([]models.Label, error) {
	var labels []models.Label
	err := c.db.Where("realm_name = ?", realm).Limit(limit).Offset(offset).Find(&labels).Error
	return labels, err
}

// UpdateLabel 更新標籤
func (c *Client) UpdateLabel(realm, key string, updates map[string]interface{}) error {
	return c.db.Model(&models.Label{}).Where("realm_name = ? AND key_name = ?", realm, key).Updates(updates).Error
}

// UpdateKey 變更標籤 Key
func (c *Client) UpdateKey(realm, oldKey, newKey string) error {
	return c.db.Model(&models.Label{}).Where("realm_name = ? AND key_name = ?", realm, oldKey).Update("key_name", newKey).Error
}

// Delete 刪除標籤
func (c *Client) DeleteLabel(realm, key string) error {
	return c.db.Where("realm_name = ? AND key_name = ?", realm, key).Delete(&models.Label{}).Error
}

// ExistsLabel 檢查標籤是否存在
func (c *Client) ExistsLabel(realm, key string) (bool, error) {
	var count int64
	err := c.db.Model(&models.Label{}).Where("realm_name = ? AND key_name = ?", realm, key).Count(&count).Error
	return count > 0, err
}

// BulkCreateOrUpdateLabel 批量新增或更新標籤
func (c *Client) BulkCreateOrUpdateLabel(realm string, labels []models.Label) ([]models.Label, error) {
	tx := c.db.Begin()
	for _, label := range labels {
		err := tx.Where("realm_name = ? AND key_name = ?", realm, label.KeyName).
			Assign("value", label.Value).
			FirstOrCreate(&label).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	return labels, tx.Commit().Error
}
