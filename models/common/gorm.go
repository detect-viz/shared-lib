package common

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
)

// 提供基礎的審計與軟刪除欄位
type AuditTimeModel struct {
	CreatedAt int64          `json:"-" gorm:"autoCreateTime"`
	UpdatedAt int64          `json:"-" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;precision:0"` // 軟刪除
}
type AuditUserModel struct {
	CreatedBy *string `json:"-" gorm:"type:varchar(36);index"`
	UpdatedBy *string `json:"-" gorm:"type:varchar(36);index"`
}

// JSONMap 讓 GORM 正確處理 JSON 欄位
type JSONMap map[string]string

func (j *JSONMap) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// SeveritySet 讓 GORM 正確處理 MySQL SET
type SeveritySet []string

// 允許的 `SET` 值
var validSeverities = map[string]bool{
	"info": true, "warn": true, "crit": true,
}

// `Scan` (從資料庫讀取)
func (s *SeveritySet) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan SeveritySet")
	}
	*s = strings.Split(str, ",") // ✅ 轉換為 []string
	return nil
}

// `Value` (存入資料庫)
func (s SeveritySet) Value() (driver.Value, error) {
	// ✅ 確保所有值都是合法的
	for _, v := range s {
		if !validSeverities[v] {
			return nil, errors.New("invalid severity value: " + v)
		}
	}
	return strings.Join(s, ","), nil // ✅ 存入 MySQL 時轉為 `info,warn` 格式
}
