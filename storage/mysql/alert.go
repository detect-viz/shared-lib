package mysql

import (
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/label"
)

// * 獲取所有告警規則，
func (c *Client) GetActiveRules(realm, resourceName string) ([]models.Rule, error) {
	var rules []models.Rule
	var resourceObjectIDs [][]byte
	tx := c.db.Begin()

	tx.Model(&models.Target{}).Where("realm_name = ? AND resource_name = ?", realm, resourceName).Pluck("id", &resourceObjectIDs)

	tx.Where("enabled = ? AND realm_name = ? AND target_id IN (?)", true, realm, resourceObjectIDs).Preload("Target").Find(&rules)

	err := tx.Commit().Error

	if err != nil {
		return nil, err
	}
	return rules, nil
}

// * 獲取所有活動的告警規則，不限制 realm 和 resourceName
func (c *Client) GetAllActiveRules() ([]models.Rule, error) {
	var rules []models.Rule

	// 使用分頁查詢，避免一次性載入過多數據
	pageSize := 1000
	page := 0

	for {
		var pageRules []models.Rule
		err := c.db.Where("enabled = ?", true).
			Preload("Target").
			Offset(page * pageSize).
			Limit(pageSize).
			Find(&pageRules).Error

		if err != nil {
			return nil, err
		}

		rules = append(rules, pageRules...)

		// 如果獲取的數量小於頁面大小，說明已經獲取完所有數據
		if len(pageRules) < pageSize {
			break
		}

		page++
	}

	return rules, nil
}

// 獲取資源群組名稱
func (c *Client) GetResourceGroupName(id []byte) (string, error) {
	var name string
	return name, c.db.Table("resource_groups").Where("id = ?", id).Select("name").Scan(&name).Error
}

// 獲取規則的聯絡人列表
func (c *Client) GetRuleContacts(ruleID []byte) ([]models.Contact, error) {
	var contactIDs [][]byte
	var contacts []models.Contact

	// 1. 獲取聯絡人 IDs
	err := c.db.Model(&models.RuleContact{}).
		Where("rule_id = ?", ruleID).
		Pluck("contact_id", &contactIDs).Error
	if err != nil {
		return nil, err
	}

	// 2. 獲取聯絡人詳情
	err = c.db.Model(&models.Contact{}).
		Where("id IN (?)", contactIDs).
		Find(&contacts).Error
	if err != nil {
		return nil, err
	}

	// 3. 為每個聯絡人獲取 Severities
	for i, contact := range contacts {
		// 直接從資料庫獲取 Contact 的 Severities
		var contactWithSeverities models.Contact
		err = c.db.Model(&models.Contact{}).
			Select("severities").
			Where("id = ?", contact.ID).
			First(&contactWithSeverities).Error
		if err != nil {
			return nil, err
		}

		contacts[i].Severities = contactWithSeverities.Severities
	}

	return contacts, nil
}

// 獲取規則的標籤
func (c *Client) GetRuleLabelByRuleID(ruleID []byte) (map[string]string, error) {
	// 假設有一個關聯表 rule_labels 連接 rules 和 label_values
	type RuleLabel struct {
		RuleID       []byte `gorm:"type:binary(16);primaryKey"`
		LabelValueID []byte `gorm:"type:binary(16);primaryKey"`
	}

	var labelValueIDs [][]byte
	// 先從關聯表獲取 label_value_id
	err := c.db.Model(&RuleLabel{}).
		Where("rule_id = ?", ruleID).
		Pluck("label_value_id", &labelValueIDs).Error
	if err != nil {
		return nil, err
	}

	// 然後獲取 label_values 及其關聯的 label_keys
	var labelValues []label.LabelValue
	err = c.db.Preload("LabelKey").
		Where("id IN (?)", labelValueIDs).
		Find(&labelValues).Error
	if err != nil {
		return nil, err
	}

	// 構建結果 map
	result := make(map[string]string)
	for _, lv := range labelValues {
		result[lv.LabelKey.KeyName] = lv.Value
	}

	return result, nil
}
