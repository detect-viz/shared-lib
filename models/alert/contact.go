package alert

import (
	"database/sql/driver"
	"errors"
	"strings"

	"github.com/detect-viz/shared-lib/models/common"
)

type Contact struct {
	ID           []byte         `json:"id" gorm:"primaryKey"`
	RealmName    string         `json:"realm_name" gorm:"default:master"`
	Name         string         `json:"name"`
	ChannelType  string         `json:"channel_type"`
	Enabled      bool           `json:"enabled" gorm:"default:1"`
	SendResolved bool           `json:"send_resolved" gorm:"default:1"`
	AutoApply    bool           `json:"auto_apply" gorm:"default:0"`
	MaxRetry     int            `json:"max_retry" gorm:"default:3"`
	RetryDelay   string         `json:"retry_delay" gorm:"default:5m"`
	Config       common.JSONMap `json:"config" gorm:"type:json"`
	Severities   SeveritySet    `json:"severities" gorm:"type:set('info','warn','crit');default:'crit'"`
	common.AuditUserModel
	common.AuditTimeModel
}

type ContactResponse struct {
	RealmName    string                 `json:"-"`
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	ChannelType  string                 `json:"channel_type"`
	Enabled      bool                   `json:"enabled"`
	SendResolved bool                   `json:"send_resolved"`
	AutoApply    bool                   `json:"auto_apply"`
	MaxRetry     int                    `json:"max_retry"`
	RetryDelay   string                 `json:"retry_delay"`
	Config       map[string]interface{} `json:"config"`
	Severities   []string               `json:"severities"`
}

// 綁定聯絡人 many2many
type RuleContact struct {
	RuleID    []byte `gorm:"type:binary(16);primaryKey"`
	ContactID []byte `gorm:"type:binary(16);primaryKey"`
}

// SeveritySet 讓 GORM 正確處理 MySQL SET
type SeveritySet []string

// 允許的 SET 值
var validSeverities = map[string]bool{
	"info": true, "warn": true, "crit": true,
}

// 從資料庫讀取
func (s *SeveritySet) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("failed to scan SeveritySet")
	}
	*s = strings.Split(str, ",")
	return nil
}

// 存入資料庫
func (s SeveritySet) Value() (driver.Value, error) {
	for _, v := range s {
		if !validSeverities[v] {
			return nil, errors.New("invalid severity value: " + v)
		}
	}
	return strings.Join(s, ","), nil //
}
