package alert

// AlertState 異常狀態記錄
type AlertState struct {
	//* 綁定關聯規則
	RuleDetailID int64 `json:"rule_detail_id"`
	//* 以下皆動態更新欄位
	SilenceStart   *int64 `json:"silence_start"`
	SilenceEnd     *int64 `json:"silence_end"`
	MuteStart      *int64 `json:"mute_start"`
	MuteEnd        *int64 `json:"mute_end"`
	ContactCounter int    `json:"contact_counter"` // 連續通知次數 > times 則進入靜音期，然后更新 silence_start 和 silence_end
	//* 狀態
	LastAlertSeverity *string `json:"last_alert_severity"` // info/warn/crit 轉為 normal => resolved
	RuleState         string  `json:"rule_state"`          // 規則狀態(alerting/resolved/normal/disabled)
	ContactState      string  `json:"contact_state"`       // 通知狀態(normal/muting/silence)
	NotifyState       string  `json:"notify_state"`        // 通知結果(sent/failed)

	FirstTriggerTime int64   `json:"first_trigger_time"`                                        // 首次觸發時間
	LastTriggerTime  int64   `json:"last_trigger_time"`                                         // 異常最後時間
	LastCheckValue   float64 `json:"last_check_value" gorm:"type:decimal(10,2);default:0.00"`   // 最後一次檢查的值 (提升精度)
	LastTriggerValue float64 `json:"last_trigger_value" gorm:"type:decimal(10,2);default:0.00"` // 前一次異常值
	//* 適用於 amplitude
	Amplitude float64 `json:"amplitude" gorm:"type:decimal(10,2);default:0.00"` // 變化幅度

	CreatedAt int64 `json:"-" form:"created_at"`
	UpdatedAt int64 `json:"-" form:"updated_at"`
}
