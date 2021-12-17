package archive

import (
	"errors"
	"fmt"
	"github.com/nickng/bibtex"
	"net/url"
	"strings"
)

const (
	DOI_PREFIX = "doi:"
	DOI_URL    = "https://doi.org/"
)

var (
	ErrInvalidReferenceUrl = errors.New("invalid URL from bibtex file")
	ErrMissingReferenceUrl = errors.New("this reference has no DOI or URL")
)

func doiToRawUrl(doi string) (rawUrl string) {
	switch {
	case strings.HasPrefix(doi, DOI_PREFIX):
		return DOI_URL + strings.TrimPrefix(doi, DOI_PREFIX)
	case strings.HasPrefix(doi, DOI_URL):
		return doi
	default:
		return DOI_URL + doi
	}
}

func UrlFromBibEntry(entry bibtex.BibEntry) (*url.URL, error) {
	var rawUrl string

	if doi, ok := entry.Fields["doi"]; ok {
		rawUrl = doiToRawUrl(doi.String())
	} else if urlField, ok := entry.Fields["url"]; ok {
		rawUrl = urlField.String()
	} else {
		return nil, ErrMissingReferenceUrl
	}

	entryUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidReferenceUrl, err)
	}

	return entryUrl, nil
}
