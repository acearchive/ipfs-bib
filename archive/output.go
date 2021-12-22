package archive

import (
	"github.com/frawleyskid/ipfs-bib/config"
)

const OutputIndent = "  "

func (l *Location) ToOutput(cfg *config.Config) (*Output, error) {
	entries := make([]EntryOutput, 0, len(l.Entries))

	for citeName, entry := range l.Entries {
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
		})
	}

	return &Output{
		Cid:     l.Root.String(),
		Entries: entries,
	}, nil
}

type EntryOutput struct {
	CiteName      string `json:"citeName"`
	FileCid       string `json:"fileCid"`
	FileName      string `json:"fileName"`
	DirectoryCid  string `json:"directoryCid"`
	DirectoryName string `json:"directoryName"`
	IpfsUrl       string `json:"ipfsUrl"`
	GatewayUrl    string `json:"gatewayUrl"`
}

type Output struct {
	Cid     string        `json:"cid"`
	Entries []EntryOutput `json:"entries"`
}
