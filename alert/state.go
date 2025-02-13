package alert

import (
	"shared-lib/interfaces"
	"shared-lib/models"

	"go.uber.org/zap"
)

// AlertStateManager 告警狀態管理
type AlertStateManager struct {
	db     interfaces.Database
	logger interfaces.Logger
}

// NewAlertStateManager 創建告警狀態管理器
func NewAlertStateManager(db interfaces.Database, logger interfaces.Logger) *AlertStateManager {
	return &AlertStateManager{
		db:     db,
		logger: logger.With(zap.String("component", "state_manager")),
	}
}

// GetAndUpdateState 獲取並更新告警狀態
func (m *AlertStateManager) GetAndUpdateState(rule models.CheckRule, value float64, timestamp int64) (models.AlertState, bool, error) {
	// 1. 獲取當前狀態
	state, err := m.db.GetAlertState(rule.RuleID, rule.ResourceName, rule.Metric)
	if err != nil {
		return state, false, err
	}

	// 2. 檢查是否超過閾值
	exceeded := CheckThreshold(rule, rule.Operator, value)

	// 3. 更新狀態
	if exceeded {
		state = m.updateExceededState(state, value, timestamp)

		// 檢查持續時間是否達到告警條件
		if state.StackDuration >= rule.Duration {
			if err := m.db.SaveAlertState(state); err != nil {
				return state, false, err
			}
			return state, true, nil
		}
	} else {
		state = m.resetState(state, value, timestamp)
	}

	if err := m.db.SaveAlertState(state); err != nil {
		return state, false, err
	}

	return state, false, nil
}

// updateExceededState 更新超過閾值的狀態
func (m *AlertStateManager) updateExceededState(state models.AlertState, value float64, timestamp int64) models.AlertState {
	if state.StartTime == 0 {
		state.StartTime = timestamp
		state.LastValue = value
		state.StackDuration = 0
	} else {
		timeDiff := timestamp - state.LastTime
		state.StackDuration += int(timeDiff / 60)
	}
	state.LastTime = timestamp
	state.LastValue = value
	return state
}

// resetState 重置狀態
func (m *AlertStateManager) resetState(state models.AlertState, value float64, timestamp int64) models.AlertState {
	state.StartTime = 0
	state.LastTime = timestamp
	state.LastValue = value
	state.StackDuration = 0
	return state
}

// GetAndUpdateAmplitudeState 獲取並更新振幅檢查狀態
func (m *AlertStateManager) GetAndUpdateAmplitudeState(rule models.CheckRule, value float64, timestamp int64) (models.AlertState, bool, error) {
	// 1. 獲取當前狀態
	state, err := m.db.GetAlertState(rule.RuleID, rule.ResourceName, rule.Metric)
	if err != nil {
		return state, false, err
	}

	// 2. 如果是第一次檢查
	if state.LastValue == 0 {
		state = m.initializeState(state, value, timestamp)
		if err := m.db.SaveAlertState(state); err != nil {
			return state, false, err
		}
		return state, false, nil
	}

	// 3. 計算變化幅度
	amplitude := ((value - state.LastValue) / state.LastValue) * 100

	// 4. 檢查是否超過閾值
	exceeded := CheckThreshold(rule, rule.Operator, amplitude)

	// 5. 更新狀態
	if exceeded {
		state = m.updateAmplitudeState(state, value, timestamp, amplitude)
	} else {
		state = m.resetAmplitudeState(state, value, timestamp)
	}

	if err := m.db.SaveAlertState(state); err != nil {
		return state, false, err
	}

	return state, exceeded, nil
}

// initializeState 初始化狀態
func (m *AlertStateManager) initializeState(state models.AlertState, value float64, timestamp int64) models.AlertState {
	state.LastTime = timestamp
	state.LastValue = value
	return state
}

// updateAmplitudeState 更新振幅狀態
func (m *AlertStateManager) updateAmplitudeState(state models.AlertState, value float64, timestamp int64, amplitude float64) models.AlertState {
	if state.StartTime == 0 {
		state.StartTime = timestamp
		state.PreviousValue = state.LastValue
		state.Amplitude = amplitude
	}
	state.LastTime = timestamp
	state.LastValue = value
	return state
}

// resetAmplitudeState 重置振幅狀態
func (m *AlertStateManager) resetAmplitudeState(state models.AlertState, value float64, timestamp int64) models.AlertState {
	state.StartTime = 0
	state.PreviousValue = 0
	state.Amplitude = 0
	state.LastTime = timestamp
	state.LastValue = value
	return state
}
