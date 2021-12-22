package config

import (
	"bytes"
	"github.com/Masterminds/sprig/v3"
	"github.com/nickng/bibtex"
	"mime"
	"path"
	"strings"
	"text/template"
)

type SourcePath struct {
	FileName      string
	DirectoryName string
}

type SourcePathTemplate struct {
	filename  template.Template
	directory template.Template
}

func NewSourcePathTemplate(cfg *Config) (*SourcePathTemplate, error) {
	filename, err := template.New("archive.file-name").Funcs(sprig.TxtFuncMap()).Parse(cfg.Archive.FileName)
	if err != nil {
		return nil, err
	}

	directory, err := template.New("archive.directory-name").Funcs(sprig.TxtFuncMap()).Parse(cfg.Archive.DirectoryName)
	if err != nil {
		return nil, err
	}

	return &SourcePathTemplate{filename: *filename, directory: *directory}, nil
}

func (s *SourcePathTemplate) Execute(entry *bibtex.BibEntry, mediaType string) (*SourcePath, error) {
	var filenameBytes bytes.Buffer
	var directoryBytes bytes.Buffer

	filenameInput, err := newFileNameTemplateInput(entry, mediaType)
	if err != nil {
		return nil, err
	}

	directoryInput := newDirectoryNameTemplateInput(entry)

	if err := s.filename.Execute(&filenameBytes, filenameInput); err != nil {
		return nil, err
	}

	if err := s.directory.Execute(&directoryBytes, directoryInput); err != nil {
		return nil, err
	}

	return &SourcePath{
		FileName:      string(filenameBytes.Bytes()),
		DirectoryName: string(directoryBytes.Bytes()),
	}, nil
}

type fileNameTemplateInput struct {
	CiteName  string
	Type      string
	Fields    map[string]interface{}
	Extension string
}

func newFileNameTemplateInput(entry *bibtex.BibEntry, mediaType string) (*fileNameTemplateInput, error) {
	mediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return nil, err
	}

	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil {
		return nil, err
	}

	var extension string

	if extensions == nil {
		extension = ""
	} else {
		extension = extensions[0]
	}

	input := fileNameTemplateInput{
		CiteName:  entry.CiteName,
		Type:      entry.Type,
		Fields:    make(map[string]interface{}),
		Extension: extension,
	}

	for key, value := range entry.Fields {
		input.Fields[key] = value.String()
	}

	return &input, nil
}

type directoryNameTemplateInput struct {
	CiteName string
	Type     string
	Fields   map[string]interface{}
}

func newDirectoryNameTemplateInput(entry *bibtex.BibEntry) *directoryNameTemplateInput {
	input := directoryNameTemplateInput{
		CiteName: entry.CiteName,
		Type:     entry.Type,
		Fields:   make(map[string]interface{}),
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

type HostnameFilter struct {
	Include []string
	Exclude []string
}

func NewProxySchemeInput(locator *SourceLocator, filter *HostnameFilter) (*ProxySchemeInput, error) {
	useProxy := len(filter.Include) == 0

	for _, includeHost := range filter.Include {
		if locator.Url.Hostname() == includeHost {
			useProxy = true
			break
		}
	}

	for _, excludeHost := range filter.Exclude {
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
