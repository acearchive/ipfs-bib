package config

import (
	"errors"
)

var ErrInvalidCarVersion = errors.New("CAR version must be \"1\" or \"2\"")

type Ipfs struct {
	Api        string `mapstructure:"api"`
	UseGateway bool   `mapstructure:"use-gateway"`
	Gateway    string `mapstructure:"gateway"`
	CarVersion string `mapstructure:"car-version"`
}

func (c Ipfs) IsCarV2() (bool, error) {
	switch c.CarVersion {
	case "1":
		return false, nil
	case "2":
		return true, nil
	default:
		return false, ErrInvalidCarVersion
	}
}

type Archive struct {
	FileName      string   `mapstructure:"file-name"`
	DirectoryName string   `mapstructure:"directory-name"`
	EmbeddedTypes []string `mapstructure:"embedded-types"`
	ExcludedTypes []string `mapstructure:"excluded-types"`
	UserAgent     string   `mapstructure:"user-agent"`
}

type Unpaywall struct {
	Enabled bool   `mapstructure:"enabled"`
	Email   string `mapstructure:"email"`
}

type Snapshot struct {
	Enabled         bool   `mapstructure:"enabled"`
	Path            string `mapstructure:"path"`
	AllowInsecure   bool   `mapstructure:"allow-insecure"`
	IncludeAudio    bool   `mapstructure:"include-audio"`
	IncludeCss      bool   `mapstructure:"include-css"`
	IncludeFonts    bool   `mapstructure:"include-fonts"`
	IncludeFrames   bool   `mapstructure:"include-frames"`
	IncludeImages   bool   `mapstructure:"include-images"`
	IncludeJs       bool   `mapstructure:"include-js"`
	IncludeVideo    bool   `mapstructure:"include-video"`
	IncludeMetadata bool   `mapstructure:"include-metadata"`
}

type Pin struct {
	Endpoint string `mapstructure:"endpoint"`
	Token    string `mapstructure:"token"`
}

type Resolver struct {
	Schemes          []string `mapstructure:"schemes"`
	IncludeHostnames []string `mapstructure:"include-hostnames"`
	ExcludeHostnames []string `mapstructure:"exclude-hostnames"`
}

type Config struct {
	Ipfs      Ipfs       `mapstructure:"ipfs"`
	Archive   Archive    `mapstructure:"archive"`
	Unpaywall Unpaywall  `mapstructure:"unpaywall"`
	Snapshot  Snapshot   `mapstructure:"snapshot"`
	Pins      []Pin      `mapstructure:"pins"`
	Resolvers []Resolver `mapstructure:"resolvers"`
}
