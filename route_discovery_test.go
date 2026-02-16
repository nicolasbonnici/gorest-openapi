package openapi

import (
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestShouldSkipRoute(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		resourcePaths map[string]bool
		want          bool
	}{
		{
			name:          "skip OpenAPI UI route",
			path:          "/openapi",
			resourcePaths: map[string]bool{},
			want:          true,
		},
		{
			name:          "skip OpenAPI JSON route",
			path:          "/openapi.json",
			resourcePaths: map[string]bool{},
			want:          true,
		},
		{
			name:          "skip resource paths",
			path:          "/users",
			resourcePaths: map[string]bool{"/users": true},
			want:          true,
		},
		{
			name:          "skip root path",
			path:          "/",
			resourcePaths: map[string]bool{},
			want:          true,
		},
		{
			name:          "skip empty path",
			path:          "",
			resourcePaths: map[string]bool{},
			want:          true,
		},
		{
			name:          "do not skip custom route",
			path:          "/auth/login",
			resourcePaths: map[string]bool{},
			want:          false,
		},
		{
			name:          "do not skip health check",
			path:          "/health",
			resourcePaths: map[string]bool{},
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSkipRoute(tt.path, tt.resourcePaths); got != tt.want {
				t.Errorf("shouldSkipRoute(%q, %v) = %v, want %v", tt.path, tt.resourcePaths, got, tt.want)
			}
		})
	}
}

func TestDetermineTag(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "auth path returns Authentication",
			path: "/auth/login",
			want: "Authentication",
		},
		{
			name: "health path returns System",
			path: "/health",
			want: "System",
		},
		{
			name: "custom path capitalizes first segment",
			path: "/users/profile",
			want: "Users",
		},
		{
			name: "single segment path",
			path: "/api",
			want: "Api",
		},
		{
			name: "empty path returns General",
			path: "",
			want: "General",
		},
		{
			name: "root path returns General",
			path: "/",
			want: "General",
		},
		{
			name: "nested path uses first segment",
			path: "/api/v1/users",
			want: "Api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineTag(tt.path); got != tt.want {
				t.Errorf("determineTag(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		method string
		want   string
	}{
		{
			name:   "GET method",
			path:   "/users",
			method: "GET",
			want:   "Get users",
		},
		{
			name:   "POST method",
			path:   "/auth/login",
			method: "POST",
			want:   "Create or execute auth login",
		},
		{
			name:   "PUT method",
			path:   "/users/:id",
			method: "PUT",
			want:   "Update users id",
		},
		{
			name:   "PATCH method",
			path:   "/products/:id",
			method: "PATCH",
			want:   "Partially update products id",
		},
		{
			name:   "DELETE method",
			path:   "/items/:id",
			method: "DELETE",
			want:   "Delete items id",
		},
		{
			name:   "unknown method",
			path:   "/custom",
			method: "OPTIONS",
			want:   "OPTIONS custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateSummary(tt.path, tt.method); got != tt.want {
				t.Errorf("generateSummary(%q, %q) = %v, want %v", tt.path, tt.method, got, tt.want)
			}
		})
	}
}

func TestGenerateDescription(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		method string
		want   string
	}{
		{
			name:   "simple path",
			path:   "/users",
			method: "GET",
			want:   "GET /users",
		},
		{
			name:   "path with parameter",
			path:   "/users/:id",
			method: "PUT",
			want:   "PUT /users/:id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateDescription(tt.path, tt.method); got != tt.want {
				t.Errorf("generateDescription(%q, %q) = %v, want %v", tt.path, tt.method, got, tt.want)
			}
		})
	}
}

func TestExtractPathParameters(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []map[string]interface{}
	}{
		{
			name: "single parameter",
			path: "/users/:id",
			want: []map[string]interface{}{
				{
					"name":        "id",
					"in":          "path",
					"required":    true,
					"description": "Path parameter: id",
					"schema":      map[string]string{"type": "string"},
				},
			},
		},
		{
			name: "multiple parameters",
			path: "/users/:userId/posts/:postId",
			want: []map[string]interface{}{
				{
					"name":        "userId",
					"in":          "path",
					"required":    true,
					"description": "Path parameter: userId",
					"schema":      map[string]string{"type": "string"},
				},
				{
					"name":        "postId",
					"in":          "path",
					"required":    true,
					"description": "Path parameter: postId",
					"schema":      map[string]string{"type": "string"},
				},
			},
		},
		{
			name: "no parameters",
			path: "/users",
			want: []map[string]interface{}(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPathParameters(tt.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractPathParameters(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestGenerateRequestBody(t *testing.T) {
	want := map[string]interface{}{
		"required": true,
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"type": "object",
				},
			},
		},
	}

	got := generateRequestBody()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("generateRequestBody() = %v, want %v", got, want)
	}
}

func TestGenerateResponses(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   map[string]interface{}
	}{
		{
			name:   "GET method responses",
			method: "GET",
			want: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
							},
						},
					},
				},
			},
		},
		{
			name:   "POST method responses",
			method: "POST",
			want: map[string]interface{}{
				"201": map[string]interface{}{
					"description": "Successfully created",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
							},
						},
					},
				},
				"400": map[string]interface{}{
					"description": "Bad request",
				},
			},
		},
		{
			name:   "PUT method responses",
			method: "PUT",
			want: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successfully updated",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "Not found",
				},
			},
		},
		{
			name:   "PATCH method responses",
			method: "PATCH",
			want: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successfully updated",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
							},
						},
					},
				},
				"404": map[string]interface{}{
					"description": "Not found",
				},
			},
		},
		{
			name:   "DELETE method responses",
			method: "DELETE",
			want: map[string]interface{}{
				"204": map[string]interface{}{
					"description": "Successfully deleted",
				},
				"404": map[string]interface{}{
					"description": "Not found",
				},
			},
		},
		{
			name:   "unknown method responses",
			method: "OPTIONS",
			want: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateResponses(tt.method)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateResponses(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestGenerateRouteSpec(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		method string
	}{
		{
			name:   "GET route without parameters",
			path:   "/health",
			method: "GET",
		},
		{
			name:   "GET route with parameters",
			path:   "/users/:id",
			method: "GET",
		},
		{
			name:   "POST route",
			path:   "/auth/login",
			method: "POST",
		},
		{
			name:   "PUT route with parameters",
			path:   "/products/:id",
			method: "PUT",
		},
		{
			name:   "DELETE route with parameters",
			path:   "/items/:id",
			method: "DELETE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateRouteSpec(tt.path, tt.method)

			// Validate basic structure
			if _, ok := got["summary"]; !ok {
				t.Error("generateRouteSpec() missing 'summary' field")
			}
			if _, ok := got["description"]; !ok {
				t.Error("generateRouteSpec() missing 'description' field")
			}
			if _, ok := got["tags"]; !ok {
				t.Error("generateRouteSpec() missing 'tags' field")
			}
			if _, ok := got["responses"]; !ok {
				t.Error("generateRouteSpec() missing 'responses' field")
			}

			// Validate parameters are included when path has parameters
			if _, hasParams := got["parameters"]; hasParams != (tt.path != "/health" && tt.path != "/auth/login") {
				t.Errorf("generateRouteSpec() parameters presence incorrect for path %q", tt.path)
			}

			// Validate requestBody is included for POST/PUT/PATCH
			shouldHaveBody := tt.method == "POST" || tt.method == "PUT" || tt.method == "PATCH"
			if _, hasBody := got["requestBody"]; hasBody != shouldHaveBody {
				t.Errorf("generateRouteSpec() requestBody presence incorrect for method %q", tt.method)
			}
		})
	}
}

func TestDiscoverNonResourceRoutes(t *testing.T) {
	tests := []struct {
		name          string
		setupRoutes   func(app *fiber.App)
		resourcePaths map[string]bool
		wantPaths     []string
		skipPaths     []string
	}{
		{
			name: "discover custom routes and skip resource routes",
			setupRoutes: func(app *fiber.App) {
				app.Get("/users", func(c *fiber.Ctx) error { return nil })
				app.Get("/health", func(c *fiber.Ctx) error { return nil })
				app.Post("/auth/login", func(c *fiber.Ctx) error { return nil })
			},
			resourcePaths: map[string]bool{"/users": true},
			wantPaths:     []string{"/health", "/auth/login"},
			skipPaths:     []string{"/users"},
		},
		{
			name: "skip OpenAPI routes",
			setupRoutes: func(app *fiber.App) {
				app.Get("/openapi", func(c *fiber.Ctx) error { return nil })
				app.Get("/openapi.json", func(c *fiber.Ctx) error { return nil })
				app.Get("/custom", func(c *fiber.Ctx) error { return nil })
			},
			resourcePaths: map[string]bool{},
			wantPaths:     []string{"/custom"},
			skipPaths:     []string{"/openapi", "/openapi.json"},
		},
		{
			name: "discover routes with path parameters",
			setupRoutes: func(app *fiber.App) {
				app.Get("/api/users/:id", func(c *fiber.Ctx) error { return nil })
				app.Put("/api/users/:id", func(c *fiber.Ctx) error { return nil })
			},
			resourcePaths: map[string]bool{},
			wantPaths:     []string{"/api/users/:id"},
			skipPaths:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			tt.setupRoutes(app)

			got := discoverNonResourceRoutes(app, tt.resourcePaths)

			// Check wanted paths are present
			for _, wantPath := range tt.wantPaths {
				if _, exists := got[wantPath]; !exists {
					t.Errorf("discoverNonResourceRoutes() missing path %q", wantPath)
				}
			}

			// Check skipped paths are not present
			for _, skipPath := range tt.skipPaths {
				if _, exists := got[skipPath]; exists {
					t.Errorf("discoverNonResourceRoutes() should skip path %q", skipPath)
				}
			}

			// Validate structure of discovered routes
			for path, methodsMap := range got {
				for method, methodSpec := range methodsMap {
					specMap, ok := methodSpec.(map[string]interface{})
					if !ok {
						t.Errorf("discoverNonResourceRoutes() path %q method %q spec not a map", path, method)
						continue
					}

					// Validate required fields
					if _, ok := specMap["summary"]; !ok {
						t.Errorf("discoverNonResourceRoutes() path %q method %q missing summary", path, method)
					}
					if _, ok := specMap["responses"]; !ok {
						t.Errorf("discoverNonResourceRoutes() path %q method %q missing responses", path, method)
					}
				}
			}
		})
	}
}
