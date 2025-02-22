package databases

import (
	"fmt"
	"time"

	"github.com/detect-viz/shared-lib/models"

	"github.com/detect-viz/shared-lib/interfaces"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MySQL 資料庫連接器
type MySQL struct {
	config *models.DatabaseConfig
	db     *gorm.DB
	logger interfaces.Logger
}

// NewDatabase 創建新的資料庫連接
func NewDatabase(cfg *models.DatabaseConfig, logger interfaces.Logger) *MySQL {
	if cfg == nil {
		logger.Error("資料庫配置為空")
		return nil
	}

	// 構建連接字串
	params := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.MySQL.User,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.DBName,
	)

	// 建立連接
	db, err := gorm.Open(mysql.Open(params), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		logger.Error("資料庫連接失敗",
			zap.String("host", cfg.MySQL.Host),
			zap.Error(err))
		return nil
	}

	// 設置連接池
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("獲取資料庫實例失敗", zap.Error(err))
		return nil
	}

	sqlDB.SetMaxIdleConns(int(cfg.MySQL.MaxIdle))
	sqlDB.SetMaxOpenConns(int(cfg.MySQL.MaxOpen))

	if lifeTime, err := time.ParseDuration(cfg.MySQL.MaxLife); err == nil {
		sqlDB.SetConnMaxLifetime(lifeTime)
	}

	logger.Info("資料庫連接成功",
		zap.String("host", cfg.MySQL.Host),
		zap.String("database", cfg.MySQL.DBName))

	return &MySQL{
		config: cfg,
		db:     db,
		logger: logger,
	}
}

// GetMetricRule 獲取指標規則定義
func (m *MySQL) GetMetricRule(id int64) (models.MetricRule, error) {
	var rule models.MetricRule
	err := m.db.First(&rule, id).Error
	return rule, err
}

// GetAlertRuleDetails 獲取告警規則詳情
func (m *MySQL) GetAlertRuleDetails(ruleID int64) ([]models.AlertRuleDetail, error) {
	var details []models.AlertRuleDetail
	err := m.db.Where("alert_rule_id = ?", ruleID).Find(&details).Error
	return details, err
}

// 獲取規則相關的標籤
func (m *MySQL) GetLabels(ruleID int64) (map[string]string, error) {
	var ruleLabelIDs []int64
	err := m.db.Model(&models.AlertRuleLabel{}).Where("rule_id = ?", ruleID).Pluck("label_id", &ruleLabelIDs).Error
	if err != nil {
		return nil, err
	}
	var labels []models.Label
	err = m.db.Model(&models.Label{}).Where("id IN (?)", ruleLabelIDs).Find(&labels).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, label := range labels {
		result[label.Key] = label.Value
	}
	return result, nil
}

// GetAlertContacts 獲取規則的聯絡人列表
func (m *MySQL) GetAlertContacts(ruleID int64) ([]models.AlertContact, error) {
	var contactIDs []int64
	var contacts []models.AlertContact

	// 1. 獲取聯絡人 IDs
	err := m.db.Model(&models.AlertRuleContact{}).
		Where("alert_rule_id = ?", ruleID).
		Pluck("alert_contact_id", &contactIDs).Error
	if err != nil {
		return nil, err
	}

	// 2. 獲取聯絡人詳情並預加載 Severities
	err = m.db.Model(&models.AlertContact{}).
		Preload(clause.Associations).
		Where("id IN (?)", contactIDs).
		Find(&contacts).Error

	for i, contact := range contacts {
		var lvl []models.AlertContactSeverity
		err = m.db.Model(&models.AlertContactSeverity{}).
			Where("alert_contact_id = ?", contact.ID).
			Find(&lvl).Error
		if err != nil {
			return nil, err
		}

		contacts[i].Severities = lvl
	}

	return contacts, err
}

// Close 關閉資料庫連接
func (db *MySQL) Close() error {
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateAlertRule 創建告警規則
func (m *MySQL) CreateAlertRule(rule *models.AlertRule) error {
	return m.db.Create(rule).Error
}

// CreateOrUpdateAlertRule 創建或更新告警規則
func (m *MySQL) CreateOrUpdateAlertRule(rule *models.AlertRule) error {
	return m.db.Save(rule).Error
}

// CreateOrUpdateAlertContact 創建或更新通知管道
func (m *MySQL) CreateOrUpdateAlertContact(contact *models.AlertContact) error {
	return m.db.Save(contact).Error
}

// CreateContact 創建聯絡人
func (m *MySQL) CreateContact(contact *models.AlertContact) error {
	return m.db.Create(contact).Error
}

// DeleteAlertRule 刪除告警規則
func (m *MySQL) DeleteAlertRule(ruleID int64) error {
	return m.db.Delete(&models.AlertRule{}, ruleID).Error
}

// DeleteContact 刪除聯絡人
func (m *MySQL) DeleteContact(contactID int64) error {
	return m.db.Delete(&models.AlertContact{}, contactID).Error
}

// GetAlertRuleByID 獲取告警規則
func (m *MySQL) GetAlertRuleByID(ruleID int64) (models.AlertRule, error) {
	var rule models.AlertRule
	return rule, m.db.First(&rule, ruleID).Error
}

// GetAlertRulesByRealm 獲取指定 Realm 的告警規則
func (m *MySQL) GetAlertRulesByRealm(realm string) ([]models.AlertRule, error) {
	var rules []models.AlertRule
	return rules, m.db.Where("realm_name = ?", realm).Find(&rules).Error
}

// UpdateAlertRule 更新告警規則
func (m *MySQL) UpdateAlertRule(rule *models.AlertRule) error {
	return m.db.Save(rule).Error
}

// UpdateContact 更新聯絡人
func (m *MySQL) UpdateContact(contact *models.AlertContact) error {
	return m.db.Save(contact).Error
}

// GetContactByID 獲取聯絡人
func (m *MySQL) GetContactByID(id int64) (models.AlertContact, error) {
	var contact models.AlertContact
	return contact, m.db.First(&contact, id).Error
}

// GetAlertRules 獲取所有告警規則，並按 realm 分組
func (m *MySQL) GetAlertRules() (map[string][]models.AlertRule, error) {
	var rules []models.AlertRule

	err := m.db.Preload(clause.Associations).
		Where("enabled = ? AND deleted_at IS NULL", true).
		Find(&rules).Error

	if err != nil {
		return nil, err
	}

	// 按 realm 分組
	rulesByRealm := make(map[string][]models.AlertRule)
	for _, rule := range rules {
		rulesByRealm[rule.RealmName] = append(rulesByRealm[rule.RealmName], rule)
	}

	return rulesByRealm, nil
}

// 獲取資源群組名稱
func (m *MySQL) GetResourceGroupName(id int64) (string, error) {
	var name string
	return name, m.db.Table("resource_groups").Where("id = ?", id).Select("name").Scan(&name).Error
}

// UpdateNotifyLog 更新通知日誌
func (m *MySQL) UpdateNotifyLog(notify models.NotifyLog) error {
	return m.db.Save(&notify).Error
}

// GetDB 獲取原始資料庫實例
func (m *MySQL) GetDB() *gorm.DB {
	return m.db
}
