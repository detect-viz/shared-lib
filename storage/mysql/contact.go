package mysql

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetContactsByRuleID 獲取規則的通知管道
func (c *Client) GetContactsByRuleID(ruleID []byte) ([]models.Contact, error) {
	var contacts []models.Contact
	err := c.db.Debug().Model(&models.Contact{}).
		Select("id, realm_name, name, channel_type, enabled, send_resolved, auto_apply, max_retry, retry_delay, config").
		Joins("JOIN rule_contacts ON rule_contacts.contact_id = contacts.id").
		Where("rule_contacts.rule_id = ?", ruleID).
		Find(&contacts).Error

	if err != nil {
		return nil, err
	}

	// 為每個聯絡人單獨獲取 severities
	for i := range contacts {
		severities, err := c.GetContactSeverities(contacts[i].ID)
		if err != nil {
			fmt.Printf("Warning: Failed to get severities for contact %x: %v\n", contacts[i].ID, err)
			// 不返回錯誤，而是使用空數組
			contacts[i].Severities = []string{}
		} else {
			contacts[i].Severities = severities
		}
	}

	return contacts, nil
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
	contact.ID = GenerateUUID16()
	exists, err := c.Exists(contact.RealmName, "contacts", "name", contact.Name)
	if err != nil {
		return nil, ParseDBError(err)
	}
	if exists {
		return nil, apierrors.ErrDuplicateEntry
	}

	// 確保 AutoApply 欄位被正確設置
	// 如果需要，可以在這裡添加日誌來檢查 AutoApply 的值
	fmt.Printf("Creating contact with AutoApply: %v\n", contact.AutoApply)

	// 使用 Select("*") 明確指定所有欄位
	if err := c.db.Select("*").Create(&contact).Error; err != nil {
		return nil, ParseDBError(err)
	}

	// 驗證 AutoApply 是否被正確保存
	var savedContact models.Contact
	if err := c.db.First(&savedContact, contact.ID).Error; err != nil {
		return contact, nil // 返回原始聯絡人，但記錄錯誤
	}

	fmt.Printf("Saved contact with AutoApply: %v\n", savedContact.AutoApply)

	return contact, nil
}

// 獲取通知管道
func (c *Client) GetContact(id []byte) (*models.Contact, error) {
	var contact models.Contact

	// 打印 ID 的十六進制表示，以便調試
	fmt.Printf("Querying contact with ID: %x\n", id)

	// 明確排除 severities 欄位
	err := c.db.Preload(clause.Associations).
		Select("id, realm_name, name, channel_type, enabled, send_resolved, auto_apply, max_retry, retry_delay, config").
		First(&contact, "id = ?", id).Error

	if err != nil {
		// 如果找不到記錄，嘗試查詢已軟刪除的記錄
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Record not found, checking if it was soft deleted\n")
			err = c.db.Unscoped().
				Select("id, realm_name, name, channel_type, enabled, send_resolved, auto_apply, max_retry, retry_delay, config").
				First(&contact, "id = ?", id).Error
			if err == nil {
				return nil, apierrors.ErrRecordDeleted
			}
		}
		return nil, ParseDBError(err)
	}

	// 單獨獲取 severities
	severities, err := c.GetContactSeverities(id)
	if err != nil {
		fmt.Printf("Warning: Failed to get severities for contact %x: %v\n", id, err)
		// 不返回錯誤，而是使用空數組
		contact.Severities = []string{}
	} else {
		contact.Severities = severities
	}

	return &contact, nil
}

// 獲取通知管道列表
func (c *Client) ListContacts(realm string, cursor int64, limit int) ([]models.Contact, int64, error) {
	var contacts []models.Contact

	// 構建基本查詢
	query := c.db.Model(&models.Contact{}).
		Where("realm_name = ?", realm)

	if cursor > 0 {
		query = query.Where("created_at > ?", cursor)
	}

	// 執行查詢，明確排除 severities 欄位
	err := query.Order("created_at ASC").
		Limit(limit).
		Select("id, realm_name, name, channel_type, enabled, send_resolved, auto_apply, max_retry, retry_delay, config").
		Find(&contacts).Error

	if err != nil {
		return nil, 0, ParseDBError(err)
	}

	// 為每個聯絡人單獨獲取 severities
	for i := range contacts {
		severities, err := c.GetContactSeverities(contacts[i].ID)
		if err != nil {
			fmt.Printf("Warning: Failed to get severities for contact %x: %v\n", contacts[i].ID, err)
			// 不返回錯誤，而是使用空數組
			contacts[i].Severities = []string{}
		} else {
			contacts[i].Severities = severities
		}
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

		// 2. 更新 Contact
		if err := tx.Model(&models.Contact{}).
			Where("id = ?", contact.ID).
			Updates(map[string]interface{}{
				"name":          contact.Name,
				"channel_type":  contact.ChannelType,
				"enabled":       contact.Enabled,
				"auto_apply":    contact.AutoApply,
				"send_resolved": contact.SendResolved,
				"max_retry":     contact.MaxRetry,
				"retry_delay":   contact.RetryDelay,
				"config":        contact.Config,
				"severities":    contact.Severities,
			}).Error; err != nil {
			return ParseDBError(err)
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
	// 檢查 ID 的長度
	if len(id) != 16 {
		return fmt.Errorf("invalid ID length: expected 16 bytes, got %d bytes", len(id))
	}

	// 打印 ID 的十六進制表示，以便調試
	fmt.Printf("Deleting contact with ID: %x\n", id)

	result := c.db.Delete(&models.Contact{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return apierrors.ErrNotFound
	}
	if result.Error != nil {
		return ParseDBError(result.Error)
	}
	return nil
}

// GetAutoApplyContacts 獲取所有設置了 AutoApply 的通知管道
func (c *Client) GetAutoApplyContacts() ([]models.Contact, error) {
	var contacts []models.Contact
	result := c.db.
		Select("id, realm_name, name, channel_type, enabled, send_resolved, auto_apply, max_retry, retry_delay, config").
		Where("auto_apply = ?", true).
		Where("enabled = ?", true).
		Find(&contacts)

	if result.Error != nil {
		return nil, result.Error
	}

	// 為每個聯絡人單獨獲取 severities
	for i := range contacts {
		severities, err := c.GetContactSeverities(contacts[i].ID)
		if err != nil {
			fmt.Printf("Warning: Failed to get severities for contact %x: %v\n", contacts[i].ID, err)
			// 不返回錯誤，而是使用空數組
			contacts[i].Severities = []string{}
		} else {
			contacts[i].Severities = severities
		}
	}

	return contacts, nil
}

// AddContactToRule 將聯絡人添加到規則
func (c *Client) AddContactToRule(ruleID []byte, contactID []byte) error {
	// 檢查是否已經存在關聯
	var count int64
	result := c.db.Table("rule_contacts").
		Where("rule_id = ? AND contact_id = ?", ruleID, contactID).
		Count(&count)

	if result.Error != nil {
		return result.Error
	}

	// 如果已經存在關聯，則不需要添加
	if count > 0 {
		return nil
	}

	// 添加關聯
	ruleContact := models.RuleContact{
		RuleID:    ruleID,
		ContactID: contactID,
	}
	result = c.db.Create(&ruleContact)
	return result.Error
}

// ContactToResponse 將 Contact 轉換為 API 響應格式
func ContactToResponse(contact models.Contact) map[string]interface{} {
	return map[string]interface{}{
		"id":            hex.EncodeToString(contact.ID),
		"name":          contact.Name,
		"channel_type":  contact.ChannelType,
		"enabled":       contact.Enabled,
		"send_resolved": contact.SendResolved,
		"auto_apply":    contact.AutoApply,
		"max_retry":     contact.MaxRetry,
		"retry_delay":   contact.RetryDelay,
		"config":        contact.Config,
		"severities":    contact.Severities,
	}
}

// GetContactSeverities 獨立函數，用於獲取聯絡人的 severities
func (c *Client) GetContactSeverities(contactID []byte) ([]string, error) {
	var db_severities string

	err := c.db.Debug().Model(&models.Contact{}).
		Where("id = ? AND deleted_at IS NULL", contactID).
		Pluck("severities", &db_severities).Error
	if err != nil {
		return nil, ParseDBError(err)
	}

	// 處理空字串
	if db_severities == "" {
		return []string{}, nil
	}

	return strings.Split(db_severities, ","), nil
}
