package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/detect-viz/shared-lib/models/common"
)

// 發送通知
func (s *Service) sendWebhook(info common.NotifyConfig) error {

	var payload map[string]interface{}
	switch info.Type {
	case "discord":
		payload = map[string]interface{}{
			"embeds": []map[string]interface{}{
				{
					"title":       info.Config["title"],
					"description": info.Config["message"],
					"color":       16711680, // 紅色
					"timestamp":   time.Now().Format(time.RFC3339),
					"footer": map[string]string{
						"text": "Alert System",
					},
				},
			},
		}
	case "slack":
		payload = map[string]interface{}{
			"text": info.Config["title"] + "\n" + info.Config["message"],
			"blocks": []map[string]interface{}{
				{
					"type": "header",
					"text": map[string]string{
						"type": "plain_text",
						"text": info.Config["title"],
					},
				},
				{
					"type": "section",
					"text": map[string]string{
						"type": "mrkdwn",
						"text": info.Config["message"],
					},
				},
			},
		}
	case "webex":
		payload = map[string]interface{}{
			"markdown": true,
			"text":     "**" + info.Config["title"] + "**\n" + info.Config["message"],
		}

	case "line":
		if info.Config["url"] == "" {
			info.Config["url"] = "https://api.line.me/v2/bot/message/push"
		}
		payload = map[string]interface{}{
			"to": info.Config["to"],
			"messages": []map[string]string{
				{
					"type": "text",
					"text": info.Config["title"] + "\n" + info.Config["message"],
				},
			},
		}
	case "teams":
		payload = map[string]interface{}{
			"@type":      "MessageCard",
			"@context":   "http://schema.org/extensions",
			"summary":    info.Config["title"],
			"themeColor": "0076D7",
			"title":      info.Config["title"],
			"text":       info.Config["message"],
		}

	case "webhook":
		payload = map[string]interface{}{
			"title":   info.Config["title"],
			"message": info.Config["message"],
		}
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	method := info.Config["method"]
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequest(method, info.Config["url"], bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token := info.Config["channel_token"]; token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message to %s: %d", info.Type, resp.StatusCode)
	}

	return nil
}
