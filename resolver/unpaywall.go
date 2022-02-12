package resolver

import (
	"context"
	"fmt"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/network"
	"net/http"
	"net/url"
)

const ContentOriginUnpaywall ContentOrigin = "unpaywall"

var unpaywallMediaTypeHint = "application/pdf"

type unpaywallLocationResponse struct {
	Url string `json:"url_for_pdf"`
}

type unpaywallResponse struct {
	BestLocation unpaywallLocationResponse `json:"best_oa_location"`
}

type UnpaywallResolver struct {
	httpClient *network.HttpClient
	auth       string
}

func NewUnpaywallResolver(httpClient *network.HttpClient, cfg config.Config) SourceResolver {
	if !cfg.File.Unpaywall.Enabled || cfg.File.Unpaywall.Email == "" {
		return &NoOpResolver{}
	}

	return &UnpaywallResolver{httpClient, cfg.File.Unpaywall.Email}
}

func (u *UnpaywallResolver) Resolve(ctx context.Context, locator config.SourceLocator) (ResolvedLocator, error) {
	if locator.Doi == nil {
		return ResolvedLocator{}, ErrNotResolved
	}

	rawUrl := fmt.Sprintf("https://api.unpaywall.org/v2/%s?email=%s", url.PathEscape(*locator.Doi), url.QueryEscape(u.auth))

	requestUrl, err := url.Parse(rawUrl)
	if err != nil {
		logging.Error.Fatal(fmt.Errorf("%w: %v", network.ErrInvalidApiUrl, err))
	}

	response, err := u.httpClient.Request(ctx, http.MethodGet, *requestUrl)
	if err != nil {
		return ResolvedLocator{}, err
	}

	apiResponse := unpaywallResponse{}

	if err := network.UnmarshalJson(response, &apiResponse); err != nil {
		return ResolvedLocator{}, err
	}

	if err := response.Body.Close(); err != nil {
		return ResolvedLocator{}, err
	}

	if apiResponse.BestLocation.Url == "" {
		return ResolvedLocator{}, ErrNotResolved
	}

	resolvedUrl, err := url.Parse(apiResponse.BestLocation.Url)
	if err != nil {
		return ResolvedLocator{}, fmt.Errorf("%w: %v", network.ErrUnmarshalResponse, err)
	}

	return ResolvedLocator{
		ResolvedUrl:   *resolvedUrl,
		OriginalUrl:   locator.Url,
		Origin:        ContentOriginUnpaywall,
		MediaTypeHint: &unpaywallMediaTypeHint,
	}, nil
}
