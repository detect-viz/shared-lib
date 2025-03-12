package archiver

import (
	"github.com/detect-viz/shared-lib/models/common"
)

// 誌輪轉器接口
type Service interface {
	// 註冊輪轉任務
	RegisterRotateTask(task common.RotateTask) error
	// 執行輪轉任務
	ExecuteRotateTask(task common.RotateTask) error
	// 註冊備份任務
	RegisterBackupTask(task common.BackupTask) error
	// 執行備份任務
	ExecuteBackupTask(task common.BackupTask) error
}
