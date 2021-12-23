package archive

import (
	"context"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/handler"
	"github.com/frawleyskid/ipfs-bib/network"
	"github.com/frawleyskid/ipfs-bib/resolver"
	"github.com/nickng/bibtex"
	"io"
	"log"
	"net/http"
)

type DownloadClient struct {
	httpClient *network.HttpClient
}

func NewDownloadClient(httpClient *network.HttpClient) *DownloadClient {
	return &DownloadClient{httpClient}
}

func (c DownloadClient) Download(ctx context.Context, locator *config.SourceLocator, downloadHandler handler.DownloadHandler, sourceResolver resolver.SourceResolver) (*handler.SourceContent, error) {
	redirectedUrl, err := c.httpClient.ResolveRedirect(ctx, &locator.Url)
	if err != nil {
		return nil, err
	}

	redirectedLocator := &config.SourceLocator{
		Doi: locator.Doi,
		Url: *redirectedUrl,
	}

	resolvedUrl, err := sourceResolver.Resolve(ctx, redirectedLocator)
	if err != nil {
		return nil, err
	}

	response, err := c.httpClient.Request(ctx, http.MethodGet, resolvedUrl)
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
		Url:    *resolvedUrl,
		Header: response.Header,
		Body:   responseBody,
	}

	content, err := downloadHandler.Handle(ctx, downloadResponse)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func FromBibtex(ctx context.Context, cfg *config.Config, bib *bibtex.BibTex) (*BibContents, error) {
	client := NewDownloadClient(network.NewClient(cfg.Archive.UserAgent))

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

		preferredLocalContent, err := ReadLocalBibSource(bibEntry, preferredMediaTypes)
		if err != nil {
			log.Println(err)
		} else if preferredLocalContent != nil {
			contentMap[BibCiteName(bibEntry.CiteName)] = *preferredLocalContent
			entryMap[BibCiteName(bibEntry.CiteName)] = *bibEntry

			continue
		}

		downloadedContent, err := client.Download(ctx, locator, downloadHandler, sourceResolver)
		if err != nil {
			log.Println(err)
		} else if downloadedContent != nil {
			contentMap[BibCiteName(bibEntry.CiteName)] = *downloadedContent
			entryMap[BibCiteName(bibEntry.CiteName)] = *bibEntry

			continue
		}

		contingencyLocalContent, err := ReadLocalBibSource(bibEntry, contingencyMediaTypes)
		if err != nil {
			log.Println(err)
		} else if contingencyLocalContent != nil {
			contentMap[BibCiteName(bibEntry.CiteName)] = *contingencyLocalContent
			entryMap[BibCiteName(bibEntry.CiteName)] = *bibEntry

			continue
		}
	}

	return &BibContents{
		Sources: contentMap,
		Entries: entryMap,
	}, nil
}
