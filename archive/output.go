package archive

import (
	"encoding/json"
	"github.com/frawleyskid/ipfs-bib/config"
)

const outputIndent = "  "

type EntryOutput struct {
	CiteName      string `json:"citeName"`
	FileCid       string `json:"fileCid"`
	FileName      string `json:"fileName"`
	DirectoryCid  string `json:"directoryCid"`
	DirectoryName string `json:"directoryName"`
	IpfsUrl       string `json:"ipfsUrl"`
	GatewayUrl    string `json:"gatewayUrl"`
	Origin        string `json:"origin"`
}

type Output struct {
	Cid     string        `json:"cid"`
	Entries []EntryOutput `json:"entries"`
}

func NewOutput(cfg *config.Config, contents *BibContents, location *Location) (*Output, error) {
	entries := make([]EntryOutput, 0, len(location.Entries))

	for citeName, entry := range location.Entries {
		entryContent := contents.Contents[citeName]

		gatewayUrl, err := entry.GatewayUrl(cfg.Ipfs.Gateway)
		if err != nil {
			return nil, err
		}

		entries = append(entries, EntryOutput{
			CiteName:      string(citeName),
			FileCid:       entry.FileCid.String(),
			FileName:      entry.FileName,
			DirectoryCid:  entry.DirectoryCid.String(),
			DirectoryName: entry.DirectoryName,
			IpfsUrl:       entry.IpfsUrl().String(),
			GatewayUrl:    gatewayUrl.String(),
			Origin:        string(entryContent.Origin),
		})
	}

	return &Output{
		Cid:     location.Root.String(),
		Entries: entries,
	}, nil
}

func (o *Output) FormatJson() (string, error) {
	marshalledOutput, err := json.MarshalIndent(o, "", outputIndent)
	if err != nil {
		return "", err
	}

	return string(marshalledOutput), nil

}