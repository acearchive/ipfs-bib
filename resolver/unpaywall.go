package resolver

import (
	"context"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/network"
	"net/http"
	"net/url"
)

type UnpaywallResolver struct {
	httpClient *network.HttpClient
	auth       string
}

func NewUnpaywallResolver(httpClient *network.HttpClient, auth string) SourceResolver {
	if auth == "" {
		return &NoOpResolver{}
	}

	return &UnpaywallResolver{httpClient, auth}
}

func (u *UnpaywallResolver) Resolve(ctx context.Context, locator *config.SourceLocator) (*url.URL, error) {
	if locator.Doi == nil {
		return nil, nil
	}

	rawUrl := fmt.Sprintf("https://api.unpaywall.org/v2/%s?email=%s", *locator.Doi, url.QueryEscape(u.auth))

	requestUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	var jsonResponse map[string]interface{}

	if err := u.httpClient.UnmarshalJson(ctx, http.MethodGet, requestUrl, &jsonResponse); err != nil {
		return nil, err
	}

	bestLocationJson, ok := jsonResponse["best_oa_location"]
	if !ok {
		return nil, nil
	}

	bestLocation := bestLocationJson.(map[string]interface{})

	rawResolvedUrl, ok := bestLocation["url"]
	if !ok {
		return nil, nil
	}

	return url.Parse(rawResolvedUrl.(string))
}
