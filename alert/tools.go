package alert

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/detect-viz/shared-lib/models"

	"go.uber.org/zap"
)

// 檢查是否處於靜音時段
func (s *Service) applySilence(rule *models.CheckRule) {
	now := time.Now().Unix()

	if rule.SilenceStart != nil && rule.SilenceEnd != nil {
		if now >= *rule.SilenceStart && now <= *rule.SilenceEnd {
			rule.ContactState = s.global.Code.State.Contact.Silence.Name
		}
	}

}

// 檢查是否處於抑制時段
func (s *Service) applyMute(rule *models.CheckRule) {
	now := time.Now().Unix()

	if rule.MuteStart != nil && rule.MuteEnd != nil {
		if now >= *rule.MuteStart && now <= *rule.MuteEnd {
			rule.ContactState = s.global.Code.State.Contact.Muting.Name
		}
	}
}

func (s *Service) parseMetricValue(value interface{}) (float64, error) {
	switch v := value.(type) {
	case string:
		floatValue, err := strconv.ParseFloat(v, 64)
		if err != nil {
			s.logger.Error("轉換指標值失敗", zap.Error(err))
			return 0, err
		}
		return floatValue, nil
	case float64:
		return v, nil
	default:
		s.logger.Error("無法解析指標值", zap.Any("value", value))
		return 0, fmt.Errorf("無法解析指標值: %v", value)
	}
}

// lockFile 鎖定檔案
func (s *Service) LockFile(path string) error {
	lockPath := path + ".lock"
	for i := 0; i < 3; i++ { // 重試3次
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			// 創建鎖檔案
			if err := os.WriteFile(lockPath, []byte{}, 0644); err != nil {
				return err
			}
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("檔案已被鎖定: %s", path)
}

// 解除檔案鎖定
func (s *Service) UnlockFile(path string) error {
	return os.Remove(path + ".lock")
}

// 帶鎖寫入檔案
func (s *Service) WriteWithLock(path string, data []byte) error {
	if err := s.LockFile(path); err != nil {
		return err
	}
	defer s.UnlockFile(path)

	return os.WriteFile(path, data, 0644)
}

// CheckName 檢查名稱是否存在
func (s *Service) CheckName(name string) bool {
	return true
}

func (s *Service) GetLogger() *zap.Logger {
	return s.logger.GetLogger()
}
