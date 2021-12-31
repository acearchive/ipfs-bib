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

type ContentMetadata struct {
	MediaType string
	FileName  string
	Origin    resolver.ContentOrigin
}

type DownloadedContent struct {
	ContentMetadata
	Content []byte
}

func (c DownloadedContent) ToMetadata() ContentMetadata {
	return c.ContentMetadata
}

type DownloadClient struct {
	httpClient *network.HttpClient
}

func NewDownloadClient(httpClient *network.HttpClient) *DownloadClient {
	return &DownloadClient{httpClient}
}

func (c DownloadClient) Download(ctx context.Context, locator config.SourceLocator, downloadHandler handler.DownloadHandler, sourceResolver resolver.SourceResolver) (DownloadedContent, error) {
	redirectedUrl, err := c.httpClient.ResolveRedirect(ctx, locator.Url)
	if err != nil {
		return DownloadedContent{}, err
	}

	redirectedLocator := config.SourceLocator{
		Doi: locator.Doi,
		Url: redirectedUrl,
	}

	resolvedLocator, err := sourceResolver.Resolve(ctx, redirectedLocator)
	if errors.Is(err, resolver.ErrNotResolved) {
		return DownloadedContent{}, ErrNoSource
	} else if err != nil {
		return DownloadedContent{}, err
	}

	response, err := c.httpClient.Request(ctx, http.MethodGet, resolvedLocator.Url)
	if err != nil {
		return DownloadedContent{}, err
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return DownloadedContent{}, err
	}

	if err := response.Body.Close(); err != nil {
		return DownloadedContent{}, err
	}

	downloadResponse := handler.DownloadResponse{
		Url:           resolvedLocator.Url,
		Header:        response.Header,
		Body:          responseBody,
		MediaTypeHint: resolvedLocator.MediaTypeHint,
	}

	sourceContent, err := downloadHandler.Handle(ctx, downloadResponse)
	if errors.Is(err, handler.ErrNotHandled) {
		return DownloadedContent{}, ErrNoSource
	} else if err != nil {
		return DownloadedContent{}, err
	}

	return DownloadedContent{
		ContentMetadata: ContentMetadata{
			MediaType: sourceContent.MediaType,
			FileName:  sourceContent.FileName,
			Origin:    resolvedLocator.Origin,
		},
		Content: sourceContent.Content,
	}, nil
}

func FromBibtex(ctx context.Context, cfg config.Config, bib bibtex.BibTex, downloadResults chan DownloadResult) {
	client := NewDownloadClient(network.NewClient(cfg.File.Archive.UserAgent))

	downloadHandler := handler.FromConfig(cfg)

	sourceResolver, err := resolver.FromConfig(cfg)
	if err != nil {
		downloadResults <- DownloadResult{Error: err}
		close(downloadResults)
		return
	}

	for _, bibEntry := range bib.Entries {
		bibContent := BibContents{Entry: *bibEntry}

		var sourceLocator *config.SourceLocator

		switch locator, err := config.LocateEntry(*bibEntry); {
		case errors.Is(err, config.ErrCouldNotLocateEntry):
			logging.Verbose.Println(err)
		case err != nil:
			downloadResults <- DownloadResult{Error: err}
			close(downloadResults)
			return
		default:
			sourceLocator = &locator
			bibContent.Doi = locator.Doi
		}

		contents, err := ReadLocalBibSource(*bibEntry, true)
		if err == nil {
			bibContent.Contents = &contents
			downloadResults <- DownloadResult{Contents: bibContent}
			continue
		} else if !errors.Is(err, ErrNoSource) {
			logging.Verbose.Println(err)
		}

		if sourceLocator != nil {
			contents, err = client.Download(ctx, *sourceLocator, downloadHandler, sourceResolver)
			if err == nil {
				bibContent.Contents = &contents
				downloadResults <- DownloadResult{Contents: bibContent}
				continue
			} else if !errors.Is(err, ErrNoSource) {
				logging.Verbose.Println(err)
			}
		}

		contents, err = ReadLocalBibSource(*bibEntry, false)
		if err == nil {
			bibContent.Contents = &contents
			downloadResults <- DownloadResult{Contents: bibContent}
			continue
		} else if !errors.Is(err, ErrNoSource) {
			logging.Verbose.Println(err)
		}

		downloadResults <- DownloadResult{Contents: bibContent}

		logging.Error.Println(fmt.Sprintf("Could not find a source for citation: %s", bibEntry.CiteName))
	}

	close(downloadResults)
}
