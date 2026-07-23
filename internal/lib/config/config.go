package config

import (
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenKey            string        `mapstructure:"TOKEN_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	Environment         string        `mapstructure:"ENVIRONMENT"`
	RbmUrl              string        `mapstructure:"RBM_URL"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}

func LoadKuberConfig() (config Config, err error) {

	return Config{
		DBSource:            os.Getenv("DB_SOURCE"),
		ServerAddress:       os.Getenv("SERVER_ADDRESS"),
		TokenKey:            os.Getenv("TOKEN_KEY"),
		AccessTokenDuration: 720 * time.Hour,
		Environment:         os.Getenv("ENVIRONMENT"),
		RbmUrl:              os.Getenv("RBM_URL"),
	}, nil
}
