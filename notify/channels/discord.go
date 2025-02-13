package channels

import (
	"bytes"
	"encoding/json"
	"fmt"

	"shared-lib/interfaces"
	"shared-lib/models"
	"shared-lib/notify/types"

	"net/http"
	"time"
)

// DiscordChannel Discord通知器
type DiscordChannel struct {
	types.BaseChannel
}

// NewDiscordChannel 創建Discord通知器
func NewDiscordChannel(config models.ChannelConfig) (interfaces.Channel, error) {
	return &DiscordChannel{
		BaseChannel: types.NewBaseChannel(config),
	}, nil
}

// Send 發送Discord通知
func (n *DiscordChannel) Send(message models.AlertMessage) error {
	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       message.Subject,
				"description": message.Body,
				"color":       16711680, // 紅色
				"timestamp":   time.Now().Format(time.RFC3339),
				"footer": map[string]string{
					"text": "Alert System",
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(n.Config.Config["webhook_url"],
		"application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message to Discord: %d", resp.StatusCode)
	}

	return nil
}

// Test 測試Discord配置
func (n *DiscordChannel) Test() error {
	msg := models.AlertMessage{
		Subject: "Test Discord",
		Body:    "This is a test message from alert system.",
	}
	return n.Send(msg)
}
