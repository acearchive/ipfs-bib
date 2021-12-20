package archive

import (
	"github.com/ipfs/go-cid"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/nickng/bibtex"
	"net/url"
	"strings"
)

const CanonicalDoiPrefix = "https://doi.org/"

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

func LocateEntry(entry *bibtex.BibEntry) (*config.SourceLocator, error) {
	var (
		sourceUrl *url.URL
		sourceDoi *string
		err       error
	)

	if rawDoi := entryField(entry, "doi"); rawDoi != nil {
		for _, doiPrefix := range DoiPrefixes {
			if strings.HasPrefix(*rawDoi, doiPrefix) {
				doi := strings.TrimPrefix(*rawDoi, doiPrefix)
				sourceDoi = &doi

				sourceUrl, err = url.Parse(CanonicalDoiPrefix + doi)
				if err != nil {
					return nil, err
				}

				break
			}
		}
	}

	if rawUrl := entryField(entry, "url"); rawUrl != nil {
		sourceUrl, err = url.Parse(*rawUrl)
		if err != nil {
			return nil, err
		}
	}

	if sourceUrl == nil {
		return nil, nil
	} else {
		return &config.SourceLocator{Url: *sourceUrl, Doi: sourceDoi}, nil
	}
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
