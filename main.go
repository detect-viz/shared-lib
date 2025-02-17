package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"shared-lib/alert"
	"shared-lib/config"
	"shared-lib/databases"
	"shared-lib/logger"
	"shared-lib/models"

	"go.uber.org/zap"
)

func main() {

	alertService := initService()

	file := models.FileInfo{
		Realm:    "master",
		Source:   "test",
		FileName: "test",
		Host:     "test",
	}

	metrics := readJsonFile("output/AL2SUB_08310400.csv.json")

	if err := alertService.ProcessFile(file, metrics); err != nil {
		fmt.Println("處理告警失敗:", err)
	}

	// 啟動 API 服務
	http.HandleFunc("/api/alert", handleAlert)
	fmt.Println("🚀 Alert API 服務啟動於 http://localhost:8081")
	http.ListenAndServe(":8081", nil)
}

func readJsonFile(filename string) map[string][]map[string]interface{} {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var result map[string][]map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	return result
}
func initService() *alert.Service {
	// 1️⃣ 讀取設定檔
	configManager := config.New()
	if err := configManager.Load("/etc/bimap-ipoc/config.yml"); err != nil {
		panic(err)
	}
	cfg := configManager.GetConfig()

	// 2️⃣ 初始化 logger
	logSvc, err := logger.NewLogger(&cfg.Logger)
	if err != nil {
		panic(err)
	}

	// 3️⃣ 初始化資料庫
	db := databases.NewDatabase(&cfg.Database, logSvc)

	// 4️⃣ 初始化各個 Service
	alertService := alert.NewService(cfg.Alert, logSvc, db)

	// 5️⃣ 啟動模組
	if err := alertService.Init(); err != nil {
		logSvc.Error("初始化告警服務失敗", zap.Error(err))
		return nil
	}

	return alertService
}

// API: 接收 `iPOC` 送來的數據並處理
func handleAlert(w http.ResponseWriter, r *http.Request) {

	// 4️⃣ 初始化各個 Service
	alertService := initService()

	var metrics map[string][]map[string]interface{}
	body, _ := io.ReadAll(r.Body)
	json.Unmarshal(body, &metrics)

	file := models.FileInfo{
		Realm:    "master",
		Source:   "api",
		FileName: "from-http",
		Host:     "remote-ipoc",
	}

	if err := alertService.ProcessFile(file, metrics); err != nil {
		http.Error(w, fmt.Sprintf("告警處理失敗: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "alert processed"}`))
}
