package mysql

import (
	"errors"
	"strings"

	"github.com/detect-viz/shared-lib/apierrors"
	"gorm.io/gorm"
)

// ParseDBError 解析 MySQL/Gorm 錯誤，轉換為標準錯誤
func ParseDBError(err error) error {
	if err == nil {
		return nil
	}

	// MySQL 唯一索引違反 (Duplicate Entry)
	if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "unique constraint") {
		return apierrors.ErrDuplicateEntry
	}

	// GORM 查無資料錯誤
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apierrors.ErrNotFound
	}

	// 其他錯誤，直接回傳
	return err
}
