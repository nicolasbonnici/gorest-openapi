package openapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestNewPlugin(t *testing.T) {
	plugin := NewPlugin()
	if plugin == nil {
		t.Fatal("NewPlugin() returned nil")
	}

	openapiPlugin, ok := plugin.(*OpenAPIPlugin)
	if !ok {
		t.Fatal("NewPlugin() did not return *OpenAPIPlugin")
	}

	if openapiPlugin.Name() != "openapi" {
		t.Errorf("plugin.Name() = %v, want 'openapi'", openapiPlugin.Name())
	}
}

func TestOpenAPIPlugin_Name(t *testing.T) {
	plugin := &OpenAPIPlugin{}
	if got := plugin.Name(); got != "openapi" {
		t.Errorf("Name() = %v, want 'openapi'", got)
	}
}

func TestOpenAPIPlugin_Initialize(t *testing.T) {
	tests := []struct {
		name    string
		cfg     map[string]interface{}
		wantErr bool
		validate func(t *testing.T, p *OpenAPIPlugin)
	}{
		{
			name: "initialize with all config values",
			cfg: map[string]interface{}{
				"pagination_limit":     30,
				"pagination_max_limit": 200,
				"dtos_directory":       "/path/to/dtos",
				"title":                "Custom API",
				"version":              "2.0.0",
				"description":          "Custom Description",
			},
			wantErr: false,
			validate: func(t *testing.T, p *OpenAPIPlugin) {
				if p.paginationLimit != 30 {
					t.Errorf("paginationLimit = %v, want 30", p.paginationLimit)
				}
				if p.paginationMaxLimit != 200 {
					t.Errorf("paginationMaxLimit = %v, want 200", p.paginationMaxLimit)
				}
				if p.dtosDirectory != "/path/to/dtos" {
					t.Errorf("dtosDirectory = %v, want '/path/to/dtos'", p.dtosDirectory)
				}
				if p.title != "Custom API" {
					t.Errorf("title = %v, want 'Custom API'", p.title)
				}
				if p.version != "2.0.0" {
					t.Errorf("version = %v, want '2.0.0'", p.version)
				}
				if p.description != "Custom Description" {
					t.Errorf("description = %v, want 'Custom Description'", p.description)
				}
			},
		},
		{
			name: "initialize with minimal config (defaults applied)",
			cfg: map[string]interface{}{
				"dtos_directory": "/path/to/dtos",
			},
			wantErr: false,
			validate: func(t *testing.T, p *OpenAPIPlugin) {
				if p.dtosDirectory != "/path/to/dtos" {
					t.Errorf("dtosDirectory = %v, want '/path/to/dtos'", p.dtosDirectory)
				}
				if p.title != "GoREST API" {
					t.Errorf("title = %v, want 'GoREST API' (default)", p.title)
				}
				if p.version != "1.0.0" {
					t.Errorf("version = %v, want '1.0.0' (default)", p.version)
				}
				if p.description != "Auto-generated REST API with full CRUD operations" {
					t.Errorf("description = %v, want default description", p.description)
				}
			},
		},
		{
			name: "error when dtos_directory is missing",
			cfg: map[string]interface{}{
				"title": "Test API",
			},
			wantErr: true,
			validate: func(t *testing.T, p *OpenAPIPlugin) {},
		},
		{
			name: "ignore invalid config types",
			cfg: map[string]interface{}{
				"dtos_directory":       "/path/to/dtos",
				"pagination_limit":     "not_an_int",
				"pagination_max_limit": "not_an_int",
				"title":                123,
				"version":              true,
			},
			wantErr: false,
			validate: func(t *testing.T, p *OpenAPIPlugin) {
				// Should use defaults when types are wrong
				if p.paginationLimit != 0 {
					t.Errorf("paginationLimit = %v, want 0 (invalid type ignored)", p.paginationLimit)
				}
				if p.title != "GoREST API" {
					t.Errorf("title = %v, want 'GoREST API' (invalid type, use default)", p.title)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &OpenAPIPlugin{}
			err := plugin.Initialize(tt.cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, plugin)
			}
		})
	}
}

func TestOpenAPIPlugin_Handler(t *testing.T) {
	plugin := &OpenAPIPlugin{}
	handler := plugin.Handler()

	if handler == nil {
		t.Fatal("Handler() returned nil")
	}

	// Test that it's a no-op middleware (calls c.Next())
	app := fiber.New()
	app.Use(handler)
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("test response")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Status code = %v, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "test response" {
		t.Errorf("Response body = %v, want 'test response'", string(body))
	}
}

func TestOpenAPIPlugin_SetupEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) (*OpenAPIPlugin, *fiber.App)
		testPath string
		validate func(t *testing.T, resp *http.Response)
	}{
		{
			name: "/openapi endpoint returns HTML",
			setup: func(t *testing.T) (*OpenAPIPlugin, *fiber.App) {
				tempDir := t.TempDir()
				plugin := &OpenAPIPlugin{
					dtosDirectory:      tempDir,
					paginationLimit:    20,
					paginationMaxLimit: 100,
					title:              "Test API",
					version:            "1.0.0",
					description:        "Test Description",
				}
				app := fiber.New()
				return plugin, app
			},
			testPath: "/openapi",
			validate: func(t *testing.T, resp *http.Response) {
				if resp.StatusCode != 200 {
					t.Errorf("Status code = %v, want 200", resp.StatusCode)
				}

				contentType := resp.Header.Get("Content-Type")
				if contentType != "text/html" {
					t.Errorf("Content-Type = %v, want 'text/html'", contentType)
				}

				body, _ := io.ReadAll(resp.Body)
				bodyStr := string(body)

				if !strings.Contains(bodyStr, "<!DOCTYPE html>") {
					t.Error("Response should contain HTML doctype")
				}

				if !strings.Contains(bodyStr, "GoREST API Documentation") {
					t.Error("Response should contain page title")
				}

				if !strings.Contains(bodyStr, "@scalar/api-reference") {
					t.Error("Response should contain Scalar API reference script")
				}

				// Validate CSP header is set
				csp := resp.Header.Get("Content-Security-Policy")
				if csp == "" {
					t.Error("Content-Security-Policy header should be set")
				}
				if !strings.Contains(csp, "https://cdn.jsdelivr.net") {
					t.Error("CSP should allow cdn.jsdelivr.net")
				}
			},
		},
		{
			name: "/openapi.json endpoint returns valid JSON spec",
			setup: func(t *testing.T) (*OpenAPIPlugin, *fiber.App) {
				tempDir := t.TempDir()

				// Create a test DTO
				dtoContent := `package dto

type UserDTO struct {
	ID   int64  ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}`
				err := os.WriteFile(filepath.Join(tempDir, "user.go"), []byte(dtoContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test DTO: %v", err)
				}

				plugin := &OpenAPIPlugin{
					dtosDirectory:      tempDir,
					paginationLimit:    20,
					paginationMaxLimit: 100,
					title:              "Test API",
					version:            "1.0.0",
					description:        "Test Description",
				}
				app := fiber.New()
				return plugin, app
			},
			testPath: "/openapi.json",
			validate: func(t *testing.T, resp *http.Response) {
				if resp.StatusCode != 200 {
					t.Errorf("Status code = %v, want 200", resp.StatusCode)
				}

				body, _ := io.ReadAll(resp.Body)
				var spec map[string]interface{}
				err := json.Unmarshal(body, &spec)
				if err != nil {
					t.Fatalf("Failed to parse JSON response: %v", err)
				}

				// Validate OpenAPI spec structure
				if spec["openapi"] != "3.0.0" {
					t.Errorf("openapi version = %v, want '3.0.0'", spec["openapi"])
				}

				info, ok := spec["info"].(map[string]interface{})
				if !ok {
					t.Fatal("spec missing info")
				}
				if info["title"] != "Test API" {
					t.Errorf("info.title = %v, want 'Test API'", info["title"])
				}

				paths, ok := spec["paths"].(map[string]interface{})
				if !ok {
					t.Fatal("spec missing paths")
				}
				if len(paths) == 0 {
					t.Error("paths should not be empty")
				}

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
			},
		},
		{
			name: "/openapi.json returns error when DTOs directory invalid",
			setup: func(t *testing.T) (*OpenAPIPlugin, *fiber.App) {
				plugin := &OpenAPIPlugin{
					dtosDirectory:      "/non/existent/directory",
					paginationLimit:    20,
					paginationMaxLimit: 100,
					title:              "Test API",
					version:            "1.0.0",
					description:        "Test Description",
				}
				app := fiber.New()
				return plugin, app
			},
			testPath: "/openapi.json",
			validate: func(t *testing.T, resp *http.Response) {
				if resp.StatusCode != 500 {
					t.Errorf("Status code = %v, want 500", resp.StatusCode)
				}

				body, _ := io.ReadAll(resp.Body)
				var errorResp map[string]interface{}
				err := json.Unmarshal(body, &errorResp)
				if err != nil {
					t.Fatalf("Failed to parse error JSON response: %v", err)
				}

				if _, exists := errorResp["error"]; !exists {
					t.Error("Error response should contain 'error' field")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, app := tt.setup(t)

			err := plugin.SetupEndpoints(app)
			if err != nil {
				t.Fatalf("SetupEndpoints() error = %v", err)
			}

			req := httptest.NewRequest("GET", tt.testPath, nil)
			req.Host = "localhost:3000"
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Test request failed: %v", err)
			}

			tt.validate(t, resp)
		})
	}
}

func TestOpenAPIPlugin_SetupEndpoints_ProtocolDetection(t *testing.T) {
	tests := []struct {
		name         string
		protocol     string
		expectedHost string
	}{
		{
			name:         "HTTP protocol",
			protocol:     "http",
			expectedHost: "localhost",
		},
		{
			name:         "HTTPS protocol",
			protocol:     "https",
			expectedHost: "secure.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create a test DTO
			dtoContent := `package dto

type TestDTO struct {
	ID int64 ` + "`json:\"id\"`" + `
}`
			err := os.WriteFile(filepath.Join(tempDir, "test.go"), []byte(dtoContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test DTO: %v", err)
			}

			plugin := &OpenAPIPlugin{
				dtosDirectory:      tempDir,
				paginationLimit:    20,
				paginationMaxLimit: 100,
				title:              "Test API",
				version:            "1.0.0",
				description:        "Test Description",
			}

			app := fiber.New()
			err = plugin.SetupEndpoints(app)
			if err != nil {
				t.Fatalf("SetupEndpoints() error = %v", err)
			}

			// Make request
			req := httptest.NewRequest("GET", "/openapi.json", nil)
			req.Host = tt.expectedHost

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Test request failed: %v", err)
			}

			body, _ := io.ReadAll(resp.Body)
			var spec map[string]interface{}
			err = json.Unmarshal(body, &spec)
			if err != nil {
				t.Fatalf("Failed to parse JSON response: %v", err)
			}

			servers, ok := spec["servers"].([]interface{})
			if !ok || len(servers) == 0 {
				t.Fatal("spec missing servers")
			}

			server := servers[0].(map[string]interface{})
			serverURL := server["url"].(string)

			expectedPrefix := "http://"
			if !strings.HasPrefix(serverURL, expectedPrefix) {
				t.Errorf("Server URL = %v, should start with %v", serverURL, expectedPrefix)
			}

			if !strings.Contains(serverURL, tt.expectedHost) {
				t.Errorf("Server URL = %v, should contain host %v", serverURL, tt.expectedHost)
			}
		})
	}
}
