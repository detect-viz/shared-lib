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

	fmt.Printf("config: %+v\n", config)
	fmt.Printf("addr: %v\n", addr)
	fmt.Printf("auth: %v\n", auth)
	fmt.Printf("config.From: %v\n", config.From)
	fmt.Printf("recipients: %v\n", recipients)
	fmt.Printf("body: %v\n", body)

	return sendMailUsingTLS(addr, auth, config.From, recipients, []byte(body), config.UseSTARTTLS)
}

// 解析郵件配置
func parseEmailConfig(config map[string]string) models.EmailSetting {
	if config["from"] == "" {
		config["from"] = config["username"]
	}

	if config["password"] != "" {
		config["use_auth"] = "true"
	}

	// 根據 `port` 自動決定 `use_tls` 和 `use_starttls`
	useTLS := config["use_tls"] == "true"
	useSTARTTLS := false

	if config["use_tls"] == "" { // 如果未手動設置，則自動判斷
		switch config["port"] {
		case "465": // SMTPS (SSL/TLS)
			useTLS = true
		case "587": // STARTTLS
			useTLS = false
			useSTARTTLS = true
		case "25": // 可選擇是否使用 STARTTLS
			useTLS = false
			useSTARTTLS = true
		default:
			useTLS = false
		}
	}

	return models.EmailSetting{
		Host:        config["host"],
		Port:        config["port"],
		Username:    config["username"],
		Password:    config["password"],
		From:        config["from"],
		To:          strings.Split(config["to"], ","),
		Cc:          strings.Split(config["cc"], ","),
		Bcc:         strings.Split(config["bcc"], ","),
		ReplyTo:     config["reply_to"],
		UseTLS:      useTLS,      // 465 用 `tls.Dial()`
		UseSTARTTLS: useSTARTTLS, // 587/25 需要 `STARTTLS`
		UseAuth:     config["use_auth"] == "true",
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

// sendMailUsingTLS 使用 TLS 發送郵件
func sendMailUsingTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, useSTARTTLS bool) error {
	var conn net.Conn
	var err error

	hostname := strings.Split(addr, ":")[0] // 獲取主機名稱
	if useSTARTTLS {
		// 587 使用 net.Dial()，稍後發送 STARTTLS
		fmt.Println("🔍 連線到 SMTP 伺服器 (STARTTLS 模式)...")
		conn, err = net.Dial("tcp", addr)
	} else {
		// 465 使用 tls.Dial() 直接建立加密連線
		fmt.Println("🔍 連線到 SMTP 伺服器 (TLS 直連)...")
		conn, err = tls.Dial("tcp", addr, &tls.Config{
			ServerName:         hostname,
			InsecureSkipVerify: true,
		})
	}

	if err != nil {
		return fmt.Errorf("SMTP 連線失敗: %w", err)
	}
	defer conn.Close()

	// 建立 SMTP 客戶端
	client, err := smtp.NewClient(conn, hostname)
	if err != nil {
		return fmt.Errorf("❌ 建立 SMTP 客戶端失敗: %w", err)
	}
	defer client.Quit()

	// 如果是 `587`，發送 `STARTTLS`
	if useSTARTTLS {
		fmt.Println("🔍 發送 STARTTLS 指令...")
		if err = client.StartTLS(&tls.Config{ServerName: hostname}); err != nil {
			return fmt.Errorf("❌ STARTTLS 失敗: %w", err)
		}
	}

	// SMTP 身份驗證
	if auth != nil {
		fmt.Println("🔍 進行 SMTP 身份驗證...")
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("❌ SMTP 認證失敗: %w", err)
		}
	}

	// 設定寄件人
	fmt.Println("📨 設定寄件人...")
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("❌ 設定寄件人失敗: %w", err)
	}

	// 設定收件人
	fmt.Println("📨 設定收件人...")
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("❌ 設定收件人 %s 失敗: %w", recipient, err)
		}
	}

	// 發送郵件內容
	fmt.Println("📨 發送郵件內容...")
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("❌ SMTP Data 指令失敗: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("❌ 郵件內容寫入失敗: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("❌ 關閉郵件內容流失敗: %w", err)
	}

	fmt.Println("✅ 郵件發送成功！")
	return nil
}
