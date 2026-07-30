# Phase 4 Implementation Guide: Multi-Language Support & Advanced AI

**Status**: Ready for Implementation  
**Date**: July 31, 2026  
**Goal**: Expand language support and add advanced AI-driven features

---

## Overview

Phase 4 expands language support to 8+ languages and adds advanced AI capabilities including test optimization, code review, and flakiness detection.

### Key Features
1. Ruby Rails Parser (routes, models, controllers)
2. Java Spring Boot Parser (controllers, models, services)
3. C# ASP.NET Core Parser (controllers, models, services)
4. TypeScript Express & NestJS Parsers
5. Rust Actix & Rocket Parsers
6. AI-Driven Test Optimization
7. Intelligent Test Generation (GPT-4)
8. Code Review Assistant
9. Test Coverage Insights
10. Test Flakiness Detection

---

## Sprint 17-18: Ruby Rails Support (Weeks 33-36)

### Task 17.1: Ruby Rails Parser

**File:** `internal/parser/ruby/rails.go`

```go
package ruby

import (
  "context"
  "regexp"
  "strings"

  "github.com/go-go-golems/gotest-agent/internal/parser/types"
  sit "github.com/smacker/go-tree-sitter"
  "github.com/smacker/go-tree-sitter/ruby"
)

type RailsParser struct {
  language *sit.Language
}

func NewRailsParser() *RailsParser {
  return &RailsParser{language: ruby.GetLanguage()}
}

func (p *RailsParser) ParseRoutes(ctx context.Context, file *types.File) ([]types.Route, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)
  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }
  var routes []types.Route
  p.findRoutes(tree.RootNode(), []byte(file.Content), &routes, file.Path)
  return routes, nil
}

func (p *RailsParser) findRoutes(node *sit.Node, source []byte, routes *[]types.Route, filePath string) {
  if node.Type() == "call" {
    methodName := p.getNodeText(node.Child(0), source)
    methodMap := map[string]string{
      "get": "GET", "post": "POST", "put": "PUT",
      "patch": "PATCH", "delete": "DELETE",
    }
    if httpMethod, ok := methodMap[methodName]; ok {
      args := node.NamedChild(0)
      if args != nil && args.NamedChildCount() > 0 {
        path := strings.Trim(p.getNodeText(args.NamedChild(0), source), "\"'")
        controller, action := "", ""
        if args.NamedChildCount() > 1 {
          actionText := p.getNodeText(args.NamedChild(1), source)
          if strings.Contains(actionText, "#") {
            parts := strings.Split(actionText, "#")
            controller, action = parts[0], parts[1]
          }
        }
        *routes = append(*routes, types.Route{
          Method: httpMethod, Path: "/" + path,
          Handler: controller + "#" + action,
          File: filePath, Line: int(node.StartPoint().Row) + 1,
        })
      }
    }
    if methodName == "resources" {
      args := node.NamedChild(0)
      if args != nil && args.NamedChildCount() > 0 {
        resourceName := strings.Trim(p.getNodeText(args.NamedChild(0), source), "\"'")
        *routes = append(*routes, p.generateResourceRoutes(resourceName, filePath, int(node.StartPoint().Row)+1)...)
      }
    }
  }
  for i := 0; i < int(node.NamedChildCount()); i++ {
    p.findRoutes(node.NamedChild(i), source, routes, filePath)
  }
}

func (p *RailsParser) generateResourceRoutes(name, file string, line int) []types.Route {
  return []types.Route{
    {Method: "GET", Path: "/" + name, Handler: name + "#index", File: file, Line: line},
    {Method: "GET", Path: "/" + name + "/new", Handler: name + "#new", File: file, Line: line},
    {Method: "POST", Path: "/" + name, Handler: name + "#create", File: file, Line: line},
    {Method: "GET", Path: "/" + name + "/:id", Handler: name + "#show", File: file, Line: line},
    {Method: "GET", Path: "/" + name + "/:id/edit", Handler: name + "#edit", File: file, Line: line},
    {Method: "PATCH", Path: "/" + name + "/:id", Handler: name + "#update", File: file, Line: line},
    {Method: "DELETE", Path: "/" + name + "/:id", Handler: name + "#destroy", File: file, Line: line},
  }
}

func (p *RailsParser) ParseModels(ctx context.Context, file *types.File) ([]types.Model, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)
  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }
  var models []types.Model
  p.findModels(tree.RootNode(), []byte(file.Content), &models, file.Path)
  return models, nil
}

func (p *RailsParser) findModels(node *sit.Node, source []byte, models *[]types.Model, filePath string) {
  if node.Type() == "class" {
    className := p.getNodeText(node.Child(0), source)
    superClass := p.findSuperClass(node, source)
    if superClass == "ApplicationRecord" || superClass == "ActiveRecord::Base" {
      *models = append(*models, types.Model{
        Name: className, Table: p.tableize(className),
        Fields: p.extractAssociations(node, source),
        File: filePath, Line: int(node.StartPoint().Row) + 1,
      })
    }
  }
  for i := 0; i < int(node.NamedChildCount()); i++ {
    p.findModels(node.NamedChild(i), source, models, filePath)
  }
}

func (p *RailsParser) extractAssociations(classNode *sit.Node, source []byte) []types.Field {
  var fields []types.Field
  body := classNode.ChildByFieldName("body")
  if body == nil {
    return fields
  }
  for i := 0; i < int(body.NamedChildCount()); i++ {
    stmt := body.NamedChild(i)
    if stmt.Type() == "call" {
      methodName := p.getNodeText(stmt.Child(0), source)
      if methodName == "belongs_to" || methodName == "has_many" || methodName == "has_one" {
        args := stmt.NamedChild(0)
        if args != nil && args.NamedChildCount() > 0 {
          fields = append(fields, types.Field{
            Name: strings.Trim(p.getNodeText(args.NamedChild(0), source), "\"'"),
            Type: methodName,
          })
        }
      }
    }
  }
  return fields
}

func (p *RailsParser) ParseControllers(ctx context.Context, file *types.File) ([]types.Handler, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)
  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }
  var handlers []types.Handler
  p.findControllers(tree.RootNode(), []byte(file.Content), &handlers, file.Path)
  return handlers, nil
}

func (p *RailsParser) findControllers(node *sit.Node, source []byte, handlers *[]types.Handler, filePath string) {
  if node.Type() == "class" {
    className := p.getNodeText(node.Child(0), source)
    if strings.HasSuffix(className, "Controller") {
      body := node.ChildByFieldName("body")
      if body != nil {
        for i := 0; i < int(body.NamedChildCount()); i++ {
          stmt := body.NamedChild(i)
          if stmt.Type() == "method" {
            methodName := p.getNodeText(stmt.Child(0), source)
            if !strings.HasPrefix(methodName, "_") && methodName != "initialize" &&
               !strings.HasPrefix(methodName, "before_") && !strings.HasPrefix(methodName, "after_") {
              *handlers = append(*handlers, types.Handler{
                Name: className + "#" + methodName,
                File: filePath, Line: int(stmt.StartPoint().Row) + 1,
              })
            }
          }
        }
      }
    }
  }
  for i := 0; i < int(node.NamedChildCount()); i++ {
    p.findControllers(node.NamedChild(i), source, handlers, filePath)
  }
}

func (p *RailsParser) getNodeText(node *sit.Node, source []byte) string {
  if node == nil { return "" }
  return string(source[node.StartByte():node.EndByte()])
}

func (p *RailsParser) findSuperClass(node *sit.Node, source []byte) string {
  sc := node.ChildByFieldName("superclass")
  if sc == nil { return "" }
  return p.getNodeText(sc, source)
}

func (p *RailsParser) tableize(s string) string {
  result := ""
  for i, r := range s {
    if i > 0 && r >= 'A' && r <= 'Z' { result += "_" }
    result += strings.ToLower(string(r))
  }
  return result + "s"
}
```

**Acceptance Criteria:**
- Rails routes extracted (including `resources` → 7 RESTful routes)
- Rails models parsed with associations (belongs_to, has_many, has_one)
- Controller actions extracted (skip private methods and callbacks)

---

## Sprint 19-20: Java & C# Support (Weeks 37-40)

### Task 19.1: Java Spring Boot Parser

**File:** `internal/parser/java/spring.go`

```go
package java

import (
  "context"
  "regexp"
  "strings"

  "github.com/go-go-golems/gotest-agent/internal/parser/types"
  sit "github.com/smacker/go-tree-sitter"
  "github.com/smacker/go-tree-sitter/java"
)

type SpringParser struct {
  language *sit.Language
}

func NewSpringParser() *SpringParser {
  return &SpringParser{language: java.GetLanguage()}
}

func (p *SpringParser) ParseRoutes(ctx context.Context, file *types.File) ([]types.Route, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)
  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }
  var routes []types.Route
  p.findRoutes(tree.RootNode(), []byte(file.Content), &routes, file.Path)
  return routes, nil
}

func (p *SpringParser) findRoutes(node *sit.Node, source []byte, routes *[]types.Route, filePath string) {
  if node.Type() == "method_declaration" {
    annotations := p.findAnnotations(node, source)
    for _, annotation := range annotations {
      method, path := p.extractRouteInfo(annotation)
      if method != "" && path != "" {
        funcName := p.getNodeText(node.ChildByFieldName("name"), source)
        *routes = append(*routes, types.Route{
          Method: method, Path: path, Handler: funcName,
          File: filePath, Line: int(node.StartPoint().Row) + 1,
        })
      }
    }
  }
  for i := 0; i < int(node.NamedChildCount()); i++ {
    p.findRoutes(node.NamedChild(i), source, routes, filePath)
  }
}

func (p *SpringParser) findAnnotations(node *sit.Node, source []byte) []string {
  var annotations []string
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    if child.Type() == "marker_annotation" || child.Type() == "annotation" {
      annotations = append(annotations, p.getNodeText(child, source))
    }
  }
  return annotations
}

func (p *SpringParser) extractRouteInfo(annotation string) (string, string) {
  annotationMap := map[string]string{
    "@GetMapping": "GET", "@PostMapping": "POST",
    "@PutMapping": "PUT", "@DeleteMapping": "DELETE",
    "@PatchMapping": "PATCH",
  }
  for annotationType, method := range annotationMap {
    if strings.Contains(annotation, annotationType) {
      re := regexp.MustCompile(`\("([^"]+)"\)`)
      matches := re.FindStringSubmatch(annotation)
      if len(matches) >= 2 {
        return method, matches[1]
      }
    }
  }
  if strings.Contains(annotation, "@RequestMapping") {
    method := "GET"
    if strings.Contains(annotation, "RequestMethod.POST") { method = "POST" }
    else if strings.Contains(annotation, "RequestMethod.PUT") { method = "PUT" }
    else if strings.Contains(annotation, "RequestMethod.DELETE") { method = "DELETE" }
    re := regexp.MustCompile(`value\s*=\s*"([^"]+)"`)
    matches := re.FindStringSubmatch(annotation)
    if len(matches) >= 2 {
      return method, matches[1]
    }
  }
  return "", ""
}

func (p *SpringParser) ParseModels(ctx context.Context, file *types.File) ([]types.Model, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)
  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }
  var models []types.Model
  p.findModels(tree.RootNode(), []byte(file.Content), &models, file.Path)
  return models, nil
}

func (p *SpringParser) findModels(node *sit.Node, source []byte, models *[]types.Model, filePath string) {
  if node.Type() == "class_declaration" {
    annotations := p.findAnnotations(node, source)
    for _, annotation := range annotations {
      if strings.Contains(annotation, "@Entity") {
        className := p.getNodeText(node.ChildByFieldName("name"), source)
        *models = append(*models, types.Model{
          Name: className, Table: p.toSnakeCase(className),
          Fields: p.extractFields(node, source),
          File: filePath, Line: int(node.StartPoint().Row) + 1,
        })
        break
      }
    }
  }
  for i := 0; i < int(node.NamedChildCount()); i++ {
    p.findModels(node.NamedChild(i), source, models, filePath)
  }
}

func (p *SpringParser) extractFields(classNode *sit.Node, source []byte) []types.Field {
  var fields []types.Field
  body := classNode.ChildByFieldName("body")
  if body == nil { return fields }
  for i := 0; i < int(body.NamedChildCount()); i++ {
    stmt := body.NamedChild(i)
    if stmt.Type() == "field_declaration" {
      fieldName := p.getNodeText(stmt.ChildByFieldName("name"), source)
      fieldType := p.extractFieldType(stmt, source)
      fields = append(fields, types.Field{Name: fieldName, Type: fieldType})
    }
  }
  return fields
}

func (p *SpringParser) extractFieldType(fieldNode *sit.Node, source []byte) string {
  typeNode := fieldNode.ChildByFieldName("type")
  if typeNode == nil { return "unknown" }
  typeText := p.getNodeText(typeNode, source)
  typeMap := map[string]string{
    "String": "string", "Integer": "integer", "Long": "integer",
    "Boolean": "boolean", "Date": "datetime",
    "LocalDate": "datetime", "LocalDateTime": "datetime",
  }
  if mapped, ok := typeMap[typeText]; ok { return mapped }
  return typeText
}

func (p *SpringParser) getNodeText(node *sit.Node, source []byte) string {
  if node == nil { return "" }
  return string(source[node.StartByte():node.EndByte()])
}

func (p *SpringParser) toSnakeCase(s string) string {
  result := ""
  for i, r := range s {
    if i > 0 && r >= 'A' && r <= 'Z' { result += "_" }
    result += strings.ToLower(string(r))
  }
  return result
}
```

**Acceptance Criteria:**
- Spring Boot routes extracted from @GetMapping, @PostMapping, etc.
- @Entity models parsed with fields and types
- Controller methods identified

---

### Task 19.2: C# ASP.NET Core Parser

**File:** `internal/parser/csharp/aspnet.go`

```go
package csharp

import (
  "context"
  "regexp"
  "strings"

  "github.com/go-go-golems/gotest-agent/internal/parser/types"
  sit "github.com/smacker/go-tree-sitter"
  "github.com/smacker/go-tree-sitter/c_sharp"
)

type AspNetParser struct {
  language *sit.Language
}

func NewAspNetParser() *AspNetParser {
  return &AspNetParser{language: c_sharp.GetLanguage()}
}

func (p *AspNetParser) ParseRoutes(ctx context.Context, file *types.File) ([]types.Route, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)
  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }
  var routes []types.Route
  p.findRoutes(tree.RootNode(), []byte(file.Content), &routes, file.Path)
  return routes, nil
}

func (p *AspNetParser) findRoutes(node *sit.Node, source []byte, routes *[]types.Route, filePath string) {
  if node.Type() == "method_declaration" {
    attributes := p.findAttributes(node, source)
    for _, attr := range attributes {
      method, path := p.extractRouteInfo(attr)
      if method != "" {
        funcName := p.getNodeText(node.ChildByFieldName("name"), source)
        *routes = append(*routes, types.Route{
          Method: method, Path: path, Handler: funcName,
          File: filePath, Line: int(node.StartPoint().Row) + 1,
        })
      }
    }
  }
  for i := 0; i < int(node.NamedChildCount()); i++ {
    p.findRoutes(node.NamedChild(i), source, routes, filePath)
  }
}

func (p *AspNetParser) findAttributes(node *sit.Node, source []byte) []string {
  var attrs []string
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    if child.Type() == "attribute_list" {
      attrs = append(attrs, p.getNodeText(child, source))
    }
  }
  return attrs
}

func (p *AspNetParser) extractRouteInfo(attr string) (string, string) {
  attrMap := map[string]string{
    "[HttpGet]": "GET", "[HttpPost]": "POST",
    "[HttpPut]": "PUT", "[HttpDelete]": "DELETE",
    "[HttpPatch]": "PATCH",
  }
  for attrType, method := range attrMap {
    if strings.Contains(attr, attrType) {
      re := regexp.MustCompile(`\("([^"]+)"\)`)
      matches := re.FindStringSubmatch(attr)
      path := "/"
      if len(matches) >= 2 { path = matches[1] }
      return method, path
    }
  }
  return "", ""
}

func (p *AspNetParser) getNodeText(node *sit.Node, source []byte) string {
  if node == nil { return "" }
  return string(source[node.StartByte():node.EndByte()])
}
```

**Acceptance Criteria:**
- ASP.NET Core routes extracted from [HttpGet], [HttpPost], etc.
- Controller methods identified
- Route paths extracted from attribute parameters

---

## Sprint 21-22: TypeScript & Rust Support (Weeks 41-44)

### Task 21.1: TypeScript Express Parser

**File:** `internal/parser/typescript/express.go`

Similar to JavaScript parser but with TypeScript type information. Uses `typescript` tree-sitter grammar.

### Task 21.2: TypeScript NestJS Parser

**File:** `internal/parser/typescript/nestjs.go`

Parse NestJS decorators: @Controller, @Get, @Post, @Put, @Delete, @Injectable

### Task 21.3: Rust Actix Parser

**File:** `internal/parser/rust/actix.go`

Parse Actix web routes: `web::get()`, `web::post()`, `.route()`, `.resource()`

### Task 21.4: Rust Rocket Parser

**File:** `internal/parser/rust/rocket.go`

Parse Rocket route macros: `#[get("/")]`, `#[post("/")]`, `#[route(GET, path = "/")]`

---

## Sprint 23-24: AI-Driven Test Optimization (Weeks 45-48)

### Task 23.1: Test Quality Analyzer

**File:** `internal/ai/quality_analyzer.go`

```go
package ai

import (
  "context"
  "fmt"

  "github.com/go-go-golems/gotest-agent/internal/parser/types"
)

type QualityAnalyzer struct {
  client Client
}

func NewQualityAnalyzer(client Client) *QualityAnalyzer {
  return &QualityAnalyzer{client: client}
}

type QualityReport struct {
  OverallScore    int              `json:"overall_score"`
  CoverageGaps    []string         `json:"coverage_gaps"`
  RedundantTests  []string         `json:"redundant_tests"`
  FlakyTests      []string         `json:"flaky_tests"`
  Suggestions     []string         `json:"suggestions"`
}

func (qa *QualityAnalyzer) Analyze(ctx context.Context, codebase *types.Codebase, tests []string) (*QualityReport, error) {
  prompt := fmt.Sprintf(`Analyze the quality of these tests for the given codebase.

CODEBASE:
Routes: %d
Models: %d
Handlers: %d

TESTS:
%s

Provide:
1. Overall quality score (0-100)
2. Coverage gaps (routes/models not tested)
3. Redundant tests (duplicate coverage)
4. Potentially flaky tests
5. Improvement suggestions

Return JSON format.`, len(codebase.Routes), len(codebase.Models), len(codebase.Handlers), formatTests(tests))

  response, err := qa.client.GenerateText(ctx, prompt)
  if err != nil {
    return nil, err
  }

  return parseQualityReport(response)
}
```

### Task 23.2: Intelligent Test Generation

**File:** `internal/ai/intelligent_generator.go`

Use GPT-4 for complex test scenarios, edge cases, negative tests, and performance tests.

### Task 23.3: Code Review Assistant

**File:** `internal/ai/code_reviewer.go`

Analyze code changes, suggest test cases, identify potential bugs.

---

## Sprint 25-26: Advanced Analytics (Weeks 49-52)

### Task 25.1: Test Coverage Insights

**File:** `internal/analytics/coverage.go`

Line coverage, branch coverage, mutation testing, coverage trends.

### Task 25.2: Test Performance Analytics

**File:** `internal/analytics/performance.go`

Execution time tracking, slow test identification, parallelization suggestions.

### Task 25.3: Test Flakiness Detection

**File:** `internal/analytics/flakiness.go`

Run tests multiple times, identify inconsistent tests, suggest fixes.

---

## Database Schema (Phase 4)

```sql
-- Language support registry
CREATE TABLE supported_languages (
    id UUID PRIMARY KEY,
    language VARCHAR(50) NOT NULL,
    framework VARCHAR(100),
    parser_status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Test quality reports
CREATE TABLE quality_reports (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL,
    overall_score INTEGER NOT NULL,
    coverage_gaps JSONB,
    redundant_tests JSONB,
    flaky_tests JSONB,
    suggestions JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Test analytics
CREATE TABLE test_analytics (
    id UUID PRIMARY KEY,
    test_name VARCHAR(255) NOT NULL,
    execution_count INTEGER DEFAULT 0,
    pass_count INTEGER DEFAULT 0,
    fail_count INTEGER DEFAULT 0,
    avg_duration_ms INTEGER,
    flakiness_score FLOAT,
    last_run_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT NOW()
);
```

---

## Summary

**Phase 4 Timeline:** 24 weeks (12 sprints)

**Key Deliverables:**
1. Ruby Rails parser
2. Java Spring Boot parser
3. C# ASP.NET Core parser
4. TypeScript Express & NestJS parsers
5. Rust Actix & Rocket parsers
6. AI-driven test optimization
7. Intelligent test generation
8. Code review assistant
9. Test coverage insights
10. Test performance analytics
11. Test flakiness detection

**Success Criteria:**
- 8+ languages supported
- 90% user satisfaction with AI-generated tests
- 95% test coverage maintained automatically
- Integration with 10+ external tools

---

*End of Phase 4 Implementation Guide*
