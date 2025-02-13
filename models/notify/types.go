package notify

// ChannelType 通知類型
type ChannelType string

const (
	EmailType   ChannelType = "email"
	TeamsType   ChannelType = "teams"
	SlackType   ChannelType = "slack"
	WebhookType ChannelType = "webhook"
	DiscordType ChannelType = "discord"
	WebexType   ChannelType = "webex"
	LineType    ChannelType = "line"
	SMSType     ChannelType = "sms"
)

// NotifyStatus 通知狀態
type NotifyStatus string

const (
	StatusPending  NotifyStatus = "pending"
	StatusSent     NotifyStatus = "sent"
	StatusFailed   NotifyStatus = "failed"
	StatusRetrying NotifyStatus = "retrying"
)
