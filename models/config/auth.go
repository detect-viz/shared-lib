package config

type KeycloakConfig struct {
	URL          string `mapstructure:"url"`
	Realm        string `mapstructure:"realm"`
	ClientID     string `mapstructure:"client_id"`
	ClientUUID   string `mapstructure:"client_uuid"`
	ClientSecret string `mapstructure:"client_secret"`
	AdminRole    string `mapstructure:"admin_role"`
}
