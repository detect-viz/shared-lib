package alert

import (
	"fmt"
	"os"

	"github.com/detect-viz/shared-lib/models"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// InitAlertRuleFromFile 初始化告警規則
func (s *Service) InitAlertRuleFromFile(configPath string) error {
	// 加載配置
	config, err := LoadRuleConfig(configPath)
	if err != nil {
		return fmt.Errorf("加載通知管道配置失敗: %w", err)
	}

	// 轉換並保存到資料庫
	for _, ruleDef := range config.Rules {
		rule := ruleDef.ToAlertRule()

		// 設置默認值
		rule.RealmName = "master"

		// 保存到資料庫
		if err := s.db.CreateOrUpdateAlertRule(rule); err != nil {
			s.logger.Error("保存告警規則失敗",
				zap.String("name", rule.Name),
				zap.Error(err))
			continue
		}
	}

	s.logger.Info("告警規則初始化完成",
		zap.Int("rule_count", len(config.Rules)))

	return nil
}

// InitContacts 初始化通知管道
func (s *Service) InitContactsFromFile(configPath string) error {
	// 加載配置
	config, err := LoadContactConfig(configPath)
	if err != nil {
		return fmt.Errorf("加載通知管道配置失敗: %w", err)
	}

	// 轉換並保存到資料庫
	for _, contactDef := range config.Contacts {
		contact := contactDef.ToAlertContact()

		// 設置默認值
		contact.RealmName = "master"

		// 保存到資料庫
		if err := s.db.CreateOrUpdateAlertContact(contact); err != nil {
			s.logger.Error("保存通知管道失敗",
				zap.String("name", contact.Name),
				zap.Error(err))
			continue
		}
	}

	s.logger.Info("通知管道初始化完成",
		zap.Int("contact_count", len(config.Contacts)))

	return nil
}

// LoadRuleConfig 從 YAML 文件加載告警規則配置
func LoadRuleConfig(filepath string) (*models.RuleConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var config models.RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadContactConfig 從 YAML 文件加載通知管道配置
func LoadContactConfig(filepath string) (*models.ContactConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var config models.ContactConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
