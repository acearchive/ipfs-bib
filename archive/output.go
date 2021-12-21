package archive

import (
	"github.com/ipfs/go-cid"
	"github.com/frawleyskid/ipfs-bib/config"
	"net/url"
)

func (l *Location) ToOutput(cfg *config.Config) (*Output, error) {
	entries := make([]EntryOutput, 0, len(l.Entries))

	for citeName, entry := range l.Entries {
		gatewayUrl, err := entry.GatewayUrl(cfg.Ipfs.Gateway)
		if err != nil {
			return nil, err
		}

		entries = append(entries, EntryOutput{
			CiteName:      citeName,
			FileCid:       entry.FileCid,
			FileName:      entry.FileName,
			DirectoryCid:  entry.DirectoryCid,
			DirectoryName: entry.DirectoryName,
			IpfsUrl:       *entry.IpfsUrl(),
			GatewayUrl:    *gatewayUrl,
		})
	}

	return &Output{
		Cid:     l.Root,
		Entries: entries,
	}, nil
}

type EntryOutput struct {
	CiteName      BibCiteName `json:"citeName"`
	FileCid       cid.Cid     `json:"fileCid"`
	FileName      string      `json:"fileName"`
	DirectoryCid  cid.Cid     `json:"directoryCid"`
	DirectoryName string      `json:"directoryName"`
	IpfsUrl       url.URL     `json:"ipfsUrl"`
	GatewayUrl    url.URL     `json:"gatewayUrl"`
}

type Output struct {
	Cid     cid.Cid       `json:"cid"`
	Entries []EntryOutput `json:"entries"`
}
