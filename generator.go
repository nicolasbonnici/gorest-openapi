package openapi

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/plugin"
)

type GeneratorConfig struct {
	DTOsDirectory      string
	PluginRegistry     *plugin.PluginRegistry
	PaginationLimit    int
	PaginationMaxLimit int
	ServerURL          string
	Title              string
	Version            string
	Description        string
}

func generateOpenAPISpec(app *fiber.App, cfg GeneratorConfig) (map[string]interface{}, error) {
	paths := map[string]interface{}{}
	components := map[string]interface{}{
		"schemas": make(map[string]interface{}),
	}

	resourcePaths := make(map[string]bool)

	var pluginResources []plugin.OpenAPIResource
	if cfg.PluginRegistry != nil {
		pluginResources = loadResourcesFromPlugins(cfg.PluginRegistry)
	}

	if len(pluginResources) > 0 {
		for _, resource := range pluginResources {
			schemaName := strings.ToUpper(resource.Name[:1]) + resource.Name[1:]

			if resource.ResponseModel != nil {
				schema := buildSchemaFromModel(resource.ResponseModel)
				components["schemas"].(map[string]interface{})[schemaName] = schema
			}

			if resource.CreateModel != nil {
				createSchemaName := "Create" + schemaName + "Request"
				schema := buildSchemaFromModel(resource.CreateModel)
				components["schemas"].(map[string]interface{})[createSchemaName] = schema
			}

			if resource.UpdateModel != nil {
				updateSchemaName := "Update" + schemaName + "Request"
				schema := buildSchemaFromModel(resource.UpdateModel)
				components["schemas"].(map[string]interface{})[updateSchemaName] = schema
			}

			base := resource.BasePath
			resourcePaths[base] = true
			resourcePaths[base+"/:id"] = true

			paths[base] = buildCollectionEndpointsFromResource(resource, schemaName, cfg)
			paths[base+"/{id}"] = buildItemEndpointsFromResource(resource, schemaName)
		}
	} else if cfg.DTOsDirectory != "" {
		resourceDTOs, err := loadResourceDTOs(cfg.DTOsDirectory)
		if err != nil {
			return nil, fmt.Errorf("failed to load DTOs: %w", err)
		}

		for _, resource := range resourceDTOs {
			mainDTO := resource.getMainDTO()
			if mainDTO == nil {
				continue
			}

			schemaName := strings.ToUpper(resource.Name[:1]) + resource.Name[1:]
			properties := buildSchemaPropertiesFromDTO(mainDTO.Fields)
			required := getRequiredFieldsFromDTO(mainDTO.Fields)

			schema := map[string]interface{}{
				"type":       "object",
				"properties": properties,
			}

			if len(required) > 0 {
				schema["required"] = required
			}

			components["schemas"].(map[string]interface{})[schemaName] = schema
		}

		for _, resource := range resourceDTOs {
			schemaName := strings.ToUpper(resource.Name[:1]) + resource.Name[1:]
			base := "/" + resource.PluralName

			resourcePaths[base] = true
			resourcePaths[base+"/:id"] = true

			paths[base] = buildCollectionEndpoints(resource, schemaName, cfg)
			paths[base+"/{id}"] = buildItemEndpoints(resource, schemaName)
		}
	}

	discoveredRoutes := discoverNonResourceRoutes(app, resourcePaths)
	for path, methods := range discoveredRoutes {
		paths[path] = methods
	}

	components["securitySchemes"] = map[string]interface{}{
		"bearerAuth": map[string]interface{}{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
			"description":  "JWT authentication token",
		},
	}

	return map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       cfg.Title,
			"version":     cfg.Version,
			"description": cfg.Description,
		},
		"servers": []map[string]string{
			{"url": cfg.ServerURL, "description": "Development server"},
		},
		"paths":      paths,
		"components": components,
		"security": []map[string]interface{}{
			{"bearerAuth": []string{}},
		},
	}, nil
}

func buildCollectionEndpoints(resource resourceDTOs, schemaName string, cfg GeneratorConfig) map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List " + resource.PluralName,
			"description": "Retrieve a list of " + resource.PluralName,
			"tags":        []string{schemaName},
			"parameters": []map[string]interface{}{
				{
					"name":        "limit",
					"in":          "query",
					"description": fmt.Sprintf("Maximum number of items to return (default: %d, max: %d)", cfg.PaginationLimit, cfg.PaginationMaxLimit),
					"schema":      map[string]interface{}{"type": "integer", "default": cfg.PaginationLimit, "maximum": cfg.PaginationMaxLimit},
				},
				{
					"name":        "offset",
					"in":          "query",
					"description": "Number of items to skip (default: 0)",
					"schema":      map[string]interface{}{"type": "integer", "default": 0, "minimum": 0},
				},
				{
					"name":        "count",
					"in":          "query",
					"description": "Include total count in response (adds hydra:totalItems field)",
					"schema":      map[string]interface{}{"type": "boolean", "default": false},
				},
				{
					"name":        "expand",
					"in":          "query",
					"description": "Comma-separated list of relations to expand",
					"schema":      map[string]string{"type": "string"},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Hydra paginated collection",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"@context":         map[string]string{"type": "string"},
									"@id":              map[string]string{"type": "string"},
									"@type":            map[string]string{"type": "string", "example": "hydra:Collection"},
									"hydra:totalItems": map[string]interface{}{"type": "integer", "description": "Total count (only present if count=true)"},
									"hydra:member": map[string]interface{}{
										"type": "array",
										"items": map[string]string{
											"$ref": "#/components/schemas/" + schemaName,
										},
									},
									"hydra:view": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"@id":            map[string]string{"type": "string"},
											"@type":          map[string]string{"type": "string"},
											"hydra:first":    map[string]string{"type": "string"},
											"hydra:last":     map[string]string{"type": "string"},
											"hydra:previous": map[string]string{"type": "string"},
											"hydra:next":     map[string]string{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"post": map[string]interface{}{
			"summary":     "Create " + resource.Name,
			"description": "Create a new " + resource.Name,
			"tags":        []string{schemaName},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]string{
							"$ref": "#/components/schemas/" + schemaName,
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"201": map[string]interface{}{
					"description": "Successfully created",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]string{
								"$ref": "#/components/schemas/" + schemaName,
							},
						},
					},
				},
			},
		},
	}
}

func buildItemEndpoints(resource resourceDTOs, schemaName string) map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get " + resource.Name + " by ID",
			"description": "Retrieve a single " + resource.Name + " by ID",
			"tags":        []string{schemaName},
			"parameters": []map[string]interface{}{
				{
					"name":        "id",
					"in":          "path",
					"required":    true,
					"description": "Resource ID",
					"schema":      map[string]string{"type": "string"},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]string{
								"$ref": "#/components/schemas/" + schemaName,
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "Resource not found",
				},
			},
		},
		"put": map[string]interface{}{
			"summary":     "Update " + resource.Name + " by ID",
			"description": "Update an existing " + resource.Name,
			"tags":        []string{schemaName},
			"parameters": []map[string]interface{}{
				{
					"name":        "id",
					"in":          "path",
					"required":    true,
					"description": "Resource ID",
					"schema":      map[string]string{"type": "string"},
				},
			},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]string{
							"$ref": "#/components/schemas/" + schemaName,
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successfully updated",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]string{
								"$ref": "#/components/schemas/" + schemaName,
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "Resource not found",
				},
			},
		},
		"delete": map[string]interface{}{
			"summary":     "Delete " + resource.Name + " by ID",
			"description": "Delete an existing " + resource.Name,
			"tags":        []string{schemaName},
			"parameters": []map[string]interface{}{
				{
					"name":        "id",
					"in":          "path",
					"required":    true,
					"description": "Resource ID",
					"schema":      map[string]string{"type": "string"},
				},
			},
			"responses": map[string]interface{}{
				"204": map[string]interface{}{
					"description": "Successfully deleted",
				},
				"404": map[string]interface{}{
					"description": "Resource not found",
				},
			},
		},
	}
}

func buildCollectionEndpointsFromResource(resource plugin.OpenAPIResource, schemaName string, cfg GeneratorConfig) map[string]interface{} {
	tags := resource.Tags
	if len(tags) == 0 {
		tags = []string{schemaName}
	}

	description := resource.Description
	if description == "" {
		description = "Retrieve a list of " + resource.PluralName
	}

	endpoints := map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List " + resource.PluralName,
			"description": description,
			"tags":        tags,
			"parameters": []map[string]interface{}{
				{
					"name":        "limit",
					"in":          "query",
					"description": fmt.Sprintf("Maximum number of items to return (default: %d, max: %d)", cfg.PaginationLimit, cfg.PaginationMaxLimit),
					"schema":      map[string]interface{}{"type": "integer", "default": cfg.PaginationLimit, "maximum": cfg.PaginationMaxLimit},
				},
				{
					"name":        "offset",
					"in":          "query",
					"description": "Number of items to skip (default: 0)",
					"schema":      map[string]interface{}{"type": "integer", "default": 0, "minimum": 0},
				},
				{
					"name":        "count",
					"in":          "query",
					"description": "Include total count in response (adds hydra:totalItems field)",
					"schema":      map[string]interface{}{"type": "boolean", "default": false},
				},
				{
					"name":        "expand",
					"in":          "query",
					"description": "Comma-separated list of relations to expand",
					"schema":      map[string]string{"type": "string"},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Hydra paginated collection",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"@context":         map[string]string{"type": "string"},
									"@id":              map[string]string{"type": "string"},
									"@type":            map[string]string{"type": "string", "example": "hydra:Collection"},
									"hydra:totalItems": map[string]interface{}{"type": "integer", "description": "Total count (only present if count=true)"},
									"hydra:member": map[string]interface{}{
										"type": "array",
										"items": map[string]string{
											"$ref": "#/components/schemas/" + schemaName,
										},
									},
									"hydra:view": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"@id":            map[string]string{"type": "string"},
											"@type":          map[string]string{"type": "string"},
											"hydra:first":    map[string]string{"type": "string"},
											"hydra:last":     map[string]string{"type": "string"},
											"hydra:previous": map[string]string{"type": "string"},
											"hydra:next":     map[string]string{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if resource.CreateModel != nil {
		createSchemaRef := "Create" + schemaName + "Request"
		endpoints["post"] = map[string]interface{}{
			"summary":     "Create " + resource.Name,
			"description": "Create a new " + resource.Name,
			"tags":        tags,
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]string{
							"$ref": "#/components/schemas/" + createSchemaRef,
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"201": map[string]interface{}{
					"description": "Successfully created",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]string{
								"$ref": "#/components/schemas/" + schemaName,
							},
						},
					},
				},
			},
		}
	}

	return endpoints
}

func buildItemEndpointsFromResource(resource plugin.OpenAPIResource, schemaName string) map[string]interface{} {
	tags := resource.Tags
	if len(tags) == 0 {
		tags = []string{schemaName}
	}

	endpoints := map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get " + resource.Name + " by ID",
			"description": "Retrieve a single " + resource.Name + " by ID",
			"tags":        tags,
			"parameters": []map[string]interface{}{
				{
					"name":        "id",
					"in":          "path",
					"required":    true,
					"description": "Resource ID",
					"schema":      map[string]string{"type": "string"},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]string{
								"$ref": "#/components/schemas/" + schemaName,
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "Resource not found",
				},
			},
		},
		"delete": map[string]interface{}{
			"summary":     "Delete " + resource.Name + " by ID",
			"description": "Delete an existing " + resource.Name,
			"tags":        tags,
			"parameters": []map[string]interface{}{
				{
					"name":        "id",
					"in":          "path",
					"required":    true,
					"description": "Resource ID",
					"schema":      map[string]string{"type": "string"},
				},
			},
			"responses": map[string]interface{}{
				"204": map[string]interface{}{
					"description": "Successfully deleted",
				},
				"404": map[string]interface{}{
					"description": "Resource not found",
				},
			},
		},
	}

	if resource.UpdateModel != nil {
		updateSchemaRef := "Update" + schemaName + "Request"
		endpoints["put"] = map[string]interface{}{
			"summary":     "Update " + resource.Name + " by ID",
			"description": "Update an existing " + resource.Name,
			"tags":        tags,
			"parameters": []map[string]interface{}{
				{
					"name":        "id",
					"in":          "path",
					"required":    true,
					"description": "Resource ID",
					"schema":      map[string]string{"type": "string"},
				},
			},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]string{
							"$ref": "#/components/schemas/" + updateSchemaRef,
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successfully updated",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]string{
								"$ref": "#/components/schemas/" + schemaName,
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "Resource not found",
				},
			},
		}
	}

	return endpoints
}
