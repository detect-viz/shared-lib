package mysql

import (
	"fmt"
	"time"

	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/models"
	"github.com/google/wire"
	"go.uber.org/zap"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLSet 提供 gorm.DB
var MySQLSet = wire.NewSet(
	NewClient,
	wire.Bind(new(Database), new(*Client)),
)

// MySQL 資料庫連接器
type Client struct {
	config *models.DatabaseConfig
	db     *gorm.DB
	logger logger.Logger
}

// NewDatabase 創建新的資料庫連接
func NewClient(cfg *models.DatabaseConfig, logger logger.Logger) *Client {
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

	return &Client{
		config: cfg,
		db:     db,
		logger: logger,
	}
}

// Close 關閉資料庫連接
func (c *Client) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
