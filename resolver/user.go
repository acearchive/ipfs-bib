package resolver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/logging"
	"github.com/frawleyskid/ipfs-bib/network"
	"net/url"
	"text/template"
)

const ContentOriginUser ContentOrigin = "resolver"

type resolverRule struct {
	Name    string
	Schemes []template.Template
	Filter  config.HostnameFilter
}

type UserResolver struct {
	httpClient *network.HttpClient
	rules      []resolverRule
}

func NewUserResolver(httpClient *network.HttpClient, cfg []config.Resolver) (SourceResolver, error) {
	if len(cfg) == 0 {
		return &NoOpResolver{}, nil
	}

	rules := make([]resolverRule, len(cfg))

	for resolverIndex, resolverCfg := range cfg {
		rule := resolverRule{}

		for schemeIndex, scheme := range resolverCfg.Schemes {
			templateName := fmt.Sprintf("resolvers.%d.schemes.%d", resolverIndex, schemeIndex)
			tmpl, err := template.New(templateName).Funcs(sprig.TxtFuncMap()).Parse(scheme)
			if err != nil {
				return nil, fmt.Errorf("%w: %v", config.ErrInvalidTemplate, err)
			}

			rule.Schemes = append(rule.Schemes, *tmpl)
		}

		rule.Filter.Include = make([]string, len(resolverCfg.IncludeHostnames))
		copy(rule.Filter.Include, resolverCfg.IncludeHostnames)

		rule.Filter.Exclude = make([]string, len(resolverCfg.ExcludeHostnames))
		copy(rule.Filter.Exclude, resolverCfg.ExcludeHostnames)

		rules[resolverIndex] = rule
	}

	return &UserResolver{httpClient, rules}, nil
}

func (u *UserResolver) Resolve(ctx context.Context, locator config.SourceLocator) (ResolvedLocator, error) {
	for _, rule := range u.rules {
		templateInput := config.NewProxySchemeInput(locator, rule.Filter)
		if templateInput == nil {
			// This source was excluded by the hostname include/exclude rules.
			continue
		}

		var rawProxyUrlBytes bytes.Buffer

		for _, scheme := range rule.Schemes {
			if err := scheme.Execute(&rawProxyUrlBytes, templateInput); err != nil {
				logging.Error.Fatal(err)
			}

			rawProxyUrl := rawProxyUrlBytes.String()

			if rawProxyUrl == "" {
				// We skip templates that resolve to an empty string.
				continue
			}

			proxyUrl, err := url.Parse(rawProxyUrl)
			if err != nil {
				continue
			}

			exists, err := u.httpClient.CheckExists(ctx, *proxyUrl)
			if err != nil {
				logging.Verbose.Println(err)
				continue
			}

			if exists {
				return ResolvedLocator{
					ResolvedUrl:   *proxyUrl,
					OriginalUrl:   locator.Url,
					Origin:        ContentOriginUser,
					MediaTypeHint: nil,
				}, nil
			}
		}
	}

	return ResolvedLocator{}, ErrNotResolved
}
