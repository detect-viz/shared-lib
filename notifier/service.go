package notifier

import (
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/notifier/validate"
	"github.com/google/wire"
)

var NotifySet = wire.NewSet(
	NewService,
	wire.Bind(new(Service), new(*serviceImpl)),
)

// Service 通知服務
type serviceImpl struct {
	validator *validate.Validator
}

// NewService 創建通知服務
func NewService() *serviceImpl {
	return &serviceImpl{
		validator: validate.New(),
	}
}

// 發送通知
func (s *serviceImpl) Send(config common.NotifySetting) error {
	switch config.Type {
	case "email":
		return s.sendEmail(config)
	default:
		return s.sendWebhook(config)
	}
}

func (s *serviceImpl) Validate(config common.NotifySetting) error {
	// 驗證配置
	if err := s.validator.Validate(config); err != nil {
		return err
	}
	return nil
}

func (s *serviceImpl) GetNotifyMethods() []string {
	return getNotifyMethods()
}

func (s *serviceImpl) GetNotifyOptions() map[string]map[string][]string {
	return getNotifyOptions()
}
