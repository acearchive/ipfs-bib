package config

import (
	"github.com/nickng/bibtex"
	"net/url"
	"path"
	"strings"
)

type SourceLocator struct {
	Url url.URL
	Doi *string
}

type ResolvedSourceLocator struct {
	Url url.URL
	Doi *string
}

type NameTemplateInput struct {
	Key    string
	Type   string
	Fields map[string]string
}

func NewNameTemplateInput(entry *bibtex.BibEntry) *NameTemplateInput {
	input := NameTemplateInput{
		Key:    entry.CiteName,
		Type:   entry.Type,
		Fields: make(map[string]string),
	}

	for key, value := range entry.Fields {
		input.Fields[key] = value.String()
	}

	return &input
}

type ProxySchemeUrl struct {
	Hostname  string
	Path      string
	Directory string
	Filename  string
}

type ProxySchemeInput struct {
	Doi *string
	Url ProxySchemeUrl
}

func NewProxySchemeInput(locator *ResolvedSourceLocator, cfg *Proxy) (*ProxySchemeInput, error) {
	useProxy := len(cfg.IncludeHostnames) == 0

	for _, includeHost := range cfg.IncludeHostnames {
		if locator.Url.Hostname() == includeHost {
			useProxy = true
			break
		}
	}

	for _, excludeHost := range cfg.ExcludeHostnames {
		if locator.Url.Hostname() == excludeHost {
			useProxy = false
			break
		}
	}

	if !useProxy {
		return nil, nil
	}

	return &ProxySchemeInput{
		Doi: locator.Doi,
		Url: ProxySchemeUrl{
			Hostname:  locator.Url.Hostname(),
			Path:      strings.TrimPrefix(locator.Url.Path, "/"),
			Directory: strings.TrimPrefix(path.Dir(locator.Url.Path), "/") + "/",
			Filename:  strings.TrimPrefix(path.Base(locator.Url.Path), "/"),
		},
	}, nil
}
