package validate

import (
	"net/mail"
	"net/url"
	"strconv"

	"github.com/detect-viz/shared-lib/models/common"
	"github.com/detect-viz/shared-lib/notifier/errors"
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

func (v *Validator) Validate(config common.NotifySetting) error {
	validator, ok := v.validators[config.Type]
	if !ok {
		return errors.NewNotifyError("Validator", "no validator for type", errors.ErrUnsupportedType)
	}
	return validator.Validate(config.Config)
}

// ValidateConfig 驗證配置
func ValidateConfig(typ string, config map[string]string) error {
	// 1. 檢查必要配置
	required := getRequiredConfig(typ)
	for _, key := range required {
		if err := NotEmpty(key, config[key]); err != nil {
			return errors.NewNotifyError("Config", err.Error(), errors.ErrMissingRequiredConf)
		}
	}

	// 2. 根據類型進行特定驗證
	validator := getValidator(typ)
	if validator != nil {
		return validator(config)
	}

	return nil
}

// getRequiredConfig 獲取必要配置項
func getRequiredConfig(typ string) []string {
	switch typ {
	case "email":
		return []string{"host", "port", "username", "password", "to"}
	case "line":
		return []string{"channel_token", "to"}
	case "slack", "discord", "teams", "webex", "webhook":
		return []string{"url"}
	default:
		return nil
	}
}

// getValidator 獲取對應的驗證器
func getValidator(typ string) func(map[string]string) error {
	switch typ {
	case "email":
		return validateEmailConfig
	case "webhook":
		return validateWebhookConfig
	case "line", "slack", "discord", "teams", "webex":
		return validateWebhookURLConfig
	default:
		return nil
	}
}

// validateEmailConfig 驗證郵件配置
func validateEmailConfig(config map[string]string) error {
	// 驗證端口
	port, err := strconv.Atoi(config["port"])
	if err != nil {
		return errors.NewNotifyError("Config", "invalid port number", err)
	}
	if err := Port(port); err != nil {
		return errors.NewNotifyError("Config", err.Error(), errors.ErrInvalidConfig)
	}

	// 驗證郵件地址
	if _, err := mail.ParseAddress(config["to"]); err != nil {
		return errors.NewNotifyError("Config", "invalid email address: "+config["to"], err)
	}

	return nil
}

// validateWebhookConfig 驗證Webhook配置
func validateWebhookConfig(config map[string]string) error {
	// 驗證URL
	if _, err := url.ParseRequestURI(config["url"]); err != nil {
		return errors.NewNotifyError("Config", "invalid webhook url: "+config["url"], err)
	}

	// 驗證HTTP方法
	if method := config["method"]; method != "" {
		if err := Method(method); err != nil {
			return errors.NewNotifyError("Config", err.Error(), errors.ErrInvalidConfig)
		}
	}

	return nil
}

// validateWebhookURLConfig 驗證Webhook URL配置
func validateWebhookURLConfig(config map[string]string) error {
	if _, err := url.ParseRequestURI(config["url"]); err != nil {
		return errors.NewNotifyError("Config", "invalid webhook url: "+config["url"], err)
	}
	return nil
}
