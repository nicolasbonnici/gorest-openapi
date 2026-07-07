package openapi

import (
	"maps"
	"sync"

	"github.com/gofiber/fiber/v3"
)

// maxCachedServerURLs bounds the marshalled-bytes cache. The servers block is
// the only request-dependent part of the spec and is derived from the Host
// header, so a hostile client could otherwise grow the cache without limit.
// Real deployments answer on a handful of hostnames, so once the bound is hit
// further hosts fall back to marshalling without being cached.
const maxCachedServerURLs = 32

type specEncoder func(any) ([]byte, error)

// specCache builds the static portion of the OpenAPI spec exactly once and
// memoises the fully marshalled JSON per server URL, so steady-state requests
// avoid both regeneration (route/DTO/reflection walks) and re-marshalling.
type specCache struct {
	build func() (map[string]interface{}, error)

	once      sync.Once
	staticDoc map[string]interface{}
	buildErr  error

	mu       sync.RWMutex
	byServer map[string][]byte
}

func newSpecCache(build func() (map[string]interface{}, error)) *specCache {
	return &specCache{
		build:    build,
		byServer: make(map[string][]byte),
	}
}

func (c *specCache) static() (map[string]interface{}, error) {
	c.once.Do(func() {
		c.staticDoc, c.buildErr = c.build()
	})
	return c.staticDoc, c.buildErr
}

// bytes returns the marshalled spec for the given server URL, generating and
// caching it on first use. The returned slice is owned by the cache and must
// not be mutated by callers.
func (c *specCache) bytes(serverURL string, encode specEncoder) ([]byte, error) {
	c.mu.RLock()
	cached, ok := c.byServer[serverURL]
	c.mu.RUnlock()
	if ok {
		return cached, nil
	}

	static, err := c.static()
	if err != nil {
		return nil, err
	}

	doc := maps.Clone(static)
	doc["servers"] = []map[string]string{
		{"url": serverURL, "description": "Development server"},
	}

	raw, err := encode(doc)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	if len(c.byServer) < maxCachedServerURLs {
		c.byServer[serverURL] = raw
	}
	c.mu.Unlock()

	return raw, nil
}

func encoderFrom(c fiber.Ctx) specEncoder {
	enc := c.App().Config().JSONEncoder
	return func(v any) ([]byte, error) { return enc(v) }
}
