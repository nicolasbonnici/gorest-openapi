package openapi

import (
	"github.com/nicolasbonnici/gorest/plugin"
)

func loadResourcesFromPlugins(registry *plugin.PluginRegistry) []plugin.OpenAPIResource {
	if registry == nil {
		return nil
	}

	var resources []plugin.OpenAPIResource

	for _, p := range registry.GetAll() {
		provider, ok := p.(plugin.OpenAPIProvider)
		if !ok {
			continue
		}

		pluginResources := provider.GetOpenAPIResources()
		resources = append(resources, pluginResources...)
	}

	return resources
}
