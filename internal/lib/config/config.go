package config

import (
	"time"

	sdkconfig "github.com/koliader/tellmi-sdk/config"
)

type Config struct {
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenKey            string        `mapstructure:"TOKEN_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	Environment         string        `mapstructure:"ENVIRONMENT"`
	RbmUrl              string        `mapstructure:"RBM_URL"`
	HealthAddress       string        `mapstructure:"HEALTH_ADDRESS"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	err := sdkconfig.Load(path, &cfg)
	return cfg, err
}
