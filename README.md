# GoREST OpenAPI Plugin

OpenAPI documentation plugin for GoREST framework.

## Installation

```bash
go get github.com/nicolasbonnici/gorest-openapi
```

## Usage

```go
import (
	"github.com/nicolasbonnici/gorest/pluginloader"
	openapiplugin "github.com/nicolasbonnici/gorest-openapi"
)

func init() {
	pluginloader.RegisterPluginFactory("openapi", openapiplugin.NewPlugin)
}
```

### Configuration

Add to your `gorest.yaml`:

```yaml
plugins:
  - name: openapi
    enabled: true
```

## Features

- Auto-generated OpenAPI 3.0 specification
- Interactive API documentation UI at `/openapi`
- OpenAPI JSON schema at `/openapi.json`
- Dynamic schema generation from database
- Scalar API reference integration

## Endpoints

- `GET /openapi` - Interactive API documentation UI
- `GET /openapi.json` - OpenAPI 3.0 JSON schema

## License

MIT License - see LICENSE file for details
