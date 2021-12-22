package archive

import (
	"context"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/handler"
	"github.com/frawleyskid/ipfs-bib/network"
	"github.com/frawleyskid/ipfs-bib/resolver"
	"github.com/nickng/bibtex"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const zoteroApiVersion = 3

var zoteroHeaders = map[string]string{
	"Zotero-API-Version": strconv.Itoa(zoteroApiVersion),
}

type ZoteroAttachment struct {
	Url       url.URL
	MediaType string
}

func (a *ZoteroAttachment) IsPreferred() bool {
	for _, preferredMediaType := range preferredMediaTypes {
		if a.MediaType == preferredMediaType {
			return true
		}
	}

	return false
}

type ZoteroKey = string

type ZoteroCitationResponse struct {
	Key ZoteroKey `json:"key"`
	Bib string    `json:"biblatex"`
}

func (r *ZoteroCitationResponse) ParseBib() (*bibtex.BibEntry, error) {
	bib, err := bibtex.Parse(strings.NewReader(r.Bib))
	if err != nil {
		return nil, err
	}

	if len(bib.Entries) == 0 {
		return nil, nil
	}

	return bib.Entries[0], nil
}

type ZoteroAttachmentData struct {
	Key       ZoteroKey `json:"parentItem"`
	Url       string    `json:"url"`
	MediaType string    `json:"contentType"`
}

type ZoteroAttachmentResponse struct {
	Data ZoteroAttachmentData `json:"data"`
}

type ZoteroCitation struct {
	Entry       bibtex.BibEntry
	Attachments []ZoteroAttachment
}

type ZoteroClient struct {
	httpClient *network.HttpClient
}

func NewZoteroClient(httpClient *network.HttpClient) *ZoteroClient {
	return &ZoteroClient{httpClient}
}

func (c *ZoteroClient) downloadCiteList(ctx context.Context, groupId string) (map[ZoteroKey]bibtex.BibEntry, error) {
	rawApiUrl := fmt.Sprintf("https://api.zotero.org/groups/%s/items?include=biblatex", url.PathEscape(groupId))

	apiUrl, err := url.Parse(rawApiUrl)
	if err != nil {
		return nil, err
	}

	apiResponse, err := c.httpClient.RequestWithHeaders(ctx, http.MethodGet, apiUrl, zoteroHeaders)
	if err != nil {
		return nil, err
	}

	var citeResponseList []ZoteroCitationResponse

	if err := network.UnmarshalJson(apiResponse, &citeResponseList); err != nil {
		return nil, err
	}

	citeMap := make(map[ZoteroKey]bibtex.BibEntry)

	for _, citeResponse := range citeResponseList {
		bib, err := citeResponse.ParseBib()
		if err != nil || bib == nil {
			continue
		}

		citeMap[citeResponse.Key] = *bib
	}

	return citeMap, nil
}

func (c *ZoteroClient) downloadAttachmentList(ctx context.Context, groupId string) (map[ZoteroKey][]ZoteroAttachment, error) {
	rawApiUrl := fmt.Sprintf("https://api.zotero.org/groups/%s/items?itemType=attachment", url.PathEscape(groupId))

	apiUrl, err := url.Parse(rawApiUrl)
	if err != nil {
		return nil, err
	}

	apiResponse, err := c.httpClient.RequestWithHeaders(ctx, http.MethodGet, apiUrl, zoteroHeaders)
	if err != nil {
		return nil, err
	}

	var attachmentResponseList []ZoteroAttachmentResponse

	if err := network.UnmarshalJson(apiResponse, &attachmentResponseList); err != nil {
		return nil, err
	}

	attachmentMap := make(map[ZoteroKey][]ZoteroAttachment)

	for _, attachmentResponse := range attachmentResponseList {
		attachmentUrl, err := url.Parse(attachmentResponse.Data.Url)
		if err != nil {
			continue
		}

		attachment := ZoteroAttachment{
			Url:       *attachmentUrl,
			MediaType: attachmentResponse.Data.MediaType,
		}

		attachmentMap[attachmentResponse.Data.Key] = append(attachmentMap[attachmentResponse.Data.Key], attachment)
	}

	return attachmentMap, nil
}

func (c *ZoteroClient) DownloadCitations(ctx context.Context, groupId string) ([]ZoteroCitation, error) {
	citeMap, err := c.downloadCiteList(ctx, groupId)
	if err != nil {
		return nil, err
	}

	attachmentMap, err := c.downloadAttachmentList(ctx, groupId)
	if err != nil {
		return nil, err
	}

	citations := make([]ZoteroCitation, 0, len(citeMap))

	for zoteroKey, bibEntry := range citeMap {
		citation := ZoteroCitation{
			Entry:       bibEntry,
			Attachments: attachmentMap[zoteroKey],
		}
		citations = append(citations, citation)
	}

	return citations, nil
}

func (c *ZoteroClient) DownloadAttachment(ctx context.Context, attachment *ZoteroAttachment) (*handler.SourceContent, error) {
	content, err := c.httpClient.Download(ctx, &attachment.Url)
	if err != nil {
		return nil, err
	}

	return &handler.SourceContent{
		Content:   content,
		MediaType: attachment.MediaType,
	}, nil
}

func FromZotero(ctx context.Context, cfg *config.Config, groupId string) (*BibContents, error) {
	httpClient := network.NewClient(cfg.Archive.UserAgent)

	zoteroClient := NewZoteroClient(httpClient)

	downloadClient := NewDownloadClient(httpClient)

	downloadHandler := handler.FromConfig(cfg)

	sourceResolver, err := resolver.FromConfig(cfg)
	if err != nil {
		return nil, err
	}

	contentMap := make(map[BibCiteName]handler.SourceContent)
	entryMap := make(map[BibCiteName]bibtex.BibEntry)

	citations, err := zoteroClient.DownloadCitations(ctx, groupId)
	if err != nil {
		return nil, err
	}

citeMap:
	for _, citation := range citations {
		entryMap[BibCiteName(citation.Entry.CiteName)] = citation.Entry

		for _, attachment := range citation.Attachments {
			if attachment.IsPreferred() {
				preferredContent, err := zoteroClient.DownloadAttachment(ctx, &attachment)
				if err != nil {
					return nil, err
				}

				contentMap[BibCiteName(citation.Entry.CiteName)] = *preferredContent

				continue citeMap
			}
		}

		locator, err := config.LocateEntry(&citation.Entry)
		if err != nil {
			return nil, err
		}

		downloadedContent, err := downloadClient.Download(ctx, locator, downloadHandler, sourceResolver)
		if err != nil {
			return nil, err
		}

		if downloadedContent != nil {
			contentMap[BibCiteName(citation.Entry.CiteName)] = *downloadedContent
			continue
		}

		if len(citation.Attachments) > 0 {
			contingencyContent, err := zoteroClient.DownloadAttachment(ctx, &citation.Attachments[0])
			if err != nil {
				return nil, err
			}

			contentMap[BibCiteName(citation.Entry.CiteName)] = *contingencyContent
		}
	}

	return &BibContents{
		Sources: contentMap,
		Entries: entryMap,
	}, nil
}
