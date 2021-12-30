package archive

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/store"
	"github.com/nickng/bibtex"
)

const downloadResultChanSize = 16

type BibCiteName = string

type BibMetadata struct {
	Entry    bibtex.BibEntry
	Doi      *string
	Contents *ContentMetadata
}

type BibContents struct {
	Entry    bibtex.BibEntry
	Doi      *string
	Contents *DownloadedContent
}

func (c BibContents) ToMetadata() BibMetadata {
	bibMetadata := BibMetadata{
		Entry: c.Entry,
		Doi:   c.Doi,
	}

	if c.Contents != nil {
		metadata := c.Contents.ToMetadata()
		bibMetadata.Contents = &metadata
	}

	return bibMetadata
}

type DownloadResult struct {
	Contents BibContents
	Error    error
}

func MetadataToBibtex(metadata []BibMetadata) bibtex.BibTex {
	bib := bibtex.NewBibTex()

	for i := range metadata {
		bib.Entries = append(bib.Entries, &metadata[i].Entry)
	}

	return *bib
}

func NewDownloadResultChan() chan DownloadResult {
	return make(chan DownloadResult, downloadResultChanSize)
}

type Location struct {
	Root    cid.Cid
	Entries map[BibCiteName]config.BibEntryLocation
}

func Store(ctx context.Context, cfg config.Config, contents chan DownloadResult, sourceStore store.SourceStore) (Location, []BibMetadata, error) {
	// We may have multiple contents with the same bibtex cite name, so we need
	// to deduplicate them by choosing the "best" contents for a given cite name.
	deduplicatedContents := DeduplicateContents(contents)

	sourcePathTemplate, err := config.NewSourcePathTemplate(cfg)
	if err != nil {
		return Location{}, nil, err
	}

	locationMap := make(map[BibCiteName]config.BibEntryLocation)

	var metadataList []BibMetadata //nolint:prealloc

	for downloadResult := range deduplicatedContents {
		if downloadResult.Error != nil {
			return Location{}, nil, err
		}

		bibContent := downloadResult.Contents

		metadataList = append(metadataList, bibContent.ToMetadata())

		if bibContent.Contents == nil {
			continue
		}

		sourcePath := sourcePathTemplate.Execute(bibContent.Entry, bibContent.Contents.FileName, bibContent.Contents.MediaType)

		bibSource := config.BibSource{
			Content:       bibContent.Contents.Content,
			FileName:      sourcePath.FileName,
			DirectoryName: sourcePath.DirectoryName,
		}

		entryLocation, err := sourceStore.AddSource(ctx, bibSource)
		if err != nil {
			return Location{}, nil, err
		}

		locationMap[bibContent.Entry.CiteName] = entryLocation
	}

	rootCid, err := sourceStore.Finalize(ctx)
	if err != nil {
		return Location{}, nil, err
	}

	return Location{
		Root:    rootCid,
		Entries: locationMap,
	}, metadataList, nil
}
