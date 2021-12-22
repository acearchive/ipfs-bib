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

func NewUnpaywallResolver(httpClient *network.HttpClient, cfg *config.Config) SourceResolver {
	if !cfg.Unpaywall.Enabled || cfg.Unpaywall.Email == "" {
		return &NoOpResolver{}
	}

	return &UnpaywallResolver{httpClient, cfg.Unpaywall.Email}
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

	response, err := u.httpClient.Request(ctx, http.MethodGet, requestUrl)
	if err != nil {
		return nil, err
	}

	if err := network.UnmarshalJson(response, &jsonResponse); err != nil {
		return nil, err
	}

	bestLocationJson, ok := jsonResponse["best_oa_location"]
	if !ok {
		return nil, nil
	}

	bestLocation, ok := bestLocationJson.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	rawResolvedUrl, ok := bestLocation["url"]
	if !ok {
		return nil, nil
	}

	return url.Parse(rawResolvedUrl.(string))
}
