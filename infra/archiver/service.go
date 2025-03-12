//go:build wireinject

package archiver

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/detect-viz/shared-lib/infra/logger"
	"github.com/detect-viz/shared-lib/infra/scheduler"
	"github.com/google/wire"

	"github.com/klauspost/pgzip"
	"go.uber.org/zap"
)

// ArchiverSet 提供 ArchiverService
var ArchiverSet = wire.NewSet(NewService,
	wire.Bind(new(Service), new(*ServiceImpl)),
)

// LogRotator 日誌輪轉器
type ServiceImpl struct {
	logger           logger.Logger
	SchedulerService scheduler.Service
}

// 創建日誌輪轉器
func NewService(logger logger.Logger, schedulerService scheduler.Service) *ServiceImpl {
	return &ServiceImpl{
		logger:           logger,
		SchedulerService: schedulerService,
	}
}

// PrepareDir 檢查並創建目標目錄
func (s *ServiceImpl) prepareDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("創建目錄失敗: %s", path)
		}
	}
	return nil
}

// 檢查目錄大小是否超過最大大小
func (r *ServiceImpl) isExceedMaxSize(dir string, maxSizeMB int64) bool {
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

// 清理舊檔案
func (r *ServiceImpl) cleanOldFiles(sourcePath string, maxAge time.Duration) error {
	files, err := r.listFiles(sourcePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if time.Since(file.ModTime()) > maxAge {
			path := filepath.Join(sourcePath, file.Name())
			if err := os.Remove(path); err != nil {
				r.logger.Error("刪除舊檔案失敗",
					zap.String("file", path),
					zap.Error(err))
			}
		}
	}
	return nil
}

// 壓縮舊日誌
func (r *ServiceImpl) compressOldLogs(sourcePath string, destPath string, compressMatchRegex, compressSaveRegex string, compressOffsetHours int) error {

	files, err := r.listCompressableFiles(sourcePath, compressMatchRegex, compressOffsetHours)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := r.compressFile(sourcePath, destPath, compressSaveRegex, file); err != nil {
			r.logger.Error("壓縮檔案失敗",
				zap.String("file", file.Name()),
				zap.Error(err))
		}
	}
	return nil
}

// 列出目錄中的所有檔案
func (r *ServiceImpl) listFiles(path string) ([]os.FileInfo, error) {
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

// 列出可壓縮的檔案
func (r *ServiceImpl) listCompressableFiles(sourcePath string, compressMatchRegex string, compressOffsetHours int) ([]os.FileInfo, error) {
	files, err := r.listFiles(sourcePath)
	if err != nil {
		return nil, err
	}

	var compressable []os.FileInfo
	matchRegex := strings.ReplaceAll(
		compressMatchRegex,
		"${YYYYMMDD}",
		time.Now().Format("20060102"),
	)

	offsetHours := time.Duration(compressOffsetHours) * time.Hour
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

// 壓縮檔案
func (r *ServiceImpl) compressFile(sourcePath string, destPath string, compressSaveRegex string, file os.FileInfo) error {
	srcPath := filepath.Join(sourcePath, file.Name())

	// 替換日期變量
	saveRegex := strings.ReplaceAll(
		compressSaveRegex,
		"${YYYYMMDD}",
		time.Now().Format("20060102"),
	)

	dstPath := filepath.Join(destPath, saveRegex)

	// 確保目標目錄存在
	if err := os.MkdirAll(destPath, 0755); err != nil {
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

// 檢查磁碟空間是否不足
func (r *ServiceImpl) isLowDiskSpace(path string, minFreeMB int64) bool {
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
