package notify

import (
	"fmt"
	"shared-lib/interfaces"
	"shared-lib/models"
	"shared-lib/notify/channels"
)

// 轉換 NotificationLog 到 AlertMessage
func toAlertMessage(notification models.NotificationLog) models.AlertMessage {
	return models.AlertMessage{
		Subject: notification.Subject,
		Body:    notification.Body,
	}
}

// NewSender 創建對應的發送器
func NewSender(notification models.NotificationLog) (interfaces.Channel, error) {
	config := models.ChannelConfig{
		Type:   notification.ChannelType,
		Name:   notification.ContactName,
		Config: notification.ContactConfig,
	}

	switch config.Type {
	case "email":
		return channels.NewEmailChannel(config)
	case "teams":
		return channels.NewTeamsChannel(config)
	case "line":
		return channels.NewLineChannel(config)
	case "slack":
		return channels.NewSlackChannel(config)
	case "discord":
		return channels.NewDiscordChannel(config)
	case "webex":
		return channels.NewWebexChannel(config)
	case "webhook":
		return channels.NewWebhookChannel(config)
	default:
		return nil, fmt.Errorf("不支援的通知類型: %s", config.Type)
	}
}
