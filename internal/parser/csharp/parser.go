package csharp

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

// Parser implements the Parser interface for C# / ASP.NET Core.
// Unlike other parsers in this project, this parser uses regex + file scanning
// because go-tree-sitter does not provide a C# language binding.
type Parser struct {
}

// NewParser creates a new C# parser.
func NewParser() *Parser {
	return &Parser{}
}

// SupportedLanguages returns the languages this parser supports.
func (p *Parser) SupportedLanguages() []string {
	return []string{"csharp"}
}

// DetectFramework detects whether the project is ASP.NET Core by scanning
// .csproj files for Microsoft.NET.Sdk.Web or Microsoft.AspNetCore references.
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return "", fmt.Errorf("failed to read root directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".csproj") {
			continue
		}

		csprojPath := filepath.Join(rootDir, entry.Name())
		content, err := os.ReadFile(csprojPath)
		if err != nil {
			continue
		}

		c := string(content)
		if strings.Contains(c, `Microsoft.NET.Sdk.Web`) ||
			strings.Contains(c, "Microsoft.AspNetCore") {
			return "aspnetcore", nil
		}
	}

	return "csharp", nil
}

// Parse analyzes a C# codebase from the root directory.
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language: "csharp",
		RootDir:  rootDir,
	}

	framework, err := p.DetectFramework(rootDir)
	if err == nil {
		codebase.Framework = framework
	}

	files, err := p.findSourceFiles(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find source files: %w", err)
	}
	codebase.FileCount = len(files)

	for _, file := range files {
		select {
		case <-ctx.Done():
			return codebase, ctx.Err()
		default:
		}

		if err := p.parseFile(codebase, file); err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", file, err)
		}
	}

	return codebase, nil
}

// ---------------------------------------------------------------------------
// File discovery
// ---------------------------------------------------------------------------

// skipDirs lists directory names that should not be traversed.
var skipDirs = map[string]bool{
	"bin":          true,
	"obj":          true,
	"node_modules": true,
	".git":         true,
	"packages":     true,
	"TestResults":  true,
	".vs":          true,
	"Properties":   true,
	"Migrations":   true,
}

func (p *Parser) findSourceFiles(rootDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(info.Name(), ".cs") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// ---------------------------------------------------------------------------
// Regex patterns
// ---------------------------------------------------------------------------

var (
	// Patterns for controller detection.
	reClassDef         = regexp.MustCompile(`^(\s*)(?:public\s+)?(?:partial\s+)?class\s+(\w+)\s*(?::\s*(\w+(?:\.\w+)*))?`)
	reApiController    = regexp.MustCompile(`\[ApiController\]`)
	reRouteAttr        = regexp.MustCompile(`\[Route\("([^"]+)"\)\]`)
	reRouteAttrBracket = regexp.MustCompile(`\[Route\(@"([^"]+)"\)\]`)

	// Patterns for HTTP method attributes on methods.
	reHttpMethodAttr = regexp.MustCompile(`^\s*\[(HttpGet|HttpPost|HttpPut|HttpDelete|HttpPatch)(?:\(\s*"(?:~?/?)?([^"]*)"\s*\))?\s*\]`)

	// Patterns for Minimal API.
	reMapMethod = regexp.MustCompile(`\b(app|api|application|endpoints)\s*\.\s*(MapGet|MapPost|MapPut|MapDelete|MapPatch|MapMethods|Map)\s*\(\s*"([^"]+)"`)

	// Patterns for model detection (class + auto-properties).
	rePublicProperty = regexp.MustCompile(`^\s*public\s+(\S+(?:<[^>]+>)?)\??\s+(\w+)\s*\{\s*get;\s*(?:private\s+)?set;\s*\}`)
)

// ---------------------------------------------------------------------------
// File parsing
// ---------------------------------------------------------------------------

func (p *Parser) parseFile(codebase *types.Codebase, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Increase buffer for long attribute lines.
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	// State for controller scanning.
	type controllerCtx struct {
		name       string
		classRoute string
		braceDepth int // depth after the opening brace was seen
		startLine  int
		entered    bool // true once we have seen the opening {
	}
	var ctl *controllerCtx

	// Recent lines buffer – we keep up to 3 preceding lines so we can back-look
	// for [ApiController] or [Route] when we encounter the class declaration.
	ring := make([]string, 3)
	ringIdx := 0

	// Track whether the current file is inside a Models/ directory.
	inModelsDir := strings.Contains(filepath.ToSlash(filePath), "/Models/") ||
		strings.Contains(filepath.ToSlash(filePath), "/Model/")

	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := scanner.Text()

		// --- Minimal API detection (file-wide, simple regex) ---
		if matches := reMapMethod.FindAllStringSubmatch(line, -1); matches != nil {
			for _, m := range matches {
				methodName := m[2] // MapGet, MapPost, …
				path := m[3]
				httpMethod := mapMethodToHTTP(methodName)
				codebase.Routes = append(codebase.Routes, types.Route{
					Method: httpMethod,
					Path:   path,
					File:   filePath,
					Line:   lineNo,
				})
			}
		}

		// --- Controller detection ---

		// If we are inside a controller, look for [Http*] attributes.
		if ctl != nil {
			// Track brace depth.
			delta := braceDelta(line)
			ctl.braceDepth += delta
			if !ctl.entered && ctl.braceDepth > 0 {
				ctl.entered = true
			}

			// Stop when the class closes (only after we've entered it).
			if ctl.entered && ctl.braceDepth <= 0 {
				ctl = nil
				goto ringPush
			}

			if m := reHttpMethodAttr.FindStringSubmatch(line); m != nil {
				verb := m[1]  // HttpGet, HttpPost, …
				mPath := m[2] // optional path argument
				fullPath := joinRoutePrefixes(ctl.classRoute, mPath)
				codebase.Routes = append(codebase.Routes, types.Route{
					Method: mapVerbToHTTP(verb),
					Path:   fullPath,
					File:   filePath,
					Line:   lineNo,
				})
			}
			goto ringPush
		}

		// Check whether this line is a class definition.
		if cm := reClassDef.FindStringSubmatch(line); cm != nil {
			className := cm[2]
			baseClass := cm[3]

			// Check recent ring lines + current line for [ApiController] or
			// [Route].
			isController := strings.EqualFold(baseClass, "ControllerBase") ||
				strings.EqualFold(baseClass, "Controller")

			var classRoute string

			// Scan the ring buffer backwards for attributes (newest first).
			for i := 0; i < 3; i++ {
				prev := ring[(ringIdx-1-i+3)%3]
				if reApiController.MatchString(prev) {
					isController = true
				}
				if rm := reRouteAttr.FindStringSubmatch(prev); rm != nil {
					classRoute = rm[1]
				}
				if rm := reRouteAttrBracket.FindStringSubmatch(prev); rm != nil {
					classRoute = rm[1]
				}
			}
			// Also check the class line itself for Route attribute (unlikely
			// but possible in attribute-on-same-line patterns).
			if rm := reRouteAttr.FindStringSubmatch(line); rm != nil {
				classRoute = rm[1]
			}

			if isController {
				// The opening brace may be on this line or on a future line.
				// We don't discard the controller until we've actually entered
				// the class body (seen an opening brace) and exited it.
				bd := braceDelta(line)
				ctl = &controllerCtx{
					name:       className,
					classRoute: classRoute,
					braceDepth: bd,
					startLine:  lineNo,
					entered:    bd > 0,
				}

				// Single-line class like "class X : ControllerBase { }" –
				// brace delta is 0, so we are both entered and immediately
				// closed. Discard.
				if ctl.entered && ctl.braceDepth <= 0 {
					ctl = nil
				}
			}
		}

	ringPush:
		// --- Model detection (only for files in Models/ directories) ---
		if inModelsDir {
			if cm := reClassDef.FindStringSubmatch(line); cm != nil {
				modelName := cm[2]
				// We will collect properties below; first emit the model stub.
				model := types.Model{
					Name: modelName,
					File: filePath,
					Line: lineNo,
				}
				codebase.Models = append(codebase.Models, model)
			}

			if pm := rePublicProperty.FindStringSubmatch(line); pm != nil {
				propType := pm[1]
				propName := pm[2]
				if len(codebase.Models) > 0 {
					last := &codebase.Models[len(codebase.Models)-1]
					last.Fields = append(last.Fields, types.Field{
						Name: propName,
						Type: propType,
					})
				}
			}
		}

		// Rotate ring buffer.
		ring[ringIdx] = line
		ringIdx = (ringIdx + 1) % 3
	}

	return scanner.Err()
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// braceDelta returns the net change in brace depth for a single line (>0 = more
// opens, <0 = more closes).
func braceDelta(line string) int {
	d := 0
	for _, ch := range line {
		switch ch {
		case '{':
			d++
		case '}':
			d--
		}
	}
	return d
}

// mapVerbToHTTP converts an attribute verb name to its HTTP method.
func mapVerbToHTTP(verb string) string {
	switch verb {
	case "HttpGet":
		return "GET"
	case "HttpPost":
		return "POST"
	case "HttpPut":
		return "PUT"
	case "HttpDelete":
		return "DELETE"
	case "HttpPatch":
		return "PATCH"
	default:
		return strings.ToUpper(strings.TrimPrefix(verb, "Http"))
	}
}

// mapMethodToHTTP converts a Minimal API method name to its HTTP method.
func mapMethodToHTTP(name string) string {
	switch name {
	case "MapGet":
		return "GET"
	case "MapPost":
		return "POST"
	case "MapPut":
		return "PUT"
	case "MapDelete":
		return "DELETE"
	case "MapPatch":
		return "PATCH"
	case "Map":
		return "ANY"
	default:
		return strings.ToUpper(strings.TrimPrefix(name, "Map"))
	}
}

// joinRoutePrefixes combines a class-level [Route] prefix with a method-level
// path. Handles leading slashes, trailing slashes, and the ~/ virtual-root
// prefix.
func joinRoutePrefixes(classRoute, methodPath string) string {
	// Strip the ~/  prefix that ASP.NET uses for relative app-root paths.
	methodPath = strings.TrimPrefix(methodPath, "~/")
	methodPath = strings.TrimPrefix(methodPath, "/")

	if classRoute == "" {
		return "/" + methodPath
	}

	classRoute = strings.TrimSuffix(classRoute, "/")

	if methodPath == "" {
		return classRoute
	}

	return classRoute + "/" + methodPath
}
