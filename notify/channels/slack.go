package channels

import (
	"bytes"
	"encoding/json"
	"fmt"

	"net/http"
	"shared-lib/interfaces"
	"shared-lib/models"
	"shared-lib/notify/types"
)

// SlackChannel Slack通知器
type SlackChannel struct {
	types.BaseChannel
}

// NewSlackChannel 創建Slack通知器
func NewSlackChannel(config models.ChannelConfig) (interfaces.Channel, error) {
	return &SlackChannel{
		BaseChannel: types.NewBaseChannel(config),
	}, nil
}

// Send 發送Slack通知
func (n *SlackChannel) Send(message models.AlertMessage) error {
	payload := map[string]interface{}{
		"text": message.Subject + "\n" + message.Body,
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": message.Subject,
				},
			},
			{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": message.Body,
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(n.BaseChannel.Config.Config["webhook_url"],
		"application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message to Slack: %d", resp.StatusCode)
	}

	return nil
}

// Test 測試Slack配置
func (n *SlackChannel) Test() error {
	msg := models.AlertMessage{
		Subject: "Test Slack",
		Body:    "This is a test message from alert system.",
	}
	return n.Send(msg)
}
