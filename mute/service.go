package mute

import (
	"fmt"
	"strconv"
	"time"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 具體實現
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

// Create 創建抑制規則
func (s *Service) Create(mute *models.Mute) error {
	startMinute, _ := strconv.ParseInt(mute.TimeIntervals[0].StartMinute, 10, 64)
	endMinute, _ := strconv.ParseInt(mute.TimeIntervals[0].EndMinute, 10, 64)

	if err := s.ValidateTimeRange(
		time.Unix(startMinute, 0),
		time.Unix(endMinute, 0),
	); err != nil {
		return err
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(mute).Error; err != nil {
			return fmt.Errorf("創建抑制規則失敗: %w", err)
		}
		if err := s.createRuleAssociations(tx, mute); err != nil {
			return err
		}
		return nil
	})

	return err
}

// Get 獲取抑制規則
func (s *Service) Get(id int64) (*models.Mute, error) {
	var mute models.Mute
	err := s.db.Preload("ResourceGroups").First(&mute, id).Error
	if err != nil {
		return nil, fmt.Errorf("獲取抑制規則失敗: %w", err)
	}
	return &mute, nil
}

// List 獲取抑制規則列表
func (s *Service) List(realm string) ([]models.Mute, error) {
	var mutes []models.Mute
	err := s.db.Where("realm_name = ?", realm).Find(&mutes).Error
	if err != nil {
		return nil, fmt.Errorf("獲取抑制規則列表失敗: %w", err)
	}

	return mutes, nil
}

// Update 更新抑制規則
func (s *Service) Update(mute *models.Mute) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(mute).Updates(map[string]interface{}{
			"name":           mute.Name,
			"years":          mute.Years,
			"time_intervals": mute.TimeIntervals,
			"repeat_type":    mute.RepeatType,
			"weekdays":       mute.Weekdays,
			"months":         mute.Months,
		}).Error; err != nil {
			return fmt.Errorf("更新抑制規則失敗: %w", err)
		}

		if err := tx.Model(mute).Association("ResourceGroups").Replace(mute.ResourceGroups); err != nil {
			return fmt.Errorf("更新規則關聯失敗: %w", err)
		}

		return nil
	})
}

// Delete 刪除抑制規則
func (s *Service) Delete(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("mute_id = ?", id).Delete(&models.MuteResourceGroup{}).Error; err != nil {
			return fmt.Errorf("刪除規則關聯失敗: %w", err)
		}
		if err := tx.Delete(&models.Mute{}, id).Error; err != nil {
			return fmt.Errorf("刪除抑制規則失敗: %w", err)
		}
		return nil
	})
}

// AddResourceGroups 批量插入
func (s *Service) AddResourceGroups(muteID int64, resourceGroupIDs []int64) error {
	var records []models.MuteResourceGroup
	for _, id := range resourceGroupIDs {
		records = append(records, models.MuteResourceGroup{
			MuteID:          muteID,
			ResourceGroupID: id,
		})
	}

	return s.db.Create(&records).Error
}

// RemoveResourceGroups 移除關聯
func (s *Service) RemoveResourceGroups(muteID int64, resourceGroupIDs []int64) error {
	if len(resourceGroupIDs) == 0 {
		return nil
	}
	return s.db.Where("mute_id = ? AND resource_group_id IN ?", muteID, resourceGroupIDs).
		Delete(&models.MuteResourceGroup{}).Error
}

// GetResourceGroups 獲取規則關聯
func (s *Service) GetResourceGroups(muteID int64) ([]models.ResourceGroup, error) {
	var resourceGroups []models.ResourceGroup
	err := s.db.Joins("JOIN mute_resource_groups ON resource_groups.id = mute_resource_groups.resource_group_id").
		Where("mute_resource_groups.mute_id = ?", muteID).
		Find(&resourceGroups).Error
	return resourceGroups, err
}

// IsRuleMuted 判斷規則是否處於靜音狀態
func (s *Service) IsRuleMuted(resourceGroupID int64, t time.Time) bool {
	var mutes []models.Mute
	err := s.db.
		Joins("JOIN mute_resource_groups ON mute_resource_groups.mute_id = mutes.id").
		Where("mute_resource_groups.resource_group_id = ?", resourceGroupID).
		Find(&mutes).Error
	if err != nil {
		s.logger.Error("查詢 Mute 失敗", zap.Error(err))
		return false
	}

	for _, mute := range mutes {
		if mute.IsMuted(t) {
			return true
		}
	}
	return false
}

// GetMutePeriod 取得 `mute_start, mute_end`
func (s *Service) GetMutePeriod(resourceGroupID int64, t time.Time) (int64, int64) {
	var mutes []models.Mute
	err := s.db.
		Joins("JOIN mute_resource_groups ON mute_resource_groups.mute_id = mutes.id").
		Where("mute_resource_groups.resource_group_id = ?", resourceGroupID).
		Find(&mutes).Error
	if err != nil {
		s.logger.Error("查詢 Mute 失敗", zap.Error(err))
		return 0, 0
	}

	for _, mute := range mutes {
		if mute.IsMuted(t) {
			muteStart := mute.GetStartTime(t)
			muteEnd := mute.GetEndTime(t)
			return muteStart, muteEnd
		}
	}
	return 0, 0
}

// GetOptions 提供 UI 選項
func (s *Service) GetOptions(typ string) []models.OptionResponse {
	switch typ {
	case "weekdays":
		return s.getWeekdaysOptions()
	case "months":
		return s.getMonthsOptions()
	default:
		return nil
	}
}

// ValidateTimeRange 驗證時間範圍
func (s *Service) ValidateTimeRange(start, end time.Time) error {
	if end.Before(start) {
		return fmt.Errorf("結束時間不能早於開始時間")
	}

	now := time.Now()
	if now.After(start) && now.Before(end) {
		s.logger.Warn("警告：當前時間已在抑制範圍內，請確認更新影響")
	}

	return nil
}

func (s *Service) createRuleAssociations(tx *gorm.DB, mute *models.Mute) error {
	if len(*mute.ResourceGroups) == 0 {
		return nil
	}

	var rules []models.MuteResourceGroup
	for _, rule := range *mute.ResourceGroups {
		rules = append(rules, models.MuteResourceGroup{
			MuteID:          mute.ID,
			ResourceGroupID: int64(rule.ID),
		})
	}

	return tx.Create(&rules).Error
}

// GetMonthsOptions 回傳月份選項
func (s *Service) getMonthsOptions() []common.OptionResponse {
	return []common.OptionResponse{
		{Text: "January", Value: "1"},
		{Text: "February", Value: "2"},
		{Text: "March", Value: "3"},
		{Text: "April", Value: "4"},
		{Text: "May", Value: "5"},
		{Text: "June", Value: "6"},
		{Text: "July", Value: "7"},
		{Text: "August", Value: "8"},
		{Text: "September", Value: "9"},
		{Text: "October", Value: "10"},
		{Text: "November", Value: "11"},
		{Text: "December", Value: "12"},
	}
}

// GetWeekdaysOptions 回傳星期選項
func (s *Service) getWeekdaysOptions() []common.OptionResponse {
	return []common.OptionResponse{
		{Text: "Monday", Value: "monday"},
		{Text: "Tuesday", Value: "tuesday"},
		{Text: "Wednesday", Value: "wednesday"},
		{Text: "Thursday", Value: "thursday"},
		{Text: "Friday", Value: "friday"},
		{Text: "Saturday", Value: "saturday"},
		{Text: "Sunday", Value: "sunday"},
	}
}

// CheckName 檢查名稱是否重複
func (s *Service) CheckName(mute models.Mute) (bool, string) {
	var existingMute models.Mute
	result := s.db.Where(&models.Mute{
		RealmName: mute.RealmName,
		Name:      mute.Name,
	}).First(&existingMute)

	if result.RowsAffected > 0 && existingMute.ID != mute.ID {
		return false, "name existed in other mute rule"
	}
	return true, ""
}
