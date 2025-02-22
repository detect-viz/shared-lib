package alert

import (
	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
)

func (s *Service) GetHistoryAlert(user models.SSOUser) ([]models.HistoryAlert, error) {
	//TODO: 實作 Trigger Log 查詢
	return nil, nil
}

func (s *Service) GetCurrentAlert(user models.SSOUser) ([]models.CurrentAlert, error) {
	//TODO: 實作 Alert State 查詢
	return nil, nil
}

func (s *Service) GetHistoryAlertMetric(user models.SSOUser, body models.HistoryAlert) ([]models.MetricResponse, error) {
	return nil, nil
}

// 發送通知
func (s *Service) Test(typ string) error {
	info := common.NotifyConfig{
		Type: typ,
		Config: map[string]string{
			"title":   "Test " + typ,
			"message": "This is a test message from alert system.",
		},
	}
	switch typ {
	case "email":
		return s.notify.Send(info)
	default:
		return s.notify.Send(info)
	}
}
