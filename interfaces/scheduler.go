package interfaces

import "time"

// TaskInfo 任務資訊
type TaskInfo struct {
	Name     string        // 任務名稱
	Schedule time.Duration // 執行間隔
	LastRun  time.Time     // 上次執行時間
	NextRun  time.Time     // 下次執行時間
	Status   string        // 任務狀態 (running/stopped)
	Error    error         // 最後一次執行錯誤
}

// Task 任務介面
type Task interface {
	Execute() error
	GetName() string
	GetSchedule() time.Duration
}

// Scheduler 排程器介面
type Scheduler interface {
	// 註冊任務
	RegisterTask(name string, schedule time.Duration, task Task) error
	// 啟動排程器
	Start() error
	// 停止排程器
	Stop()
	// 列出所有任務
	ListTasks() map[string]TaskInfo
	// 獲取特定任務資訊
	GetTaskInfo(name string) (TaskInfo, bool)
	// 暫停任務
	PauseTask(name string) error
	// 恢復任務
	ResumeTask(name string) error
	// 立即執行任務
	RunTaskNow(name string) error
}
