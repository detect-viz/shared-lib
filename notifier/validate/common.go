package validate

import (
	"fmt"
	"regexp"
)

// NotEmpty 驗證非空字段
func NotEmpty(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", field)
	}
	return nil
}

// Length 驗證字段長度
func Length(field, value string, min, max int) error {
	l := len(value)
	if l < min || l > max {
		return fmt.Errorf("%s length must be between %d and %d", field, min, max)
	}
	return nil
}

// Range 驗證數值範圍
func Range[T int | int64 | float64](field string, value, min, max T) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %v and %v", field, min, max)
	}
	return nil
}

// Port 驗證端口號
func Port(port int) error {
	return Range("port", port, 1, 65535)
}

// Method 驗證HTTP方法
func Method(method string) error {
	switch method {
	case "GET", "POST", "PUT", "DELETE", "PATCH":
		return nil
	default:
		return fmt.Errorf("invalid HTTP method: %s", method)
	}
}

// InList 驗證值是否在列表中
func InList[T comparable](field string, value T, list []T) error {
	for _, item := range list {
		if value == item {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of %v", field, list)
}

// Pattern 驗證字符串是否匹配正則表達式
func Pattern(field, value, pattern string) error {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return fmt.Errorf("invalid pattern: %v", err)
	}
	if !matched {
		return fmt.Errorf("%s does not match pattern %s", field, pattern)
	}
	return nil
}
