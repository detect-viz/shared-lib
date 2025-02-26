package notifier

// GetNotifyMethods 提供固定的通知方式
func getNotifyMethods() []string {
	return []string{"email", "slack", "discord", "teams", "webex", "webhook", "line"}
}

// GetNotifyOptions 提供前端可選擇的通知類型與必填欄位
func getNotifyOptions() map[string]map[string][]string {
	return map[string]map[string][]string{
		"email": {
			"required": {"host", "port", "username", "password", "to"},
			"optional": {"from", "cc", "bcc", "reply_to"},
		},
		"line": {
			"required": {"channel_token", "to"},
			"optional": {},
		},
		"slack": {
			"required": {"url"},
			"optional": {},
		},
		"discord": {
			"required": {"url"},
			"optional": {},
		},
		"teams": {
			"required": {"url"},
			"optional": {},
		},
		"webex": {
			"required": {"url"},
			"optional": {},
		},
		"webhook": {
			"required": {"url"},
			"optional": {},
		},
	}
}
