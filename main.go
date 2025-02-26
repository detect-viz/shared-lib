// @title Shared-lib API
// @version 1.0
// @description 提供告警、通知、規則管理等功能
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
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

	// ✅ 3. 統一初始化 Database
	dbcfg := cfg.GetDatabaseConfig()
	db := mysql.NewClient(&dbcfg, log)
	defer db.Close()

	// ✅ 4. 統一初始化 Keycloak
	keycloakcfg := cfg.GetKeycloakConfig()
	keycloakC, err := keycloak.NewClient(&keycloakcfg)
	if err != nil {
		fmt.Println("初始化 Keycloak 失敗:", err)
		return
	}

	alertService, err := alert.InitializeAlertService(
		cfg.GetAlertConfig(),
		cfg.GetGlobalConfig(),
		db,
		log,
		keycloakC.(*keycloak.Client),
	)
	if err != nil {
		log.Error("初始化告警服務失敗", zap.Error(err))
		return
	}

	// 註冊 Alert API
	r := api.RegisterRoutes(alertService)

	// 啟動 API 服務
	r.Run(":8080")
}
