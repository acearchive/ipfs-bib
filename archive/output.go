package archive

import (
	"encoding/json"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/nickng/bibtex"
)

const outputIndent = "  "

type ArchivedOutput struct {
	CiteName      string `json:"citeName"`
	Doi           string `json:"doi"`
	FileCid       string `json:"fileCid"`
	FileName      string `json:"fileName"`
	DirectoryCid  string `json:"directoryCid"`
	DirectoryName string `json:"directoryName"`
	IpfsUrl       string `json:"ipfsUrl"`
	GatewayUrl    string `json:"gatewayUrl"`
	ContentOrigin string `json:"contentOrigin"`
}

type NotArchivedOutput struct {
	CiteName string `json:"citeName"`
	Doi      string `json:"doi"`
}

type Output struct {
	Cid           string              `json:"cid"`
	TotalEntries  int                 `json:"totalEntries"`
	TotalArchived int                 `json:"totalArchived"`
	Archived      []ArchivedOutput    `json:"archived"`
	NotArchived   []NotArchivedOutput `json:"notArchived"`
}

func doiFromEntry(entry *bibtex.BibEntry) string {
	if doi := config.BibEntryField(entry, "doi"); doi != nil {
		return *doi
	} else {
		return ""
	}
}

func NewOutput(cfg *config.Config, contents []BibContents, location *Location) (*Output, error) {
	var (
		archivedEntries    []ArchivedOutput
		notArchivedEntries []NotArchivedOutput
	)

	for _, bibContent := range contents {
		bibLocation, hasLocation := location.Entries[bibContent.Entry.CiteName]

		if hasLocation && bibContent.Contents != nil {
			gatewayUrl, err := bibLocation.GatewayUrl(cfg.Ipfs.Gateway)
			if err != nil {
				return nil, err
			}

			archivedEntries = append(archivedEntries, ArchivedOutput{
				CiteName:      bibContent.Entry.CiteName,
				Doi:           doiFromEntry(&bibContent.Entry),
				FileCid:       bibLocation.FileCid.String(),
				FileName:      bibLocation.FileName,
				DirectoryCid:  bibLocation.DirectoryCid.String(),
				DirectoryName: bibLocation.DirectoryName,
				IpfsUrl:       bibLocation.IpfsUrl().String(),
				GatewayUrl:    gatewayUrl.String(),
				ContentOrigin: string(bibContent.Contents.Origin),
			})
		} else {
			notArchivedEntries = append(notArchivedEntries, NotArchivedOutput{
				CiteName: bibContent.Entry.CiteName,
				Doi:      doiFromEntry(&bibContent.Entry),
			})
		}
	}

	return &Output{
		Cid:           location.Root.String(),
		TotalEntries:  len(contents),
		TotalArchived: len(archivedEntries),
		Archived:      archivedEntries,
		NotArchived:   notArchivedEntries,
	}, nil
}

func (o *Output) FormatJson() (string, error) {
	marshalledOutput, err := json.MarshalIndent(o, "", outputIndent)
	if err != nil {
		return "", err
	}

	return string(marshalledOutput), nil

}
