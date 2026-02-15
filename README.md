# GoREST OpenAPI Plugin

[![CI](https://github.com/nicolasbonnici/gorest-openapi/actions/workflows/ci.yml/badge.svg)](https://github.com/nicolasbonnici/gorest-openapi/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nicolasbonnici/gorest-openapi)](https://goreportcard.com/report/github.com/nicolasbonnici/gorest-openapi)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

OpenAPI documentation plugin for GoREST framework.

## Installation

```bash
go get github.com/nicolasbonnici/gorest-openapi
```


## Development Environment

To set up your development environment:

```bash
make install
```

This will:
- Install Go dependencies
- Install development tools (golangci-lint)
- Set up git hooks (pre-commit linting and tests)

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
    config:
      # Required
      dtos_directory: "./dtos"  # Path to your DTOs directory

      # Optional API information (with defaults shown)
      title: "My API"                                    # default: "GoREST API"
      version: "1.0.0"                                   # default: "1.0.0"
      description: "My awesome API documentation"        # default: "Auto-generated REST API with full CRUD operations"

      # Optional pagination settings (with defaults shown)
      pagination_limit: 20      # default: 20
      pagination_max_limit: 100 # default: 100
```

#### Minimal Configuration

```yaml
plugins:
  - name: openapi
    enabled: true
    config:
      dtos_directory: "./dtos"
```

**Note:** The server URL is automatically detected from incoming requests, so it works with any port your application runs on.

## Features

- Auto-generated OpenAPI 3.0 specification
- Interactive API documentation UI at `/openapi`
- OpenAPI JSON schema at `/openapi.json`
- Dynamic schema generation from database
- Scalar API reference integration

## Endpoints

- `GET /openapi` - Interactive API documentation UI
- `GET /openapi.json` - OpenAPI 3.0 JSON schema

---

## Git Hooks

This directory contains git hooks for the GoREST plugin to maintain code quality.

### Available Hooks

#### pre-commit

Runs before each commit to ensure code quality:
- **Linting**: Runs `make lint` to check code style and potential issues
- **Tests**: Runs `make test` to verify all tests pass

### Installation

#### Automatic Installation

Run the install script from the project root:

```bash
./.githooks/install.sh
```

#### Manual Installation

Copy the hooks to your `.git/hooks` directory:

```bash
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---


## License

MIT License - see LICENSE file for details
