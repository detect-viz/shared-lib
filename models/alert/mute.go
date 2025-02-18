package alert

import (
	"shared-lib/models/common"

	"gorm.io/gorm"
)

// AlertMute 告警抑制規則
type AlertMute struct {
	common.Common
	RealmName      string         `json:"realm_name" gorm:"default:master"`
	ID             int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Name           string         `json:"name"`                             // 抑制規則名稱
	Description    *string        `json:"description"`                      // 描述
	Enabled        bool           `json:"enabled" gorm:"default:true"`      // 是否啟用
	StartTime      int64          `json:"start_time"`                       // 執行開始時間
	EndTime        int64          `json:"end_time"`                         // 執行結束時間
	State          string         `json:"state" gorm:"default:disabled"`    // 抑制狀態
	CronExpression *string        `json:"cron_expression"`                  // Cron 表達式
	CronPeriod     *int64         `json:"cron_period"`                      // 抑制持續時間(秒)
	RepeatTime     RepeatTime     `json:"repeat_time" gorm:"embedded"`      // 重複時間設定
	RepeatType     string         `json:"repeat_type" gorm:"default:never"` // 重複類型: never, daily, weekly, monthly
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	AlertRules     []AlertRule    `json:"alert_rules" gorm:"many2many:alert_mute_rules"`
}

type RepeatTime struct {
	Never   bool `json:"never" gorm:"default:true"` // 不重複
	Daily   bool `json:"daily"`                     // 每天重複
	Weekly  bool `json:"weekly"`                    // 每週重複
	Monthly bool `json:"monthly"`                   // 每月重複
	Mon     bool `json:"mon" gorm:"default:false"`  // 週一
	Tue     bool `json:"tue" gorm:"default:false"`  // 週二
	Wed     bool `json:"wed" gorm:"default:false"`  // 週三
	Thu     bool `json:"thu" gorm:"default:false"`  // 週四
	Fri     bool `json:"fri" gorm:"default:false"`  // 週五
	Sat     bool `json:"sat" gorm:"default:false"`  // 週六
	Sun     bool `json:"sun" gorm:"default:false"`  // 週日
}

// AlertMuteRule 告警抑制規則關聯
type AlertMuteRule struct {
	AlertMuteID int64 `json:"alert_mute_id" gorm:"primaryKey"`
	AlertRuleID int64 `json:"alert_rule_id" gorm:"primaryKey"`
}
