package alert

import (
	"github.com/detect-viz/shared-lib/models/config"
)

// * 當前告警
type CurrentAlert struct {
	Time          int         `json:"timestamp"`      //* 觸發時間
	Status        config.Code `json:"status"`         //* 狀態
	Severity      config.Code `json:"severity"`       //* 告警等級
	RuleName      string      `json:"rule_name"`      //* 規則名稱
	ResourceGroup string      `json:"resource_group"` //* 群組名稱
	Resource      string      `json:"resource"`       //* 主機名稱
	IP            string      `json:"ip"`             //* IP(伺服器標籤)
	Partition     string      `json:"partition"`      //* 分區名稱
	Threshold     string      `json:"threshold"`      //* 觸發閥值
	Duration      string      `json:"duration"`       //* 持續時間
}

// * 歷史告警
type HistoryAlert struct {
	UUID          string      `json:"uuid"`           //* 告警事件UUID
	AlertTime     int         `json:"alert_time"`     //* 觸發時間
	ResolveTime   int         `json:"resolve_time"`   //* 解決時間
	Status        config.Code `json:"status"`         //* 狀態
	ContactState  config.Code `json:"contact_state"`  //* 告警通知狀態
	Severity      config.Code `json:"severity"`       //* 告警等級(取最高)
	RuleName      string      `json:"rule_name"`      //* 規則名稱
	RuleID        int         `json:"rule_id"`        //* 規則ID
	ResourceGroup string      `json:"resource_group"` //* 群組名稱
	Resource      string      `json:"resource"`       //* 主機名稱
	IP            string      `json:"ip"`             //* IP(伺服器標籤)
	Partition     string      `json:"partition"`      //* 分區名稱
	Duration      string      `json:"duration"`       //* 持續時間
	Threshold     string      `json:"threshold"`      //* 觸發值(取最大或最小)
	Operator      string      `json:"operator"`
	Contacts      string      `json:"contacts"` // * 通知管道 (逗號)
}
