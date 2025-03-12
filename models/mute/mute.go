package mute

import (
	"fmt"

	"strconv"
	"strings"
	"time"

	"github.com/detect-viz/shared-lib/models/resource"
)

// 抑制規則
type Mute struct {
	ID             string                   `json:"id" gorm:"primaryKey"`
	RealmName      string                   `json:"realm_name" gorm:"default:master"`
	Name           string                   `json:"name"`                             // 抑制規則名稱
	Years          []string                 `json:"years" gorm:"type:json"`           // 限制特定年份
	TimeIntervals  []TimeRange              `json:"times" gorm:"type:json"`           // 允許一天內多個時間範圍
	RepeatType     string                   `json:"repeat_type" gorm:"default:never"` // 重複類型: never, daily, weekly, monthly
	Weekdays       []string                 `json:"weekdays" gorm:"type:json"`        // 支援 `"monday:wednesday"`
	Months         []string                 `json:"months" gorm:"type:json"`          // `"may:august"`
	ResourceGroups []resource.ResourceGroup `gorm:"many2many:mute_resource_groups"`
	CreatedAt      int64                    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      int64                    `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      int64                    `json:"-" gorm:"index"`
}

// MuteResourceGroup 告警抑制規則關聯
type MuteResourceGroup struct {
	MuteID          string `json:"mute_id" gorm:"primaryKey"`
	ResourceGroupID int64  `json:"resource_group_id" gorm:"primaryKey"`
}

// 時間範圍 (允許一天內多個時段)
type TimeRange struct {
	StartMinute string `json:"start_time"`
	EndMinute   string `json:"end_time"`
}

// IsMuted 判斷當前時間是否符合 mute 規則
func (m *Mute) IsMuted(t time.Time) bool {
	year := fmt.Sprintf("%d", t.Year())
	month := t.Month().String()
	weekday := t.Weekday().String()
	hourMinute := fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())

	// 判斷年份
	if len(m.Years) > 0 && !matchRangeString(year, m.Years) {
		return false
	}

	// 判斷月份
	if len(m.Months) > 0 && !matchRangeString(month, m.Months) {
		return false
	}

	// 判斷星期
	if len(m.Weekdays) > 0 && !matchRangeString(weekday, m.Weekdays) {
		return false
	}

	// 判斷時間範圍
	for _, tr := range m.TimeIntervals {
		if hourMinute >= tr.StartMinute && hourMinute <= tr.EndMinute {
			return true
		}
	}

	return false
}

// matchRangeString 判斷 `value` 是否符合 `ranges` 定義
func matchRangeString(value string, ranges []string) bool {
	for _, r := range ranges {
		// 直接匹配單個值
		if r == value {
			return true
		}

		// 判斷範圍 (e.g. "monday:wednesday", "1:5", "2020:2022")
		if strings.Contains(r, ":") {
			parts := strings.Split(r, ":")
			if len(parts) != 2 {
				continue
			}

			start, end := parts[0], parts[1]

			// 嘗試將 start 和 end 解析為數字（用於判斷 "1:5" 或 "2020:2022"）
			startInt, errStart := strconv.Atoi(start)
			endInt, errEnd := strconv.Atoi(end)
			valueInt, errValue := strconv.Atoi(value)

			if errStart == nil && errEnd == nil && errValue == nil {
				if valueInt >= startInt && valueInt <= endInt {
					return true
				}
			} else { // 字串比較 (月份, 星期)
				if value >= start && value <= end {
					return true
				}
			}
		}
	}
	return false
}

// GetStartTime 取得當前時間對應的 Mute Start
func (m *Mute) GetStartTime(t time.Time) int64 {
	for _, tr := range m.TimeIntervals {
		if inTimeRange(tr, t) {
			return t.Unix()
		}
	}
	return 0
}

// GetEndTime 取得當前時間對應的 Mute End
func (m *Mute) GetEndTime(t time.Time) int64 {
	for _, tr := range m.TimeIntervals {
		if inTimeRange(tr, t) {
			return t.Unix() + 60*60 // 假設 `mute` 持續 1 小時
		}
	}
	return 0
}

// inTimeRange 判斷當前時間是否落在 `TimeRange`
func inTimeRange(tr TimeRange, t time.Time) bool {
	hourMinute := fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
	return hourMinute >= tr.StartMinute && hourMinute <= tr.EndMinute
}
