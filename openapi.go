package openapi

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/plugin"
)

type OpenAPIPlugin struct {
	paginationLimit    int
	paginationMaxLimit int
	dtosDirectory      string
}

func NewPlugin() plugin.Plugin {
	return &OpenAPIPlugin{}
}

func (p *OpenAPIPlugin) Name() string {
	return "openapi"
}

func (p *OpenAPIPlugin) Initialize(cfg map[string]interface{}) error {
	if limit, ok := cfg["pagination_limit"].(int); ok {
		p.paginationLimit = limit
	}
	if maxLimit, ok := cfg["pagination_max_limit"].(int); ok {
		p.paginationMaxLimit = maxLimit
	}
	if dtosDir, ok := cfg["dtos_directory"].(string); ok {
		p.dtosDirectory = dtosDir
	} else {
		return fmt.Errorf("dtos_directory required in plugin config")
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
		spec, err := generateOpenAPISpec(app, GeneratorConfig{
			DTOsDirectory:      p.dtosDirectory,
			PaginationLimit:    p.paginationLimit,
			PaginationMaxLimit: p.paginationMaxLimit,
		})

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to generate OpenAPI spec: %v", err),
			})
		}

		return c.JSON(spec)
	})

	return nil
}
