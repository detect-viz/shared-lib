package mysql

import (
	"github.com/detect-viz/shared-lib/models"
)

// GetContactsByRuleID 獲取規則的通知管道
func (c *Client) GetContactsByRuleID(ruleID int64) ([]models.Contact, error) {
	var contacts []models.Contact
	err := c.db.Model(&models.Contact{}).
		Where("rule_id = ?", ruleID).
		Find(&contacts).Error
	return contacts, err
}

// IsUsedByRules 檢查通知管道是否被規則使用
func (c *Client) IsUsedByRules(id int64) (bool, error) {
	var count int64
	err := c.db.Model(&models.Rule{}).
		Where("contact_id = ?", id).
		Count(&count).Error

	return count > 0, err
}

// CreateContact 創建通知管道
func (c *Client) CreateContact(contact *models.Contact) error {
	return c.db.Create(contact).Error
}

// GetContact 獲取通知管道
func (c *Client) GetContact(id int64) (*models.Contact, error) {
	var contact models.Contact
	err := c.db.First(&contact, id).Error
	return &contact, err
}

// ListContacts 獲取通知管道列表
func (c *Client) ListContacts() ([]models.Contact, error) {
	var contacts []models.Contact
	err := c.db.Find(&contacts).Error
	return contacts, err
}

// UpdateContact 更新通知管道
func (c *Client) UpdateContact(contact *models.Contact) error {
	return c.db.Save(contact).Error
}

// DeleteContact 刪除通知管道
func (c *Client) DeleteContact(id int64) error {
	return c.db.Delete(&models.Contact{}, id).Error
}

// CheckContactName 檢查名稱是否重複
func (c *Client) CheckContactName(contact models.Contact) (bool, string) {
	var existingContact models.Contact
	result := c.db.Where(&models.Contact{
		RealmName: contact.RealmName,
		Name:      contact.Name,
	}).First(&existingContact)

	if result.Error == nil && existingContact.ID != contact.ID {
		return false, "名稱已存在"
	}
	return true, ""
}
