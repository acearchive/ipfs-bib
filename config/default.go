package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/spf13/viper"
)

//go:embed config.toml
var defaultConfig []byte

var ErrInvalidConfigFile = errors.New("could not parse config file")

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
		logging.Error.Fatalf("could not parse default config: %w", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfigFile, err)
	}

	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		logging.Error.Fatal("could not unmarshal config file: %w", err)
	}

	return &config, nil
}
