package validate

// LineValidator Line配置驗證器
type LineValidator struct{}

func (v *LineValidator) Validate(config map[string]string) error {
	if err := NotEmpty("channel_token", config["channel_token"]); err != nil {
		return err
	}
	if err := NotEmpty("to", config["to"]); err != nil {
		return err
	}
	return nil
}
