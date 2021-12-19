package config

import (
	"github.com/frawleyskid/ipfs-bib/config/pattern"
	"github.com/pelletier/go-toml"
	"os"
)

type Ipfs struct {
	Api        string `toml:"api" default:"http://127.0.0.1:5001"`
	UseGateway bool   `toml:"use-gateway" default:"true"`
	Gateway    string `toml:"gateway" default:"dweb.link"`
	CarVersion string `toml:"car-version" default:"1"`
}

type Archive struct {
	NamePatterns  []pattern.Pattern `toml:"name-patterns"`
	NameCommand   *string           `toml:"name-command"`
	ExcludedTypes []string          `toml:"excluded-types"`
}

type Handler struct {
	MediaTypes []string `toml:"media-types"`
	Command    string   `toml:"command"`
	UseStdin   bool     `toml:"use-stdin"`
	OutputType string   `toml:"output-type"`
}

type Proxy struct {
	Schemes   []pattern.Pattern `toml:"schemes"`
	Doi       bool              `toml:"doi"`
	Hostnames []string          `toml:"hostnames"`
}

type Config struct {
	Ipfs     Ipfs      `toml:"ipfs"`
	Archive  Archive   `toml:"bib"`
	Handlers []Handler `toml:"handlers"`
	Proxies  []Proxy   `toml:"proxies"`
}

func FromToml(file string) (*Config, error) {
	configBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	config := Config{}
	if err := toml.Unmarshal(configBytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
