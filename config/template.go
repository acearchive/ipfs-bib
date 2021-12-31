package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/nickng/bibtex"
	"mime"
	"path"
	"strings"
	"text/template"
)

const DefaultMediaType = "application/octet-stream"

var ErrInvalidTemplate = errors.New("malformed config template")

type SourcePath struct {
	FileName      string
	DirectoryName string
}

type SourcePathTemplate struct {
	filename  template.Template
	directory template.Template
}

func NewSourcePathTemplate(cfg Config) (SourcePathTemplate, error) {
	filename, err := template.New("archive.file-name").Funcs(sprig.TxtFuncMap()).Parse(cfg.File.Archive.FileName)
	if err != nil {
		return SourcePathTemplate{}, fmt.Errorf("%w: %v", ErrInvalidTemplate, err)
	}

	directory, err := template.New("archive.directory-name").Funcs(sprig.TxtFuncMap()).Parse(cfg.File.Archive.DirectoryName)
	if err != nil {
		return SourcePathTemplate{}, fmt.Errorf("%w: %v", ErrInvalidTemplate, err)
	}

	return SourcePathTemplate{filename: *filename, directory: *directory}, nil
}

func (s SourcePathTemplate) Execute(entry bibtex.BibEntry, originalFileName string, mediaType string) SourcePath {
	var filenameBytes bytes.Buffer
	var directoryBytes bytes.Buffer

	filenameInput := newFileNameTemplateInput(entry, originalFileName, mediaType)
	directoryInput := newDirectoryNameTemplateInput(entry)

	if err := s.filename.Execute(&filenameBytes, filenameInput); err != nil {
		logging.Error.Fatal(err)
	}

	if err := s.directory.Execute(&directoryBytes, directoryInput); err != nil {
		logging.Error.Fatal(err)
	}

	return SourcePath{
		FileName:      filenameBytes.String(),
		DirectoryName: directoryBytes.String(),
	}
}

type fileNameTemplateInput struct {
	Original  string
	CiteName  string
	Type      string
	Fields    map[string]interface{}
	Extension string
}

func newFileNameTemplateInput(entry bibtex.BibEntry, originalFileName string, mediaType string) fileNameTemplateInput {
	mediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		mediaType = DefaultMediaType
	}

	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil {
		extensions = nil
	}

	var extension string

	if extensions == nil {
		extension = ""
	} else {
		extension = extensions[0]
	}

	input := fileNameTemplateInput{
		Original:  originalFileName,
		CiteName:  entry.CiteName,
		Type:      entry.Type,
		Fields:    make(map[string]interface{}),
		Extension: extension,
	}

	for key, value := range entry.Fields {
		input.Fields[key] = value.String()
	}

	return input
}

type directoryNameTemplateInput struct {
	CiteName string
	Type     string
	Fields   map[string]interface{}
}

func newDirectoryNameTemplateInput(entry bibtex.BibEntry) directoryNameTemplateInput {
	input := directoryNameTemplateInput{
		CiteName: entry.CiteName,
		Type:     entry.Type,
		Fields:   make(map[string]interface{}),
	}

	for key, value := range entry.Fields {
		input.Fields[key] = value.String()
	}

	return input
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

func NewProxySchemeInput(locator SourceLocator, filter HostnameFilter) *ProxySchemeInput {
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
		return nil
	}

	return &ProxySchemeInput{
		Doi: locator.Doi,
		Url: ProxySchemeUrl{
			Hostname:  locator.Url.Hostname(),
			Path:      strings.TrimPrefix(locator.Url.Path, "/"),
			Directory: strings.TrimPrefix(path.Dir(locator.Url.Path), "/") + "/",
			Filename:  strings.TrimPrefix(path.Base(locator.Url.Path), "/"),
		},
	}
}
