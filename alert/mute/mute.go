package mute

import (
	"fmt"
	"shared-lib/models"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 將接口重命名為 MuteService
type MuteService interface {
	// 基本 CRUD
	Create(mute *models.AlertMute) error
	Get(id int64) (*models.AlertMute, error)
	List(realm string) ([]models.AlertMute, error)
	Update(mute *models.AlertMute) error
	Delete(id int64) error

	// 抑制狀態管理
	Enable(id int64) error
	Disable(id int64) error
	IsActive(id int64) bool

	// 規則關聯管理
	AddRules(muteID int64, ruleIDs []int64) error
	RemoveRules(muteID int64, ruleIDs []int64) error
	GetRules(muteID int64) ([]models.AlertRule, error)

	// 檢查操作
	IsRuleMuted(ruleID int64, t time.Time) bool
	ValidateTimeRange(start, end time.Time) error
	CheckName(mute models.AlertMute) (bool, string)
}

// 具體實現
type Service struct {
	db      *gorm.DB
	logger  *zap.Logger
	cron    *cron.Cron
	cronMap map[int64]cron.EntryID
}

func NewService(db *gorm.DB, logger *zap.Logger, cronjob *cron.Cron) *Service {
	return &Service{
		db:      db,
		logger:  logger,
		cron:    cronjob,
		cronMap: make(map[int64]cron.EntryID),
	}
}

// Create 創建抑制規則
func (s *Service) Create(mute *models.AlertMute) error {
	if err := s.ValidateTimeRange(
		time.Unix(int64(mute.StartTime), 0),
		time.Unix(int64(mute.EndTime), 0),
	); err != nil {
		return err
	}

	mute.State = s.initialState(mute.Enabled)
	expr := s.generateCronExpr(mute)
	mute.CronExpression = &expr

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(mute).Error; err != nil {
			return fmt.Errorf("創建抑制規則失敗: %w", err)
		}

		if err := s.createRuleAssociations(tx, mute); err != nil {
			return err
		}

		return nil
	})

	if err == nil && mute.Enabled {
		s.scheduleMute(mute)
	}

	return err
}

// Get 獲取抑制規則
func (s *Service) Get(id int64) (*models.AlertMute, error) {
	var mute models.AlertMute
	err := s.db.Preload("AlertRules").First(&mute, id).Error
	if err != nil {
		return nil, fmt.Errorf("獲取抑制規則失敗: %w", err)
	}
	return &mute, nil
}

// List 獲取抑制規則列表
func (s *Service) List(realm string) ([]models.AlertMute, error) {
	var mutes []models.AlertMute
	err := s.db.Preload("AlertRules").
		Where("realm_name = ?", realm).
		Find(&mutes).Error
	if err != nil {
		return nil, fmt.Errorf("獲取抑制規則列表失敗: %w", err)
	}
	return mutes, nil
}

// Update 更新抑制規則
func (s *Service) Update(mute *models.AlertMute) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 更新基本信息
		if err := tx.Model(mute).Updates(map[string]interface{}{
			"name":            mute.Name,
			"description":     mute.Description,
			"enabled":         mute.Enabled,
			"start_time":      mute.StartTime,
			"end_time":        mute.EndTime,
			"state":           mute.State,
			"cron_expression": mute.CronExpression,
			"cron_period":     mute.CronPeriod,
			"repeat_time":     mute.RepeatTime,
		}).Error; err != nil {
			return fmt.Errorf("更新抑制規則失敗: %w", err)
		}

		// 更新規則關聯
		if err := tx.Model(mute).Association("AlertRules").Replace(mute.AlertRules); err != nil {
			return fmt.Errorf("更新規則關聯失敗: %w", err)
		}

		return nil
	})
}

// Delete 刪除抑制規則
func (s *Service) Delete(id int64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("alert_mute_id = ?", id).Delete(&models.AlertMuteRule{}).Error; err != nil {
			return fmt.Errorf("刪除規則關聯失敗: %w", err)
		}
		if err := tx.Delete(&models.AlertMute{}, id).Error; err != nil {
			return fmt.Errorf("刪除抑制規則失敗: %w", err)
		}
		return nil
	})
}

// Enable 啟用抑制規則
func (s *Service) Enable(id int64) error {
	mute, err := s.Get(id)
	if err != nil {
		return err
	}

	mute.Enabled = true
	mute.State = "scheduled"

	if err := s.Update(mute); err != nil {
		return err
	}

	s.scheduleMute(mute)
	return nil
}

// Disable 禁用抑制規則
func (s *Service) Disable(id int64) error {
	if entryID, ok := s.cronMap[id]; ok {
		s.cron.Remove(entryID)
		delete(s.cronMap, id)
	}

	return s.db.Model(&models.AlertMute{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"enabled": false,
			"state":   "disabled",
		}).Error
}

// IsActive 檢查抑制規則是否處於活動狀態
func (s *Service) IsActive(id int64) bool {
	var mute models.AlertMute
	if err := s.db.First(&mute, id).Error; err != nil {
		return false
	}
	return mute.State == "active"
}

// AddRules 添加規則關聯
func (s *Service) AddRules(muteID int64, ruleIDs []int64) error {
	var rules []models.AlertMuteRule
	for _, ruleID := range ruleIDs {
		rules = append(rules, models.AlertMuteRule{
			AlertMuteID: muteID,
			AlertRuleID: ruleID,
		})
	}
	return s.db.Create(&rules).Error
}

// RemoveRules 移除規則關聯
func (s *Service) RemoveRules(muteID int64, ruleIDs []int64) error {
	return s.db.Where("alert_mute_id = ? AND alert_rule_id IN ?", muteID, ruleIDs).
		Delete(&models.AlertMuteRule{}).Error
}

// GetRules 獲取關聯的規則
func (s *Service) GetRules(muteID int64) ([]models.AlertRule, error) {
	var rules []models.AlertRule
	err := s.db.Joins("JOIN alert_mute_rules ON alert_rules.id = alert_mute_rules.alert_rule_id").
		Where("alert_mute_rules.alert_mute_id = ?", muteID).
		Find(&rules).Error
	return rules, err
}

// IsRuleMuted 檢查規則是否被抑制
func (s *Service) IsRuleMuted(ruleID int64, t time.Time) bool {
	var count int64
	err := s.db.Model(&models.AlertMute{}).
		Joins("JOIN alert_mute_rules ON alert_mutes.id = alert_mute_rules.alert_mute_id").
		Where("alert_mute_rules.alert_rule_id = ? AND alert_mutes.enabled = ? AND alert_mutes.state = ?",
			ruleID, true, "active").
		Count(&count).Error
	return err == nil && count > 0
}

// ValidateTimeRange 驗證時間範圍
func (s *Service) ValidateTimeRange(start, end time.Time) error {
	if end.Before(start) {
		return fmt.Errorf("結束時間不能早於開始時間")
	}
	if time.Now().After(start) && time.Now().Before(end) {
		return fmt.Errorf("當前時間不能在抑制時間範圍內")
	}
	return nil
}

// 內部輔助方法

func (s *Service) initialState(enabled bool) string {
	if enabled {
		return "scheduled"
	}
	return "disabled"
}

func (s *Service) generateCronExpr(mute *models.AlertMute) string {
	t := time.Unix(int64(mute.StartTime), 0)

	switch {
	case mute.RepeatTime.Never:
		return fmt.Sprintf("0 %d %d %d %d *", t.Minute(), t.Hour(), t.Day(), t.Month())
	case mute.RepeatTime.Daily:
		return fmt.Sprintf("0 %d %d * * *", t.Minute(), t.Hour())
	case mute.RepeatTime.Weekly:
		return fmt.Sprintf("0 %d %d * * %s", t.Minute(), t.Hour(), s.getWeekdays(mute.RepeatTime))
	case mute.RepeatTime.Monthly:
		return fmt.Sprintf("0 %d %d %d * *", t.Minute(), t.Hour(), t.Day())
	default:
		return ""
	}
}

func (s *Service) getWeekdays(repeat models.RepeatTime) string {
	var days []string
	if repeat.Mon {
		days = append(days, "1")
	}
	if repeat.Tue {
		days = append(days, "2")
	}
	if repeat.Wed {
		days = append(days, "3")
	}
	if repeat.Thu {
		days = append(days, "4")
	}
	if repeat.Fri {
		days = append(days, "5")
	}
	if repeat.Sat {
		days = append(days, "6")
	}
	if repeat.Sun {
		days = append(days, "0")
	}
	if len(days) == 0 {
		return "*"
	}
	return strings.Join(days, ",")
}

func (s *Service) createRuleAssociations(tx *gorm.DB, mute *models.AlertMute) error {
	if len(mute.AlertRules) == 0 {
		return nil
	}

	var rules []models.AlertMuteRule
	for _, rule := range mute.AlertRules {
		rules = append(rules, models.AlertMuteRule{
			AlertMuteID: mute.ID,
			AlertRuleID: rule.ID,
		})
	}

	return tx.Create(&rules).Error
}

func (s *Service) scheduleMute(mute *models.AlertMute) {
	entryID, err := s.cron.AddFunc(*mute.CronExpression, func() {
		s.handleMuteActivation(mute)
	})

	if err != nil {
		s.logger.Error("設置抑制定時任務失敗",
			zap.Int64("mute_id", mute.ID),
			zap.Error(err))
		return
	}

	s.cronMap[mute.ID] = entryID
}

func (s *Service) handleMuteActivation(mute *models.AlertMute) {
	// 更新狀態為活動中
	if err := s.db.Model(mute).Update("state", "active").Error; err != nil {
		s.logger.Error("更新抑制狀態失敗",
			zap.Int64("mute_id", mute.ID),
			zap.Error(err))
		return
	}

	// 設置結束定時器
	time.AfterFunc(time.Duration(*mute.CronPeriod)*time.Second, func() {
		newState := "scheduled"
		if mute.RepeatTime.Never {
			newState = "ended"
		}

		if err := s.db.Model(mute).Update("state", newState).Error; err != nil {
			s.logger.Error("更新抑制結束狀態失敗",
				zap.Int64("mute_id", mute.ID),
				zap.Error(err))
		}
	})
}

// 檢查名稱是否重複
func (s *Service) CheckName(mute models.AlertMute) (bool, string) {
	var existingMute models.AlertMute
	result := s.db.Where(&models.AlertMute{
		RealmName: mute.RealmName,
		Name:      mute.Name,
	}).First(&existingMute)

	if result.RowsAffected > 0 && existingMute.ID != mute.ID {
		return false, "name existed in other alert_mute"
	}
	return true, ""
}
