package errors

import "fmt"

// NotifyError 通知錯誤
type NotifyError struct {
	Component string // 錯誤發生的組件
	Message   string // 錯誤信息
	Err       error  // 原始錯誤
}

// Error 實現 error 接口
func (e *NotifyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Component, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Component, e.Message)
}

// Unwrap 實現 errors.Unwrap 接口
func (e *NotifyError) Unwrap() error {
	return e.Err
}

// NewNotifyError 創建通知錯誤
func NewNotifyError(component, message string, err error) error {
	return &NotifyError{
		Component: component,
		Message:   message,
		Err:       err,
	}
}

// 預定義錯誤
var (
	// ErrChannelNotFound 通知器未找到
	ErrChannelNotFound = fmt.Errorf("channel not found")
	// ErrUnsupportedType 不支持的通知類型
	ErrUnsupportedType = fmt.Errorf("unsupported channel type")
	// ErrMissingRequiredConf 缺少必要配置
	ErrMissingRequiredConf = fmt.Errorf("missing required config")
	// ErrInvalidConfig 無效的配置
	ErrInvalidConfig = fmt.Errorf("invalid config")
	// ErrSendFailed 發送失敗
	ErrSendFailed = fmt.Errorf("send notify failed")
)
