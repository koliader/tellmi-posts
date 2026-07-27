package config

import (
	"os"
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
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	err := sdkconfig.LoadConfig(path, &cfg)
	return cfg, err
}

func LoadKuberConfig() (Config, error) {
	return Config{
		DBSource:            os.Getenv("DB_SOURCE"),
		ServerAddress:       os.Getenv("SERVER_ADDRESS"),
		TokenKey:            os.Getenv("TOKEN_KEY"),
		AccessTokenDuration: 720 * time.Hour,
		Environment:         os.Getenv("ENVIRONMENT"),
		RbmUrl:              os.Getenv("RBM_URL"),
	}, nil
}
