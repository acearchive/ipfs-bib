package archive

import (
	"errors"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/network"
	"github.com/frawleyskid/ipfs-bib/resolver"
	"github.com/nickng/bibtex"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	bibtexFileSeparator      = ";"
	bibtexFileFieldSeparator = ":"
	bibtexFileFields         = 3
	defaultBibtexPermissions = 0644
)

const ContentOriginLocal resolver.ContentOrigin = "local"

const stdinFileName = "-"

var ErrParseBibtex = errors.New("error parsing bibtex")

func ParseBibtex(bibPath string) (bibtex.BibTex, error) {
	var (
		bibFile io.ReadCloser
		err     error
	)

	if bibPath == stdinFileName {
		bibFile = os.Stdin
	} else {
		bibFile, err = os.Open(bibPath)
		if err != nil {
			return bibtex.BibTex{}, err
		}
	}

	bib, err := bibtex.Parse(bibFile)
	if err != nil {
		return bibtex.BibTex{}, fmt.Errorf("%w: %v", ErrParseBibtex, err)
	}

	if err := bibFile.Close(); err != nil {
		return bibtex.BibTex{}, err
	}

	return *bib, nil
}

func ReadLocalBibSource(entry bibtex.BibEntry, includeSnapshots bool) (DownloadedContent, error) {
	rawField := config.BibEntryField(entry, "file")
	if rawField == nil {
		return DownloadedContent{}, ErrNoSource
	}

	rawFiles := strings.Split(*rawField, bibtexFileSeparator)

	for _, rawFile := range rawFiles {
		rawFileFields := strings.Split(rawFile, bibtexFileFieldSeparator)
		if len(rawFileFields) != bibtexFileFields {
			continue
		}

		bibFilePath, bibMediaType := rawFileFields[1], rawFileFields[2]

		if bibMediaType == network.HtmlMediaType && !includeSnapshots {
			continue
		}

		fileContent, err := os.ReadFile(bibFilePath)
		if errors.Is(err, os.ErrNotExist) {
			logging.Verbose.Println(fmt.Sprintf("Local source file does not exist: %s", bibFilePath))
			continue
		} else if err != nil {
			return DownloadedContent{}, err
		}

		bibFileName := filepath.Base(bibFilePath)
		if bibFileName == "." {
			bibFileName = ""
		}

		return DownloadedContent{
			Content: fileContent,
			ContentMetadata: ContentMetadata{
				MediaType: bibMediaType,
				FileName:  bibFileName,
				Origin:    ContentOriginLocal,
			},
		}, nil
	}

	return DownloadedContent{}, ErrNoSource
}

func UpdateBib(bib bibtex.BibTex, gateway *string, location Location) error {
	for _, entry := range bib.Entries {
		entryLocation, ok := location.Entries[entry.CiteName]
		if !ok {
			continue
		}

		var (
			updatedUrl url.URL
			err        error
		)

		if gateway == nil {
			updatedUrl = entryLocation.IpfsUrl()
		} else {
			updatedUrl, err = entryLocation.GatewayUrl(*gateway)
			if err != nil {
				return err
			}
		}

		entry.Fields["url"] = bibtex.NewBibConst(updatedUrl.String())
	}

	return nil
}

func WriteBib(bib bibtex.BibTex, file string) error {
	return os.WriteFile(file, []byte(bib.PrettyString()), defaultBibtexPermissions)
}
