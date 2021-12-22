package config

import (
	"bytes"
	_ "embed"
	"github.com/spf13/viper"
)

//go:embed config.toml
var defaultConfig []byte

func LoadDefault() (*Config, error) {
	viper.SetConfigName("default")
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		return nil, err
	}

	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func FromToml(file string) (*Config, error) {
	viper.SetConfigName("default")
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		return nil, err
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
