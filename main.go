package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/detect-viz/shared-lib/alert"
	"github.com/detect-viz/shared-lib/config"
	"github.com/detect-viz/shared-lib/databases"
	"github.com/detect-viz/shared-lib/logger"
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/mute"
	"github.com/detect-viz/shared-lib/notify"
	"github.com/detect-viz/shared-lib/scheduler"
	"github.com/detect-viz/shared-lib/templates"

	"go.uber.org/zap"
)

func main() {
	// 初始化服務
	alertService := initService()
	if alertService == nil {
		return
	}

	// 啟動 API 服務
	http.HandleFunc("/api/alert", handleAlert)
	fmt.Println("Alert API 服務啟動於 http://localhost:8081")
	http.ListenAndServe(":8081", nil)
}

func initService() *alert.Service {
	// 1️⃣ 讀取設定檔
	configManager := config.New()
	if err := configManager.Load("./config.yaml"); err != nil {
		panic(err)
	}
	cfg := configManager.GetConfig()

	// 2️⃣ 初始化 logger
	logSvc, err := logger.NewLogger(&cfg.Logger)
	if err != nil {
		panic(err)
	}

	// 1️⃣ 初始化排程器
	scheduler := scheduler.NewScheduler(&cfg.Scheduler, logSvc)

	// 2️⃣ 初始化日誌管理器
	logMgr := logger.NewLogRotator(logSvc)

	// 3️⃣ 初始化資料庫
	db := databases.NewDatabase(&cfg.Database, logSvc)

	// 3️⃣ 初始化通知服務
	notifySvc := notify.NewService()

	// 初始化 MuteService
	muteService := mute.NewService(
		db.GetDB(),         // 獲取 *gorm.DB 實例
		logSvc.GetLogger(), // 獲取 *zap.Logger 實例
	)

	templateSvc := templates.NewService(logSvc)
	// 4️⃣ 初始化告警服務
	alertService := alert.NewService(
		cfg.Alert,   // AlertConfig
		cfg.Mapping, // MappingConfig
		db,          // Database
		logSvc,      // Logger
		logMgr,      // LogRotator
		notifySvc,   // NotifyService
		scheduler,   // Scheduler
		muteService, // MuteService
		templateSvc, // TemplateService
	)
	if err := alertService.Init(); err != nil {
		logSvc.Error("初始化告警服務失敗", zap.Error(err))
		return nil
	}

	return alertService
}

// handleAlert 處理來自 iPOC 的告警數據
func handleAlert(w http.ResponseWriter, r *http.Request) {
	alertService := initService()
	if alertService == nil {
		http.Error(w, "服務初始化失敗", http.StatusInternalServerError)
		return
	}

	var metrics map[string][]map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "解析請求數據失敗", http.StatusBadRequest)
		return
	}

	file := models.FileInfo{
		Realm:    "master",
		Source:   "logman",
		FileName: "from-http",
		Host:     "SRVECCDV01",
	}

	if err := alertService.Process(file, metrics); err != nil {
		http.Error(w, fmt.Sprintf("告警處理失敗: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alert processed"})
}
