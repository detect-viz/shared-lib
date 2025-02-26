package mysql

import (
	"fmt"

	"github.com/detect-viz/shared-lib/models"
	"gorm.io/gorm/clause"
)

// GetMetricRule 獲取指標規則定義
func (c *Client) GetMetricRule(id int64) (models.MetricRule, error) {
	var rule models.MetricRule
	err := c.db.First(&rule, id).Error
	return rule, err
}

// GetAlertRuleDetails 獲取告警規則詳情
func (c *Client) GetAlertRuleDetails(ruleID int64) ([]models.AlertRuleDetail, error) {
	var details []models.AlertRuleDetail
	err := c.db.Where("alert_rule_id = ?", ruleID).Find(&details).Error
	return details, err
}

// GetRules 獲取所有告警規則，並按 realm 分組
func (c *Client) GetRules() (map[string][]models.Rule, error) {
	var rules []models.Rule

	err := c.db.Preload(clause.Associations).
		Where("enabled = ? AND deleted_at IS NULL", true).
		Find(&rules).Error

	if err != nil {
		return nil, err
	}

	// 按 realm 分組
	rulesByRealm := make(map[string][]models.Rule)
	for _, rule := range rules {
		rulesByRealm[rule.RealmName] = append(rulesByRealm[rule.RealmName], rule)
	}

	return rulesByRealm, nil
}

// 獲取資源群組名稱
func (c *Client) GetResourceGroupName(id int64) (string, error) {
	var name string
	return name, c.db.Table("resource_groups").Where("id = ?", id).Select("name").Scan(&name).Error
}

// CreateOrUpdateAlertRule 創建或更新告警規則
func (c *Client) CreateOrUpdateAlertRule(rule *models.Rule) error {
	return c.db.Save(rule).Error
}

// CreateOrUpdateAlertContact 創建或更新通知管道
func (c *Client) CreateOrUpdateAlertContact(contact *models.Contact) error {
	return c.db.Save(contact).Error
}

// 獲取規則的聯絡人列表
func (c *Client) GetRuleContacts(ruleID int64) ([]models.Contact, error) {
	var contactIDs []int64
	var contacts []models.Contact

	// 1. 獲取聯絡人 IDs
	err := c.db.Model(&models.AlertRuleContact{}).
		Where("alert_rule_id = ?", ruleID).
		Pluck("alert_contact_id", &contactIDs).Error
	if err != nil {
		return nil, err
	}

	// 2. 獲取聯絡人詳情並預加載 Severities
	err = c.db.Model(&models.Contact{}).
		Preload(clause.Associations).
		Where("id IN (?)", contactIDs).
		Find(&contacts).Error

	for i, contact := range contacts {
		var lvl []models.AlertContactSeverity
		err = c.db.Model(&models.AlertContactSeverity{}).
			Where("alert_contact_id = ?", contact.ID).
			Find(&lvl).Error
		if err != nil {
			return nil, err
		}

		contacts[i].Severities = lvl
	}

	return contacts, err
}

// GetLabelByRuleID 獲取規則的標籤
func (c *Client) GetRuleLabels(ruleID int64) (map[string]string, error) {
	var labelIDs []int64
	err := c.db.Model(&models.AlertRuleLabel{}).Where("rule_id = ?", ruleID).Pluck("label_id", &labelIDs).Error
	if err != nil {
		return nil, err
	}
	var labels []models.Label
	err = c.db.Where("id IN (?)", labelIDs).Find(&labels).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, label := range labels {
		result[label.KeyName] = label.Value.String()
	}
	return result, nil
}

// CreateRule 創建抑制規則
func (c *Client) CreateRule(rule *models.Rule) error {
	return c.db.Create(rule).Error
}

// GetRule 獲取抑制規則
func (c *Client) GetRule(id int64) (*models.Rule, error) {
	var rule models.Rule
	if err := c.db.Preload("AlertRuleDetails").First(&rule, id).Error; err != nil {
		return nil, fmt.Errorf("獲取抑制規則失敗: %w", err)
	}
	return &rule, nil
}

// ListRules 獲取抑制規則列表
func (c *Client) ListRules(realm string) ([]models.Rule, error) {
	var rules []models.Rule
	if err := c.db.Preload("AlertRuleDetails").Where("realm_name = ?", realm).Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("獲取抑制規則列表失敗: %w", err)
	}
	return rules, nil
}

// UpdateRule 更新抑制規則
func (c *Client) UpdateRule(rule *models.Rule) error {
	return c.db.Model(rule).Updates(rule).Error
}

// DeleteRule 刪除抑制規則
func (c *Client) DeleteRule(id int64) error {
	return c.db.Delete(&models.Rule{}, id).Error
}
