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

// WebhookChannel Webhook通知器
type WebhookChannel struct {
	types.BaseChannel
}

// NewWebhookChannel 創建Webhook通知器
func NewWebhookChannel(config models.ChannelConfig) (interfaces.Channel, error) {
	return &WebhookChannel{
		BaseChannel: types.NewBaseChannel(config),
	}, nil
}

// Send 發送Webhook通知
func (n *WebhookChannel) Send(message models.AlertMessage) error {
	payload := map[string]interface{}{
		"subject": message.Subject,
		"body":    message.Body,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	method := n.BaseChannel.Config.Config["method"]
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequest(method, n.BaseChannel.Config.Config["url"], bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token := n.BaseChannel.Config.Config["token"]; token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message to Webhook: %d", resp.StatusCode)
	}

	return nil
}

// Test 測試Webhook配置
func (n *WebhookChannel) Test() error {
	msg := models.AlertMessage{
		Subject: "Test Webhook",
		Body:    "This is a test message from alert system.",
	}
	return n.Send(msg)
}
