package config

import (
	"errors"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/network"
	"github.com/ipfs/go-cid"
	"github.com/nickng/bibtex"
	"mime"
	"net/http"
	"net/url"
	"path"
	"regexp"
)

const (
	canonicalDoiUrlPrefix = "https://doi.org/"
)

const doiRegexMatchGroup = 4

var doiRegex = regexp.MustCompile(`^(doi:|(https?://)?(dx\.)?doi\.org/)?(10\.[0-9]{4,}(\.[0-9]+)*/\S+)$`)

var (
	ErrCouldNotLocateEntry = errors.New("bibtex entry has no URL or DOI")
	ErrMalformedGateway    = errors.New("IPFS gateway is malformed")
)

func fileNameFromUrl(sourceUrl *url.URL, mediaType string) *string {
	if sourceUrl == nil || mediaType == "" {
		return nil
	}

	fileName := path.Base(sourceUrl.Path)
	if fileName == "." {
		return nil
	}

	fileExtension := path.Ext(fileName)
	if mime.TypeByExtension(fileExtension) == mediaType {
		return &fileName
	} else {
		return nil
	}
}

func fileNameFromContentDisposition(header http.Header) *string {
	_, params, err := mime.ParseMediaType(header.Get(network.ContentDispositionHeader))
	if err != nil {
		return nil
	}

	if filename, ok := params["filename"]; ok {
		return &filename
	} else {
		return nil
	}
}

func InferFileName(sourceUrl *url.URL, mediaType string, header http.Header) string {
	if filename := fileNameFromContentDisposition(header); filename != nil {
		return *filename
	}

	if filename := fileNameFromUrl(sourceUrl, mediaType); filename != nil {
		return *filename
	}

	return ""
}

type SourceLocator struct {
	Url url.URL
	Doi *string
}

func BibEntryField(entry bibtex.BibEntry, field string) *string {
	if value, ok := entry.Fields[field]; ok {
		stringValue := value.String()
		return &stringValue
	}

	return nil
}

func LocateEntry(entry bibtex.BibEntry) (SourceLocator, error) {
	var (
		sourceUrl *url.URL
		sourceDoi *string
		err       error
	)

	if rawDoi := BibEntryField(entry, "doi"); rawDoi != nil {
		if matches := doiRegex.FindStringSubmatch(*rawDoi); matches != nil {
			sourceDoi = &matches[doiRegexMatchGroup]
		}
	}

	if rawUrl := BibEntryField(entry, "url"); rawUrl != nil {
		sourceUrl, err = url.Parse(*rawUrl)
		if err != nil {
			logging.Verbose.Printf("malformed bibtex URL: %s", *rawUrl)
			sourceUrl = nil
		}

		if sourceUrl != nil && sourceDoi == nil {
			// Attempt to extract the DOI from the URL.
			if matches := doiRegex.FindStringSubmatch(*rawUrl); matches != nil {
				sourceDoi = &matches[doiRegexMatchGroup]
			}
		}
	}

	if sourceUrl == nil && sourceDoi != nil {
		sourceUrl, err = url.Parse(canonicalDoiUrlPrefix + url.PathEscape(*sourceDoi))
		if err != nil {
			logging.Error.Fatal(err)
		}
	}

	if sourceUrl == nil {
		return SourceLocator{}, fmt.Errorf("%w: %s", ErrCouldNotLocateEntry, entry.CiteName)
	} else {
		return SourceLocator{Url: *sourceUrl, Doi: sourceDoi}, nil
	}
}

type BibSource struct {
	Content       []byte
	DirectoryName string
	FileName      string
}

type BibEntryLocation struct {
	FileCid       cid.Cid
	FileName      string
	DirectoryCid  cid.Cid
	DirectoryName string
}

func (l *BibEntryLocation) IpfsUrl() url.URL {
	ipfsUrl, err := url.Parse(fmt.Sprintf("ipfs://%s/?filename=%s", l.FileCid.String(), url.QueryEscape(l.FileName)))
	if err != nil {
		logging.Error.Fatal(err)
	}

	return *ipfsUrl
}

func (l *BibEntryLocation) GatewayUrl(gateway string) (url.URL, error) {
	gatewayUrl, err := url.Parse(fmt.Sprintf("https://%s/ipfs/%s/?filename=%s", gateway, l.FileCid.String(), url.QueryEscape(l.FileName)))
	if err != nil {
		return url.URL{}, fmt.Errorf("%w: %s", ErrMalformedGateway, gateway)
	}

	return *gatewayUrl, nil
}
