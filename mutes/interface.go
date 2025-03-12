package mutes

import (
	"time"

	"github.com/detect-viz/shared-lib/models"
)

type Service interface {
	// 基本 CRUD
	Create(mute *models.Mute) error
	Get(id int64) (*models.Mute, error)
	List(realm string) ([]models.Mute, error)
	Update(mute *models.Mute) error
	Delete(id int64) error

	// 抑制規則選項
	GetOptions(typ string) []models.OptionResponse

	// 規則關聯管理
	GetResourceGroups(muteID int64) ([]models.ResourceGroup, error)

	// 檢查操作
	IsRuleMuted(ruleID int64, t time.Time) bool
	GetMutePeriod(resourceGroupID int64, t time.Time) (int64, int64)
	ValidateTimeRange(start, end time.Time) error
}
