package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"shared-lib/alert"
	"shared-lib/config"
	"shared-lib/databases"
	"shared-lib/logger"
	"shared-lib/models"

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

	logMgr, err := logger.NewLogManager(&cfg.Logger)
	if err != nil {
		panic(err)
	}

	// 3️⃣ 初始化資料庫
	db := databases.NewDatabase(&cfg.Database, logSvc)

	// 4️⃣ 初始化告警服務
	alertService := alert.NewService(cfg.Alert, db, logSvc, logMgr)
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

	if err := alertService.ProcessFile(file, metrics); err != nil {
		http.Error(w, fmt.Sprintf("告警處理失敗: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alert processed"})
}
