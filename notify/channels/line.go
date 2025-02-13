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

// LineChannel Line通知器
type LineChannel struct {
	types.BaseChannel
}

// NewLineChannel 創建Line通知器
func NewLineChannel(config models.ChannelConfig) (interfaces.Channel, error) {
	return &LineChannel{
		BaseChannel: types.NewBaseChannel(config),
	}, nil
}

// Send 發送Line通知
func (n *LineChannel) Send(message models.AlertMessage) error {
	payload := map[string]interface{}{
		"to": n.BaseChannel.Config.Config["to"],
		"messages": []map[string]string{
			{
				"type": "text",
				"text": message.Subject + "\n" + message.Body,
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/push",
		bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.BaseChannel.Config.Config["channel_token"])

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message to Line: %d", resp.StatusCode)
	}

	return nil
}

// Test 測試Line配置
func (n *LineChannel) Test() error {
	msg := models.AlertMessage{
		Subject: "Test Line",
		Body:    "This is a test message from alert system.",
	}
	return n.Send(msg)
}
