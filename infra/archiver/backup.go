package archiver

import (
	"fmt"

	"github.com/detect-viz/shared-lib/models/common"
)

// NewBackupTask 建立 `BackupTask`
func NewBackupTask(service *ServiceImpl) *BackupTask {
	return &BackupTask{service: service}
}

// BackupTask 管理備份
type BackupTask struct {
	service *ServiceImpl
}

// Execute 執行備份
func (b *BackupTask) Execute(setting common.BackupSetting) error {
	// 1️⃣ 檢查目標目錄
	if err := b.service.prepareDir(setting.SourcePath); err != nil {
		return err
	}

	// 2️⃣ 磁碟空間檢查
	if !b.service.isLowDiskSpace(setting.SourcePath, setting.MinDiskFreeMB) {
		return fmt.Errorf("磁碟空間不足")
	}

	// 3️⃣ 清理舊備份
	if err := b.service.cleanOldFiles(setting.SourcePath, setting.MaxAge); err != nil {
		return err
	}

	// 4️⃣ 執行備份
	fmt.Println("🔄 執行備份: ", setting.BackupType)
	if setting.BackupType == "database" {
		fmt.Println("📌 備份數據庫:", setting.BackupType)
		// TODO: 執行 DB 備份
	} else {
		fmt.Println("📌 備份檔案:", setting.SourcePath)
		// TODO: 執行文件備份
	}

	fmt.Println("✅ 備份完成")
	return nil
}
