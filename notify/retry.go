package notify

import (
	"shared-lib/notify/errors"
	"time"
)

// RetryPolicy 重試策略
type RetryPolicy struct {
	MaxRetries  int           // 最大重試次數
	InitialWait time.Duration // 初始等待時間
	MaxWait     time.Duration // 最大等待時間
}

// DefaultRetryPolicy 默認重試策略
var DefaultRetryPolicy = RetryPolicy{
	MaxRetries:  3,
	InitialWait: time.Second,
	MaxWait:     30 * time.Second,
}

// ShouldRetry 判斷是否應該重試
func (p *RetryPolicy) ShouldRetry(attempt int, err error) bool {
	if attempt >= p.MaxRetries {
		return false
	}

	// 根據錯誤類型判斷是否重試
	switch err.(type) {
	case *errors.NotifyError:
		return true
	default:
		return false
	}
}

// GetBackoff 獲取退避時間
func (p *RetryPolicy) GetBackoff(attempt int) time.Duration {
	// 指數退避
	wait := p.InitialWait * time.Duration(1<<uint(attempt))
	if wait > p.MaxWait {
		wait = p.MaxWait
	}
	return wait
}

// RetryWithPolicy 使用重試策略執行
func RetryWithPolicy(policy RetryPolicy, fn func() error) error {
	var err error
	for attempt := 0; attempt < policy.MaxRetries; attempt++ {
		if err = fn(); err == nil {
			return nil
		}

		if !policy.ShouldRetry(attempt, err) {
			return err
		}

		time.Sleep(policy.GetBackoff(attempt))
	}
	return err
}
