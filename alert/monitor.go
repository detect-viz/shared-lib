package alert

import (
	"github.com/detect-viz/shared-lib/models"
)

func (s *Service) ListAlertState(realm string) ([]models.CurrentAlert, error) {
	//TODO: 實作 Alert State 查詢
	return nil, nil
}

func (s *Service) ListAlertHistory(realm string) ([]models.HistoryAlert, error) {
	//TODO: 實作 Trigger Log 查詢
	return nil, nil
}
