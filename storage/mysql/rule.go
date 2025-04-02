package mysql

import (
	"github.com/detect-viz/shared-lib/apierrors"
	"github.com/detect-viz/shared-lib/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// * 創建規則 + Alert State
func (c *Client) CreateRule(rule *models.Rule) (*models.Rule, error) {
	rule.ID = GenerateUUID16()
	err := c.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(rule).Error; err != nil {
			return ParseDBError(err)
		}

		//* 檢查 RuleState 是否存在
		var exists bool
		err := tx.Model(&models.RuleState{}).
			Where("rule_id = ?", rule.ID).
			Select("count(*) > 0").
			Find(&exists).Error
		if err != nil {
			return ParseDBError(err)
		}

		if !exists {
			ruleState := models.RuleState{
				RuleID: rule.ID,
			}
			if err := tx.Select("*").Create(&ruleState).Error; err != nil {
				return ParseDBError(err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return rule, nil
}

// 獲取規則
func (c *Client) GetRule(id []byte) (*models.Rule, error) {
	var rule models.Rule
	err := c.db.Preload(clause.Associations).First(&rule, "id = ?", id).Error
	if err != nil {
		return nil, ParseDBError(err)
	}
	return &rule, nil
}

// 獲取規則列表
func (c *Client) ListRules(realm string, cursor int64, limit int) ([]models.Rule, int64, error) {
	var rules []models.Rule
	// 只回傳 UI 需要的欄位
	query := c.db.Model(&models.Rule{}).
		Where("realm_name = ?", realm)

	if cursor > 0 {
		query = query.Where("created_at > ?", cursor)
	}

	err := query.Order("created_at ASC").
		Limit(limit).
		Find(&rules).Error

	if err != nil {
		return nil, 0, ParseDBError(err)
	}

	// 計算 next_cursor
	nextCursor := int64(-1)
	if len(rules) > 0 && len(rules) >= limit {
		var lastCreatedAt int64
		c.db.Model(&models.Rule{}).
			Where("id = ?", rules[len(rules)-1].ID).
			Pluck("created_at", &lastCreatedAt)
		nextCursor = lastCreatedAt
	}

	return rules, nextCursor, nil
}

// 更新規則
func (c *Client) UpdateRule(rule *models.Rule) (*models.Rule, error) {
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// 1. 檢查規則是否存在
		var exists bool
		err := tx.Model(&models.Rule{}).
			Where("id = ?", rule.ID).
			Select("count(*) > 0").
			Find(&exists).Error
		if err != nil {
			return ParseDBError(err)
		}

		if !exists {
			return apierrors.ErrNotFound
		}

		// 2. 更新 Rule 本身（排除關聯）
		if err := tx.Omit("Contacts").Model(&models.Rule{}).
			Where("id = ?", rule.ID).
			Updates(map[string]interface{}{
				"target_id":       rule.TargetID,
				"metric_rule_uid": rule.MetricRuleUID,
				"create_type":     rule.CreateType,
				"auto_apply":      rule.AutoApply,
				"enabled":         rule.Enabled,
				"info_threshold":  rule.InfoThreshold,
				"warn_threshold":  rule.WarnThreshold,
				"crit_threshold":  rule.CritThreshold,
				"times":           rule.Times,
				"duration":        rule.Duration,
				"silence_period":  rule.SilencePeriod,
			}).Error; err != nil {
			return ParseDBError(err)
		}

		// 3. 更新 Contacts 關聯
		if err := tx.Model(rule).Association("Contacts").Replace(rule.Contacts); err != nil {
			return ParseDBError(err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 重新查詢最新的 Rule，確保返回最新數據
	return c.GetRule(rule.ID)
}

// 刪除規則
func (c *Client) DeleteRule(id []byte) error {
	result := c.db.Delete(&models.Rule{}, "id = ?", id)
	if result.RowsAffected == 0 {
		return apierrors.ErrNotFound
	}
	if result.Error != nil {
		return ParseDBError(result.Error)
	}
	return nil
}

// 獲取自動匹配的規則
func (c *Client) GetAutoApplyRulesByMetricRuleUIDs(realm string, metricRuleUIDs []string) ([]models.Rule, error) {
	var rules []models.Rule

	err := c.db.Where("realm_name = ? AND metric_rule_uid IN ? AND enabled = ? AND auto_apply = ?", realm, metricRuleUIDs, true, true).Find(&rules).Error

	if err != nil {
		return nil, ParseDBError(err)
	}

	return rules, nil
}

// 獲取與特定 target 相關的規則
func (c *Client) GetRulesByTarget(realm, resourceName, partitionName string) ([]models.Rule, error) {
	var rules []models.Rule

	// 首先獲取符合條件的 target
	var target models.Target
	err := c.db.Where("realm_name = ? AND resource_name = ? AND partition_name = ?",
		realm, resourceName, partitionName).First(&target).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []models.Rule{}, nil
		}
		return nil, ParseDBError(err)
	}

	// 然後獲取與該 target 相關的規則
	err = c.db.Preload(clause.Associations).
		Where("target_id = ? AND enabled = ?", target.ID, true).
		Find(&rules).Error

	if err != nil {
		return nil, ParseDBError(err)
	}

	return rules, nil
}

func (c *Client) CreateRules(rules []models.Rule) error {
	// 為每個規則生成 ID
	for i := range rules {
		rules[i].ID = GenerateUUID16()
	}

	err := c.db.Transaction(func(tx *gorm.DB) error {
		// 創建規則
		if err := tx.Create(&rules).Error; err != nil {
			return ParseDBError(err)
		}

		// 為每個規則創建 RuleState
		for _, rule := range rules {
			ruleState := models.RuleState{
				RuleID:       rule.ID,
				State:        "normal",
				ContactState: "normal",
			}
			if err := tx.Create(&ruleState).Error; err != nil {
				return ParseDBError(err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
