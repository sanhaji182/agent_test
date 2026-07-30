# GoTest Agent Parsers Documentation

This document describes the multi-language parser implementations in GoTest Agent.

## Overview

GoTest Agent includes comprehensive parsers for multiple programming languages and frameworks. These parsers analyze your codebase to extract routes, models, and handlers, enabling intelligent test generation.

## Supported Languages and Frameworks

### JavaScript/TypeScript
- **Frameworks**: Express.js, Next.js, NestJS
- **Parser**: `internal/parser/javascript/parser.go`
- **Capabilities**:
  - Extract Express routes (`app.get()`, `app.post()`, etc.)
  - Extract route parameters (`:id`, `:userId`, etc.)
  - Extract middleware
  - Extract handlers and controllers

### Go
- **Frameworks**: Chi, Gin, Echo, Fiber
- **Parser**: `internal/parser/go/parser.go`
- **Capabilities**:
  - Extract Chi routes (`r.Get()`, `r.Post()`, etc.)
  - Extract Gin routes (`router.GET()`, `router.POST()`, etc.)
  - Extract Echo routes (`e.GET()`, `e.POST()`, etc.)
  - Extract Fiber routes (`app.Get()`, `app.Post()`, etc.)
  - Extract route parameters and middleware

### PHP
- **Frameworks**: Laravel, Symfony
- **Parser**: `internal/parser/php/parser.go`
- **Capabilities**:
  - Extract Laravel routes (`Route::get()`, `Route::post()`, etc.)
  - Extract route parameters and middleware
  - Extract controllers and methods
  - Extract Eloquent models and relationships

### Python
- **Frameworks**: Django, Flask, FastAPI
- **Parser**: `internal/parser/python/parser.go`
- **Capabilities**:
  - **Django**: Extract URL patterns from `urls.py`, extract models from `models.py`, extract views from `views.py`
  - **Flask**: Extract `@app.route()` decorators, extract HTTP methods from decorator arguments
  - **FastAPI**: Extract `@app.get()`, `@app.post()` decorators, extract route handlers

## Parser Architecture

### Common Interface

All parsers implement a common interface defined in `internal/parser/interface.go`:

```go
type Parser interface {
    // Parse analyzes the codebase and extracts routes, models, and handlers
    Parse(ctx context.Context, rootDir string) (*Codebase, error)
    
    // DetectFramework detects the framework used in the codebase
    DetectFramework(rootDir string) (string, error)
    
    // GetSupportedFrameworks returns the list of supported frameworks
    GetSupportedFrameworks() []string
}
```

### Codebase Structure

The `Codebase` type represents the parsed codebase:

```go
type Codebase struct {
    Language   string    // Programming language (e.g., "javascript", "go", "python")
    Framework  string    // Framework (e.g., "express", "gin", "django")
    Routes     []Route   // Extracted routes
    Models     []Model   // Extracted models
    Handlers   []Handler // Extracted handlers
    Metadata   map[string]interface{} // Additional metadata
}
```

### Route Structure

```go
type Route struct {
    Method     string            // HTTP method (GET, POST, PUT, DELETE, etc.)
    Path       string            // Route path (e.g., "/users/:id")
    Handler    string            // Handler function name
    Middleware []string          // Middleware applied to the route
    Parameters []RouteParameter  // Route parameters
}

type RouteParameter struct {
    Name string // Parameter name (e.g., "id")
    Type string // Parameter type (if known)
}
```

### Model Structure

```go
type Model struct {
    Name        string            // Model name
    Table       string            // Database table name
    Fields      []Field           // Model fields
    Relations   []Relation        // Model relationships
}

type Field struct {
    Name     string // Field name
    Type     string // Field type
    Required bool   // Whether the field is required
}

type Relation struct {
    Type       string // Relation type (e.g., "belongs_to", "has_many")
    Model      string // Related model name
    ForeignKey string // Foreign key field
}
```

### Handler Structure

```go
type Handler struct {
    Name       string   // Handler function name
    File       string   // File where the handler is defined
    Line       int      // Line number
    Parameters []string // Handler parameters
}
```

## Usage

### Basic Usage

```go
import (
    "context"
    "github.com/go-go-golems/gotest-agent/internal/parser"
)

// Create parser registry
registry := parser.NewRegistry()

// Parse codebase
codebase, err := registry.Parse(context.Background(), "/path/to/project")
if err != nil {
    log.Fatal(err)
}

// Use extracted routes
for _, route := range codebase.Routes {
    fmt.Printf("Route: %s %s -> %s\n", route.Method, route.Path, route.Handler)
}

// Use extracted models
for _, model := range codebase.Models {
    fmt.Printf("Model: %s (table: %s)\n", model.Name, model.Table)
}

// Use extracted handlers
for _, handler := range codebase.Handlers {
    fmt.Printf("Handler: %s (file: %s:%d)\n", handler.Name, handler.File, handler.Line)
}
```

### Framework Detection

```go
// Detect framework
framework, err := registry.DetectFramework("/path/to/project")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Detected framework: %s\n", framework)
```

## Installation

### Tree-sitter Dependencies

The parsers use tree-sitter for AST parsing. You need to install the tree-sitter dependencies:

```bash
# Run the installation script
./scripts/install-tree-sitter-deps.sh
```

Or install manually:

```bash
go get github.com/smacker/go-tree-sitter
go get github.com/smacker/go-tree-sitter/javascript
go get github.com/smacker/go-tree-sitter/golang
go get github.com/smacker/go-tree-sitter/python
go get github.com/smacker/go-tree-sitter/php
```

## Testing

Run parser tests:

```bash
# Run all parser tests
go test ./internal/parser/... -v

# Run specific parser tests
go test ./internal/parser/javascript -v
go test ./internal/parser/go -v
go test ./internal/parser/python -v
go test ./internal/parser/php -v
```

## Future Enhancements

### Planned Language Support
- **Ruby**: Ruby on Rails, Sinatra
- **Java**: Spring Boot, Spring MVC
- **C#**: ASP.NET Core, ASP.NET MVC
- **Rust**: Actix, Rocket, Axum

### Planned Features
- **Database Schema Extraction**: Extract database schemas from migration files
- **API Documentation Extraction**: Extract API documentation from code comments
- **Test Coverage Analysis**: Analyze existing test coverage
- **Code Quality Metrics**: Extract code quality metrics
- **Dependency Analysis**: Analyze project dependencies

## Performance

Parser performance varies by project size:

- **Small projects** (< 100 files): 1-5 seconds
- **Medium projects** (100-500 files): 5-15 seconds
- **Large projects** (500+ files): 15-60 seconds

## Limitations

- **Dynamic Routes**: Some dynamic route patterns may not be detected
- **Complex Middleware**: Complex middleware chains may not be fully extracted
- **Custom Frameworks**: Custom or non-standard frameworks may not be supported
- **Minified Code**: Minified or obfuscated code may not be parsed correctly

## Troubleshooting

### Parser Setup Failed

If you see "missing go.sum entry" errors:

```bash
go mod tidy
go mod download github.com/smacker/go-tree-sitter
```

### Parser Not Detecting Framework

If the parser is not detecting your framework:

1. Ensure your project structure follows standard conventions
2. Check that your framework is supported
3. Verify that your code is not minified or obfuscated
4. Check parser logs for errors

### Parser Performance Issues

If parsing is slow:

1. Exclude unnecessary directories (e.g., `node_modules`, `vendor`)
2. Use `.gitignore` patterns to exclude large directories
3. Consider using a subset of your codebase for testing

## Contributing

To add support for a new language or framework:

1. Create a new parser file in `internal/parser/<language>/parser.go`
2. Implement the `Parser` interface
3. Add tests in `internal/parser/<language>/parser_test.go`
4. Update this documentation
5. Submit a pull request

## References

- [Tree-sitter Documentation](https://tree-sitter.github.io/tree-sitter/)
- [Tree-sitter Go Bindings](https://github.com/smacker/go-tree-sitter)
- [Express.js Documentation](https://expressjs.com/)
- [Gin Framework Documentation](https://gin-gonic.com/)
- [Django Documentation](https://docs.djangoproject.com/)
- [Flask Documentation](https://flask.palletsprojects.com/)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Laravel Documentation](https://laravel.com/docs)

## License

MIT License - see LICENSE file for details.
