package archive

import (
	"context"
	"errors"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/handler"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/network"
	"github.com/frawleyskid/ipfs-bib/resolver"
	"github.com/nickng/bibtex"
	"io"
	"net/http"
)

var ErrNoSource = errors.New("source not found")

type DownloadedContent struct {
	Content   []byte
	MediaType string
	FileName  string
	Origin    resolver.ContentOrigin
}

type DownloadClient struct {
	httpClient *network.HttpClient
}

func NewDownloadClient(httpClient *network.HttpClient) *DownloadClient {
	return &DownloadClient{httpClient}
}

func (c DownloadClient) Download(ctx context.Context, locator *config.SourceLocator, downloadHandler handler.DownloadHandler, sourceResolver resolver.SourceResolver) (*DownloadedContent, error) {
	redirectedUrl, err := c.httpClient.ResolveRedirect(ctx, &locator.Url)
	if err != nil {
		return nil, err
	}

	redirectedLocator := &config.SourceLocator{
		Doi: locator.Doi,
		Url: *redirectedUrl,
	}

	resolvedLocator, err := sourceResolver.Resolve(ctx, redirectedLocator)
	if errors.Is(err, resolver.ErrNotResolved) {
		return nil, ErrNoSource
	} else if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Request(ctx, http.MethodGet, &resolvedLocator.Url)
	if err != nil {
		return nil, err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := response.Body.Close(); err != nil {
		return nil, err
	}

	downloadResponse := &handler.DownloadResponse{
		Url:           resolvedLocator.Url,
		Header:        response.Header,
		Body:          responseBody,
		MediaTypeHint: resolvedLocator.MediaTypeHint,
	}

	sourceContent, err := downloadHandler.Handle(ctx, downloadResponse)
	if errors.Is(err, handler.ErrNotHandled) {
		return nil, ErrNoSource
	} else if err != nil {
		return nil, err
	}

	return &DownloadedContent{
		Content:   sourceContent.Content,
		MediaType: sourceContent.MediaType,
		FileName:  sourceContent.FileName,
		Origin:    resolvedLocator.Origin,
	}, nil
}

func FromBibtex(ctx context.Context, cfg *config.Config, bib *bibtex.BibTex) ([]BibContents, error) {
	client := NewDownloadClient(network.NewClient(cfg.Archive.UserAgent))

	downloadHandler := handler.FromConfig(cfg)

	sourceResolver, err := resolver.FromConfig(cfg)
	if err != nil {
		return nil, err
	}

	bibContentsList := make([]BibContents, len(bib.Entries))

	for entryIndex, bibEntry := range bib.Entries {
		locator, err := config.LocateEntry(bibEntry)
		if errors.Is(err, config.ErrCouldNotLocateEntry) {
			logging.Verbose.Println(err)
			continue
		} else if err != nil {
			return nil, err
		}

		bibContent := BibContents{Entry: *bibEntry, Doi: locator.Doi}

		bibContent.Contents, err = ReadLocalBibSource(bibEntry, preferredMediaTypes)
		if err == nil {
			bibContentsList[entryIndex] = bibContent
			continue
		} else if !errors.Is(err, ErrNoSource) {
			logging.Verbose.Println(err)
		}

		if locator != nil {
			bibContent.Contents, err = client.Download(ctx, locator, downloadHandler, sourceResolver)
			if err == nil {
				bibContentsList[entryIndex] = bibContent
				continue
			} else if !errors.Is(err, ErrNoSource) {
				logging.Verbose.Println(err)
			}
		}

		bibContent.Contents, err = ReadLocalBibSource(bibEntry, contingencyMediaTypes)
		if err == nil {
			bibContentsList[entryIndex] = bibContent
			continue
		} else if !errors.Is(err, ErrNoSource) {
			logging.Verbose.Println(err)
		}

		bibContent.Contents = nil
		bibContentsList[entryIndex] = bibContent

		logging.Error.Println(fmt.Sprintf("Could not find a source for citation: %s", bibEntry.CiteName))
	}

	return bibContentsList, nil
}
