package utils

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/notifier/validate"

	"time"
)

// DefaultHTTPClient 默認HTTP客戶端
var DefaultHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// ValidateChannelType 驗證通知器類型
func ValidateChannelType(typ string) error {
	switch typ {
	case "email", "line", "slack", "discord", "teams", "webex", "webhook":
		return nil
	default:
		return fmt.Errorf("unsupported channel type: %s", typ)
	}
}

// ValidateChannelConfig 驗證通知器配置
func ValidateChannelConfig(c common.NotifySetting) error {
	// 1. 驗證類型
	if err := ValidateChannelType(string(c.Type)); err != nil {
		return err
	}

	// 2. 驗證名稱
	if c.Name == "" {
		return fmt.Errorf("channel name is required")
	}

	// 3. 驗證必要配置
	return validate.ValidateConfig(c.Type, c.Config)
}

// RetryWithBackoff 重試邏輯
func RetryWithBackoff(maxRetries int, fn func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}
		// 指數退避
		time.Sleep(time.Duration(1<<uint(i)) * time.Second)
	}
	return fmt.Errorf("max retries reached: %v", err)
}
