package alert

// AlertMessage 告警訊息
type AlertMessage struct {
	Title   string `json:"title"`   // 標題
	Message string `json:"message"` // 內容
}

// AlertMessageTemplate 告警訊息模板
type AlertMessageTemplate struct {
	Type            string `json:"type"`             // 通知類型
	State           string `json:"state"`            // 告警狀態
	TitleTemplate   string `json:"title_template"`   // 標題模板
	MessageTemplate string `json:"message_template"` // 內容模板
}
