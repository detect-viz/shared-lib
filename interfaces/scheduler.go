package interfaces

import (
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
)

// Scheduler 排程器介面
type Scheduler interface {
	// 註冊 Cron 任務
	RegisterCronJob(job models.SchedulerJob) error
	// 註冊任務
	RegisterTask(name string, schedule string, task LogRotator) error
	// 啟動排程器
	Start() error
	// 停止排程器
	Stop()
	// 列出所有任務
	ListTasks() map[string]common.TaskInfo
	// 獲取特定任務資訊
	GetTaskInfo(name string) (common.TaskInfo, bool)
	// 暫停任務
	PauseTask(name string) error
	// 恢復任務
	ResumeTask(name string) error
	// 立即執行任務
	RunTaskNow(name string) error
}
