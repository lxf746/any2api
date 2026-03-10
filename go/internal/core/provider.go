package core

import (
	"fmt"
	"sort"
	"strings"
)

func normalizeProviderID(name string) string {
	return strings.TrimSpace(name)
}

type Provider interface {
	ID() string
	Capabilities() ProviderCapabilities
	Models() []ModelInfo
	BuildUpstreamPreview(req UnifiedRequest) map[string]interface{}
	GenerateReply(req UnifiedRequest) string
}

type Registry struct {
	defaultProvider string
	providers       map[string]Provider
}

func NewRegistry(defaultProvider string) *Registry {
	return &Registry{defaultProvider: normalizeProviderID(defaultProvider), providers: map[string]Provider{}}
}

func (r *Registry) Register(p Provider) {
	r.providers[p.ID()] = p
}

func (r *Registry) Resolve(name string) (Provider, error) {
	name = normalizeProviderID(name)
	if name == "" {
		name = r.defaultProvider
	}
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return p, nil
}

func (r *Registry) Providers() []string {
	keys := make([]string, 0, len(r.providers))
	for key := range r.providers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (r *Registry) Models(providerFilter string) ([]ModelInfo, error) {
	if providerFilter != "" {
		p, err := r.Resolve(providerFilter)
		if err != nil {
			return nil, err
		}
		return p.Models(), nil
	}

	var models []ModelInfo
	for _, key := range r.Providers() {
		models = append(models, r.providers[key].Models()...)
	}
	return models, nil
}

// CredentialSchemaProvider is an optional interface that Providers can implement
// to expose their credential field definitions. Admin API checks for this via
// type assertion to dynamically render credential forms.
type CredentialSchemaProvider interface {
	CredentialSchema() map[string]FieldSchema
}

// FieldSchema describes a single credential field for a Provider.
type FieldSchema struct {
	Type      string `json:"type"`      // "string", "boolean"
	Label     string `json:"label"`     // display name
	Sensitive bool   `json:"sensitive"` // whether the field is sensitive
	Required  bool   `json:"required"`  // whether the field is required
}
