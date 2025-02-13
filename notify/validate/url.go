package validate

import (
	"fmt"
	"net/url"
	"regexp"
)

var (
	// urlRegex URL正則
	urlRegex = regexp.MustCompile(`^https?://`)
)

// URL 驗證URL
func URL(rawURL string) error {
	// 解析URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	// 檢查協議
	if !urlRegex.MatchString(rawURL) {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	// 檢查主機名
	if u.Host == "" {
		return fmt.Errorf("missing host in URL")
	}

	return nil
}
