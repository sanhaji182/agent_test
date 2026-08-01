package rust

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

var (
	frameworkIndicators = []struct {
		dependency string
		framework  string
	}{
		{"actix-web", "actix-web"},
		{"axum", "axum"},
		{"rocket", "rocket"},
		{"warp", "warp"},
		{"poem", "poem"},
		{"salvo", "salvo"},
	}

	// Actix/Rocket-style route macros: #[get("/users/{id}")], #[post("/users")], etc.
	routeMacroRe = regexp.MustCompile(`#\[\s*(get|post|put|delete|patch|head|options)\s*\(\s*"([^"]+)"`)

	// Actix route macro with explicit method: #[route("/users", method = "GET")].
	routeWithMethodRe = regexp.MustCompile(`#\[\s*route\s*\(\s*"([^"]+)"\s*,\s*method\s*=\s*"([A-Za-z]+)"`)

	// Function signature used for handlers.
	functionSigRe = regexp.MustCompile(`(?m)(?:pub(?:\([^)]*\))?\s+)?(?:async\s+)?fn\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*(?:->\s*([^\{\n]+))?`)

	// Axum route method chain: route("/users", get(list).post(create)).
	axumMethodCallRe = regexp.MustCompile(`(?:routing::)?(get|post|put|delete|patch|head|options|any)\s*\(\s*([A-Za-z_][A-Za-z0-9_:]*)`)

	// Rust structs with named fields. Tuple/unit structs are intentionally ignored.
	structStartRe = regexp.MustCompile(`(?m)(?:pub\s+)?struct\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:<[^>{]+>)?\s*\{`)
	fieldLineRe   = regexp.MustCompile(`^\s*(?:pub(?:\([^)]*\))?\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*:\s*([^,]+),?\s*(?://.*)?$`)
)

// Parser implements parser.Parser for Rust web services.
// It uses lightweight lexical scanning instead of tree-sitter so it works without
// native grammar dependencies.
type Parser struct{}

// NewParser creates a new Rust parser.
func NewParser() *Parser {
	return &Parser{}
}

// SupportedLanguages returns the language supported by this parser.
func (p *Parser) SupportedLanguages() []string {
	return []string{"rust"}
}

// DetectFramework detects common Rust web frameworks from Cargo.toml and source imports.
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	return p.detectFramework(rootDir), nil
}

func (p *Parser) detectFramework(rootDir string) string {
	cargoPath := filepath.Join(rootDir, "Cargo.toml")
	if content, err := os.ReadFile(cargoPath); err == nil {
		cargo := strings.ToLower(string(content))
		for _, indicator := range frameworkIndicators {
			if strings.Contains(cargo, indicator.dependency) {
				return indicator.framework
			}
		}
	}

	files, err := p.findRustFiles(rootDir)
	if err != nil {
		return "rust"
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		lower := strings.ToLower(string(content))
		for _, indicator := range frameworkIndicators {
			if strings.Contains(lower, strings.ReplaceAll(indicator.dependency, "-", "_")) ||
				strings.Contains(lower, indicator.dependency) {
				return indicator.framework
			}
		}
	}

	return "rust"
}

// Parse analyzes a Rust codebase and extracts routes, models, and handlers.
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		RootDir:    rootDir,
		Language:   "rust",
		Framework:  p.detectFramework(rootDir),
		Routes:     make([]types.Route, 0),
		Models:     make([]types.Model, 0),
		Handlers:   make([]types.Handler, 0),
		AnalyzedAt: time.Now(),
	}

	files, err := p.findRustFiles(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find Rust files: %w", err)
	}
	codebase.FileCount = len(files)

	for _, file := range files {
		select {
		case <-ctx.Done():
			return codebase, ctx.Err()
		default:
		}
		if err := p.parseFile(file, codebase); err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", file, err)
		}
	}

	return codebase, nil
}

func (p *Parser) findRustFiles(rootDir string) ([]string, error) {
	var files []string
	skipDirs := map[string]bool{
		".git":         true,
		"target":       true,
		"node_modules": true,
		"vendor":       true,
		".idea":        true,
		".vscode":      true,
	}

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".rs") && !strings.HasSuffix(d.Name(), "_test.rs") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func (p *Parser) parseFile(filePath string, codebase *types.Codebase) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	text := string(content)

	p.extractMacroRoutes(text, filePath, codebase)
	p.extractAxumRoutes(text, filePath, codebase)
	p.extractModels(text, filePath, codebase)
	p.extractHandlers(text, filePath, codebase)

	return nil
}

func (p *Parser) extractMacroRoutes(content, filePath string, codebase *types.Codebase) {
	for _, match := range routeMacroRe.FindAllStringSubmatchIndex(content, -1) {
		method := strings.ToUpper(content[match[2]:match[3]])
		path := content[match[4]:match[5]]
		handler := nextFunctionName(content, match[1])
		codebase.Routes = append(codebase.Routes, types.Route{
			Method:  method,
			Path:    path,
			Handler: handler,
			Params:  rustPathParams(path),
			File:    filePath,
			Line:    lineNumber(content, match[0]),
		})
	}

	for _, match := range routeWithMethodRe.FindAllStringSubmatchIndex(content, -1) {
		path := content[match[2]:match[3]]
		method := strings.ToUpper(content[match[4]:match[5]])
		handler := nextFunctionName(content, match[1])
		codebase.Routes = append(codebase.Routes, types.Route{
			Method:  method,
			Path:    path,
			Handler: handler,
			Params:  rustPathParams(path),
			File:    filePath,
			Line:    lineNumber(content, match[0]),
		})
	}
}

func (p *Parser) extractAxumRoutes(content, filePath string, codebase *types.Codebase) {
	for _, call := range findRouteCalls(content) {
		path := firstQuotedString(call.body)
		if path == "" {
			continue
		}
		for _, match := range axumMethodCallRe.FindAllStringSubmatch(call.body, -1) {
			method := strings.ToUpper(match[1])
			if method == "ANY" {
				method = "*"
			}
			handler := match[2]
			codebase.Routes = append(codebase.Routes, types.Route{
				Method:  method,
				Path:    path,
				Handler: handler,
				Params:  rustPathParams(path),
				File:    filePath,
				Line:    lineNumber(content, call.start),
			})
		}
	}
}

func (p *Parser) extractModels(content, filePath string, codebase *types.Codebase) {
	matches := structStartRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		openBrace := strings.LastIndex(content[match[0]:match[1]], "{")
		if openBrace == -1 {
			continue
		}
		bodyStart := match[0] + openBrace + 1
		bodyEnd := findMatchingBrace(content, bodyStart-1)
		if bodyEnd == -1 {
			continue
		}

		model := types.Model{
			Name:  name,
			Table: snakePlural(name),
			File:  filePath,
			Line:  lineNumber(content, match[0]),
		}
		body := content[bodyStart:bodyEnd]
		for _, line := range strings.Split(body, "\n") {
			line = strings.TrimSpace(stripLineComment(line))
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if fm := fieldLineRe.FindStringSubmatch(line); fm != nil {
				model.Fields = append(model.Fields, types.Field{
					Name:     fm[1],
					Type:     strings.TrimSpace(fm[2]),
					Required: !isOptionalRustType(fm[2]),
				})
			}
		}
		codebase.Models = append(codebase.Models, model)
	}
}

func (p *Parser) extractHandlers(content, filePath string, codebase *types.Codebase) {
	matches := functionSigRe.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		name := content[match[2]:match[3]]
		params := ""
		if match[4] >= 0 && match[5] >= 0 {
			params = content[match[4]:match[5]]
		}
		returnType := ""
		if len(match) >= 8 && match[6] >= 0 && match[7] >= 0 {
			returnType = strings.TrimSpace(content[match[6]:match[7]])
		}

		body := functionBody(content, match[1])
		handler := types.Handler{
			Name:          name,
			Parameters:    parseRustParams(params),
			ReturnType:    returnType,
			HasValidation: strings.Contains(body, ".validate(") || strings.Contains(body, "validate(") || strings.Contains(body, "validator::"),
			DatabaseOps:   detectDatabaseOps(body),
			ExternalCalls: detectExternalCalls(body),
			File:          filePath,
			Line:          lineNumber(content, match[0]),
		}
		codebase.Handlers = append(codebase.Handlers, handler)
	}
}

type routeCall struct {
	start int
	body  string
}

func findRouteCalls(content string) []routeCall {
	var calls []routeCall
	searchFrom := 0
	for {
		idx := strings.Index(content[searchFrom:], ".route(")
		if idx == -1 {
			break
		}
		start := searchFrom + idx
		open := start + len(".route")
		end := findMatchingParen(content, open)
		if end == -1 {
			searchFrom = open + 1
			continue
		}
		calls = append(calls, routeCall{start: start, body: content[open+1 : end]})
		searchFrom = end + 1
	}
	return calls
}

func nextFunctionName(content string, from int) string {
	if from >= len(content) {
		return ""
	}
	if match := functionSigRe.FindStringSubmatch(content[from:]); match != nil {
		return match[1]
	}
	return ""
}

func firstQuotedString(value string) string {
	start := strings.Index(value, "\"")
	if start == -1 {
		return ""
	}
	end := strings.Index(value[start+1:], "\"")
	if end == -1 {
		return ""
	}
	return value[start+1 : start+1+end]
}

func functionBody(content string, sigEnd int) string {
	open := strings.Index(content[sigEnd:], "{")
	if open == -1 {
		return ""
	}
	open += sigEnd
	close := findMatchingBrace(content, open)
	if close == -1 || close <= open {
		return ""
	}
	return content[open+1 : close]
}

func findMatchingBrace(content string, open int) int {
	return findMatchingDelimiter(content, open, '{', '}')
}

func findMatchingParen(content string, open int) int {
	return findMatchingDelimiter(content, open, '(', ')')
}

func findMatchingDelimiter(content string, open int, left, right byte) int {
	if open < 0 || open >= len(content) || content[open] != left {
		return -1
	}
	depth := 0
	inString := false
	escaped := false
	for i := open; i < len(content); i++ {
		ch := content[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		if ch == left {
			depth++
		}
		if ch == right {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func parseRustParams(params string) []types.Parameter {
	if strings.TrimSpace(params) == "" {
		return nil
	}
	parts := splitTopLevel(params, ',')
	result := make([]types.Parameter, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "self" || part == "&self" || part == "&mut self" {
			continue
		}
		idx := strings.LastIndex(part, ":")
		if idx == -1 {
			result = append(result, types.Parameter{Name: part, Required: true})
			continue
		}
		name := strings.TrimSpace(part[:idx])
		paramType := strings.TrimSpace(part[idx+1:])
		result = append(result, types.Parameter{
			Name:     cleanRustParamName(name),
			Type:     paramType,
			Required: !isOptionalRustType(paramType),
		})
	}
	return result
}

func splitTopLevel(value string, sep rune) []string {
	var result []string
	start := 0
	angleDepth := 0
	parenDepth := 0
	bracketDepth := 0
	for i, r := range value {
		switch r {
		case '<':
			angleDepth++
		case '>':
			if angleDepth > 0 {
				angleDepth--
			}
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case sep:
			if angleDepth == 0 && parenDepth == 0 && bracketDepth == 0 {
				result = append(result, value[start:i])
				start = i + len(string(r))
			}
		}
	}
	result = append(result, value[start:])
	return result
}

func cleanRustParamName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "mut ")
	name = strings.TrimSpace(name)
	name = strings.Trim(name, "()")
	if idx := strings.LastIndex(name, " "); idx != -1 {
		name = name[idx+1:]
	}
	return strings.TrimSpace(name)
}

func isOptionalRustType(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), "Option<")
}

func detectDatabaseOps(body string) []string {
	lower := strings.ToLower(body)
	ops := make([]string, 0, 4)
	add := func(op string) {
		for _, existing := range ops {
			if existing == op {
				return
			}
		}
		ops = append(ops, op)
	}
	if strings.Contains(lower, "insert") || strings.Contains(lower, "create") || strings.Contains(lower, ".save") {
		add("create")
	}
	if strings.Contains(lower, "select") || strings.Contains(lower, "fetch") || strings.Contains(lower, "find") || strings.Contains(lower, "load") {
		add("read")
	}
	if strings.Contains(lower, "update") || strings.Contains(lower, ".set(") {
		add("update")
	}
	if strings.Contains(lower, "delete") || strings.Contains(lower, "remove") {
		add("delete")
	}
	return ops
}

func detectExternalCalls(body string) []string {
	lower := strings.ToLower(body)
	var calls []string
	for _, marker := range []string{"reqwest", "hyper::client", "awc::client", "ureq"} {
		if strings.Contains(lower, marker) {
			calls = append(calls, marker)
		}
	}
	return calls
}

func rustPathParams(path string) map[string]string {
	params := make(map[string]string)
	for _, segment := range strings.Split(path, "/") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		if strings.HasPrefix(segment, ":") {
			params[strings.TrimPrefix(segment, ":")] = "string"
			continue
		}
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			params[strings.Trim(segment, "{}")] = "string"
		}
	}
	if len(params) == 0 {
		return nil
	}
	return params
}

func lineNumber(content string, index int) int {
	if index <= 0 {
		return 1
	}
	if index > len(content) {
		index = len(content)
	}
	return strings.Count(content[:index], "\n") + 1
}

func stripLineComment(line string) string {
	if idx := strings.Index(line, "//"); idx != -1 {
		return line[:idx]
	}
	return line
}

func snakePlural(name string) string {
	var b strings.Builder
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteRune('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	value := b.String()
	if strings.HasSuffix(value, "s") {
		return value
	}
	return value + "s"
}
