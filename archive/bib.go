package archive

import (
	"errors"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/resolver"
	"github.com/nickng/bibtex"
	"os"
	"strings"
)

const (
	bibtexFileSeparator      = ";"
	bibtexFileFieldSeparator = ":"
	bibtexFileFields         = 3
)

const ContentOriginLocal resolver.ContentOrigin = "local"

var (
	preferredMediaTypes   = []string{"application/pdf"}
	contingencyMediaTypes = []string{"text/html"}
)

func ParseBibtex(bibPath string) (*bibtex.BibTex, error) {
	bibFile, err := os.Open(bibPath)
	if err != nil {
		return nil, err
	}

	bib, err := bibtex.Parse(bibFile)
	if err != nil {
		return nil, err
	}

	if err := bibFile.Close(); err != nil {
		return nil, err
	}

	return bib, nil
}

func ReadLocalBibSource(entry *bibtex.BibEntry, mediaTypes []string) (*DownloadedContent, error) {
	rawField := config.BibEntryField(entry, "file")
	if rawField == nil {
		return nil, nil
	}

	rawFiles := strings.Split(*rawField, bibtexFileSeparator)

	for _, rawFile := range rawFiles {
		rawFileFields := strings.Split(rawFile, bibtexFileFieldSeparator)
		if len(rawFileFields) != bibtexFileFields {
			continue
		}

		bibFilePath, bibMediaType := rawFileFields[1], rawFileFields[2]

		for _, mediaType := range mediaTypes {
			if bibMediaType == mediaType {
				fileContent, err := os.ReadFile(bibFilePath)
				if errors.Is(err, os.ErrNotExist) {
					logging.Verbose.Println(fmt.Sprintf("Local source file does not exist: %s", bibFilePath))
					continue
				} else if err != nil {
					return nil, err
				}

				return &DownloadedContent{
					Content:   fileContent,
					MediaType: bibMediaType,
					Origin:    ContentOriginLocal,
				}, nil
			}
		}
	}

	return nil, nil
}

func UpdateBib(bib *bibtex.BibTex, gateway *string, location *Location) error {
	for _, entry := range bib.Entries {
		entryLocation, ok := location.Entries[BibCiteName(entry.CiteName)]
		if !ok {
			continue
		}

		if gateway == nil {
			entry.Fields["url"] = bibtex.NewBibConst(entryLocation.IpfsUrl().String())
		} else {
			gatewayUrl, err := entryLocation.GatewayUrl(*gateway)
			if err != nil {
				return err
			}

			entry.Fields["url"] = bibtex.NewBibConst(gatewayUrl.String())
		}
	}

	return nil
}

func WriteBib(bib *bibtex.BibTex, file string) error {
	return os.WriteFile(file, []byte(bib.PrettyString()), 0644)
}
