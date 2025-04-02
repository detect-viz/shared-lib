package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/detect-viz/shared-lib/models/common"
)

// 發送通知
func (s *serviceImpl) sendWebhook(info common.NotifySetting) error {

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

		// 處理訊息格式，確保換行符號正確顯示
		lineMessage := info.Config["title"] + "\n\n" + info.Config["message"]
		// 移除多餘的換行符號
		lineMessage = strings.ReplaceAll(lineMessage, "\n\n\n", "\n\n")

		payload = map[string]interface{}{
			"to": info.Config["to"],
			"messages": []map[string]interface{}{
				{
					"type": "text",
					"text": lineMessage,
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

	// 設置請求頭
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "DetectViz-Notifier/1.0")

	// 根據通知類型設置特定請求頭
	switch info.Type {
	case "line":
		// 只有當 channel_token 不為空時才添加到請求頭
		if token, ok := info.Config["channel_token"]; ok && token != "" {
			// 確保 token 格式正確
			token = strings.TrimSpace(token)
			req.Header.Set("Authorization", "Bearer "+token)
			fmt.Printf("LINE 通知 Authorization 頭: Bearer %s\n", token)
		} else {
			fmt.Println("警告: LINE 通知缺少 channel_token")
		}

	case "slack":
		// ... existing code ...
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return fmt.Errorf(
			"failed to send message to %s: %d, config: %v, payload: %s, response: %s",
			info.Type, resp.StatusCode, info.Config, string(data), string(bodyBytes),
		)
	}

	return nil
}
