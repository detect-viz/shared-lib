package validate

import (
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
)

var (
	// emailRegex 郵件地址正則
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// Email 驗證郵件地址
func Email(email string) error {
	// 使用 net/mail 包驗證
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format: %v", err)
	}

	// 使用正則進行額外驗證
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateEmailConfig 驗證郵件配置
func ValidateEmailConfig(config map[string]string) error {
	required := []string{"host", "port", "from", "to"}
	for _, field := range required {
		if err := NotEmpty(field, config[field]); err != nil {
			return err
		}
	}

	// 驗證端口
	if port, err := strconv.Atoi(config["port"]); err != nil {
		return fmt.Errorf("invalid port number")
	} else if port < 1 || port > 65535 {
		return fmt.Errorf("port number out of range")
	}

	// 驗證郵件地址
	addresses := []struct {
		name  string
		value string
	}{
		{"from", config["from"]},
		{"to", config["to"]},
		{"cc", config["cc"]},
		{"bcc", config["bcc"]},
		{"reply_to", config["reply_to"]},
	}

	for _, addr := range addresses {
		if addr.value != "" {
			if _, err := mail.ParseAddress(addr.value); err != nil {
				return fmt.Errorf("invalid %s address: %v", addr.name, err)
			}
		}
	}

	return nil
}

type EmailValidator struct{}

func (v *EmailValidator) Validate(config map[string]string) error {
	// 驗證郵件配置
	if err := NotEmpty("host", config["host"]); err != nil {
		return err
	}
	if err := NotEmpty("port", config["port"]); err != nil {
		return err
	}
	if err := Email(config["to"]); err != nil {
		return err
	}
	return nil
}
