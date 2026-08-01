package typescript

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

var (
	// Express route method calls: .get(, .post(, .put(, etc.
	// Matches the method call opening paren and captures the HTTP method.
	expressMethodRe = regexp.MustCompile(
		`\.\s*(get|post|put|delete|patch|head|options|use)\s*\(`,
	)

	// NestJS @Controller decorator, optionally with a prefix path.
	nestControllerRe = regexp.MustCompile(
		`@Controller\s*(?:\(\s*["']([^"']*)["']\s*\))?`,
	)

	// NestJS HTTP method decorators: @Get, @Post, @Put, @Delete, @Patch
	// Optionally with a path argument.
	nestMethodRe = regexp.MustCompile(
		`@(Get|Post|Put|Delete|Patch)\s*(?:\(\s*["']([^"']*)["']\s*\))?`,
	)

	// NestJS @Param('name') variable: type — captures param name, variable name, and type.
	nestParamRe = regexp.MustCompile(
		`@Param\s*\(\s*["'](\w+)["']\s*\)\s*(\w+)(?:\s*:\s*(\w+))?`,
	)

	// NestJS @Body() variable: type.
	nestBodyRe = regexp.MustCompile(
		`@Body\s*\(\s*\)\s*(\w+)(?:\s*:\s*(\w+))?`,
	)

	// NestJS @Query('name') variable: type.
	nestQueryRe = regexp.MustCompile(
		`@Query\s*(?:\(\s*["'](\w+)["']\s*\))?\s*(\w+)(?:\s*:\s*(\w+))?`,
	)

	// Exported TypeScript interface declarations.
	interfaceRe = regexp.MustCompile(
		`export\s+interface\s+(\w+)`,
	)

	// Interface field definitions: name?: type;
	interfaceFieldRe = regexp.MustCompile(
		`(\w+)(\?)?\s*:\s*([^;]+?)\s*;`,
	)

	// Class declaration: export class Name
	classRe = regexp.MustCompile(
		`export\s+(?:abstract\s+)?class\s+(\w+)`,
	)

	// Method name after decorators and access modifiers.
	methodNameRe = regexp.MustCompile(
		`(?:public|private|protected|async\s+)*\s*(\w+)\s*\(`,
	)

	// NestJS dependency indicator.
	nestJSIndicatorRe = regexp.MustCompile(
		`@nestjs/(?:core|common)`,
	)
)

// Parser implements the Parser interface for TypeScript.
// Uses regex-based scanning since no tree-sitter TypeScript grammar is available.
type Parser struct{}

// NewParser creates a new TypeScript parser.
func NewParser() *Parser {
	return &Parser{}
}

// SupportedLanguages returns the languages this parser supports.
func (p *Parser) SupportedLanguages() []string {
	return []string{"typescript"}
}

// DetectFramework detects the TypeScript framework used.
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	// Check for NestJS first (more specific).
	if p.isNestJS(rootDir) {
		return "nestjs", nil
	}

	// Check package.json for Express.
	packagePath := filepath.Join(rootDir, "package.json")
	if content, err := os.ReadFile(packagePath); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, `"express"`) || strings.Contains(contentStr, `"express`) {
			return "express", nil
		}
	}

	return "typescript", nil
}

// isNestJS checks if the project uses NestJS.
func (p *Parser) isNestJS(rootDir string) bool {
	found := false
	// Scan for @nestjs/* in source files.
	filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "dist" || name == ".next" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".ts" && ext != ".tsx" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if nestJSIndicatorRe.Match(content) {
			found = true
			return filepath.SkipAll
		}
		return nil
	})

	return found
}

// Parse analyzes a TypeScript codebase.
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language:   "typescript",
		RootDir:    rootDir,
		AnalyzedAt: time.Now(),
	}

	// Detect framework.
	framework, err := p.DetectFramework(rootDir)
	if err == nil {
		codebase.Framework = framework
	}

	// Find all TypeScript source files.
	files, err := p.findSourceFiles(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find source files: %w", err)
	}
	codebase.FileCount = len(files)

	// Parse each file.
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

// findSourceFiles finds all TypeScript files, skipping standard ignorable directories.
func (p *Parser) findSourceFiles(rootDir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "dist" || name == ".next" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".ts" || ext == ".tsx" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// parseFile parses a single TypeScript file for routes, models, and handlers.
func (p *Parser) parseFile(filePath string, codebase *types.Codebase) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)

	// Extract Express-style routes.
	expressRoutes := p.extractExpressRoutes(contentStr, filePath)
	codebase.Routes = append(codebase.Routes, expressRoutes...)

	// Extract NestJS controllers (routes + handlers).
	p.extractNestJSControllers(contentStr, filePath, codebase)

	// Extract NestJS modules.
	nestModules := p.extractNestJSModules(contentStr, filePath)
	codebase.Models = append(codebase.Models, nestModules...)

	// Extract interface/model definitions.
	models := p.extractInterfaces(contentStr, filePath)
	codebase.Models = append(codebase.Models, models...)

	return nil
}

// extractExpressRoutes finds Express-style route definitions using regex.
func (p *Parser) extractExpressRoutes(content string, filePath string) []types.Route {
	var routes []types.Route

	locs := expressMethodRe.FindAllStringSubmatchIndex(content, -1)
	for _, loc := range locs {
		if len(loc) < 4 {
			continue
		}

		method := content[loc[2]:loc[3]]

		// The opening paren is the last char of the full match.
		openParen := loc[1] - 1
		closeParen := findClosingDelim(content, openParen, '(', ')')
		if closeParen == -1 {
			continue
		}

		argsContent := content[openParen+1 : closeParen]

		// Extract the first argument (path string).
		path, rest := extractFirstArg(argsContent)
		if path == "" {
			continue
		}
		path = strings.Trim(path, "'\"`")

		// Extract handler/middleware names from remaining arguments.
		handlerNames := extractHandlerNames(rest)

		// Map the method name to HTTP method.
		httpMethod := mapExpressMethod(method)

		// Compute line number.
		line := lineNumber(content, openParen)

		routes = append(routes, types.Route{
			Method:     httpMethod,
			Path:       path,
			Middleware: handlerNames,
			File:       filePath,
			Line:       line,
		})
	}

	return routes
}

// extractNestJSControllers finds NestJS @Controller classes and their route methods.
func (p *Parser) extractNestJSControllers(content string, filePath string, codebase *types.Codebase) {
	controllerLocs := nestControllerRe.FindAllStringSubmatchIndex(content, -1)

	for _, cLoc := range controllerLocs {
		basePath := "/"
		if cLoc[2] != -1 {
			basePath = "/" + strings.TrimPrefix(content[cLoc[2]:cLoc[3]], "/")
		}

		// Find the class name that follows this @Controller decorator.
		// Look for "export class ClassName" after the decorator.
		afterController := content[cLoc[1]:]
		classMatch := classRe.FindStringSubmatchIndex(afterController)
		if classMatch == nil {
			continue
		}
		className := afterController[classMatch[2]:classMatch[3]]

		// Find the opening brace of the class body.
		classBodyStart := findOpeningBrace(afterController, classMatch[1])
		if classBodyStart == -1 {
			continue
		}
		classBodyEnd := findClosingDelim(afterController, classBodyStart, '{', '}')
		if classBodyEnd == -1 {
			continue
		}
		classBody := afterController[classBodyStart+1 : classBodyEnd]

		// Find method decorators within the class body.
		methodLocs := nestMethodRe.FindAllStringSubmatchIndex(classBody, -1)
		for _, mLoc := range methodLocs {
			if len(mLoc) < 4 {
				continue
			}

			httpMethod := strings.ToUpper(classBody[mLoc[2]:mLoc[3]])

			methodPath := "/"
			if mLoc[4] != -1 {
				methodPath = "/" + strings.TrimPrefix(classBody[mLoc[4]:mLoc[5]], "/")
			}

			fullPath := combineNestPaths(basePath, methodPath)

			// Find the method name after the decorator.
			afterMethodDecorator := classBody[mLoc[1]:]
			methodMatch := methodNameRe.FindStringSubmatchIndex(afterMethodDecorator)
			if methodMatch == nil {
				continue
			}
			methodName := afterMethodDecorator[methodMatch[2]:methodMatch[3]]

			// Calculate line number relative to the original content.
			classBodyOffset := cLoc[1] + classBodyStart + 1
			methodOffset := classBodyOffset + mLoc[0]
			line := lineNumber(content, methodOffset)

			// Extract parameter info from the method signature.
			// methodEnd should be the end of the class body in the original content.
			params, hasBody := p.extractNestMethodParams(content, methodOffset, cLoc[1]+classBodyEnd)

			// Build the route.
			route := types.Route{
				Method:  httpMethod,
				Path:    fullPath,
				Handler: methodName,
				Params:  params,
				File:    filePath,
				Line:    line,
			}
			codebase.Routes = append(codebase.Routes, route)

			// Build the handler.
			handler := types.Handler{
				Name:       methodName,
				Controller: className,
				Method:     httpMethod,
				File:       filePath,
				Line:       line,
			}
			if hasBody {
				handler.HasValidation = true
			}
			codebase.Handlers = append(codebase.Handlers, handler)
		}
	}
}

// extractNestMethodParams extracts @Param and @Body from around a NestJS method.
func (p *Parser) extractNestMethodParams(content string, methodStart int, methodEnd int) (map[string]string, bool) {
	if methodEnd <= methodStart {
		return nil, false
	}

	region := content[methodStart:methodEnd]
	params := make(map[string]string)
	hasBody := false

	// Extract @Param decorators.
	paramMatches := nestParamRe.FindAllStringSubmatch(region, -1)
	for _, pm := range paramMatches {
		paramName := pm[1]
		paramType := "string"
		if pm[3] != "" {
			paramType = pm[3]
		}
		params[paramName] = paramType
	}

	// Check for @Body.
	if nestBodyRe.MatchString(region) {
		hasBody = true
	}

	return params, hasBody
}

// extractInterfaces finds exported TypeScript interfaces and converts them to models.
func (p *Parser) extractInterfaces(content string, filePath string) []types.Model {
	var models []types.Model

	ifaceLocs := interfaceRe.FindAllStringSubmatchIndex(content, -1)
	for _, iLoc := range ifaceLocs {
		if len(iLoc) < 4 {
			continue
		}

		name := content[iLoc[2]:iLoc[3]]

		// Find the opening brace of the interface body.
		openBrace := findOpeningBrace(content, iLoc[1])
		if openBrace == -1 {
			continue
		}
		closeBrace := findClosingDelim(content, openBrace, '{', '}')
		if closeBrace == -1 {
			continue
		}

		body := content[openBrace+1 : closeBrace]

		// Extract field definitions.
		fieldMatches := interfaceFieldRe.FindAllStringSubmatch(body, -1)
		var fields []types.Field
		for _, fm := range fieldMatches {
			fieldName := fm[1]
			isOptional := fm[2] == "?"
			fieldType := strings.TrimSpace(fm[3])

			fields = append(fields, types.Field{
				Name:     fieldName,
				Type:     fieldType,
				Required: !isOptional,
			})
		}

		line := lineNumber(content, iLoc[0])

		models = append(models, types.Model{
			Name:   name,
			Fields: fields,
			File:   filePath,
			Line:   line,
		})
	}

	return models
}

// extractNestJSModules finds @Module() decorators in TypeScript files.
func (p *Parser) extractNestJSModules(content string, filePath string) []types.Model {
	// @Module({...}) pattern
	moduleRe := regexp.MustCompile(`@Module\s*\(\s*\{`)
	var modules []types.Model

	locs := moduleRe.FindAllStringIndex(content, -1)
	for _, loc := range locs {
		openBrace := loc[1] - 1 // position of {
		closeBrace := findClosingDelim(content, openBrace, '{', '}')
		if closeBrace == -1 {
			continue
		}

		body := content[openBrace+1 : closeBrace]
		line := lineNumber(content, loc[0])

		// Extract module name from class that follows @Module
		afterModule := content[loc[1]:]
		classMatch := classRe.FindStringSubmatchIndex(afterModule)
		moduleName := "Module"
		if classMatch != nil {
			moduleName = afterModule[classMatch[2]:classMatch[3]]
		}

		// Extract controllers, providers, imports, exports
		var fields []types.Field
		fieldExtractors := []string{"controllers", "providers", "imports", "exports"}
		for _, fieldName := range fieldExtractors {
			re := regexp.MustCompile(fieldName + `\s*:\s*\[([^\]]*)\]`)
			match := re.FindStringSubmatch(body)
			if match != nil {
				items := strings.TrimSpace(match[1])
				fields = append(fields, types.Field{
					Name: fieldName,
					Type: items,
				})
			}
		}

		modules = append(modules, types.Model{
			Name:   moduleName,
			Table:  "module",
			Fields: fields,
			File:   filePath,
			Line:   line,
		})
	}

	return modules
}

// --- Helper functions ---

// findClosingDelim finds the matching closing delimiter starting from an opening delimiter position.
func findClosingDelim(s string, start int, open, close byte) int {
	depth := 0
	for i := start; i < len(s); i++ {
		if s[i] == open {
			depth++
		} else if s[i] == close {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// extractFirstArg extracts the first argument from a function call arguments string.
// Returns the first arg and the rest of the arguments.
func extractFirstArg(s string) (string, string) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return "", ""
	}

	// Check if it's a string literal
	if s[0] == '\'' || s[0] == '"' || s[0] == '`' {
		quote := s[0]
		for i := 1; i < len(s); i++ {
			if s[i] == '\\' {
				i++ // skip escaped char
				continue
			}
			if s[i] == quote {
				first := s[0 : i+1]
				rest := strings.TrimSpace(s[i+1:])
				if len(rest) > 0 && rest[0] == ',' {
					rest = strings.TrimSpace(rest[1:])
				}
				return first, rest
			}
		}
		return s, ""
	}

	// Non-string argument: find the first comma outside of nested structures
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ',':
			if depth == 0 {
				return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])
			}
		case '(', '{', '[':
			depth++
		case ')', '}', ']':
			depth--
		}
	}
	return s, ""
}

// extractHandlerNames extracts handler/middleware names from comma-separated arguments.
func extractHandlerNames(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var names []string
	for s != "" {
		// Skip arrow functions and anonymous functions
		if strings.HasPrefix(s, "(") || strings.HasPrefix(s, "async") || strings.HasPrefix(s, "function") {
			// Find the end of this argument
			_, s = extractFirstArg(s)
			continue
		}

		first, rest := extractFirstArg(s)
		if first == "" {
			break
		}

		// Skip string literals (they're usually path arguments)
		if len(first) > 0 && (first[0] == '\'' || first[0] == '"' || first[0] == '`') {
			s = rest
			continue
		}

		names = append(names, first)
		s = rest
	}

	return names
}

// mapExpressMethod maps Express.js method names to uppercase HTTP methods.
func mapExpressMethod(method string) string {
	mapping := map[string]string{
		"get":     "GET",
		"post":    "POST",
		"put":     "PUT",
		"delete":  "DELETE",
		"patch":   "PATCH",
		"head":    "HEAD",
		"options": "OPTIONS",
		"use":     "USE",
		"all":     "ALL",
	}
	if m, ok := mapping[method]; ok {
		return m
	}
	return strings.ToUpper(method)
}

// lineNumber returns the 1-based line number for the given byte offset.
func lineNumber(s string, offset int) int {
	if offset < 0 || offset > len(s) {
		return 1
	}
	line := 1
	for i := 0; i < offset && i < len(s); i++ {
		if s[i] == '\n' {
			line++
		}
	}
	return line
}

// findOpeningBrace finds the position of the next '{' after the given offset.
func findOpeningBrace(s string, offset int) int {
	for i := offset; i < len(s); i++ {
		if s[i] == '{' {
			return i
		}
	}
	return -1
}

// combineNestPaths combines a base path and method path.
func combineNestPaths(base, method string) string {
	b := strings.TrimRight(base, "/")
	if b == "" {
		b = "/"
	}

	if method == "/" {
		if b == "/" {
			return "/"
		}
		return b
	}

	m := strings.TrimLeft(method, "/")
	return b + "/" + m
}
