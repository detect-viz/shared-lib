package interfaces

import "github.com/detect-viz/shared-lib/models"

// Channel 通知器接口
type Channel interface {
	Send(message models.AlertMessage) error
	Test() error
}
