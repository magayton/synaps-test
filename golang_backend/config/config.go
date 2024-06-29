package config

import (
	"github.com/spf13/viper"
)

func LoadConfig() error {
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
