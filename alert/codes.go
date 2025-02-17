package alert

import (
	"shared-lib/models"
)

// GetSeverityByName 根據名稱獲取告警等級 (使用 map 代替 switch-case)
func (s *Service) GetSeverityByName(name string) models.Code {
	// 直接讀取 `Level`，避免多次存取 `s.config.AlertCodes.Level`
	level := s.config.AlertCodes.Level

	// 建立映射表 (Map)
	severityMap := map[string]models.Code{
		level.Crit.Name: level.Crit,
		level.Warn.Name: level.Warn,
		level.Info.Name: level.Info, // 預設等級
	}

	// 查找等級，找不到則回傳 `Info`
	if severity, ok := severityMap[name]; ok {
		return severity
	}
	return level.Info
}
