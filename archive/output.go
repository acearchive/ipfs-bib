package archive

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/logging"
	"strconv"
)

const outputIndent = "  "

type ArchivedOutput struct {
	CiteName      string  `json:"citeName"`
	Doi           *string `json:"doi"`
	MediaType     string  `json:"mediaType"`
	FileCid       string  `json:"fileCid"`
	FileName      string  `json:"fileName"`
	DirectoryCid  string  `json:"directoryCid"`
	DirectoryName string  `json:"directoryName"`
	IpfsUrl       string  `json:"ipfsUrl"`
	GatewayUrl    string  `json:"gatewayUrl"`
	ContentOrigin string  `json:"contentOrigin"`
}

type NotArchivedOutput struct {
	CiteName string  `json:"citeName"`
	Doi      *string `json:"doi"`
}

type Output struct {
	Cid           string              `json:"cid"`
	TotalEntries  int                 `json:"totalEntries"`
	TotalArchived int                 `json:"totalArchived"`
	Archived      []ArchivedOutput    `json:"archived"`
	NotArchived   []NotArchivedOutput `json:"notArchived"`
}

func NewOutput(cfg config.Config, metadata []BibMetadata, location Location) (Output, error) {
	archivedEntries := make([]ArchivedOutput, 0, len(metadata))
	notArchivedEntries := make([]NotArchivedOutput, 0, len(metadata))

	for _, bibMetadata := range metadata {
		bibLocation, hasLocation := location.Entries[bibMetadata.Entry.CiteName]

		if hasLocation && bibMetadata.Contents != nil {
			gatewayUrl, err := bibLocation.GatewayUrl(cfg.File.Ipfs.Gateway)
			if err != nil {
				return Output{}, err
			}

			ipfsUrl := bibLocation.IpfsUrl()

			archivedEntries = append(archivedEntries, ArchivedOutput{
				CiteName:      bibMetadata.Entry.CiteName,
				Doi:           bibMetadata.Doi,
				MediaType:     bibMetadata.Contents.MediaType,
				FileCid:       bibLocation.FileCid.String(),
				FileName:      bibLocation.FileName,
				DirectoryCid:  bibLocation.DirectoryCid.String(),
				DirectoryName: bibLocation.DirectoryName,
				IpfsUrl:       ipfsUrl.String(),
				GatewayUrl:    gatewayUrl.String(),
				ContentOrigin: string(bibMetadata.Contents.Origin),
			})
		} else {
			notArchivedEntries = append(notArchivedEntries, NotArchivedOutput{
				CiteName: bibMetadata.Entry.CiteName,
				Doi:      bibMetadata.Doi,
			})
		}
	}

	return Output{
		Cid:           location.Root.String(),
		TotalEntries:  len(metadata),
		TotalArchived: len(archivedEntries),
		Archived:      archivedEntries,
		NotArchived:   notArchivedEntries,
	}, nil
}

func prettyPrintLine(title string, value string) {
	titleFunc := color.New(color.Bold).SprintFunc()
	if _, err := fmt.Fprintf(color.Output, "%s: %s\n", titleFunc(title), value); err != nil {
		logging.Error.Fatal(err)
	}
}

func (o Output) PrettyPrint() {
	good := color.New(color.FgGreen).SprintFunc()
	bad := color.New(color.FgRed).SprintFunc()

	prettyPrintLine("Root CID", o.Cid)
	prettyPrintLine("Total entries", strconv.Itoa(o.TotalEntries))
	prettyPrintLine("Entries archived", good(o.TotalArchived))
	prettyPrintLine("Entries not archived", bad(o.TotalEntries-o.TotalArchived))
}

func (o Output) JsonPrint() {
	marshalledOutput, err := json.MarshalIndent(o, "", outputIndent)
	if err != nil {
		logging.Error.Fatal(err)
	}
	fmt.Println(string(marshalledOutput))
}
