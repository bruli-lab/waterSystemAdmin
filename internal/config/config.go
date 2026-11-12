package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	ServerURL    string `envconfig:"SERVER_URL" required:"true"`
	ApiUrl       string `envconfig:"API_URL" required:"true"`
	ApiAuthKey   string `envconfig:"WS_AUTH_TOKEN" required:"true"`
	PasswordFile string `envconfig:"PASSWORD_FILE" required:"true"`
	LogLevel     string `envconfig:"LOG_LEVEL" default:"info"`
	SessionKey   string `envconfig:"SESSION_KEY"`
}

func New() (*Config, error) {
	var conf Config
	if err := envconfig.Process("", &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
