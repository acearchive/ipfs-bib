package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//go:embed config.toml
var defaultConfig []byte

var ErrInvalidConfigFile = errors.New("could not parse config file")

func loadDefault() (File, error) {
	cfg := viper.New()

	cfg.SetConfigName("default")
	cfg.SetConfigType("toml")
	if err := cfg.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		return File{}, err
	}

	config := File{}
	if err := cfg.Unmarshal(&config); err != nil {
		return File{}, err
	}

	return config, nil
}

func fromToml(file string) (File, error) {
	cfg := viper.New()

	cfg.SetConfigName("default")
	cfg.SetConfigType("toml")
	if err := cfg.ReadConfig(bytes.NewReader(defaultConfig)); err != nil {
		logging.Error.Fatalf("could not parse default config: %v", err)
	}

	cfg.SetConfigName("config")
	cfg.SetConfigType("toml")
	cfg.SetConfigFile(file)
	if err := cfg.ReadInConfig(); err != nil {
		return File{}, fmt.Errorf("%w: %v", ErrInvalidConfigFile, err)
	}

	config := File{}
	if err := cfg.Unmarshal(&config); err != nil {
		logging.Error.Fatalf("could not unmarshal config file: %v", err)
	}

	return config, nil
}

func Load(flagSet *pflag.FlagSet) (Config, error) {
	flagsCfg := viper.New()

	if err := flagsCfg.BindPFlags(flagSet); err != nil {
		return Config{}, fmt.Errorf("could not parse CLI flags: %v", err)
	}

	flags := Flags{}
	if err := flagsCfg.Unmarshal(&flags); err != nil {
		logging.Error.Fatalf("could not unmarshal CLI flags: %v", err)
	}

	var (
		file File
		err  error
	)

	if flags.MaybeConfigPath() == nil {
		file, err = loadDefault()
		if err != nil {
			return Config{}, err
		}
	} else {
		file, err = fromToml(flags.ConfigPath)
		if err != nil {
			return Config{}, err
		}
	}

	return Config{
		File:  file,
		Flags: flags,
	}, nil
}
