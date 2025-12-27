package openapi

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type GeneratorConfig struct {
	DTOsDirectory      string
	PaginationLimit    int
	PaginationMaxLimit int
}

func generateOpenAPISpec(app *fiber.App, cfg GeneratorConfig) (map[string]interface{}, error) {
	resourceDTOs, err := loadResourceDTOs(cfg.DTOsDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to load DTOs: %w", err)
	}

	paths := map[string]interface{}{}
	components := map[string]interface{}{
		"schemas": make(map[string]interface{}),
	}

	resourcePaths := make(map[string]bool)

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
			"title":       "GoREST API",
			"version":     "1.0.0",
			"description": "Auto-generated REST API with full CRUD operations",
		},
		"servers": []map[string]string{
			{"url": "http://localhost:3000", "description": "Development server"},
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
