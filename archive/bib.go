package archive

import (
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/nickng/bibtex"
	"net/url"
	"strings"
)

const (
	DoiPrefix = "doi:"
	DoiUrl    = "https://doi.org/"
)

var (
	ErrInvalidReferenceUrl = errors.New("invalid URL from bibtex file")
	ErrMissingReferenceUrl = errors.New("this reference has no DOI or URL")
)

func doiToRawUrl(doi string) (rawUrl string) {
	switch {
	case strings.HasPrefix(doi, DoiPrefix):
		return DoiUrl + strings.TrimPrefix(doi, DoiPrefix)
	case strings.HasPrefix(doi, DoiUrl):
		return doi
	default:
		return DoiUrl + doi
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

type BibSource struct {
	Content       []byte
	DirectoryName string
	FileName      string
}

type BibSourceId struct {
	ContentCid   cid.Cid
	DirectoryCid cid.Cid
}
