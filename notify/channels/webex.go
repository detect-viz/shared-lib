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

// WebexChannel Webex通知器
type WebexChannel struct {
	types.BaseChannel
}

// NewWebexChannel 創建Webex通知器
func NewWebexChannel(config models.ChannelConfig) (interfaces.Channel, error) {
	return &WebexChannel{
		BaseChannel: types.NewBaseChannel(config),
	}, nil
}

// Send 發送Webex通知
func (n *WebexChannel) Send(message models.AlertMessage) error {
	payload := map[string]interface{}{
		"markdown": true,
		"text":     "**" + message.Subject + "**\n" + message.Body,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(n.BaseChannel.Config.Config["webhook_url"], "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message to Webex: %d", resp.StatusCode)
	}

	return nil
}

// Test 測試Webex配置
func (n *WebexChannel) Test() error {
	msg := models.AlertMessage{
		Subject: "Test Webex",
		Body:    "This is a test message from alert system.",
	}
	return n.Send(msg)
}
