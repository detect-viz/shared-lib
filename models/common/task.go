package common

import (
	"time"

	"github.com/robfig/cron/v3"
)

// Task 定義一個可調度的任務
type Task struct {
	Enabled     bool          `json:"enabled"`     // 是否啟用
	Name        string        `json:"name"`        // 任務名稱
	Timezone    string        `json:"timezone"`    // 時區
	Description string        `json:"description"` // 任務描述
	Spec        string        `json:"spec"`        // CRON 表達式 或 @every 1h
	Duration    time.Duration `json:"duration"`    // 任務執行時間
	Type        string        `json:"type"`        // 任務類型 (ex: "alert", "backup", "rotate")
	RetryCount  int           `json:"retry_count"` // 最大重試次數
	RetryDelay  time.Duration `json:"retry_delay"` // 重試延遲時間
	ExecFunc    func() error  `json:"-"`           // 執行函數 (不存 DB)
}

// 任務資訊
type TaskInfo struct {
	ID      cron.EntryID `json:"id"`       // 任務 ID (唯一標識)
	Name    string       `json:"name"`     // 任務名稱 (唯一標識)
	Spec    string       `json:"spec"`     // 執行間隔
	LastRun time.Time    `json:"last_run"` // 上次執行時間
	NextRun time.Time    `json:"next_run"` // 下次執行時間
	Status  string       `json:"status"`   // 任務狀態 (running/stopped)
	Error   error        `json:"error"`    // 最後一次執行錯誤
}
