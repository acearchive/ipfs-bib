package archive

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/handler"
	"github.com/frawleyskid/ipfs-bib/network"
	"github.com/frawleyskid/ipfs-bib/resolver"
	"github.com/frawleyskid/ipfs-bib/store"
	"github.com/nickng/bibtex"
	"os"
)

type BibCiteName string

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

type BibContents struct {
	Entries map[BibCiteName]bibtex.BibEntry
	Sources map[BibCiteName]handler.SourceContent
}

func Download(ctx context.Context, cfg *config.Config, bib *bibtex.BibTex) (*BibContents, error) {
	client := NewClient(network.NewClient(cfg.Archive.UserAgent))

	downloadHandler := handler.FromConfig(cfg)

	sourceResolver, err := resolver.FromConfig(cfg)
	if err != nil {
		return nil, err
	}

	contentMap := make(map[BibCiteName]handler.SourceContent)
	entryMap := make(map[BibCiteName]bibtex.BibEntry)

	for _, bibEntry := range bib.Entries {
		locator, err := config.LocateEntry(bibEntry)
		if err != nil {
			return nil, err
		}

		content, err := client.Download(ctx, locator, downloadHandler, sourceResolver)
		if err != nil {
			return nil, err
		} else if content == nil {
			continue
		}

		contentMap[BibCiteName(bibEntry.CiteName)] = *content
		entryMap[BibCiteName(bibEntry.CiteName)] = *bibEntry
	}

	return &BibContents{
		Sources: contentMap,
		Entries: entryMap,
	}, nil
}

type Location struct {
	Root    cid.Cid
	Entries map[BibCiteName]config.BibEntryLocation
}

func storeContents(ctx context.Context, cfg *config.Config, contents *BibContents, sourceStore *store.SourceStore) (*Location, error) {
	sourcePathTemplate, err := config.NewSourcePathTemplate(cfg)
	if err != nil {
		return nil, err
	}

	locationMap := make(map[BibCiteName]config.BibEntryLocation)

	for citeName, source := range contents.Sources {
		bibEntry := contents.Entries[citeName]

		sourcePath, err := sourcePathTemplate.Execute(&bibEntry, source.MediaType)
		if err != nil {
			return nil, err
		}

		bibSource := &config.BibSource{
			Content:       source.Content,
			FileName:      sourcePath.FileName,
			DirectoryName: sourcePath.DirectoryName,
		}

		entryLocation, err := sourceStore.AddSource(ctx, bibSource)
		if err != nil {
			return nil, err
		}

		locationMap[citeName] = *entryLocation
	}

	rootCid, err := sourceStore.Write(ctx)
	if err != nil {
		return nil, err
	}

	return &Location{
		Root:    rootCid,
		Entries: locationMap,
	}, nil
}

func ToCar(ctx context.Context, cfg *config.Config, carPath string, contents *BibContents) (*Location, error) {
	dagService, err := store.CarService()
	if err != nil {
		return nil, err
	}

	sourceStore, err := store.NewSourceStore(ctx, dagService)
	if err != nil {
		return nil, err
	}

	location, err := storeContents(ctx, cfg, contents, sourceStore)
	if err != nil {
		return nil, err
	}

	if err := store.WriteCar(ctx, dagService, location.Root, carPath, cfg.Ipfs.CarVersion == "2"); err != nil {
		return nil, err
	}

	return location, nil
}

func ToNode(ctx context.Context, cfg *config.Config, pin bool, contents *BibContents) (*Location, error) {
	ipfsApi, err := store.IpfsClient(cfg.Ipfs.Api)
	if err != nil {
		return nil, err
	}

	sourceStore, err := store.NewSourceStore(ctx, ipfsApi.Dag())
	if err != nil {
		return nil, err
	}

	location, err := storeContents(ctx, cfg, contents, sourceStore)
	if err != nil {
		return nil, err
	}

	if pin {
		if err := store.Pin(ctx, ipfsApi, location.Root, true); err != nil {
			return nil, err
		}
	}

	return location, nil
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
