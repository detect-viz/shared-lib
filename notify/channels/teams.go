package channels

import (
	"bytes"
	"encoding/json"
	"fmt"

	"shared-lib/interfaces"
	"shared-lib/models"
	"shared-lib/notify/types"

	"net/http"
)

// TeamsChannel Teams通知器
type TeamsChannel struct {
	types.BaseChannel
}

// NewTeamsChannel 創建Teams通知器
func NewTeamsChannel(config models.ChannelConfig) (interfaces.Channel, error) {
	return &TeamsChannel{
		BaseChannel: types.NewBaseChannel(config),
	}, nil
}

// Send 發送Teams通知
func (n *TeamsChannel) Send(message models.AlertMessage) error {
	payload := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"summary":    message.Subject,
		"themeColor": "0076D7",
		"title":      message.Subject,
		"text":       message.Body,
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
		return fmt.Errorf("failed to send message to Teams: %d", resp.StatusCode)
	}

	return nil
}

// Test 測試Teams配置
func (n *TeamsChannel) Test() error {
	msg := models.AlertMessage{
		Subject: "Test Teams",
		Body:    "This is a test message from alert system.",
	}
	return n.Send(msg)
}
