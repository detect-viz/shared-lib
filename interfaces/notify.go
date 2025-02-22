package interfaces

import (
	"github.com/detect-viz/shared-lib/models"
)

// 通知服務介面
type NotifyService interface {
	Send(notify models.NotifyConfig) error
	Validate(config models.NotifyConfig) error
	GetNotifyMethods() []string
	GetNotifyOptions() map[string]map[string][]string
}
