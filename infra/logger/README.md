✅ RotateSetting & RotateTask 設計最佳化

你的 RotateSetting & RotateTask 設計已經涵蓋 日誌/檔案輪轉 所需的 壓縮、清理、容量管理 需求，且有 優先級判斷 (MinDiskFreeMB > MaxSizeMB > MaxAge)，這樣的設計是正確的。

🔥 1️⃣ RotateSetting 是否足夠完整？

✅ Enabled → 控制是否啟用檔案輪轉
✅ Schedule → 使用 cron 來定時執行輪轉
✅ CompressEnabled → 控制是否壓縮舊日誌
✅ CompressMatchRegex → 指定要壓縮的文件匹配規則 (*.log / *.csv)
✅ CompressOffsetHours → 壓縮多少小時前的日誌 (24h 前的壓縮)
✅ CompressSaveRegex → 指定壓縮後保留的文件格式 (*.tar.gz / *.zip)
✅ MaxAge → 刪除超過 N 天的文件 (7 天前的刪除)
✅ MaxSizeMB → 限制單個目錄的最大容量 (超過 5GB 就刪除舊檔案)
✅ MinDiskFreeMB → 如果磁碟空間過低 (小於 1GB)，則刪除舊檔案

🔥 2️⃣ RotateTask 是否需要 Path？

是的，因為 不同模組的日誌路徑不同，所以 RotateTask 需要 Path 來區分 不同模組的日誌清理設定。

📌 ✅ 你的 RotateTask 已經符合最佳實踐
	•	Path → 確保不同模組 (Notify, Alert, Parser) 可設定不同目錄
	•	RotateSetting → 讓 logger 內部統一管理 壓縮 & 清理

🔥 3️⃣ RotateTask 設計更新

📌 加入 JobID 以便 Logger 進行管理


🔥 4️⃣ Logger 註冊 RotateTask

📌 logger/scheduler.go

package logger

import (
	"fmt"
	"time"
)

// 內部維護所有的 `RotateTask`
var rotateTasks = []RotateTask{}

// 註冊 `RotateTask`
func RegisterRotateTask(task RotateTask) {
	rotateTasks = append(rotateTasks, task)
	fmt.Printf("✅ 註冊檔案輪轉任務: [%s] 目錄=%s, 排程=%s\n", task.JobID, task.Path, task.RotateSetting.Schedule)
}

// 啟動排程
func StartRotateScheduler() {
	for _, task := range rotateTasks {
		go runRotateTask(task)
	}
}

// 執行 `RotateTask`
func runRotateTask(task RotateTask) {
	for {
		time.Sleep(time.Hour) // 每小時執行一次 (可改 cron job)
		fmt.Printf("🔄 執行檔案輪轉: [%s] 目錄=%s\n", task.JobID, task.Path)

		// 1️⃣ 檢查磁碟空間
		if isLowDiskSpace(task.RotateSetting.MinDiskFreeMB) {
			deleteOldFiles(task.Path, task.RotateSetting.MaxAge)
		}

		// 2️⃣ 檢查目錄大小
		if isExceedMaxSize(task.Path, task.RotateSetting.MaxSizeMB) {
			deleteOldFiles(task.Path, task.RotateSetting.MaxAge)
		}

		// 3️⃣ 壓縮舊日誌
		if task.RotateSetting.CompressEnabled {
			compressOldLogs(task.Path, task.RotateSetting.CompressMatchRegex, task.RotateSetting.CompressOffsetHours)
		}
	}
}

// **判斷磁碟空間是否低於 `MinDiskFreeMB`**
func isLowDiskSpace(minDiskMB int64) bool {
	// (模擬) 假設磁碟剩餘 800MB
	diskFreeMB := 800
	return diskFreeMB < int(minDiskMB)
}

// **判斷目錄大小是否超過 `MaxSizeMB`**
func isExceedMaxSize(path string, maxSizeMB int64) bool {
	// (模擬) 假設目錄大小 6GB
	dirSizeMB := 6000
	return dirSizeMB > int(maxSizeMB)
}

// **刪除超過 `MaxAge` 的舊日誌**
func deleteOldFiles(path string, maxAge int) {
	fmt.Printf("🗑 刪除超過 %d 天的日誌: %s\n", maxAge, path)
}

// **壓縮超過 `CompressOffsetHours` 的日誌**
func compressOldLogs(path, pattern string, hoursAgo int) {
	fmt.Printf("📦 壓縮 %d 小時前的日誌: %s (匹配: %s)\n", hoursAgo, path, pattern)
}

🔥 5️⃣ NotifyService 註冊 RotateTask

📌 notify/service.go

package notify

import (
	"fmt"
	"shared-lib/logger"
	"shared-lib/models"
)

// NotifyService 負責通知管理
type NotifyService struct {
	Config models.NotifyConfig
}

// 初始化 Notify Service
func NewNotifyService(config models.NotifyConfig) *NotifyService {
	service := &NotifyService{
		Config: config,
	}

	// 註冊日誌輪轉
	rotateTask := logger.RotateTask{
		JobID: "notify_log_rotation",
		Path:  config.NotifyPath,
		RotateSetting: models.RotateSetting{
			Enabled:             true,
			Schedule:            "0 3 * * *", // 每天 3 點清理
			CompressEnabled:     true,
			CompressMatchRegex:  "*.log",
			CompressOffsetHours: 24,
			CompressSaveRegex:   "notify_backup.tar.gz",
			MaxAge:              7,
			MaxSizeMB:           5000,
			MinDiskFreeMB:       1024,
		},
	}
	logger.RegisterRotateTask(rotateTask)

	fmt.Println("✅ NotifyService 初始化完成")
	return service
}

🔥 6️⃣ config.yaml 配置範例

notify:
  max_retry: 3
  retry_interval: 10
  notify_path: "/var/log/ipoc/notify"
  backup_path: "/var/log/ipoc/backup"
  failed_path: "/var/log/ipoc/failed"

  rotate:
    enabled: true
    schedule: "0 3 * * *"
    compress_enabled: true
    compress_match_regex: "*.log"
    compress_offset_hours: 24
    compress_save_regex: "notify_backup.tar.gz"
    max_age: 7
    max_size_mb: 5000
    min_disk_free_mb: 1024

🚀 最終結論

📌 1️⃣ RotateSetting 內含 壓縮 & 清理，並提供 優先級機制 (MinDiskFreeMB > MaxSizeMB > MaxAge)
📌 2️⃣ NotifyService 註冊 RotateTask，確保 通知日誌 適時清理
📌 3️⃣ Logger 負責執行 壓縮 & 清理，由 Scheduler 控制排程
📌 4️⃣ config.yaml 讓不同模組可獨立配置 輪轉機制

這樣的設計確保 iPOC 日誌/通知/監控數據都能長期穩定運行，不會無限增長！ 🚀🚀🚀