package config

type KeycloakConfig struct {
	URL          string `mapstructure:"url"`
	Realm        string `mapstructure:"realm"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	ClientID     string `mapstructure:"client_id"`
	ClientUUID   string `mapstructure:"client_uuid"`
	ClientSecret string `mapstructure:"client_secret"`
	AdminRole    string `mapstructure:"admin_role"`
}
