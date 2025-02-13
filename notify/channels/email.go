package channels

import (
	"fmt"
	"net/smtp"
	"strings"

	"shared-lib/interfaces"
	"shared-lib/models"
	"shared-lib/notify/types"
)

// EmailChannel 郵件通知器
type EmailChannel struct {
	types.BaseChannel
}

// NewEmailChannel 創建郵件通知器
func NewEmailChannel(config models.ChannelConfig) (interfaces.Channel, error) {
	return &EmailChannel{
		BaseChannel: types.NewBaseChannel(config),
	}, nil
}

// Send 發送郵件通知
func (n *EmailChannel) Send(message models.AlertMessage) error {
	config := parseEmailConfig(n.BaseChannel.Config.Config)

	// 設置郵件頭
	headers := make(map[string]string)
	headers["From"] = config.From
	headers["To"] = strings.Join(config.To, ",")
	headers["Cc"] = strings.Join(config.Cc, ",")
	headers["Subject"] = message.Subject
	headers["Reply-To"] = config.ReplyTo
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 組合郵件內容
	var body string
	for k, v := range headers {
		body += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	body += "\r\n" + message.Body

	// 合併所有收件人
	recipients := mergeRecipients(config.To, config.Cc, config.Bcc)

	// 建立認證
	var auth smtp.Auth
	if config.UseAuth {
		auth = smtp.PlainAuth("", config.Username, config.Password, config.Host)
	}

	// 發送郵件
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	if config.UseTLS {
		return sendMailUsingTLS(addr, auth, config.From, recipients, []byte(body))
	}

	return smtp.SendMail(addr, auth, config.From, recipients, []byte(body))
}

// 解析郵件配置
func parseEmailConfig(config map[string]string) models.EmailSetting {
	return models.EmailSetting{
		Host:     config["host"],
		Port:     config["port"],
		Username: config["username"],
		Password: config["password"],
		From:     config["from"],
		To:       strings.Split(config["to"], ","),
		Cc:       strings.Split(config["cc"], ","),
		Bcc:      strings.Split(config["bcc"], ","),
		ReplyTo:  config["reply_to"],
		UseTLS:   config["use_tls"] == "true",
		UseAuth:  config["use_auth"] == "true",
	}
}

// 合併收件人列表
func mergeRecipients(to, cc, bcc []string) []string {
	var recipients []string

	// 過濾空值並合併
	for _, list := range [][]string{to, cc, bcc} {
		for _, addr := range list {
			if addr = strings.TrimSpace(addr); addr != "" {
				recipients = append(recipients, addr)
			}
		}
	}

	return recipients
}

// Test 測試郵件配置
func (n *EmailChannel) Test() error {
	msg := models.AlertMessage{
		Subject: "Test Email",
		Body:    "This is a test email from alert system.",
	}
	return n.Send(msg)
}
