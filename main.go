// @title Shared-lib API
// @version 1.0
// @description 提供告警、通知、規則管理等功能
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @host localhost:8080
// @name Authorization

package main

import (
	"fmt"

	"github.com/detect-viz/shared-lib/alert"
	"github.com/detect-viz/shared-lib/api"
	"github.com/detect-viz/shared-lib/auth/keycloak"

	"github.com/detect-viz/shared-lib/infra/config"
	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/storage/mysql"
	"go.uber.org/zap"
)

func main() {
	// ✅ 1. 統一初始化 Config
	cfg := config.NewConfigManager()

	// ✅ 2. 統一初始化 Logger
	logcfg := cfg.GetLoggerConfig()
	log, err := logger.NewService(&logcfg, logger.WithCallerSkip(1))
	if err != nil {
		fmt.Println("初始化 Logger 失敗:", err)
		return
	}

	// ✅ 3. 統一初始化 Keycloak
	keycloakConfig := cfg.GetKeycloakConfig()
	keycloakClient, err := keycloak.NewClient(&keycloakConfig)
	if err != nil {
		log.Error("❌ InitializeKeycloakClient 發生錯誤", zap.Error(err))
		panic(err)
	}

	// ✅ 4. 統一初始化 Database
	dbcfg := cfg.GetDatabaseConfig()
	db := mysql.NewClient(&dbcfg, log)
	defer db.Close()

	alertService, err := alert.InitializeAlertService(
		cfg.GetAlertConfig(),
		cfg.GetGlobalConfig(),
		db,
		log,
		keycloakClient.(*keycloak.Client),
	)
	if err != nil {
		log.Error("❌ InitializeAlertService 發生錯誤", zap.Error(err))
		panic(err)
	}
	if alertService == nil {
		panic("🚨 alertService 是 nil，請檢查 wire 依賴！")
	}
	keycloakC := keycloakClient.(*keycloak.Client)
	// 註冊 Alert API
	r := api.RegisterRoutes(alertService, *keycloakC)

	// 啟動 API 服務
	r.Run(":8080")
}
