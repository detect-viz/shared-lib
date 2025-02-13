package common

import (
	"fmt"
	"time"
)

// ValidateTimestamp 驗證時間戳
func ValidateTimestamp(timestamp int64) error {
	if timestamp <= 0 {
		return fmt.Errorf(ErrInvalidTime, timestamp)
	}
	if timestamp > time.Now().Unix() {
		return fmt.Errorf(ErrInvalidTime, timestamp)
	}
	if (time.Now().Unix() - timestamp) > MaxDataAge {
		return fmt.Errorf(ErrExpiredData, timestamp)
	}
	return nil
}

// ValidateName 驗證名稱
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf(ErrInvalidValue, "empty name")
	}
	return nil
}

// ConvertUnit 單位轉換
func ConvertUnit(value float64, fromUnit string, toUnit string) float64 {
	switch {
	case fromUnit == "KB" && toUnit == "B":
		return value * BytesPerKB
	case fromUnit == "MB" && toUnit == "B":
		return value * BytesPerMB
	case fromUnit == "GB" && toUnit == "B":
		return value * BytesPerGB
	case fromUnit == "ms" && toUnit == "s":
		return value / MilliToSec
	default:
		return value
	}
}

// CopyMap 複製 map
func CopyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{})
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
