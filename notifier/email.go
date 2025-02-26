package notifier

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
)

// Send 發送郵件通知
func (s *serviceImpl) sendEmail(info common.NotifySetting) error {
	config := parseEmailConfig(info.Config)

	// 設置郵件頭
	headers := make(map[string]string)
	headers["From"] = config.From
	headers["To"] = strings.Join(config.To, ",")
	headers["Cc"] = strings.Join(config.Cc, ",")
	headers["Subject"] = info.Config["title"]
	headers["Reply-To"] = config.ReplyTo
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 組合郵件內容
	var body string
	for k, v := range headers {
		body += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	body += "\r\n" + info.Config["message"]

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

// Dial 建立 TLS 連接
func dialTLS(addr string) (*smtp.Client, error) {
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, fmt.Errorf("tls dial error: %v", err)
	}

	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

// sendMailUsingTLS 使用 TLS 發送郵件
func sendMailUsingTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	client, err := dialTLS(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth error: %v", err)
			}
		}
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	if _, err = w.Write(msg); err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	return client.Quit()
}
