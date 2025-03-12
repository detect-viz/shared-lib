package mysql

import (
	"database/sql"
	"fmt"

	"github.com/detect-viz/shared-lib/models"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (c *Client) LoadAlertMigrate(migratePath string) error {
	host := c.config.MySQL.Host
	port := c.config.MySQL.Port
	user := c.config.MySQL.User
	password := c.config.MySQL.Password
	dbname := c.config.MySQL.DBName

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
		fmt.Println("❌ 遷移失敗", err)
		return err
	}

	c.logger.Info(
		fmt.Sprintf("alert migrate %v", mi.Up()),
	)
	return nil
}

// AlertMenuMigration 遷移
func (c *Client) AlertMenuMigration() error {
	// 1. 聯絡人表
	if err := c.db.AutoMigrate(&models.Contact{}); err != nil {
		return err
	}

	// 2. 規則與聯絡人關聯表
	if err := c.db.AutoMigrate(&models.RuleContact{}); err != nil {
		return err
	}

	return nil
}
