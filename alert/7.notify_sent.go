package alert

import "github.com/detect-viz/shared-lib/models/common"

// 發送通知
func (s *Service) Test(typ string) error {
	info := common.NotifySetting{
		Type: typ,
		Config: map[string]string{
			"title":   "Test " + typ,
			"message": "This is a test message from alert system.",
		},
	}

	return s.notifyService.Send(info)

}
