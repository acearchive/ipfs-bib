package config

import (
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
	NameTemplate  string   `toml:"name-template" default:"{{ coalesce (get .Fields \"doi\") .Key }}"`
	ExcludedTypes []string `toml:"excluded-types"`
}

type EmbedHandler struct {
	Enabled    bool     `toml:"enabled" default:"true"`
	MediaTypes []string `toml:"media-types"`
}

type MonolithHandler struct {
	Enabled         bool   `toml:"enabled" default:"true"`
	Path            string `toml:"path" default:"monolith"`
	AllowInsecure   bool   `toml:"allow-insecure" default:"false"`
	IncludeAudio    bool   `toml:"include-audio" default:"true"`
	IncludeCss      bool   `toml:"include-css" default:"true"`
	IncludeFonts    bool   `toml:"include-fonts" default:"true"`
	IncludeFrames   bool   `toml:"include-frames" default:"true"`
	IncludeImages   bool   `toml:"include-images" default:"true"`
	IncludeJs       bool   `toml:"include-js" default:"true"`
	IncludeVideo    bool   `toml:"include-video" default:"true"`
	IncludeMetadata bool   `toml:"include-metadata" default:"true"`
}

type Handlers struct {
	Embed    EmbedHandler    `toml:"embed"`
	Monolith MonolithHandler `toml:"monolith"`
}

type Proxy struct {
	Schemes          []string `toml:"schemes"`
	IncludeHostnames []string `toml:"include-hostnames"`
	ExcludeHostnames []string `toml:"exclude-hostnames"`
}

type Config struct {
	Ipfs     Ipfs     `toml:"ipfs"`
	Archive  Archive  `toml:"bib"`
	Handlers Handlers `toml:"handlers"`
	Proxies  []Proxy  `toml:"proxies"`
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
