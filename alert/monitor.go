package alert

import (
	"shared-lib/models"
)

func (s *Service) GetHistoryAlert(user models.SSOUser) ([]models.HistoryAlert, error) {
	return nil, nil
}

func (s *Service) GetCurrentAlert(user models.SSOUser) ([]models.CurrentAlert, error) {
	return nil, nil
}

func (s *Service) GetHistoryAlertMetric(user models.SSOUser, body models.HistoryAlert) ([]models.MetricResponse, error) {
	return nil, nil
}
