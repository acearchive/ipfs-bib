package config

import (
	"errors"
)

var (
	ErrInvalidCarVersion = errors.New("CAR version must be \"1\" or \"2\"")
	ErrMfsAndCar         = errors.New("can not add sources to MFS if exporting them as a CAR")
	ErrPinAndCar         = errors.New("can not pin sources if exporting them as a CAR")
)

type Ipfs struct {
	Api        string `mapstructure:"api"`
	UseGateway bool   `mapstructure:"use-gateway"`
	Gateway    string `mapstructure:"gateway"`
	CarVersion string `mapstructure:"car-version"`
}

func (c Ipfs) MaybeGateway() *string {
	if c.UseGateway {
		return &c.Gateway
	} else {
		return nil
	}
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

type File struct {
	Ipfs      Ipfs       `mapstructure:"ipfs"`
	Archive   Archive    `mapstructure:"archive"`
	Unpaywall Unpaywall  `mapstructure:"unpaywall"`
	Snapshot  Snapshot   `mapstructure:"snapshot"`
	Pins      []Pin      `mapstructure:"pins"`
	Resolvers []Resolver `mapstructure:"resolvers"`
}

type Flags struct {
	CarPath       string `mapstructure:"car"`
	ConfigPath    string `mapstructure:"config"`
	DryRun        bool   `mapstructure:"dry-run"`
	JsonOutput    bool   `mapstructure:"json"`
	MfsPath       string `mapstructure:"mfs"`
	OutputPath    string `mapstructure:"output"`
	PinLocal      bool   `mapstructure:"pin"`
	PinRemoteName string `mapstructure:"pin-remote"`
	Verbose       bool   `mapstructure:"verbose"`
	UseZotero     bool   `mapstructure:"zotero"`
}

func (f Flags) MaybeCarPath() *string {
	if f.CarPath == "" {
		return nil
	} else {
		return &f.CarPath
	}
}

func (f Flags) MaybeConfigPath() *string {
	if f.ConfigPath == "" {
		return nil
	} else {
		return &f.ConfigPath
	}
}

func (f Flags) MaybeMfsPath() *string {
	if f.MfsPath == "" {
		return nil
	} else {
		return &f.MfsPath
	}
}

func (f Flags) MaybeOutputPath() *string {
	if f.OutputPath == "" {
		return nil
	} else {
		return &f.OutputPath
	}
}

func (f Flags) MaybePinRemoteName() *string {
	if f.PinRemoteName == "" {
		return nil
	} else {
		return &f.PinRemoteName
	}
}

func (f Flags) Validate() error {
	if f.MaybeCarPath() != nil && f.MaybeMfsPath() != nil {
		return ErrMfsAndCar
	}

	if f.MaybeCarPath() != nil && (f.PinLocal || f.MaybePinRemoteName() != nil) {
		return ErrPinAndCar
	}

	return nil
}

type Config struct {
	File  File
	Flags Flags
}
