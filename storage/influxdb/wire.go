package influxdb

import (
	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/storage/influxdb/interfaces"
	v2 "github.com/detect-viz/shared-lib/storage/influxdb/v2"
	v3 "github.com/detect-viz/shared-lib/storage/influxdb/v3"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// 提供 InfluxDB 依賴
var InfluxDBSet = wire.NewSet(
	NewClient,
)

// NewClient 創建 InfluxDB 客戶端
func NewClient(cfg *models.InfluxDBConfig, logger logger.Logger) interfaces.Database {
	if cfg == nil {
		logger.Error("InfluxDB 配置為空")
		return nil
	}

	switch cfg.Version {
	case "v2":
		client, err := v2.NewClient(cfg)
		if err != nil {
			logger.Error("創建 InfluxDB v2 客戶端失敗",
				zap.String("url", cfg.URL),
				zap.Error(err))
			return nil
		}
		logger.Info("InfluxDB v2 客戶端連接成功",
			zap.String("url", cfg.URL))
		return client

	case "v3":
		client, err := v3.NewClient(cfg)
		if err != nil {
			logger.Error("創建 InfluxDB v3 客戶端失敗",
				zap.String("url", cfg.URL),
				zap.Error(err))
			return nil
		}
		logger.Info("InfluxDB v3 客戶端連接成功",
			zap.String("url", cfg.URL))
		return client

	default:
		logger.Error("不支援的 InfluxDB 版本",
			zap.String("version", cfg.Version))
		return nil
	}
}
