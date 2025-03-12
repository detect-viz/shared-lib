package mysql

import (
	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateLabel 新增標籤
func (c *Client) CreateLabel(label *models.LabelKey, values []string) (*models.LabelKey, error) {

	exists, err := c.Exists(label.RealmName, "label_keys", "key_name", label.KeyName)
	if err != nil {
		return nil, ParseDBError(err)
	}
	if exists {
		return nil, apierrors.ErrDuplicateEntry
	}

	if err := c.db.Create(&label).Error; err != nil {
		return nil, ParseDBError(err)
	}

	var labelValues []models.LabelValue
	if len(values) > 0 {
		for _, v := range values {
			labelValues = append(labelValues, models.LabelValue{LabelKeyID: label.ID, Value: v})
		}
		if err := c.db.Create(&labelValues).Error; err != nil {
			return nil, ParseDBError(err)
		}
	}

	return label, nil
}

// Get 查詢單個標籤
func (c *Client) GetLabel(id int64) (*models.LabelKey, error) {
	var label models.LabelKey
	err := c.db.Preload(clause.Associations).First(&label, "id = ?", id).Error
	if err != nil {
		return nil, ParseDBError(err)
	}
	return &label, nil
}

// ListLabels 列出標籤（支援分頁）
func (c *Client) ListLabels(realm string, cursor int64, limit int) ([]models.LabelKey, int64, error) {
	var labels []models.LabelKey
	err := c.db.Preload(clause.Associations).
		Where("realm_name = ? AND created_at > ?", realm, cursor).
		Order("created_at ASC").
		Limit(limit).
		Find(&labels).Error
	if err != nil {
		return nil, 0, ParseDBError(err)
	}
	// 計算 next_cursor
	nextCursor := int64(-1)
	if len(labels) > 0 {
		nextCursor = labels[len(labels)-1].CreatedAt
	}

	return labels, nextCursor, nil
}

// 更新標籤
func (c *Client) UpdateLabel(input models.LabelKey, values []string) (*models.LabelKey, error) {
	//* 使用 Transaction 確保數據一致性
	err := c.db.Transaction(func(tx *gorm.DB) error {

		//* 1. 檢查 key 是否已存在（排除自身 key）
		exists, err := c.ExistsExcludeSelf(input.RealmName, "label_keys", "key_name", input.KeyName, input.ID)
		if err != nil {
			return ParseDBError(err)
		}
		if exists {
			return apierrors.ErrDuplicateEntry
		}

		//* 2. 刪除舊的 LabelValues
		if err := tx.Where("label_key_id = ?", input.ID).Delete(&models.LabelValue{}).Error; err != nil {
			return err
		}

		//* 3. 批量插入新的 LabelValues
		var labelValues []models.LabelValue
		for _, v := range values {
			labelValues = append(labelValues, models.LabelValue{LabelKeyID: input.ID, Value: v})
		}

		if len(labelValues) > 0 {
			if err := tx.Create(&labelValues).Error; err != nil {
				return ParseDBError(err)
			}
		}

		return nil
	})

	//* 如果 Transaction 失敗，回傳 nil, error
	if err != nil {
		return nil, err
	}

	//* 重新查詢最新的 Label，確保返回最新數據
	return c.GetLabel(input.ID)
}

// 變更標籤 Key
func (c *Client) UpdateLabelKeyName(realm, oldKey, newKey string) (*models.LabelKey, error) {
	exists, err := c.Exists(realm, "label_keys", "key_name", newKey)
	if err != nil {
		return nil, ParseDBError(err)
	}
	if exists {
		return nil, apierrors.ErrDuplicateEntry
	}
	result := c.db.Model(&models.LabelKey{}).Where("realm_name = ? AND key_name = ?", realm, oldKey).Update("key_name", newKey)
	if result.Error != nil {
		return nil, ParseDBError(result.Error)
	}

	var label models.LabelKey
	err = c.db.Preload(clause.Associations).First(&label, "key_name = ?", newKey).Error
	if err != nil {
		return nil, ParseDBError(err)
	}
	return &label, nil
}

// Delete 刪除標籤
func (c *Client) DeleteLabel(id int64) error {
	result := c.db.Where("id = ?", id).Delete(&models.LabelKey{})
	if result.RowsAffected == 0 {
		return apierrors.ErrNotFound
	}
	if result.Error != nil {
		return ParseDBError(result.Error)
	}
	return nil
}

// 批量新增或更新標籤
func (c *Client) BulkCreateOrUpdateLabel(realm string, labels []models.LabelDTO) error {

	for _, label := range labels {
		var existing models.LabelKey

		// 查詢是否已存在相同的 Key
		result := c.db.Debug().Table("label_keys").Where("realm_name = ? AND key_name = ?", realm, label.Key).First(&existing)

		if result.Error != nil {
			return ParseDBError(result.Error)
		}
		if result.RowsAffected == 0 {
			c.CreateLabel(&models.LabelKey{
				RealmName: realm,
				KeyName:   label.Key,
			}, label.Values)
		} else {
			// 如果存在，更新 Values
			_, err := c.UpdateLabel(models.LabelKey{
				ID:        existing.ID,
				RealmName: realm,
				KeyName:   label.Key,
			}, label.Values)
			if err != nil {
				return ParseDBError(err)
			}

		}
	}

	return nil
}
