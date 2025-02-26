package notifier

import (
	"github.com/detect-viz/shared-lib/models"
)

// 通知服務介面
type Service interface {
	Send(notify models.NotifySetting) error
	Validate(config models.NotifySetting) error
	GetNotifyMethods() []string
	GetNotifyOptions() map[string]map[string][]string
}
