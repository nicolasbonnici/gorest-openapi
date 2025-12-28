package openapi

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestBuildCollectionEndpoints(t *testing.T) {
	resource := resourceDTOs{
		Name:       "user",
		PluralName: "users",
	}
	schemaName := "User"
	cfg := GeneratorConfig{
		PaginationLimit:    20,
		PaginationMaxLimit: 100,
	}

	got := buildCollectionEndpoints(resource, schemaName, cfg)

	// Validate GET endpoint
	getEndpoint, ok := got["get"].(map[string]interface{})
	if !ok {
		t.Fatal("buildCollectionEndpoints() missing GET endpoint")
	}

	if summary, ok := getEndpoint["summary"].(string); !ok || summary != "List users" {
		t.Errorf("GET summary = %v, want 'List users'", summary)
	}

	if tags, ok := getEndpoint["tags"].([]string); !ok || len(tags) != 1 || tags[0] != "User" {
		t.Errorf("GET tags = %v, want ['User']", tags)
	}

	// Validate GET parameters
	params, ok := getEndpoint["parameters"].([]map[string]interface{})
	if !ok {
		t.Fatal("GET endpoint missing parameters")
	}

	expectedParams := []string{"limit", "offset", "count", "expand"}
	if len(params) != len(expectedParams) {
		t.Errorf("GET parameters count = %d, want %d", len(params), len(expectedParams))
	}

	// Validate POST endpoint
	postEndpoint, ok := got["post"].(map[string]interface{})
	if !ok {
		t.Fatal("buildCollectionEndpoints() missing POST endpoint")
	}

	if summary, ok := postEndpoint["summary"].(string); !ok || summary != "Create user" {
		t.Errorf("POST summary = %v, want 'Create user'", summary)
	}

	requestBody, ok := postEndpoint["requestBody"].(map[string]interface{})
	if !ok {
		t.Fatal("POST endpoint missing requestBody")
	}

	if required, ok := requestBody["required"].(bool); !ok || !required {
		t.Error("POST requestBody should be required")
	}

	// Validate responses
	responses, ok := postEndpoint["responses"].(map[string]interface{})
	if !ok {
		t.Fatal("POST endpoint missing responses")
	}

	if _, ok := responses["201"]; !ok {
		t.Error("POST responses missing 201 status")
	}
}

func TestBuildItemEndpoints(t *testing.T) {
	resource := resourceDTOs{
		Name:       "user",
		PluralName: "users",
	}
	schemaName := "User"

	got := buildItemEndpoints(resource, schemaName)

	// Validate GET endpoint
	getEndpoint, ok := got["get"].(map[string]interface{})
	if !ok {
		t.Fatal("buildItemEndpoints() missing GET endpoint")
	}

	if summary, ok := getEndpoint["summary"].(string); !ok || summary != "Get user by ID" {
		t.Errorf("GET summary = %v, want 'Get user by ID'", summary)
	}

	// Validate GET parameters
	params, ok := getEndpoint["parameters"].([]map[string]interface{})
	if !ok || len(params) != 1 {
		t.Fatal("GET endpoint should have 1 parameter")
	}

	if params[0]["name"] != "id" {
		t.Errorf("GET parameter name = %v, want 'id'", params[0]["name"])
	}

	// Validate GET responses
	getResponses, ok := getEndpoint["responses"].(map[string]interface{})
	if !ok {
		t.Fatal("GET endpoint missing responses")
	}

	if _, ok := getResponses["200"]; !ok {
		t.Error("GET responses missing 200 status")
	}

	if _, ok := getResponses["404"]; !ok {
		t.Error("GET responses missing 404 status")
	}

	// Validate PUT endpoint
	putEndpoint, ok := got["put"].(map[string]interface{})
	if !ok {
		t.Fatal("buildItemEndpoints() missing PUT endpoint")
	}

	if summary, ok := putEndpoint["summary"].(string); !ok || summary != "Update user by ID" {
		t.Errorf("PUT summary = %v, want 'Update user by ID'", summary)
	}

	// Validate PUT requestBody
	requestBody, ok := putEndpoint["requestBody"].(map[string]interface{})
	if !ok {
		t.Fatal("PUT endpoint missing requestBody")
	}

	if required, ok := requestBody["required"].(bool); !ok || !required {
		t.Error("PUT requestBody should be required")
	}

	// Validate DELETE endpoint
	deleteEndpoint, ok := got["delete"].(map[string]interface{})
	if !ok {
		t.Fatal("buildItemEndpoints() missing DELETE endpoint")
	}

	if summary, ok := deleteEndpoint["summary"].(string); !ok || summary != "Delete user by ID" {
		t.Errorf("DELETE summary = %v, want 'Delete user by ID'", summary)
	}

	// Validate DELETE responses
	deleteResponses, ok := deleteEndpoint["responses"].(map[string]interface{})
	if !ok {
		t.Fatal("DELETE endpoint missing responses")
	}

	if _, ok := deleteResponses["204"]; !ok {
		t.Error("DELETE responses missing 204 status")
	}

	if _, ok := deleteResponses["404"]; !ok {
		t.Error("DELETE responses missing 404 status")
	}
}

func TestGenerateOpenAPISpec(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) (string, *fiber.App, GeneratorConfig)
		validate  func(t *testing.T, spec map[string]interface{})
		wantErr   bool
	}{
		{
			name: "generate spec with DTOs",
			setupFunc: func(t *testing.T) (string, *fiber.App, GeneratorConfig) {
				tempDir := t.TempDir()

				// Create user.go
				userContent := `package dto

type UserDTO struct {
	ID    int64  ` + "`json:\"id\"`" + `
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}`
				err := os.WriteFile(filepath.Join(tempDir, "user.go"), []byte(userContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create user.go: %v", err)
				}

				app := fiber.New()
				cfg := GeneratorConfig{
					DTOsDirectory:      tempDir,
					PaginationLimit:    20,
					PaginationMaxLimit: 100,
					ServerURL:          "http://localhost:3000",
					Title:              "Test API",
					Version:            "1.0.0",
					Description:        "Test Description",
				}

				return tempDir, app, cfg
			},
			validate: func(t *testing.T, spec map[string]interface{}) {
				// Validate top-level structure
				if version, ok := spec["openapi"].(string); !ok || version != "3.0.0" {
					t.Errorf("openapi version = %v, want '3.0.0'", version)
				}

				// Validate info
				info, ok := spec["info"].(map[string]interface{})
				if !ok {
					t.Fatal("spec missing info")
				}
				if info["title"] != "Test API" {
					t.Errorf("info.title = %v, want 'Test API'", info["title"])
				}
				if info["version"] != "1.0.0" {
					t.Errorf("info.version = %v, want '1.0.0'", info["version"])
				}

				// Validate servers
				servers, ok := spec["servers"].([]map[string]string)
				if !ok || len(servers) != 1 {
					t.Fatal("spec missing or invalid servers")
				}
				if servers[0]["url"] != "http://localhost:3000" {
					t.Errorf("server url = %v, want 'http://localhost:3000'", servers[0]["url"])
				}

				// Validate paths
				paths, ok := spec["paths"].(map[string]interface{})
				if !ok {
					t.Fatal("spec missing paths")
				}

				// Should have /users and /users/{id}
				if _, exists := paths["/users"]; !exists {
					t.Error("spec missing /users path")
				}
				if _, exists := paths["/users/{id}"]; !exists {
					t.Error("spec missing /users/{id} path")
				}

				// Validate components
				components, ok := spec["components"].(map[string]interface{})
				if !ok {
					t.Fatal("spec missing components")
				}

				// Validate schemas
				schemas, ok := components["schemas"].(map[string]interface{})
				if !ok {
					t.Fatal("components missing schemas")
				}

				userSchema, exists := schemas["User"]
				if !exists {
					t.Fatal("schemas missing User")
				}

				userSchemaMap, ok := userSchema.(map[string]interface{})
				if !ok {
					t.Fatal("User schema not a map")
				}

				if userSchemaMap["type"] != "object" {
					t.Errorf("User schema type = %v, want 'object'", userSchemaMap["type"])
				}

				properties, ok := userSchemaMap["properties"].(map[string]interface{})
				if !ok {
					t.Fatal("User schema missing properties")
				}

				if len(properties) != 3 {
					t.Errorf("User schema has %d properties, want 3", len(properties))
				}

				// Validate security schemes
				securitySchemes, ok := components["securitySchemes"].(map[string]interface{})
				if !ok {
					t.Fatal("components missing securitySchemes")
				}

				if _, exists := securitySchemes["bearerAuth"]; !exists {
					t.Error("securitySchemes missing bearerAuth")
				}

				// Validate security
				security, ok := spec["security"].([]map[string]interface{})
				if !ok || len(security) != 1 {
					t.Fatal("spec missing or invalid security")
				}
			},
			wantErr: false,
		},
		{
			name: "generate spec with multiple resources",
			setupFunc: func(t *testing.T) (string, *fiber.App, GeneratorConfig) {
				tempDir := t.TempDir()

				// Create user.go
				userContent := `package dto

type UserDTO struct {
	ID int64 ` + "`json:\"id\"`" + `
}`
				err := os.WriteFile(filepath.Join(tempDir, "user.go"), []byte(userContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create user.go: %v", err)
				}

				// Create product.go
				productContent := `package dto

type ProductDTO struct {
	ID int64 ` + "`json:\"id\"`" + `
}`
				err = os.WriteFile(filepath.Join(tempDir, "product.go"), []byte(productContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create product.go: %v", err)
				}

				app := fiber.New()
				cfg := GeneratorConfig{
					DTOsDirectory:      tempDir,
					PaginationLimit:    20,
					PaginationMaxLimit: 100,
					ServerURL:          "http://localhost:3000",
					Title:              "Test API",
					Version:            "1.0.0",
					Description:        "Test Description",
				}

				return tempDir, app, cfg
			},
			validate: func(t *testing.T, spec map[string]interface{}) {
				paths, ok := spec["paths"].(map[string]interface{})
				if !ok {
					t.Fatal("spec missing paths")
				}

				// Should have paths for both resources
				expectedPaths := []string{"/users", "/users/{id}", "/products", "/products/{id}"}
				for _, path := range expectedPaths {
					if _, exists := paths[path]; !exists {
						t.Errorf("spec missing path %q", path)
					}
				}

				// Validate schemas
				components, ok := spec["components"].(map[string]interface{})
				if !ok {
					t.Fatal("spec missing components")
				}

				schemas, ok := components["schemas"].(map[string]interface{})
				if !ok {
					t.Fatal("components missing schemas")
				}

				if _, exists := schemas["User"]; !exists {
					t.Error("schemas missing User")
				}
				if _, exists := schemas["Product"]; !exists {
					t.Error("schemas missing Product")
				}
			},
			wantErr: false,
		},
		{
			name: "error when DTOs directory doesn't exist",
			setupFunc: func(t *testing.T) (string, *fiber.App, GeneratorConfig) {
				app := fiber.New()
				cfg := GeneratorConfig{
					DTOsDirectory:      "/non/existent/directory",
					PaginationLimit:    20,
					PaginationMaxLimit: 100,
					ServerURL:          "http://localhost:3000",
					Title:              "Test API",
					Version:            "1.0.0",
					Description:        "Test Description",
				}

				return "", app, cfg
			},
			validate: func(t *testing.T, spec map[string]interface{}) {},
			wantErr:  true,
		},
		{
			name: "generate spec with discovered routes",
			setupFunc: func(t *testing.T) (string, *fiber.App, GeneratorConfig) {
				tempDir := t.TempDir()

				// Create minimal DTO
				userContent := `package dto

type UserDTO struct {
	ID int64 ` + "`json:\"id\"`" + `
}`
				err := os.WriteFile(filepath.Join(tempDir, "user.go"), []byte(userContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create user.go: %v", err)
				}

				app := fiber.New()
				// Add custom routes that should be discovered
				app.Get("/health", func(c *fiber.Ctx) error { return c.SendString("OK") })
				app.Post("/auth/login", func(c *fiber.Ctx) error { return c.SendString("login") })

				cfg := GeneratorConfig{
					DTOsDirectory:      tempDir,
					PaginationLimit:    20,
					PaginationMaxLimit: 100,
					ServerURL:          "http://localhost:3000",
					Title:              "Test API",
					Version:            "1.0.0",
					Description:        "Test Description",
				}

				return tempDir, app, cfg
			},
			validate: func(t *testing.T, spec map[string]interface{}) {
				paths, ok := spec["paths"].(map[string]interface{})
				if !ok {
					t.Fatal("spec missing paths")
				}

				// Should have discovered custom routes
				if _, exists := paths["/health"]; !exists {
					t.Error("spec missing discovered /health path")
				}
				if _, exists := paths["/auth/login"]; !exists {
					t.Error("spec missing discovered /auth/login path")
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, app, cfg := tt.setupFunc(t)

			spec, err := generateOpenAPISpec(app, cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateOpenAPISpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, spec)
			}
		})
	}
}

func TestGeneratorConfig(t *testing.T) {
	cfg := GeneratorConfig{
		DTOsDirectory:      "/path/to/dtos",
		PaginationLimit:    25,
		PaginationMaxLimit: 200,
		ServerURL:          "https://api.example.com",
		Title:              "My API",
		Version:            "2.0.0",
		Description:        "API Description",
	}

	// Validate all fields are set correctly
	if cfg.DTOsDirectory != "/path/to/dtos" {
		t.Errorf("DTOsDirectory = %v, want '/path/to/dtos'", cfg.DTOsDirectory)
	}
	if cfg.PaginationLimit != 25 {
		t.Errorf("PaginationLimit = %v, want 25", cfg.PaginationLimit)
	}
	if cfg.PaginationMaxLimit != 200 {
		t.Errorf("PaginationMaxLimit = %v, want 200", cfg.PaginationMaxLimit)
	}
	if cfg.ServerURL != "https://api.example.com" {
		t.Errorf("ServerURL = %v, want 'https://api.example.com'", cfg.ServerURL)
	}
	if cfg.Title != "My API" {
		t.Errorf("Title = %v, want 'My API'", cfg.Title)
	}
	if cfg.Version != "2.0.0" {
		t.Errorf("Version = %v, want '2.0.0'", cfg.Version)
	}
	if cfg.Description != "API Description" {
		t.Errorf("Description = %v, want 'API Description'", cfg.Description)
	}
}

func TestBuildCollectionEndpointsWithDifferentConfigs(t *testing.T) {
	tests := []struct {
		name string
		cfg  GeneratorConfig
	}{
		{
			name: "default pagination config",
			cfg: GeneratorConfig{
				PaginationLimit:    20,
				PaginationMaxLimit: 100,
			},
		},
		{
			name: "custom pagination config",
			cfg: GeneratorConfig{
				PaginationLimit:    50,
				PaginationMaxLimit: 500,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := resourceDTOs{Name: "item", PluralName: "items"}
			endpoints := buildCollectionEndpoints(resource, "Item", tt.cfg)

			getEndpoint := endpoints["get"].(map[string]interface{})
			params := getEndpoint["parameters"].([]map[string]interface{})

			// Find limit parameter and validate
			for _, param := range params {
				if param["name"] == "limit" {
					schema := param["schema"].(map[string]interface{})
					if !reflect.DeepEqual(schema["default"], tt.cfg.PaginationLimit) {
						t.Errorf("limit default = %v, want %v", schema["default"], tt.cfg.PaginationLimit)
					}
					if !reflect.DeepEqual(schema["maximum"], tt.cfg.PaginationMaxLimit) {
						t.Errorf("limit maximum = %v, want %v", schema["maximum"], tt.cfg.PaginationMaxLimit)
					}
				}
			}
		})
	}
}
