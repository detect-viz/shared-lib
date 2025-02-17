package interfaces

import (
	"shared-lib/models"
	"time"
)

// 通知服務介面
type NotifyService interface {
	// 獲取待處理的文件
	GetPendingFiles() ([]models.FileInfo, error)

	// 通知發送
	Send(notification models.NotificationLog) error
	SendNotification(notification *models.NotificationLog) error

	// 通知管理
	InitPaths(basePath string)
	ProcessNotifications() error

	// 重試機制
	RetryWithPolicy(policy RetryPolicy, fn func() error) error
}

// 通知器工廠介面
type ChannelFactory interface {
	Register(typ models.ChannelType, creator ChannelCreator)
	Create(config models.ChannelConfig) (Channel, error)
}

// RetryPolicy 定義重試策略
type RetryPolicy struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// ChannelCreator 定義通知器創建函數類型
type ChannelCreator func(config models.ChannelConfig) (Channel, error)
