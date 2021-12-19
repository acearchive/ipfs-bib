package config

import (
	"github.com/nickng/bibtex"
	"net/url"
	"path"
	"strings"
)

var DoiPrefixes = []string{
	"doi:",
	"https://doi.org/",
	"http://doi.org/",
	"doi.org/",
}

func entryField(entry *bibtex.BibEntry, field string) *string {
	if value, ok := entry.Fields[field]; ok {
		stringValue := value.String()
		return &stringValue
	}

	return nil
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
	Url *ProxySchemeUrl
}

func NewProxySchemeInput(entry *bibtex.BibEntry, config *Proxy) (*ProxySchemeInput, error) {
	input := ProxySchemeInput{}

	if rawDoi := entryField(entry, "doi"); rawDoi != nil {
		for _, doiPrefix := range DoiPrefixes {
			if strings.HasPrefix(*rawDoi, doiPrefix) {
				doi := strings.TrimPrefix(*rawDoi, doiPrefix)
				input.Doi = &doi
				break
			}
		}
	}

	if rawUrl := entryField(entry, "url"); rawUrl != nil {
		entryUrl, err := url.Parse(*rawUrl)
		if err != nil {
			return nil, err
		}

		includeUrl := len(config.IncludeHostnames) == 0

		for _, includeHost := range config.IncludeHostnames {
			if entryUrl.Hostname() == includeHost {
				includeUrl = true
				break
			}
		}

		for _, excludeHost := range config.ExcludeHostnames {
			if entryUrl.Hostname() == excludeHost {
				includeUrl = false
				break
			}
		}

		if includeUrl {
			input.Url = &ProxySchemeUrl{
				Hostname:  entryUrl.Hostname(),
				Path:      strings.TrimPrefix(entryUrl.Path, "/"),
				Directory: strings.TrimPrefix(path.Dir(entryUrl.Path), "/") + "/",
				Filename:  strings.TrimPrefix(path.Base(entryUrl.Path), "/"),
			}
		}
	}

	return &input, nil
}
