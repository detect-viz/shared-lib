package mysql

import (
	"database/sql/driver"
	"errors"
	"strings"

	"github.com/google/uuid"
)

// 正確的方式
func GenerateUUID16() []byte {
	id := uuid.New()
	return id[:16] // 直接取 16-byte binary
}

// SeveritySet 讓 GORM 正確處理 MySQL SET
type SeveritySet []string

// 允許的 SET 值
var ValidSeverities = map[string]bool{
	"info": true, "warn": true, "crit": true,
}

// 從資料庫讀取
func (s *SeveritySet) Scan(value interface{}) error {
	// 處理 nil 值
	if value == nil {
		*s = []string{}
		return nil
	}

	// 處理字串值
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return errors.New("failed to scan SeveritySet: unsupported type")
	}

	// 處理空字串
	if str == "" {
		*s = []string{}
		return nil
	}

	// 分割字串並設置值
	*s = strings.Split(str, ",")
	return nil
}

// 存入資料庫
func (s SeveritySet) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "", nil
	}

	for _, v := range s {
		if !ValidSeverities[v] {
			return nil, errors.New("invalid severity value: " + v)
		}
	}
	return strings.Join(s, ","), nil
}
