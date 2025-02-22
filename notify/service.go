package notify

import (
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/notify/validate"
)

// Service 通知服務
type Service struct {
	validator *validate.Validator
}

// NewService 創建通知服務
func NewService() *Service {
	return &Service{
		validator: validate.New(),
	}
}

// 發送通知
func (s *Service) Send(config common.NotifyConfig) error {
	switch config.Type {
	case "email":
		return s.sendEmail(config)
	default:
		return s.sendWebhook(config)
	}
}

func (s *Service) Validate(config common.NotifyConfig) error {
	// 驗證配置
	if err := s.validator.Validate(config); err != nil {
		return err
	}
	return nil
}

func (s *Service) GetNotifyMethods() []string {
	return getNotifyMethods()
}

func (s *Service) GetNotifyOptions() map[string]map[string][]string {
	return getNotifyOptions()
}
