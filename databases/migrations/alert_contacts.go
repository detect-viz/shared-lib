package migrations

import (
	"shared-lib/models"

	"gorm.io/gorm"
)

// AlertContactsMigration 聯絡人相關表格遷移
func AlertContactsMigration(db *gorm.DB) error {
	// 1. 聯絡人表
	if err := db.AutoMigrate(&models.AlertContact{}); err != nil {
		return err
	}

	// 2. 規則與聯絡人關聯表
	if err := db.AutoMigrate(&models.AlertRuleContact{}); err != nil {
		return err
	}

	// 3. 觸發日誌表
	if err := db.AutoMigrate(&models.TriggerLog{}); err != nil {
		return err
	}

	// 4. 通知日誌表
	if err := db.AutoMigrate(&models.NotificationLog{}); err != nil {
		return err
	}

	// 5. 通知與觸發日誌關聯表
	if err := db.AutoMigrate(&models.NotificationTriggerLog{}); err != nil {
		return err
	}

	return nil
}
