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
	"net/url"
	"strconv"
	"strings"
)

const (
	zoteroApiVersion = 3
	apiPageLimit     = 50
)

var zoteroHeaders = map[string]string{
	"Zotero-API-Version": strconv.Itoa(zoteroApiVersion),
}

type ZoteroLinkMode string

const (
	LinkModeImportedFile ZoteroLinkMode = "imported_file"
	LinkModeImportedUrl  ZoteroLinkMode = "imported_url"
	LinkModeLinkedFile   ZoteroLinkMode = "linked_file"
	LinkModeLinkedUrl    ZoteroLinkMode = "linked_url"
)

const ContentOriginZotero resolver.ContentOrigin = "zotero"

type ZoteroAttachment struct {
	Key       ZoteroKey
	LinkMode  ZoteroLinkMode
	Url       *url.URL
	MediaType string
	FileName  string
}

func (a ZoteroAttachment) IsPreferred() bool {
	return IsPreferredMediaType(a.MediaType)
}

type ZoteroKey = string

type zoteroCitationResponse struct {
	Key ZoteroKey `json:"key"`
	Bib string    `json:"biblatex"`
}

func (r zoteroCitationResponse) ParseBib() (bibtex.BibEntry, error) {
	bib, err := bibtex.Parse(strings.NewReader(r.Bib))
	if err != nil {
		return bibtex.BibEntry{}, fmt.Errorf("%w: %v", network.ErrUnmarshalResponse, err)
	}

	if len(bib.Entries) == 0 {
		return bibtex.BibEntry{}, fmt.Errorf("%w: %s", network.ErrUnmarshalResponse, "invalid bibtex entry")
	}

	return *bib.Entries[0], nil
}

type zoteroAttachmentDataResponse struct {
	CitationKey ZoteroKey      `json:"parentItem"`
	Url         string         `json:"url"`
	LinkMode    ZoteroLinkMode `json:"linkMode"`
	MediaType   string         `json:"contentType"`
	FileName    string         `json:"filename"`
}

type zoteroAttachmentResponse struct {
	Key  ZoteroKey                    `json:"key"`
	Data zoteroAttachmentDataResponse `json:"data"`
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
	var citeResponseList []zoteroCitationResponse

	startIndex := 0

	for {
		rawApiUrl := fmt.Sprintf("https://api.zotero.org/groups/%s/items?include=biblatex&start=%d&limit=%d", url.PathEscape(groupId), startIndex, apiPageLimit)

		apiUrl, err := url.Parse(rawApiUrl)
		if err != nil {
			logging.Error.Fatal(fmt.Errorf("%w: %v", network.ErrInvalidApiUrl, err))
		}

		apiResponse, err := c.httpClient.RequestWithHeaders(ctx, http.MethodGet, *apiUrl, zoteroHeaders)
		if err != nil {
			return nil, err
		}

		var currentResponseList []zoteroCitationResponse

		if err := network.UnmarshalJson(apiResponse, &currentResponseList); err != nil {
			return nil, err
		}

		if err := apiResponse.Body.Close(); err != nil {
			return nil, err
		}

		startIndex += len(currentResponseList)
		citeResponseList = append(citeResponseList, currentResponseList...)

		if len(currentResponseList) < apiPageLimit {
			break
		}
	}

	citeMap := make(map[ZoteroKey]bibtex.BibEntry)

	for _, citeResponse := range citeResponseList {
		bib, err := citeResponse.ParseBib()
		if err != nil {
			logging.Verbose.Println(err)
			continue
		}

		citeMap[citeResponse.Key] = bib
	}

	return citeMap, nil
}

func (c *ZoteroClient) downloadAttachmentList(ctx context.Context, groupId string) (map[ZoteroKey][]ZoteroAttachment, error) {
	var attachmentResponseList []zoteroAttachmentResponse

	startIndex := 0

	for {
		rawApiUrl := fmt.Sprintf("https://api.zotero.org/groups/%s/items?itemType=attachment&start=%d&limit=%d", url.PathEscape(groupId), startIndex, apiPageLimit)

		apiUrl, err := url.Parse(rawApiUrl)
		if err != nil {
			logging.Error.Fatal(fmt.Errorf("%w: %v", network.ErrInvalidApiUrl, err))
		}

		apiResponse, err := c.httpClient.RequestWithHeaders(ctx, http.MethodGet, *apiUrl, zoteroHeaders)
		if err != nil {
			return nil, err
		}

		var currentResponseList []zoteroAttachmentResponse

		if err := network.UnmarshalJson(apiResponse, &currentResponseList); err != nil {
			return nil, err
		}

		if err := apiResponse.Body.Close(); err != nil {
			return nil, err
		}

		startIndex += len(currentResponseList)
		attachmentResponseList = append(attachmentResponseList, currentResponseList...)

		if len(currentResponseList) < apiPageLimit {
			break
		}
	}

	attachmentMap := make(map[ZoteroKey][]ZoteroAttachment)

	for _, attachmentResponse := range attachmentResponseList {
		var (
			attachmentUrl *url.URL
			err           error
		)

		attachmentUrl, err = url.Parse(attachmentResponse.Data.Url)
		if attachmentResponse.Data.Url == "" || err != nil {
			attachmentUrl = nil
		}

		attachment := ZoteroAttachment{
			Key:       attachmentResponse.Key,
			Url:       attachmentUrl,
			LinkMode:  attachmentResponse.Data.LinkMode,
			MediaType: attachmentResponse.Data.MediaType,
			FileName:  attachmentResponse.Data.FileName,
		}

		attachmentMap[attachmentResponse.Data.CitationKey] = append(attachmentMap[attachmentResponse.Data.CitationKey], attachment)
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

func (c *ZoteroClient) DownloadAttachment(ctx context.Context, groupId string, attachment ZoteroAttachment) (DownloadedContent, error) {
	var (
		downloadUrl *url.URL
		err         error
	)

	switch attachment.LinkMode {
	case LinkModeLinkedUrl, LinkModeImportedUrl:
		if attachment.Url == nil {
			return DownloadedContent{}, ErrNoSource
		} else {
			downloadUrl = attachment.Url
		}
	case LinkModeLinkedFile, LinkModeImportedFile:
		rawApiUrl := fmt.Sprintf("https://api.zotero.org/groups/%s/items/%s/file", url.PathEscape(groupId), url.PathEscape(attachment.Key))

		downloadUrl, err = url.Parse(rawApiUrl)
		if err != nil {
			logging.Error.Fatal(fmt.Errorf("%w: %v", network.ErrInvalidApiUrl, err))
		}
	}

	downloadResponse, err := c.httpClient.RequestWithHeaders(ctx, http.MethodGet, *downloadUrl, zoteroHeaders)
	if err != nil {
		return DownloadedContent{}, err
	}

	content, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		return DownloadedContent{}, fmt.Errorf("%w: %v", network.ErrHttp, err)
	}

	if err := downloadResponse.Body.Close(); err != nil {
		return DownloadedContent{}, fmt.Errorf("%w: %v", network.ErrHttp, err)
	}

	filename := attachment.FileName
	if filename == "" {
		filename = config.InferFileName(attachment.Url, attachment.MediaType, downloadResponse.Header)
	}

	return DownloadedContent{
		Content:   content,
		MediaType: attachment.MediaType,
		Origin:    ContentOriginZotero,
		FileName:  filename,
	}, nil
}

func FromZotero(ctx context.Context, cfg config.Config, groupId string) ([]BibContents, error) {
	httpClient := network.NewClient(cfg.Archive.UserAgent)

	zoteroClient := NewZoteroClient(httpClient)

	downloadClient := NewDownloadClient(httpClient)

	downloadHandler := handler.FromConfig(cfg)

	sourceResolver, err := resolver.FromConfig(cfg)
	if err != nil {
		return nil, err
	}

	citations, err := zoteroClient.DownloadCitations(ctx, groupId)
	if err != nil {
		return nil, err
	}

	bibContentsList := make([]BibContents, len(citations))

citeMap:
	for citationIndex, citation := range citations {
		bibContent := BibContents{Entry: citation.Entry}

		var sourceLocator *config.SourceLocator

		switch locator, err := config.LocateEntry(citation.Entry); {
		case errors.Is(err, config.ErrCouldNotLocateEntry):
			logging.Verbose.Println(err)
		case err != nil:
			return nil, err
		default:
			sourceLocator = &locator
			bibContent.Doi = locator.Doi
		}

		for _, attachment := range citation.Attachments {
			if attachment.IsPreferred() {
				contents, err := zoteroClient.DownloadAttachment(ctx, groupId, attachment)
				if err == nil {
					bibContent.Contents = &contents
					bibContentsList[citationIndex] = bibContent
					continue citeMap
				} else if !errors.Is(err, ErrNoSource) {
					logging.Verbose.Println(err)
				}
			}
		}

		if sourceLocator != nil {
			contents, err := downloadClient.Download(ctx, *sourceLocator, downloadHandler, sourceResolver)
			if err == nil {
				bibContent.Contents = &contents
				bibContentsList[citationIndex] = bibContent
				continue
			} else if !errors.Is(err, ErrNoSource) {
				logging.Verbose.Println(err)
			}
		}

		if len(citation.Attachments) > 0 {
			contents, err := zoteroClient.DownloadAttachment(ctx, groupId, citation.Attachments[0])
			if err == nil {
				bibContent.Contents = &contents
				bibContentsList[citationIndex] = bibContent
				continue
			} else if !errors.Is(err, ErrNoSource) {
				logging.Verbose.Println(err)
			}
		}

		bibContentsList[citationIndex] = bibContent

		logging.Error.Println(fmt.Sprintf("Could not find a source for citation: %s", citation.Entry.CiteName))
	}

	return bibContentsList, nil
}
