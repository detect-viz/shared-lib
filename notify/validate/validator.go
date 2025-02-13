package validate

import (
	"shared-lib/models"
	"shared-lib/notify/errors"
)

// ConfigValidator 配置驗證器介面
type ConfigValidator interface {
	Validate(config map[string]string) error
}

type Validator struct {
	validators map[string]ConfigValidator
}

func New() *Validator {
	v := &Validator{
		validators: make(map[string]ConfigValidator),
	}

	// 註冊內建驗證器
	v.Register("email", &EmailValidator{})
	v.Register("teams", &WebhookValidator{})
	v.Register("line", &LineValidator{})
	v.Register("slack", &WebhookValidator{})
	v.Register("discord", &WebhookValidator{})
	v.Register("webhook", &WebhookValidator{})
	v.Register("webex", &WebhookValidator{})

	return v
}

func (v *Validator) Register(typ string, validator ConfigValidator) {
	v.validators[typ] = validator
}

func (v *Validator) Validate(config models.ChannelConfig) error {
	validator, ok := v.validators[config.Type]
	if !ok {
		return errors.NewNotifyError("Validator", "no validator for type", errors.ErrUnsupportedType)
	}
	return validator.Validate(config.Config)
}
