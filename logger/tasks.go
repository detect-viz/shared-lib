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

	"github.com/detect-viz/shared-lib/interfaces"
	"github.com/detect-viz/shared-lib/models/common"

	"github.com/klauspost/pgzip"
	"go.uber.org/zap"
)

// LogRotator 日誌輪轉器
type LogRotator struct {
	logger interfaces.Logger
}

// NewLogRotator 創建日誌輪轉器
func NewLogRotator(logger interfaces.Logger) interfaces.LogRotator {
	return &LogRotator{
		logger: logger,
	}
}

// GetLogger 獲取日誌記錄器
func (r *LogRotator) GetLogger() interfaces.Logger {
	return r.logger
}

// Start 啟動日誌輪轉器
func (r *LogRotator) Start() error {
	return nil
}

// Stop 停止日誌輪轉器
func (r *LogRotator) Stop() {
}

// ExecuteRotateTask 實現輪轉任務
func (r *LogRotator) ExecuteRotateTask(task common.RotateTask) error {
	// 1. 檢查並創建目標目錄
	if err := os.MkdirAll(task.DestPath, 0755); err != nil {
		return fmt.Errorf("創建目標目錄失敗: %w", err)
	}

	// 2. 檢查磁碟空間
	if r.isLowDiskSpace(task.SourcePath, task.RotateSetting.MinDiskFreeMB) {
		r.logger.Warn("磁碟空間不足，執行緊急清理",
			zap.String("path", task.SourcePath),
			zap.Int64("min_free_mb", task.RotateSetting.MinDiskFreeMB))
		if err := r.cleanOldFiles(task); err != nil {
			return fmt.Errorf("緊急清理失敗: %w", err)
		}
	}

	// 3. 檢查目錄大小並清理
	if r.isExceedMaxSize(task.SourcePath, task.RotateSetting.MaxSizeMB) {
		if err := r.cleanOldFiles(task); err != nil {
			return fmt.Errorf("清理超大目錄失敗: %w", err)
		}
	}

	// 4. 壓縮舊日誌
	if task.RotateSetting.CompressEnabled {
		if err := r.compressOldLogs(task); err != nil {
			return fmt.Errorf("壓縮舊日誌失敗: %w", err)
		}
	}

	r.logger.Info("輪轉任務執行完成",
		zap.String("source", task.SourcePath),
		zap.String("dest", task.DestPath))

	return nil
}

// 輔助方法
func (r *LogRotator) isExceedMaxSize(dir string, maxSizeMB int64) bool {
	var size int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size > maxSizeMB*1024*1024
}

func (r *LogRotator) cleanOldFiles(task common.RotateTask) error {
	files, err := r.listFiles(task.SourcePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if time.Since(file.ModTime()) > task.RotateSetting.MaxAge {
			path := filepath.Join(task.SourcePath, file.Name())
			if err := os.Remove(path); err != nil {
				r.logger.Error("刪除舊檔案失敗",
					zap.String("file", path),
					zap.Error(err))
			}
		}
	}
	return nil
}

func (r *LogRotator) compressOldLogs(task common.RotateTask) error {
	files, err := r.listCompressableFiles(task)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := r.compressFile(task, file); err != nil {
			r.logger.Error("壓縮檔案失敗",
				zap.String("file", file.Name()),
				zap.Error(err))
		}
	}
	return nil
}

// 其他輔助方法保持不變...
func (r *LogRotator) listFiles(path string) ([]os.FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []os.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			r.logger.Error("獲取檔案資訊失敗",
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

func (r *LogRotator) listCompressableFiles(task common.RotateTask) ([]os.FileInfo, error) {
	files, err := r.listFiles(task.SourcePath)
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

func (r *LogRotator) compressFile(task common.RotateTask, file os.FileInfo) error {
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
		r.logger.Error("刪除原始檔案失敗",
			zap.String("file", srcPath),
			zap.Error(err))
	}

	r.logger.Info("檔案壓縮完成",
		zap.String("source", file.Name()),
		zap.String("target", filepath.Base(dstPath)))

	return nil
}

// isLowDiskSpace 檢查磁碟空間是否不足
func (r *LogRotator) isLowDiskSpace(path string, minFreeMB int64) bool {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		r.logger.Error("檢查磁碟空間失敗",
			zap.String("path", path),
			zap.Error(err))
		return false
	}

	// 計算可用空間(MB)
	freeSpace := (stat.Bavail * uint64(stat.Bsize)) / (1024 * 1024)
	return freeSpace < uint64(minFreeMB)
}
