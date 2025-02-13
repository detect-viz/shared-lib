package factory

import (
	"shared-lib/interfaces"
	"shared-lib/models"
	"shared-lib/notify/errors"
	"shared-lib/notify/validate"
)

// ChannelCreator 通知器創建函數類型
type ChannelCreator func(config models.ChannelConfig) (interfaces.Channel, error)

type ChannelFactory struct {
	creators  map[models.ChannelType]ChannelCreator
	validator *validate.Validator
}

func NewFactory() *ChannelFactory {
	return &ChannelFactory{
		creators:  make(map[models.ChannelType]ChannelCreator),
		validator: validate.New(),
	}
}

func (f *ChannelFactory) Register(typ models.ChannelType, creator ChannelCreator) {
	f.creators[typ] = creator
}

func (f *ChannelFactory) Create(config models.ChannelConfig) (interfaces.Channel, error) {
	// 驗證配置
	if err := f.validator.Validate(config); err != nil {
		return nil, err
	}

	// 創建通知器
	creator, ok := f.creators[models.ChannelType(config.Type)]
	if !ok {
		return nil, errors.NewNotifyError("Factory", "unknown channel type: "+string(config.Type), errors.ErrUnsupportedType)
	}

	return creator(config)
}
