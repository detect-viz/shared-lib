package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"shared-lib/interfaces"
	"shared-lib/models/common"

	"github.com/klauspost/pgzip"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// RotateManager 輪轉管理器
type RotateManager struct {
	tasks  []common.RotateTask
	logger interfaces.Logger
	cron   *cron.Cron
}

// NewRotateManager 創建輪轉管理器
func NewRotateManager(logger interfaces.Logger) *RotateManager {
	return &RotateManager{
		logger: logger,
		cron:   cron.New(cron.WithSeconds()),
	}
}

// RegisterTask 註冊輪轉任務
func (m *RotateManager) RegisterTask(task common.RotateTask) error {
	m.logger.Info("註冊檔案輪轉任務",
		zap.String("job_id", task.JobID),
		zap.String("source_path", task.SourcePath),
		zap.String("dest_path", task.DestPath))

	m.tasks = append(m.tasks, task)
	return nil
}

// RunTask 執行輪轉任務
func (m *RotateManager) RunTask(task common.RotateTask) error {
	// 1. 檢查磁碟空間
	if m.isLowDiskSpace(task.SourcePath, task.RotateSetting.MinDiskFreeMB) {
		if err := m.cleanOldFiles(task); err != nil {
			return fmt.Errorf("清理舊檔案失敗: %w", err)
		}
	}

	// 2. 檢查目錄大小
	if m.isExceedMaxSize(task.SourcePath, task.RotateSetting.MaxSizeMB) {
		if err := m.cleanOldFiles(task); err != nil {
			return fmt.Errorf("清理超大目錄失敗: %w", err)
		}
	}

	// 3. 壓縮舊日誌
	if task.RotateSetting.CompressEnabled {
		if err := m.compressOldLogs(task); err != nil {
			return fmt.Errorf("壓縮舊日誌失敗: %w", err)
		}
	}

	m.logger.Info("輪轉任務執行完成",
		zap.String("job_id", task.JobID),
		zap.String("source", task.SourcePath),
		zap.String("dest", task.DestPath))

	return nil
}

// listFiles 列出目錄下的檔案
func (m *RotateManager) listFiles(path string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []os.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			m.logger.Error("獲取檔案資訊失敗",
				zap.String("name", entry.Name()),
				zap.Error(err))
			continue
		}
		if !info.IsDir() {
			files = append(files, info)
		}
	}

	// 按修改時間排序
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	return files, nil
}

// isLowDiskSpace 檢查磁碟空間是否不足
func (m *RotateManager) isLowDiskSpace(path string, minDiskMB int64) bool {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		m.logger.Error("獲取磁碟資訊失敗", zap.Error(err))
		return false
	}

	freeSpace := (stat.Bavail * uint64(stat.Bsize)) / (1024 * 1024)
	return freeSpace < uint64(minDiskMB)
}

// isExceedMaxSize 檢查目錄大小是否超過限制
func (m *RotateManager) isExceedMaxSize(path string, maxSizeMB int64) bool {
	var totalSize int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		m.logger.Error("計算目錄大小失敗", zap.Error(err))
		return false
	}

	return totalSize > maxSizeMB*1024*1024
}

// Start 啟動輪轉管理器
func (m *RotateManager) Start() error {
	for _, task := range m.tasks {
		if !task.RotateSetting.Enabled {
			continue
		}

		if err := m.scheduleTask(task); err != nil {
			return err
		}
	}

	m.cron.Start()
	return nil
}

// Stop 停止輪轉管理器
func (m *RotateManager) Stop() {
	m.cron.Stop()
}

// scheduleTask 排程任務
func (m *RotateManager) scheduleTask(task common.RotateTask) error {
	_, err := m.cron.AddFunc(task.RotateSetting.Schedule, func() {
		if err := m.RunTask(task); err != nil {
			m.logger.Error("執行輪轉任務失敗",
				zap.String("job_id", task.JobID),
				zap.Error(err))
		}
	})
	return err
}

// cleanOldFiles 清理舊檔案
func (m *RotateManager) cleanOldFiles(task common.RotateTask) error {
	files, err := m.listFiles(task.SourcePath)
	if err != nil {
		return err
	}

	maxAge := task.RotateSetting.MaxAge
	now := time.Now()

	for _, file := range files {
		age := now.Sub(file.ModTime())
		if age > maxAge {
			if err := os.Remove(filepath.Join(task.SourcePath, file.Name())); err != nil {
				m.logger.Error("刪除檔案失敗",
					zap.String("file", file.Name()),
					zap.Error(err))
				continue
			}
			m.logger.Info("已刪除過期檔案",
				zap.String("file", file.Name()),
				zap.Duration("age", age))
		}
	}
	return nil
}

// compressOldLogs 壓縮舊日誌
func (m *RotateManager) compressOldLogs(task common.RotateTask) error {
	// 獲取需要壓縮的檔案
	files, err := m.listCompressableFiles(task)
	if err != nil {
		return fmt.Errorf("獲取可壓縮檔案失敗: %w", err)
	}

	if len(files) == 0 {
		return nil
	}

	// 按時間排序，優先壓縮舊檔案
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	// 執行壓縮
	for _, file := range files {
		if err := m.compressFile(task, file); err != nil {
			m.logger.Error("壓縮檔案失敗",
				zap.String("file", file.Name()),
				zap.Error(err))
			continue
		}
	}

	return nil
}

// listCompressableFiles 獲取需要壓縮的檔案
func (m *RotateManager) listCompressableFiles(task common.RotateTask) ([]os.FileInfo, error) {
	files, err := m.listFiles(task.SourcePath)
	if err != nil {
		return nil, err
	}

	var compressable []os.FileInfo
	matchRegex := strings.ReplaceAll(
		task.RotateSetting.CompressMatchRegex,
		"${YYYYMMDD}",
		time.Now().Format("20060102"),
	)

	offsetHours := time.Duration(task.RotateSetting.CompressOffsetHours) * time.Hour
	now := time.Now()

	for _, file := range files {
		// 檢查是否符合壓縮條件
		if matched, _ := filepath.Match(matchRegex, file.Name()); !matched {
			continue
		}

		// 檢查檔案年齡
		age := now.Sub(file.ModTime())
		if age < offsetHours {
			continue
		}

		compressable = append(compressable, file)
	}

	return compressable, nil
}

// compressFile 壓縮單個檔案
func (m *RotateManager) compressFile(task common.RotateTask, file os.FileInfo) error {
	srcPath := filepath.Join(task.SourcePath, file.Name())

	// 替換日期變量
	saveRegex := strings.ReplaceAll(
		task.RotateSetting.CompressSaveRegex,
		"${YYYYMMDD}",
		time.Now().Format("20060102"),
	)

	dstPath := filepath.Join(task.DestPath, saveRegex)

	// 確保目標目錄存在
	if err := os.MkdirAll(task.DestPath, 0755); err != nil {
		return fmt.Errorf("創建目標目錄失敗: %w", err)
	}

	// 開啟源檔案
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("開啟源檔案失敗: %w", err)
	}
	defer src.Close()

	// 創建目標檔案
	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("創建壓縮檔案失敗: %w", err)
	}
	defer dst.Close()

	// 創建 gzip writer
	gw := pgzip.NewWriter(dst)
	defer gw.Close()

	// 設置檔案資訊
	gw.Name = file.Name()
	gw.ModTime = file.ModTime()

	// 複製內容
	if _, err := io.Copy(gw, src); err != nil {
		return fmt.Errorf("壓縮檔案內容失敗: %w", err)
	}

	// 關閉 gzip writer
	if err := gw.Close(); err != nil {
		return fmt.Errorf("關閉壓縮寫入器失敗: %w", err)
	}

	// 刪除原始檔案
	if err := os.Remove(srcPath); err != nil {
		m.logger.Error("刪除原始檔案失敗",
			zap.String("file", srcPath),
			zap.Error(err))
	}

	m.logger.Info("檔案壓縮完成",
		zap.String("source", file.Name()),
		zap.String("target", filepath.Base(dstPath)))

	return nil
}
