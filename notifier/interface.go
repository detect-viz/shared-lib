package notifier

import (
	"github.com/detect-viz/shared-lib/models/common"
)

// 通知服務介面
type Service interface {
	Send(notify common.NotifySetting) error
	Validate(config common.NotifySetting) error
}
