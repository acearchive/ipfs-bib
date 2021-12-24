package config

import (
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/nickng/bibtex"
	"mime"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const canonicalDoiUrlPrefix = "https://doi.org/"

var doiPrefixes = []string{
	"doi:",
	"http://doi.org/",
	"https://doi.org/",
	"doi.org/",
	"http://dx.doi.org/",
	"https://dx.doi.org/",
	"dx.doi.org/",
}

const doiUrlRegexMatchGroup = 3

var doiUrlRegex = regexp.MustCompile(`^(https?://)?(dx\.)?doi\.org/(10\.[0-9]{4,}(\.[0-9]+)*/\S+)$`)

func FileNameFromUrl(sourceUrl *url.URL, mediaType string) string {
	if sourceUrl == nil || mediaType == "" {
		return ""
	}

	fileName := path.Base(sourceUrl.Path)
	if fileName == "." {
		return ""
	}

	fileExtension := path.Ext(fileName)
	if mime.TypeByExtension(fileExtension) == mediaType {
		return fileName
	} else {
		return ""
	}
}

type SourceLocator struct {
	Url url.URL
	Doi *string
}

func BibEntryField(entry *bibtex.BibEntry, field string) *string {
	if value, ok := entry.Fields[field]; ok {
		stringValue := value.String()
		return &stringValue
	}

	return nil
}

func LocateEntry(entry *bibtex.BibEntry) (*SourceLocator, error) {
	var (
		sourceUrl *url.URL
		sourceDoi *string
		err       error
	)

	if rawDoi := BibEntryField(entry, "doi"); rawDoi != nil {
		doi := *rawDoi

		for _, doiPrefix := range doiPrefixes {
			if strings.HasPrefix(*rawDoi, doiPrefix) {
				doi = strings.TrimPrefix(*rawDoi, doiPrefix)
				break
			}
		}

		sourceUrl, err = url.Parse(canonicalDoiUrlPrefix + doi)
		if err != nil {
			return nil, err
		}

		sourceDoi = &doi
	}

	if rawUrl := BibEntryField(entry, "url"); rawUrl != nil {
		sourceUrl, err = url.Parse(*rawUrl)
		if err != nil {
			return nil, err
		}

		if sourceDoi == nil {
			// Attempt to extract the DOI from the URL.
			if matches := doiUrlRegex.FindStringSubmatch(*rawUrl); matches != nil {
				sourceDoi = &matches[doiUrlRegexMatchGroup]
			}
		}
	}

	if sourceUrl == nil {
		return nil, nil
	} else {
		return &SourceLocator{Url: *sourceUrl, Doi: sourceDoi}, nil
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

func (l *BibEntryLocation) IpfsUrl() *url.URL {
	ipfsUrl, err := url.Parse(fmt.Sprintf("ipfs://%s/?filename=%s", l.FileCid.String(), url.QueryEscape(l.FileName)))
	if err != nil {
		panic("failed to parse ipfs:// URL")
	}

	return ipfsUrl
}

func (l *BibEntryLocation) GatewayUrl(gateway string) (*url.URL, error) {
	switch l.FileCid.Version() {
	case 0:
		return url.Parse(fmt.Sprintf("https://%s/ipfs/%s/?filename=%s", gateway, l.FileCid.String(), url.QueryEscape(l.FileName)))
	default:
		return url.Parse(fmt.Sprintf("https://%s.ipfs.%s/?filename=%s", l.FileCid.String(), gateway, url.QueryEscape(l.FileName)))
	}
}
