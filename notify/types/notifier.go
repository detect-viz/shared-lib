package types

import "shared-lib/models"

// BaseChannel 基礎通知器實現
type BaseChannel struct {
	Config models.ChannelConfig
}

// NewBaseChannel 創建基礎通知器
func NewBaseChannel(config models.ChannelConfig) BaseChannel {
	return BaseChannel{Config: config}
}

// GetConfig 獲取配置
func (n *BaseChannel) GetConfig() models.ChannelConfig {
	return n.Config
}
