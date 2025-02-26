package mysql

import (
	"os"

	"github.com/detect-viz/shared-lib/models"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// GetTemplate 根據 `contactType` 獲取適用的模板
func (c *Client) GetTemplate(realm string, RuleState string, format string) (models.Template, error) {

	var template models.Template

	err := c.db.Where("format = ?", format).
		Where("is_default = ?", true).
		First(&template).Error

	if err != nil {
		c.logger.Warn("找不到適用的通知模板，將使用 text 格式",
			zap.String("realm", realm),
			zap.String("rule_state", RuleState),
			zap.String("format", format))
		// template.Format = "text"
		// template.Title = "【告警通知】"
		// template.Message = "未找到對應的模板，請聯繫管理員"
	}

	return template, nil
}

// 初始化時載入 YAML 預設模板
func (c *Client) LoadDefaultTemplates(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var templates []models.Template
	if err := yaml.Unmarshal(data, &templates); err != nil {
		return err
	}

	// 將模板寫入 DB，避免重複插入
	for _, tmpl := range templates {
		var count int64
		c.db.Model(&models.Template{}).Where("name = ?", tmpl.Name).Count(&count)
		if count == 0 {
			c.db.Create(&tmpl)
		}
	}

	c.logger.Info("成功載入 YAML 預設模板")
	return nil
}
func (c *Client) UpdateTemplate(id string, updatedTemplate models.Template) error {
	return c.db.Model(&models.Template{}).Where("id = ?", id).Updates(updatedTemplate).Error
}
