package databases

import (
	"database/sql"
	"fmt"

	"github.com/detect-viz/shared-lib/models"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (m *MySQL) LoadAlertMigrate(migratePath string) error {
	host := m.config.MySQL.Host
	port := m.config.MySQL.Port
	user := m.config.MySQL.User
	password := m.config.MySQL.Password
	dbname := m.config.MySQL.DBName

	db, _ := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, dbname))
	driver, _ := mysql.WithInstance(db, &mysql.Config{
		MigrationsTable: "alert_schema_migrations",
	})

	mi, err := migrate.NewWithDatabaseInstance(
		migratePath,
		"mysql",
		driver,
	)
	if err != nil {
		return err
	}

	m.logger.Info(
		fmt.Sprintf("alert migrate %v", mi.Up()),
	)
	return nil
}

// AlertMenuMigration 遷移
func (m *MySQL) AlertMenuMigration() error {
	// 1. 聯絡人表
	if err := m.db.AutoMigrate(&models.AlertContact{}); err != nil {
		return err
	}

	// 2. 規則與聯絡人關聯表
	if err := m.db.AutoMigrate(&models.AlertRuleContact{}); err != nil {
		return err
	}

	return nil
}
