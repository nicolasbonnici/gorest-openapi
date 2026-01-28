package openapi

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/logger"
	"github.com/nicolasbonnici/gorest/plugin"
)

type OpenAPIPlugin struct {
	paginationLimit    int
	paginationMaxLimit int
	dtosDirectory      string
	pluginRegistry     *plugin.PluginRegistry
	title              string
	version            string
	description        string
}

func NewPlugin() plugin.Plugin {
	return &OpenAPIPlugin{}
}

func (p *OpenAPIPlugin) Name() string {
	return "openapi"
}

func (p *OpenAPIPlugin) Initialize(cfg map[string]interface{}) error {
	if dtosDir, ok := cfg["dtos_directory"].(string); ok {
		p.dtosDirectory = dtosDir
	}

	if limit, ok := cfg["pagination_limit"].(int); ok {
		p.paginationLimit = limit
	}
	if maxLimit, ok := cfg["pagination_max_limit"].(int); ok {
		p.paginationMaxLimit = maxLimit
	}

	if registry, ok := cfg["plugin_registry"].(*plugin.PluginRegistry); ok {
		p.pluginRegistry = registry
	}

	if title, ok := cfg["title"].(string); ok {
		p.title = title
	} else {
		p.title = "GoREST API"
	}
	if version, ok := cfg["version"].(string); ok {
		p.version = version
	} else {
		p.version = "1.0.0"
	}
	if description, ok := cfg["description"].(string); ok {
		p.description = description
	} else {
		p.description = "Auto-generated REST API with full CRUD operations"
	}

	return nil
}

// Handler returns a no-op middleware
func (p *OpenAPIPlugin) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

// SetupEndpoints implements the EndpointSetup interface
func (p *OpenAPIPlugin) SetupEndpoints(app *fiber.App) error {
	// Setup OpenAPI UI endpoint
	app.Get("/openapi", func(c *fiber.Ctx) error {
		// Override CSP to allow loading external scripts and styles for Scalar UI
		c.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; "+
				"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; "+
				"font-src 'self' https://cdn.jsdelivr.net data:; "+
				"img-src 'self' data: https:; "+
				"connect-src 'self' https:;")

		html := `<!DOCTYPE html>
<html>
<head>
    <title>GoREST API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
</head>
<body>
    <script id="api-reference" data-url="/openapi.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	app.Get("/openapi.json", func(c *fiber.Ctx) error {
		// Build server URL from request
		protocol := "http"
		if c.Protocol() == "https" {
			protocol = "https"
		}
		serverURL := fmt.Sprintf("%s://%s", protocol, c.Hostname())

		spec, err := generateOpenAPISpec(app, GeneratorConfig{
			DTOsDirectory:      p.dtosDirectory,
			PluginRegistry:     p.pluginRegistry,
			PaginationLimit:    p.paginationLimit,
			PaginationMaxLimit: p.paginationMaxLimit,
			ServerURL:          serverURL,
			Title:              p.title,
			Version:            p.version,
			Description:        p.description,
		})

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to generate OpenAPI spec: %v", err),
			})
		}

		return c.JSON(spec)
	})

	logger.Log.Info("Api spec available", "url", fmt.Sprintf("http://localhost:%s/%s", "8000", "openapi"))
	logger.Log.Info("Api spec available (json format)", "url", fmt.Sprintf("http://localhost:%s/%s", "8000", "openapi.json"))

	return nil
}
