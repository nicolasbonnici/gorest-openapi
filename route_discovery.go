package openapi

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func discoverNonResourceRoutes(app *fiber.App, resourcePaths map[string]bool) map[string]map[string]interface{} {
	routes := app.GetRoutes(true)
	discovered := make(map[string]map[string]interface{})

	for _, route := range routes {
		path := route.Path
		method := strings.ToUpper(route.Method)

		if shouldSkipRoute(path, resourcePaths) {
			continue
		}

		if discovered[path] == nil {
			discovered[path] = make(map[string]interface{})
		}

		discovered[path][strings.ToLower(method)] = generateRouteSpec(path, method)
	}

	return discovered
}

func shouldSkipRoute(path string, resourcePaths map[string]bool) bool {
	if path == "/openapi" || path == "/openapi.json" {
		return true
	}

	if resourcePaths[path] {
		return true
	}

	if path == "" || path == "/" {
		return true
	}

	return false
}

func generateRouteSpec(path, method string) map[string]interface{} {
	tag := determineTag(path)
	summary := generateSummary(path, method)
	description := generateDescription(path, method)

	spec := map[string]interface{}{
		"summary":     summary,
		"description": description,
		"tags":        []string{tag},
	}

	if strings.Contains(path, ":") {
		spec["parameters"] = extractPathParameters(path)
	}

	if method == "POST" || method == "PUT" || method == "PATCH" {
		spec["requestBody"] = generateRequestBody()
	}

	spec["responses"] = generateResponses(method)

	return spec
}

func determineTag(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "General"
	}

	segment := parts[0]

	switch segment {
	case "auth":
		return "Authentication"
	case "health":
		return "System"
	default:
		return strings.ToUpper(segment[:1]) + segment[1:]
	}
}

func generateSummary(path, method string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	action := ""

	switch method {
	case "GET":
		action = "Get"
	case "POST":
		action = "Create or execute"
	case "PUT":
		action = "Update"
	case "PATCH":
		action = "Partially update"
	case "DELETE":
		action = "Delete"
	default:
		action = method
	}

	pathName := strings.Join(parts, " ")
	pathName = strings.ReplaceAll(pathName, ":", "")

	return fmt.Sprintf("%s %s", action, pathName)
}

func generateDescription(path, method string) string {
	return fmt.Sprintf("%s %s", method, path)
}

func extractPathParameters(path string) []map[string]interface{} {
	var params []map[string]interface{}

	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			paramName := strings.TrimPrefix(part, ":")
			params = append(params, map[string]interface{}{
				"name":        paramName,
				"in":          "path",
				"required":    true,
				"description": fmt.Sprintf("Path parameter: %s", paramName),
				"schema":      map[string]string{"type": "string"},
			})
		}
	}

	return params
}

func generateRequestBody() map[string]interface{} {
	return map[string]interface{}{
		"required": true,
		"content": map[string]interface{}{
			"application/json": map[string]interface{}{
				"schema": map[string]interface{}{
					"type": "object",
				},
			},
		},
	}
}

func generateResponses(method string) map[string]interface{} {
	responses := map[string]interface{}{}

	switch method {
	case "GET":
		responses["200"] = map[string]interface{}{
			"description": "Successful response",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
					},
				},
			},
		}
	case "POST":
		responses["201"] = map[string]interface{}{
			"description": "Successfully created",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
					},
				},
			},
		}
		responses["400"] = map[string]interface{}{
			"description": "Bad request",
		}
	case "PUT", "PATCH":
		responses["200"] = map[string]interface{}{
			"description": "Successfully updated",
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
					},
				},
			},
		}
		responses["404"] = map[string]interface{}{
			"description": "Not found",
		}
	case "DELETE":
		responses["204"] = map[string]interface{}{
			"description": "Successfully deleted",
		}
		responses["404"] = map[string]interface{}{
			"description": "Not found",
		}
	default:
		responses["200"] = map[string]interface{}{
			"description": "Successful response",
		}
	}

	return responses
}
