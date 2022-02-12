package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/network"
	"github.com/nickng/bibtex"
	"mime"
	"path"
	"strings"
	"text/template"
)

var ErrInvalidTemplate = errors.New("malformed config template")

type SourcePath struct {
	FileName      string
	DirectoryName string
}

type SourcePathTemplate struct {
	filename        template.Template
	directory       template.Template
	occurrenceCount map[string]int
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

	return SourcePathTemplate{
		filename:        *filename,
		directory:       *directory,
		occurrenceCount: make(map[string]int),
	}, nil
}

func (s SourcePathTemplate) Execute(entry bibtex.BibEntry, originalFileName string, mediaType string) SourcePath {
	var filenameBytes bytes.Buffer
	var directoryBytes bytes.Buffer

	filenameInput := newFileNameTemplateInput(entry, originalFileName, mediaType)
	directoryInput := newDirectoryNameTemplateInput(entry, 0)

	if err := s.filename.Execute(&filenameBytes, filenameInput); err != nil {
		logging.Error.Fatal(err)
	}

	if err := s.directory.Execute(&directoryBytes, directoryInput); err != nil {
		logging.Error.Fatal(err)
	}

	// We execute the directory template with an ordinal value of `0` first, so
	// we can determine whether it has been duplicated or not.
	ordinal := s.occurrenceCount[directoryBytes.String()]
	s.occurrenceCount[directoryBytes.String()] = ordinal + 1

	if ordinal > 0 {
		directoryInput.Ordinal = ordinal
		directoryBytes.Reset()

		if err := s.directory.Execute(&directoryBytes, directoryInput); err != nil {
			logging.Error.Fatal(err)
		}
	}

	sanitizedFileName := strings.ReplaceAll(filenameBytes.String(), "/", "-")
	sanitizedDirectoryName := strings.ReplaceAll(directoryBytes.String(), "/", "-")

	return SourcePath{
		FileName:      sanitizedFileName,
		DirectoryName: sanitizedDirectoryName,
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
		mediaType = network.DefaultMediaType
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
	Ordinal  int
}

func newDirectoryNameTemplateInput(entry bibtex.BibEntry, ordinal int) directoryNameTemplateInput {
	input := directoryNameTemplateInput{
		CiteName: entry.CiteName,
		Type:     entry.Type,
		Fields:   make(map[string]interface{}),
		Ordinal:  ordinal,
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
