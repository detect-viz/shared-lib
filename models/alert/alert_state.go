package alert

import "shared-lib/models/common"

// AlertState 異常狀態記錄
type AlertState struct {
	common.Common
	RuleID        int64   `json:"rule_id" gorm:"primaryKey;autoIncrement"`
	ResourceName  string  `json:"resource_name"`
	MetricName    string  `json:"metric_name"`
	StartTime     int64   `json:"start_time"`                                            // 異常開始時間
	LastTime      int64   `json:"last_time"`                                             // 最後一次異常時間
	LastValue     float64 `json:"last_value" gorm:"type:decimal(10,2);default:0.00"`     // 最後一次異常值 (提升精度)
	StackDuration int     `json:"stack_duration" gorm:"default:0"`                       // 異常持續時間 (適用於 absolute)
	PreviousValue float64 `json:"previous_value" gorm:"type:decimal(10,2);default:0.00"` // 前一次異常值 (適用於 amplitude)
	Amplitude     float64 `json:"amplitude" gorm:"type:decimal(10,2);default:0.00"`      // 變化幅度 (適用於 amplitude)
}
