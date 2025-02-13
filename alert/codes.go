package alert

import (
	"shared-lib/models"
)

// GetSeverityByName 根據名稱獲取告警等級
func (s *Service) GetSeverityByName(name string) models.Code {
	level := s.config.AlertCodes.Level
	switch name {
	case level.Crit.Name:
		return level.Crit
	case level.Warn.Name:
		return level.Warn
	default:
		return level.Info
	}
}
