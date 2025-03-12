package mysql

import (
	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetContactsByRuleID 獲取規則的通知管道
func (c *Client) GetContactsByRuleID(ruleID []byte) ([]models.Contact, error) {
	var contacts []models.Contact
	err := c.db.Model(&models.Contact{}).
		Joins("JOIN rule_contacts ON rule_contacts.contact_id = contacts.id").
		Where("rule_contacts.rule_id = ?", ruleID).
		Find(&contacts).Error
	return contacts, err
}

// IsUsedByRules 檢查通知管道是否被規則使用
func (c *Client) IsUsedByRules(contactID []byte) (bool, error) {
	var count int64
	err := c.db.Model(&models.RuleContact{}).
		Where("contact_id = ?", contactID).
		Count(&count).Error
	if err != nil {
		return false, ParseDBError(err)
	}
	return count > 0, nil
}

// 創建通知管道
func (c *Client) CreateContact(contact *models.Contact) (*models.Contact, error) {
	exists, err := c.Exists(contact.RealmName, "contacts", "name", contact.Name)
	if err != nil {
		return nil, ParseDBError(err)
	}
	if exists {
		return nil, apierrors.ErrDuplicateEntry
	}

	// 創建聯絡人，不包含 Severities
	if err := c.db.Omit("Severities").Create(&contact).Error; err != nil {
		return nil, ParseDBError(err)
	}

	// 更新 Severities
	if len(contact.Severities) > 0 {
		if err := c.db.Model(&contact).Update("severities", contact.Severities).Error; err != nil {
			return nil, ParseDBError(err)
		}
	}

	return contact, nil
}

// 獲取通知管道
func (c *Client) GetContact(id []byte) (*models.Contact, error) {
	var contact models.Contact
	err := c.db.Preload(clause.Associations).First(&contact, id).Error
	if err != nil {
		return nil, ParseDBError(err)
	}
	return &contact, nil
}

// 獲取通知管道列表
func (c *Client) ListContacts(realm string, cursor int64, limit int) ([]models.Contact, int64, error) {
	var contacts []models.Contact
	// 只回傳 UI 需要的欄位
	query := c.db.Model(&models.Contact{}).
		Where("realm_name = ?", realm)

	if cursor > 0 {
		query = query.Where("created_at > ?", cursor)
	}

	err := query.Order("created_at ASC").
		Limit(limit).
		Find(&contacts).Error

	if err != nil {
		return nil, 0, ParseDBError(err)
	}

	// 計算 next_cursor
	nextCursor := int64(-1)
	if len(contacts) > 0 && len(contacts) >= limit {
		var lastCreatedAt int64
		c.db.Model(&models.Contact{}).
			Where("id = ?", contacts[len(contacts)-1].ID).
			Pluck("created_at", &lastCreatedAt)
		nextCursor = lastCreatedAt
	}

	return contacts, nextCursor, nil
}

// 更新通知管道
func (c *Client) UpdateContact(contact *models.Contact) (*models.Contact, error) {
	// 使用 Transaction 確保數據一致性
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// 1. 檢查 name 是否已存在（排除自身 ID）
		var count int64
		err := tx.Model(&models.Contact{}).
			Where("realm_name = ? AND name = ? AND id != ?", contact.RealmName, contact.Name, contact.ID).
			Count(&count).Error
		if err != nil {
			return ParseDBError(err)
		}
		if count > 0 {
			return apierrors.ErrDuplicateEntry
		}

		// 2. 更新 Contact 本身（排除 Severities）
		if err := tx.Omit("Severities").Model(&models.Contact{}).
			Where("id = ?", contact.ID).
			Updates(map[string]interface{}{
				"name":          contact.Name,
				"channel_type":  contact.ChannelType,
				"enabled":       contact.Enabled,
				"send_resolved": contact.SendResolved,
				"max_retry":     contact.MaxRetry,
				"retry_delay":   contact.RetryDelay,
				"config":        contact.Config,
			}).Error; err != nil {
			return ParseDBError(err)
		}

		// 3. 更新 Severities
		if len(contact.Severities) > 0 {
			if err := tx.Model(&contact).Update("severities", contact.Severities).Error; err != nil {
				return ParseDBError(err)
			}
		}

		return nil
	})

	// 如果 Transaction 失敗，回傳 nil, error
	if err != nil {
		return nil, err
	}

	// 重新查詢最新的 Contact，確保返回最新數據
	return c.GetContact(contact.ID)
}

// 刪除通知管道
func (c *Client) DeleteContact(id []byte) error {
	result := c.db.Delete(&models.Contact{}, id)
	if result.RowsAffected == 0 {
		return apierrors.ErrNotFound
	}
	if result.Error != nil {
		return ParseDBError(result.Error)
	}
	return nil
}
