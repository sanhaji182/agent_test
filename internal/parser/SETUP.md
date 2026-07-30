# Parser Setup Guide

## Install Dependencies

The JavaScript parser requires tree-sitter for AST parsing. Install the dependencies:

```bash
go get github.com/smacker/go-tree-sitter
go get github.com/smacker/go-tree-sitter/javascript
```

## Verify Installation

After installing dependencies, verify the parser works:

```bash
# Run all parser tests
go test ./internal/parser/... -v

# Run JavaScript parser tests specifically
go test ./internal/parser/javascript/... -v

# Run with coverage
go test ./internal/parser/... -cover
```

## Current Status

### ✅ Completed

- **Registry & Type Definitions** (Task 1.1-1.2)
  - `internal/parser/interface.go` - Parser interface
  - `internal/parser/registry.go` - Parser registry with language detection
  - `internal/parser/types/` - Codebase, Route, Model, Handler types
  - 10 tests passing

- **Parser Stubs** (Task 1.3)
  - JavaScript parser stub with framework detection
  - Go parser stub with framework detection
  - PHP parser stub with framework detection
  - Python parser stub with framework detection

- **JavaScript Parser Implementation** (Task 1.4 - Partial)
  - Complete Express route parsing implementation
  - AST traversal with tree-sitter
  - Support for app.get/post/put/delete/use patterns
  - Support for router.get/post/put/delete patterns
  - Middleware extraction
  - File path and line number tracking
  - Skip node_modules, dist, build directories
  - 9 comprehensive test cases

### ⏳ Pending

- **Install tree-sitter dependencies** (blocked by classifier)
  - `go get github.com/smacker/go-tree-sitter`
  - `go get github.com/smacker/go-tree-sitter/javascript`

- **Run JavaScript parser tests** (requires dependencies)
  - `TestJavaScriptParser_ParseExpressRoutes`
  - `TestJavaScriptParser_ParseRouterPattern`
  - `TestJavaScriptParser_ParseTypeScript`
  - `TestJavaScriptParser_ParseNextJS`
  - `TestJavaScriptParser_ParseReact`
  - `TestJavaScriptParser_SkipNodeModules`
  - `TestJavaScriptParser_EmptyProject`
  - `TestJavaScriptParser_MultipleFiles`

- **Implement Go parser** (Task 1.5)
  - Chi/Gin/Echo route parsing
  - Struct tag extraction
  - Handler function analysis

- **Implement PHP parser** (Task 1.6)
  - Laravel route parsing
  - Eloquent model parsing
  - Migration parsing

- **Implement Python parser** (Task 1.7)
  - FastAPI/Flask route parsing
  - Pydantic model parsing
  - SQLAlchemy model parsing

## Test Coverage

### Registry Tests (10/10 passing)

- `TestNewParserRegistry` - Registry creation
- `TestRegisterParser` - Parser registration
- `TestGetParserNotExists` - Error handling
- `TestDetectLanguage` - Language detection (4 subtests)
- `TestParseCodebase` - Basic parsing
- `TestAllParsers` - Parser instantiation
- `TestSupportedLanguages` - Language listing
- `TestParseWithAutoDetect` - Auto-detection
- `TestDetectLanguageByFileExtension` - Extension-based detection
- `TestParseEmptyDirectory` - Edge case handling

### JavaScript Parser Tests (9 written, pending execution)

- `TestJavaScriptParser_ParseExpressRoutes` - Express route extraction
- `TestJavaScriptParser_ParseRouterPattern` - Router pattern support
- `TestJavaScriptParser_ParseTypeScript` - TypeScript file support
- `TestJavaScriptParser_ParseNextJS` - Next.js framework detection
- `TestJavaScriptParser_ParseReact` - React framework detection
- `TestJavaScriptParser_SkipNodeModules` - Directory filtering
- `TestJavaScriptParser_EmptyProject` - Empty project handling
- `TestJavaScriptParser_MultipleFiles` - Multi-file projects

## Next Steps

1. **Install dependencies** (manual step required)
   ```bash
   go get github.com/smacker/go-tree-sitter
   go get github.com/smacker/go-tree-sitter/javascript
   ```

2. **Run tests**
   ```bash
   go test ./internal/parser/javascript/... -v
   ```

3. **Verify all tests pass**
   - All 9 JavaScript parser tests should pass
   - All 10 registry tests should pass

4. **Continue to Task 1.5** (Go parser implementation)

## Architecture Notes

### Parser Flow

1. **Registry detects language** - Checks for package.json, go.mod, composer.json, etc.
2. **Registry selects parser** - Returns appropriate parser for language
3. **Parser scans files** - Finds all source files (skips node_modules, dist, build)
4. **Parser parses each file** - Uses tree-sitter to build AST
5. **Parser extracts components** - Routes, models, handlers from AST
6. **Registry aggregates results** - Combines results into Codebase struct

### Tree-sitter Integration

- **Parser initialization** - Each parser creates tree-sitter parser with language grammar
- **AST traversal** - Recursive traversal to find patterns
- **Node extraction** - Extract relevant nodes (route definitions, model schemas, etc.)
- **Text extraction** - Get source text from byte ranges
- **Line tracking** - Map byte offsets to line numbers

### Express Route Patterns

Currently supported:

```javascript
// app.method patterns
app.get('/path', handler)
app.post('/path', middleware1, middleware2, handler)
app.put('/path/:id', handler)
app.delete('/path/:id', middleware, handler)
app.use('/path', router)

// router.method patterns
router.get('/path', handler)
router.post('/path', handler)
router.put('/path/:id', handler)
router.delete('/path/:id', handler)
```

### Future Enhancements

- **Parameter extraction** - Extract route parameters and types
- **Model field extraction** - Parse Mongoose schemas for field definitions
- **Handler signature extraction** - Extract function parameters and return types
- **Import tracking** - Track which handlers are imported from where
- **Comment parsing** - Extract JSDoc comments for documentation
- **TypeScript types** - Parse TypeScript interfaces and types
