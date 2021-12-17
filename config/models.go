package config

import "github.com/frawleyskid/ipfs-bib/config/pattern"

type Ipfs struct {
	Api        string `json:"api"`
	UseGateway bool   `json:"use-gateway"`
	Gateway    string `json:"gateway"`
}

type Bib struct {
	NamePatterns []pattern.Pattern `json:"name-patterns"`
	NameCommand  *string           `json:"name-command"`
}

type Handler struct {
	MediaTypes []string `json:"media-types"`
	Command    string   `json:"command"`
}

type Proxy struct {
	Schemes   []pattern.Pattern `json:"schemes"`
	Doi       bool              `json:"doi"`
	Hostnames []string          `json:"hostnames"`
}

type Config struct {
	Ipfs     Ipfs      `json:"ipfs"`
	Bib      Bib       `json:"bib"`
	Handlers []Handler `json:"handlers"`
	Proxies  []Proxy   `json:"proxies"`
}
