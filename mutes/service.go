package mutes

import (
	"fmt"
	"time"

	"github.com/detect-viz/shared-lib/models"
	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/storage/mysql"
	"github.com/google/wire"

	"go.uber.org/zap"
)

var MuteSet = wire.NewSet(
	NewService,
	wire.Bind(new(Service), new(*serviceImpl)),
)

// 具體實現
type serviceImpl struct {
	mysql  *mysql.Client
	logger *zap.Logger
}

func NewService(mysql *mysql.Client, logger *zap.Logger) *serviceImpl {
	return &serviceImpl{
		mysql:  mysql,
		logger: logger,
	}
}

func (s *serviceImpl) Create(mute *models.Mute) error {
	return s.mysql.CreateMute(mute)
}

func (s *serviceImpl) Get(id int64) (*models.Mute, error) {
	return s.mysql.GetMute(id)
}

func (s *serviceImpl) List(realm string) ([]models.Mute, error) {
	return s.mysql.ListMutes(realm)
}

func (s *serviceImpl) Update(mute *models.Mute) error {
	return s.mysql.UpdateMute(mute)
}

func (s *serviceImpl) Delete(id int64) error {
	return s.mysql.DeleteMute(id)
}

// GetResourceGroups 獲取規則關聯
func (s *serviceImpl) GetResourceGroups(muteID int64) ([]models.ResourceGroup, error) {
	return s.mysql.GetMuteResourceGroups(muteID)
}

// IsRuleMuted 判斷規則是否處於靜音狀態
func (s *serviceImpl) IsRuleMuted(resourceGroupID int64, t time.Time) bool {
	mutes, err := s.mysql.GetMutesByResourceGroup(resourceGroupID)
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
func (s *serviceImpl) GetMutePeriod(resourceGroupID int64, t time.Time) (int64, int64) {
	mutes, err := s.mysql.GetMutesByResourceGroup(resourceGroupID)
	if err != nil {
		s.logger.Error("查詢 Mute 失敗", zap.Error(err))
		return 0, 0
	}

	for _, mute := range mutes {
		if mute.IsMuted(t) {
			return mute.GetStartTime(t), mute.GetEndTime(t)
		}
	}
	return 0, 0
}

// GetOptions 提供 UI 選項
func (s *serviceImpl) GetOptions(typ string) []models.OptionResponse {
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
func (s *serviceImpl) ValidateTimeRange(start, end time.Time) error {
	if end.Before(start) {
		return fmt.Errorf("結束時間不能早於開始時間")
	}

	now := time.Now()
	if now.After(start) && now.Before(end) {
		s.logger.Warn("警告：當前時間已在抑制範圍內，請確認更新影響")
	}

	return nil
}

// GetMonthsOptions 回傳月份選項
func (s *serviceImpl) getMonthsOptions() []common.OptionResponse {
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
func (s *serviceImpl) getWeekdaysOptions() []common.OptionResponse {
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
