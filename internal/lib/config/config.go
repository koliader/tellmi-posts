package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
	Environment   string `mapstructure:"ENVIRONMENT"`
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
		DBSource:      os.Getenv("DB_SOURCE"),
		ServerAddress: os.Getenv("SERVER_ADDRESS"),
		Environment:   os.Getenv("ENVIRONMENT"),
	}, nil
}
