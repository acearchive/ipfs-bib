package archive

import (
	"context"
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

type DownloadedContent struct {
	Content   []byte
	MediaType string
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
	if err != nil {
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
		Url:    resolvedLocator.Url,
		Header: response.Header,
		Body:   responseBody,
	}

	sourceContent, err := downloadHandler.Handle(ctx, downloadResponse)
	if err != nil {
		return nil, err
	} else if sourceContent == nil {
		return nil, nil
	}

	return &DownloadedContent{
		Content:   sourceContent.Content,
		MediaType: sourceContent.MediaType,
		Origin:    resolvedLocator.Origin,
	}, nil
}

func FromBibtex(ctx context.Context, cfg *config.Config, bib *bibtex.BibTex) (*BibContents, error) {
	client := NewDownloadClient(network.NewClient(cfg.Archive.UserAgent))

	downloadHandler := handler.FromConfig(cfg)

	sourceResolver, err := resolver.FromConfig(cfg)
	if err != nil {
		return nil, err
	}

	contentMap := make(map[BibCiteName]DownloadedContent)
	entryMap := make(map[BibCiteName]bibtex.BibEntry)

	for _, bibEntry := range bib.Entries {
		preferredLocalContent, err := ReadLocalBibSource(bibEntry, preferredMediaTypes)
		if err != nil {
			logging.Verbose.Println(err)
		} else if preferredLocalContent != nil {
			contentMap[BibCiteName(bibEntry.CiteName)] = *preferredLocalContent
			entryMap[BibCiteName(bibEntry.CiteName)] = *bibEntry

			continue
		}

		locator, err := config.LocateEntry(bibEntry)
		if err != nil {
			return nil, err
		}

		if locator != nil {
			downloadedContent, err := client.Download(ctx, locator, downloadHandler, sourceResolver)
			if err != nil {
				logging.Verbose.Println(err)
			} else if downloadedContent != nil {
				contentMap[BibCiteName(bibEntry.CiteName)] = *downloadedContent
				entryMap[BibCiteName(bibEntry.CiteName)] = *bibEntry

				continue
			}
		}

		contingencyLocalContent, err := ReadLocalBibSource(bibEntry, contingencyMediaTypes)
		if err != nil {
			logging.Verbose.Println(err)
		} else if contingencyLocalContent != nil {
			contentMap[BibCiteName(bibEntry.CiteName)] = *contingencyLocalContent
			entryMap[BibCiteName(bibEntry.CiteName)] = *bibEntry

			continue
		}

		logging.Error.Println(fmt.Sprintf("Could not find a source for citation: %s", bibEntry.CiteName))
	}

	return &BibContents{
		Contents: contentMap,
		Entries:  entryMap,
	}, nil
}
