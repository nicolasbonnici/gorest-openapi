package openapi

import "strings"

// Security requirements attached to operations. OpenAPI treats an empty slice on
// an operation as an explicit override meaning "no authentication required",
// while a requirement naming bearerAuth means a JWT bearer token is needed.
//
// The spec previously declared a single global bearerAuth requirement and no
// per-operation overrides, so every operation, including public ones such as
// POST /auth/login and read endpoints, appeared to require a token. These
// helpers let each operation state its real requirement.

// securedRequirement marks an operation as requiring a bearer token.
func securedRequirement() []map[string]interface{} {
	return []map[string]interface{}{{"bearerAuth": []string{}}}
}

// publicRequirement marks an operation as explicitly not requiring auth.
func publicRequirement() []map[string]interface{} {
	return []map[string]interface{}{}
}

// isPublicPath reports whether a path is always reachable without a token,
// regardless of HTTP method: the authentication endpoints (which issue or rotate
// tokens) and the documentation/health surface.
func isPublicPath(path string) bool {
	p := strings.ToLower(strings.TrimRight(path, "/"))
	if p == "" {
		p = "/"
	}

	if strings.HasPrefix(p, "/auth/") || p == "/auth" {
		return true
	}

	switch p {
	case "/", "/openapi", "/openapi.json", "/status", "/health", "/healthz", "/ready", "/live", "/metrics/prometheus":
		return true
	}

	return false
}

// operationSecurity returns the security requirement for a non-resource route.
// The authentication and documentation surfaces are public; otherwise safe,
// side-effect-free reads (GET/HEAD/OPTIONS) are treated as public and
// state-changing methods require a bearer token. This mirrors the common GoREST
// pattern where reads are open and writes are gated by the auth and RBAC
// middleware.
func operationSecurity(path, method string) []map[string]interface{} {
	if isPublicPath(path) {
		return publicRequirement()
	}

	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS":
		return publicRequirement()
	default:
		return securedRequirement()
	}
}
