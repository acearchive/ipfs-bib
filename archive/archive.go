package archive

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/store"
	"github.com/nickng/bibtex"
)

type BibCiteName = string

type BibContents struct {
	Entry    bibtex.BibEntry
	Doi      *string
	Contents *DownloadedContent
}

func ContentsToBibtex(contents []BibContents) *bibtex.BibTex {
	bib := bibtex.NewBibTex()

	for i := range contents {
		bib.Entries = append(bib.Entries, &contents[i].Entry)
	}

	return bib
}

type Location struct {
	Root    cid.Cid
	Entries map[BibCiteName]config.BibEntryLocation
}

func storeContents(ctx context.Context, cfg *config.Config, contents []BibContents, sourceStore *store.SourceStore) (*Location, error) {
	sourcePathTemplate, err := config.NewSourcePathTemplate(cfg)
	if err != nil {
		return nil, err
	}

	locationMap := make(map[BibCiteName]config.BibEntryLocation)

	for _, bibContent := range contents {
		if bibContent.Contents == nil {
			continue
		}

		sourcePath, err := sourcePathTemplate.Execute(&bibContent.Entry, bibContent.Contents.MediaType)
		if err != nil {
			return nil, err
		}

		bibSource := &config.BibSource{
			Content:       bibContent.Contents.Content,
			FileName:      sourcePath.FileName,
			DirectoryName: sourcePath.DirectoryName,
		}

		entryLocation, err := sourceStore.AddSource(ctx, bibSource)
		if err != nil {
			return nil, err
		}

		locationMap[bibContent.Entry.CiteName] = *entryLocation
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

func ToCar(ctx context.Context, cfg *config.Config, carPath string, contents []BibContents) (*Location, error) {
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

func ToNode(ctx context.Context, cfg *config.Config, pin bool, contents []BibContents) (*Location, error) {
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
