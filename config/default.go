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

func LoadDefault() (Config, error) {
	viper.SetConfigName("default")
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		return Config{}, err
	}

	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func FromToml(file string) (Config, error) {
	viper.SetConfigName("default")
	viper.SetConfigType("toml")
	if err := viper.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		logging.Error.Fatalf("could not parse default config: %v", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrInvalidConfigFile, err)
	}

	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		logging.Error.Fatalf("could not unmarshal config file: %v", err)
	}

	return config, nil
}
