package databases

import (
	"errors"
	"fmt"
	"time"

	"shared-lib/models"

	"shared-lib/interfaces"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQL struct {
	db *gorm.DB
}

func NewMySQL(db *gorm.DB) *MySQL {
	return &MySQL{db: db}
}

func NewDatabase(cfg *models.DatabaseConfig, logger interfaces.Logger) *MySQL {
	// get env config
	host := cfg.MySQL.Host
	port := cfg.MySQL.Port
	user := cfg.MySQL.User
	password := cfg.MySQL.Password
	params := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, cfg.MySQL.DBName)

	err := errors.New("mock error")
	var db *gorm.DB
	for err != nil {
		db, err = gorm.Open(mysql.Open(params), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		time.Sleep(1 * time.Second)

		if cfg.MySQL.Level == "debug" {
			db = db.Debug()
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error(
			fmt.Sprintf("mysql [%v] connection error: %v", params, err.Error()),
			zap.String("service", "mysql"),
		)
		return nil
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(int(cfg.MySQL.MaxIdle))

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(int(cfg.MySQL.MaxOpen))

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	lifeTime, _ := time.ParseDuration(cfg.MySQL.MaxLife)
	sqlDB.SetConnMaxLifetime(lifeTime)

	logger.Info(
		fmt.Sprintf("mysql [%v] connection success", cfg.MySQL.DBName),
		zap.String("service", "mysql"),
	)

	return &MySQL{db: db}
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

// GetCustomLabels 獲取規則相關的標籤
func (m *MySQL) GetCustomLabels(ruleID int64) (map[string]string, error) {
	var labels []models.Label
	err := m.db.Where("resource_id = ? AND type = ?", ruleID, "rule").Find(&labels).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, label := range labels {
		result[label.Key] = label.Value
	}
	return result, nil
}

// GetAlertState 獲取告警狀態
func (m *MySQL) GetAlertState(ruleID int64, resourceName string, metricName string) (models.AlertState, error) {
	var state models.AlertState
	err := m.db.Where(
		"rule_id = ? AND resource_name = ? AND metric_name = ?",
		ruleID, resourceName, metricName,
	).First(&state).Error

	if err == gorm.ErrRecordNotFound {
		// 如果不存在，返回新的狀態
		return models.AlertState{
			RuleID:       ruleID,
			ResourceName: resourceName,
			MetricName:   metricName,
		}, nil
	}
	return state, err
}

// SaveAlertState 保存告警狀態
func (m *MySQL) SaveAlertState(state models.AlertState) error {
	// 使用 Upsert 操作
	return m.db.Save(&state).Error
}

// GetAlertContacts 獲取規則的聯絡人列表
func (m *MySQL) GetAlertContacts(ruleID int64) ([]models.AlertContact, error) {
	var contacts []models.AlertContact

	// 1. 查詢規則關聯的聯絡人
	err := m.db.Table("alert_rule_contacts").
		Select("alert_contacts.*").
		Joins("LEFT JOIN alert_contacts ON alert_rule_contacts.contact_id = alert_contacts.id").
		Where("alert_rule_contacts.rule_id = ? AND alert_contacts.status = ?", ruleID, "enabled").
		Find(&contacts).Error

	if err != nil {
		return nil, fmt.Errorf("查詢聯絡人失敗: %v", err)
	}

	return contacts, nil
}

// WriteTriggeredLog 寫入觸發日誌
func (m *MySQL) WriteTriggeredLog(trigger models.TriggerLog) error {
	return m.db.Create(&trigger).Error
}

// WriteNotificationLog 寫入通知日誌
func (m *MySQL) WriteNotificationLog(notification models.NotificationLog) error {
	// 開啟交易
	tx := m.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 1. 寫入通知日誌
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. 寫入關聯表
	for _, trigger := range notification.TriggerLogs {
		relation := models.NotificationTriggerLog{
			NotificationID: notification.ID,
			TriggerLogID:   int64(trigger.ID),
		}
		if err := tx.Create(&relation).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交交易
	return tx.Commit().Error
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
	return rules, m.db.Where("realm = ?", realm).Find(&rules).Error
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
