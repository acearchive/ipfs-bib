package resolver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/frawleyskid/ipfs-bib/config"
	"github.com/frawleyskid/ipfs-bib/network"
	"net/url"
	"text/template"
)

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
				return nil, err
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

func (u *UserResolver) Resolve(ctx context.Context, locator *config.SourceLocator) (*ResolvedLocator, error) {
	var resolvedLocators []ResolvedLocator

	for _, rule := range u.rules {
		templateInput, err := config.NewProxySchemeInput(locator, &rule.Filter)
		if err != nil {
			return nil, err
		} else if templateInput == nil {
			// This source was excluded by the hostname include/exclude rules.
			continue
		}

		var rawProxyUrlBytes bytes.Buffer

		for _, scheme := range rule.Schemes {
			if err := scheme.Execute(&rawProxyUrlBytes, templateInput); err != nil {
				return nil, err
			}

			if rawProxyUrlBytes.Len() == 0 {
				// We skip templates that resolve to an empty string.
				continue
			}

			rawProxyUrl := string(rawProxyUrlBytes.Bytes())

			proxyUrl, err := url.Parse(rawProxyUrl)
			if err != nil {
				continue
			}

			resolvedLocator := ResolvedLocator{Url: *proxyUrl, Origin: ContentOrigin(rule.Name)}

			resolvedLocators = append(resolvedLocators, resolvedLocator)
		}
	}

	for _, resolvedLocator := range resolvedLocators {
		exists, err := u.httpClient.CheckExists(ctx, &resolvedLocator.Url)
		if err != nil {
			return nil, err
		}

		if exists {
			return &resolvedLocator, nil
		}
	}

	return nil, nil
}
