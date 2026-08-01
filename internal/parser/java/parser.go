package java

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
	// Controller annotations
	controllerAnnotationRe = regexp.MustCompile(`@(?:Rest)?Controller\b`)

	// Class declaration
	classRe = regexp.MustCompile(`class\s+(\w+)`)

	// @RequestMapping (class-level and method-level)
	// Captures value= or bare string path, and optionally method=RequestMethod.XXX
	requestMappingRe = regexp.MustCompile(
		`@RequestMapping\s*\(` +
			`(?:[^)]*?(?:value\s*=\s*)?["']([^"']+)["'])?` +
			`(?:[^)]*?method\s*=\s*\{?\s*RequestMethod\.(\w+)\s*}?)?` +
			`[^)]*\)`,
	)

	// @GetMapping, @PostMapping, @PutMapping, @DeleteMapping, @PatchMapping
	// Matches annotations WITH an explicit path: @GetMapping("/path")
	methodMappingRe = regexp.MustCompile(
		`@(GetMapping|PostMapping|PutMapping|DeleteMapping|PatchMapping)\s*\(` +
			`[^)]*?(?:value\s*=\s*)?["']([^"']+)["'][^)]*\)`,
	)

	// Matches @GetMapping, @GetMapping(), etc. WITHOUT an explicit path (defaults to "/")
	methodMappingNoPathRe = regexp.MustCompile(
		`@(GetMapping|PostMapping|PutMapping|DeleteMapping|PatchMapping)(?:\s*\(\s*\))?`,
	)

	// Method signature: finds the next public method after an annotation.
	// Matches optional other annotations, then "public ReturnType methodName(".
	// Used on text that starts AFTER the mapping annotation has been consumed.
	methodSigRe = regexp.MustCompile(
		`(?:@\w+(?:\s*\([^)]*\))?\s*)*public\s+\S+\s+(\w+)\s*\(`,
	)

	// @Entity
	entityRe = regexp.MustCompile(`@Entity\b`)

	// @Table
	tableRe = regexp.MustCompile(`@Table\s*\([^)]*?(?:name\s*=\s*)?["']([^"']+)["'][^)]*\)`)

	// @Column with field declaration
	columnRe = regexp.MustCompile(
		`@Column\s*\([^)]*?(?:name\s*=\s*)?["']([^"']+)["'][^)]*\)\s*` +
			`(?:@\w+\s*\([^)]*\)\s*)*` + // other annotations
			`(?:private|protected|public)\s+(\S+)\s+(\w+)\s*;`,
	)

	// @Repository on interface
	repositoryRe = regexp.MustCompile(`@Repository\b[\s\S]*?public\s+interface\s+(\w+)`)

	// Spring Boot dependency indicators
	springBootIndicatorRe = regexp.MustCompile(
		`(?:spring-boot-starter|org\.springframework\.boot)`,
	)
)

// Parser implements the Parser interface for Java/Spring Boot.
// Uses regex-based scanning since no tree-sitter Java grammar is available.
type Parser struct{}

// NewParser creates a new Java parser.
func NewParser() *Parser {
	return &Parser{}
}

// SupportedLanguages returns the languages this parser supports.
func (p *Parser) SupportedLanguages() []string {
	return []string{"java"}
}

// DetectFramework detects the Java framework used.
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	return p.detectFramework(rootDir), nil
}

// detectFramework checks project structure for framework indicators.
func (p *Parser) detectFramework(rootDir string) string {
	// Check for Maven (pom.xml)
	pomPath := filepath.Join(rootDir, "pom.xml")
	if content, err := os.ReadFile(pomPath); err == nil {
		if springBootIndicatorRe.Match(content) {
			return "spring-boot"
		}
		return "maven"
	}

	// Check for Gradle (build.gradle or build.gradle.kts)
	for _, gradleFile := range []string{"build.gradle", "build.gradle.kts"} {
		gradlePath := filepath.Join(rootDir, gradleFile)
		if content, err := os.ReadFile(gradlePath); err == nil {
			if springBootIndicatorRe.Match(content) {
				return "spring-boot"
			}
			return "gradle"
		}
	}

	// Scan for Spring annotations in source files
	if p.hasSpringAnnotations(rootDir) {
		return "spring-boot"
	}

	return "java"
}

// hasSpringAnnotations scans .java files for Spring annotations.
func (p *Parser) hasSpringAnnotations(rootDir string) bool {
	files, err := p.findJavaFiles(rootDir)
	if err != nil {
		return false
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if controllerAnnotationRe.Match(content) ||
			entityRe.Match(content) ||
			repositoryRe.Match(content) {
			return true
		}
	}

	return false
}

// Parse analyzes a Java codebase.
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		RootDir:    rootDir,
		Language:   "java",
		Framework:  p.detectFramework(rootDir),
		Routes:     make([]types.Route, 0),
		Models:     make([]types.Model, 0),
		Handlers:   make([]types.Handler, 0),
		AnalyzedAt: time.Now(),
	}

	files, err := p.findJavaFiles(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find Java files: %w", err)
	}

	for _, file := range files {
		if err := p.parseFile(file, codebase); err != nil {
			// Log warning but continue
			fmt.Printf("Warning: failed to parse %s: %v\n", file, err)
		}
	}

	codebase.FileCount = len(files)
	return codebase, nil
}

// findJavaFiles walks the project and returns .java source files (excluding tests).
func (p *Parser) findJavaFiles(rootDir string) ([]string, error) {
	var files []string

	skipDirs := map[string]bool{
		".git": true, "target": true, "build": true, "node_modules": true,
		".gradle": true, "bin": true, "out": true, ".idea": true,
	}

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if strings.HasSuffix(name, ".java") &&
			!strings.HasSuffix(name, "Test.java") &&
			!strings.HasSuffix(name, "Tests.java") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// parseFile parses a single .java file and extracts components into the codebase.
func (p *Parser) parseFile(filePath string, codebase *types.Codebase) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	fileContent := string(content)

	p.extractSpringControllers(fileContent, filePath, codebase)
	p.extractSpringEntities(fileContent, filePath, codebase)
	p.extractSpringRepositories(fileContent, filePath, codebase)

	return nil
}

// controllerInfo holds extracted controller metadata.
type controllerInfo struct {
	ClassName string
	BasePath  string
	BodyStart int // index in content where class body starts
	BodyEnd   int // index in content where class body ends
	FileLine  int // line number of the controller annotation
}

// extractSpringControllers finds @RestController/@Controller classes and their routes.
func (p *Parser) extractSpringControllers(content, filePath string, codebase *types.Codebase) {
	// Find all controller annotation positions
	matches := controllerAnnotationRe.FindAllStringIndex(content, -1)

	for _, match := range matches {
		ctrl := p.findControllerInfo(content, match[0])
		if ctrl == nil {
			continue
		}

		classBody := content[ctrl.BodyStart:ctrl.BodyEnd]

		// Find all method-level mapping annotations and their associated methods
		p.extractMethodRoutes(classBody, ctrl, filePath, codebase)
	}
}

// findControllerInfo locates the class declaration and class body for a controller.
func (p *Parser) findControllerInfo(content string, annotationPos int) *controllerInfo {
	// Search forward from the annotation for 'class ClassName'
	remaining := content[annotationPos:]
	classMatch := classRe.FindStringSubmatchIndex(remaining)
	if classMatch == nil {
		return nil
	}

	className := remaining[classMatch[2]:classMatch[3]]
	classEnd := annotationPos + classMatch[1]

	// Look for @RequestMapping between the annotation and class declaration
	basePath := ""
	betweenClass := content[annotationPos:classEnd]
	if rmMatch := requestMappingRe.FindStringSubmatch(betweenClass); rmMatch != nil {
		basePath = rmMatch[1]
	}

	// Find the class body by matching braces
	afterClass := content[classEnd:]
	openBrace := strings.Index(afterClass, "{")
	if openBrace == -1 {
		return nil
	}

	bodyContent := afterClass[openBrace+1:]
	bodyLen := p.balancedBlockLength(bodyContent)
	if bodyLen < 0 {
		return nil
	}

	bodyStart := classEnd + openBrace + 1
	bodyEnd := bodyStart + bodyLen

	// Calculate line number (1-based)
	line := strings.Count(content[:annotationPos], "\n") + 1

	return &controllerInfo{
		ClassName: className,
		BasePath:  basePath,
		BodyStart: bodyStart,
		BodyEnd:   bodyEnd,
		FileLine:  line,
	}
}

// extractMethodRoutes finds method-level mapping annotations within a class body.
func (p *Parser) extractMethodRoutes(classBody string, ctrl *controllerInfo, filePath string, codebase *types.Codebase) {
	// Track which positions have already been processed (to avoid duplicates
	// when a no-path @GetMapping also matches the path version)
	processed := make(map[int]bool)

	// Step 1: Find @GetMapping("/path"), @PostMapping("/path"), etc. with explicit paths
	methodMatches := methodMappingRe.FindAllStringSubmatchIndex(classBody, -1)
	for _, mm := range methodMatches {
		processed[mm[0]] = true
		annotationType := classBody[mm[2]:mm[3]]
		methodPath := classBody[mm[4]:mm[5]]

		// Find the method name after the annotation
		afterAnnotation := classBody[mm[1]:]
		sigMatch := methodSigRe.FindStringSubmatchIndex(afterAnnotation)
		methodName := ""
		if sigMatch != nil {
			methodName = afterAnnotation[sigMatch[2]:sigMatch[3]]
		}

		httpMethod := mapSpringMethod(annotationType)
		fullPath := combinePaths(ctrl.BasePath, methodPath)

		lineOffset := strings.Count(classBody[:mm[0]], "\n")

		p.addRouteAndHandler(codebase, ctrl, filePath, httpMethod, fullPath, methodName, lineOffset)
	}

	// Step 2: Find @GetMapping, @PostMapping, etc. WITHOUT explicit paths (default to "/")
	noPathMatches := methodMappingNoPathRe.FindAllStringSubmatchIndex(classBody, -1)
	for _, nm := range noPathMatches {
		// Skip if this position was already handled as a path-annotation
		if processed[nm[0]] {
			continue
		}
		// Skip if this annotation has parens with content (already handled above)
		matchText := classBody[nm[0]:nm[1]]
		if strings.Contains(matchText, `"`) || strings.Contains(matchText, `'`) {
			continue
		}

		annotationType := classBody[nm[2]:nm[3]]

		// Find the method name after the annotation
		afterAnnotation := classBody[nm[1]:]
		sigMatch := methodSigRe.FindStringSubmatchIndex(afterAnnotation)
		methodName := ""
		if sigMatch != nil {
			methodName = afterAnnotation[sigMatch[2]:sigMatch[3]]
		}

		httpMethod := mapSpringMethod(annotationType)
		fullPath := combinePaths(ctrl.BasePath, "/")

		lineOffset := strings.Count(classBody[:nm[0]], "\n")

		p.addRouteAndHandler(codebase, ctrl, filePath, httpMethod, fullPath, methodName, lineOffset)
	}

	// Step 3: Find @RequestMapping methods (which may have method=RequestMethod.GET)
	reqMapMatches := requestMappingRe.FindAllStringSubmatchIndex(classBody, -1)
	for _, rm := range reqMapMatches {
		methodPath := classBody[rm[2]:rm[3]]
		methodAttr := ""
		if rm[4] >= 0 && rm[5] >= 0 {
			methodAttr = classBody[rm[4]:rm[5]]
		}

		// Skip if no path (class-level @RequestMapping handled elsewhere)
		if methodPath == "" {
			continue
		}

		httpMethod := "REQUEST"
		if methodAttr != "" {
			httpMethod = strings.ToUpper(methodAttr)
		}

		// Find the method name
		afterAnnotation := classBody[rm[1]:]
		sigMatch := methodSigRe.FindStringSubmatchIndex(afterAnnotation)
		methodName := ""
		if sigMatch != nil {
			methodName = afterAnnotation[sigMatch[2]:sigMatch[3]]
		}

		fullPath := combinePaths(ctrl.BasePath, methodPath)
		lineOffset := strings.Count(classBody[:rm[0]], "\n")

		p.addRouteAndHandler(codebase, ctrl, filePath, httpMethod, fullPath, methodName, lineOffset)
	}
}

// addRouteAndHandler adds a route and handler to the codebase.
func (p *Parser) addRouteAndHandler(codebase *types.Codebase, ctrl *controllerInfo, filePath, httpMethod, fullPath, methodName string, lineOffset int) {
	route := types.Route{
		Method: httpMethod,
		Path:   fullPath,
		Handler: func() string {
			if methodName != "" {
				return ctrl.ClassName + "." + methodName
			}
			return ctrl.ClassName
		}(),
		File: filePath,
		Line: ctrl.FileLine + lineOffset,
	}
	codebase.Routes = append(codebase.Routes, route)

	// Also add as a handler
	if methodName != "" {
		handler := types.Handler{
			Name:       methodName,
			Controller: ctrl.ClassName,
			Method:     httpMethod,
			File:       filePath,
			Line:       ctrl.FileLine + lineOffset,
		}
		codebase.Handlers = append(codebase.Handlers, handler)
	}
}

// extractSpringEntities finds @Entity classes with @Table and @Column annotations.
func (p *Parser) extractSpringEntities(content, filePath string, codebase *types.Codebase) {
	// Find all @Entity annotations
	matches := entityRe.FindAllStringIndex(content, -1)

	for _, match := range matches {
		remaining := content[match[0]:]

		// Find the class declaration
		classMatch := classRe.FindStringSubmatchIndex(remaining)
		if classMatch == nil {
			continue
		}

		className := remaining[classMatch[2]:classMatch[3]]
		classEndOffset := match[0] + classMatch[1]

		// Look for @Table annotation between @Entity and class
		tableName := ""
		between := content[match[0]:classEndOffset]
		if tMatch := tableRe.FindStringSubmatch(between); tMatch != nil {
			tableName = tMatch[1]
		}

		// Find the class body
		afterClass := content[classEndOffset:]
		openBrace := strings.Index(afterClass, "{")
		if openBrace == -1 {
			continue
		}

		bodyContent := afterClass[openBrace+1:]
		bodyLen := p.balancedBlockLength(bodyContent)
		if bodyLen < 0 {
			continue
		}

		classBodyStart := classEndOffset + openBrace + 1
		classBodyEnd := classBodyStart + bodyLen
		classBody := content[classBodyStart:classBodyEnd]

		// Extract @Column fields
		var fields []types.Field
		columnMatches := columnRe.FindAllStringSubmatchIndex(classBody, -1)
		for _, cm := range columnMatches {
			colName := classBody[cm[2]:cm[3]]
			fieldType := classBody[cm[4]:cm[5]]
			fieldName := classBody[cm[6]:cm[7]]

			fields = append(fields, types.Field{
				Name: fieldName,
				Type: fieldType,
			})

			_ = colName // database column name (unused for now, can be added later)
		}

		// If no @Column annotations found, try plain field declarations
		if len(fields) == 0 {
			fields = p.extractPlainFields(classBody)
		}

		line := strings.Count(content[:match[0]], "\n") + 1

		model := types.Model{
			Name:   className,
			Table:  tableName,
			Fields: fields,
			File:   filePath,
			Line:   line,
		}
		codebase.Models = append(codebase.Models, model)
	}
}

// extractPlainFields finds simple field declarations (private Type name;) without annotations.
func (p *Parser) extractPlainFields(classBody string) []types.Field {
	fieldRe := regexp.MustCompile(`(?:private|protected|public)\s+(\S+)\s+(\w+)\s*;`)

	var fields []types.Field
	matches := fieldRe.FindAllStringSubmatch(classBody, -1)
	for _, m := range matches {
		fieldType := m[1]
		fieldName := m[2]

		// Skip serialVersionUID and logger fields
		if fieldName == "serialVersionUID" || fieldType == "Logger" || fieldType == "log" {
			continue
		}

		fields = append(fields, types.Field{
			Name: fieldName,
			Type: fieldType,
		})
	}

	return fields
}

// extractSpringRepositories finds @Repository interfaces and adds them as handlers.
func (p *Parser) extractSpringRepositories(content, filePath string, codebase *types.Codebase) {
	matches := repositoryRe.FindAllStringSubmatchIndex(content, -1)

	for _, m := range matches {
		interfaceName := content[m[2]:m[3]]

		line := strings.Count(content[:m[0]], "\n") + 1

		handler := types.Handler{
			Name:       interfaceName,
			Controller: interfaceName + "Repository",
			Method:     "REPOSITORY",
			File:       filePath,
			Line:       line,
		}
		codebase.Handlers = append(codebase.Handlers, handler)
	}
}

// balancedBlockLength returns the length of content up to and excluding the matching closing brace.
// Returns -1 if no matching brace is found.
func (p *Parser) balancedBlockLength(content string) int {
	depth := 1
	for i, ch := range content {
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// combinePaths joins a base path and method path, normalizing slashes.
func combinePaths(base, method string) string {
	base = strings.TrimRight(base, "/")
	// Root path: no method path to append
	if method == "/" || method == "" {
		if base == "" {
			return "/"
		}
		return base
	}
	if !strings.HasPrefix(method, "/") {
		method = "/" + method
	}
	return base + method
}

// mapSpringMethod converts a Spring annotation type to an HTTP method string.
func mapSpringMethod(annotation string) string {
	switch annotation {
	case "GetMapping":
		return "GET"
	case "PostMapping":
		return "POST"
	case "PutMapping":
		return "PUT"
	case "DeleteMapping":
		return "DELETE"
	case "PatchMapping":
		return "PATCH"
	}
	return "REQUEST"
}
