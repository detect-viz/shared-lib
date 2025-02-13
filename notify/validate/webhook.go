package validate

// WebhookValidator Webhook配置驗證器
type WebhookValidator struct{}

func (v *WebhookValidator) Validate(config map[string]string) error {
	if err := NotEmpty("url", config["url"]); err != nil {
		return err
	}
	if err := URL(config["url"]); err != nil {
		return err
	}
	return nil
}
