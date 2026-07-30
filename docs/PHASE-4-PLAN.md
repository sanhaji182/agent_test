# Phase 4: Advanced AI Features & Multi-Language Support

**Timeline:** 12 weeks (6 sprints)  
**Target:** Expand AI capabilities, add more language support, integrate with external tools  
**Success Criteria:** Support 8+ languages, AI-driven test optimization, 90% user satisfaction

---

## Overview

Phase 4 focuses on making GoTest Agent a truly advanced AI testing platform. We'll:

1. **Expand language support** beyond JavaScript, Go, PHP to include Python, Ruby, Java, C#, Rust, TypeScript
2. **Enhance AI capabilities** with advanced test optimization, code review, and intelligent suggestions
3. **Integrate with external tools** (CI/CD, code quality tools, monitoring)
4. **Provide advanced analytics** and insights for test quality
5. **Add enterprise features** for team collaboration and governance

---

## Sprint 17-18: Python & Ruby Support (Weeks 33-36)

### Goal
Add full support for Python (Django, Flask, FastAPI) and Ruby (Rails, Sinatra)

### Tasks

#### Task 17.1: Python Django Parser

**Objective:** Parse Django routes, models, and views

**Implementation:**
```go
// internal/parser/python/django.go
package python

import (
  "context"
  "regexp"
  "strings"

  "github.com/go-go-golems/gotest-agent/internal/parser/types"
  sit "github.com/smacker/go-tree-sitter"
)

type DjangoParser struct {
  language *sit.Language
}

func NewDjangoParser() *DjangoParser {
  return &DjangoParser{
    language: python.GetLanguage(),
  }
}

// ParseRoutes extracts Django URL patterns from urls.py
func (p *DjangoParser) ParseRoutes(ctx context.Context, file *types.File) ([]types.Route, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)

  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }

  var routes []types.Route

  // Find urlpatterns list
  urlpatterns := p.findURLPatterns(tree.RootNode(), []byte(file.Content))

  for _, pattern := range urlpatterns {
    route := types.Route{
      Method: pattern.Method,
      Path:   pattern.Path,
      Handler: pattern.View,
      File:   file.Path,
      Line:   pattern.Line,
    }
    routes = append(routes, route)
  }

  return routes, nil
}

// URLPattern represents a Django URL pattern
type URLPattern struct {
  Method string
  Path   string
  View   string
  Line   int
}

// findURLPatterns finds all URL patterns in Django urls.py
func (p *DjangoParser) findURLPatterns(node *sit.Node, source []byte) []URLPattern {
  var patterns []URLPattern

  // Find path() or re_path() calls
  if node.Type() == "call" {
    funcName := p.getNodeText(node.Child(0), source)
    
    if funcName == "path" || funcName == "re_path" {
      // Extract path and view
      args := node.NamedChild(0)
      if args != nil && args.NamedChildCount() >= 2 {
        pathNode := args.NamedChild(0)
        viewNode := args.NamedChild(1)

        path := p.getNodeText(pathNode, source)
        view := p.getNodeText(viewNode, source)

        // Clean up strings
        path = strings.Trim(path, "\"'")
        view = strings.TrimSpace(view)

        patterns = append(patterns, URLPattern{
          Method: "GET", // Django default
          Path:   "/" + path,
          View:   view,
          Line:   int(node.StartPoint().Row) + 1,
        })
      }
    }
  }

  // Recursively search children
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    childPatterns := p.findURLPatterns(child, source)
    patterns = append(patterns, childPatterns...)
  }

  return patterns
}

// ParseModels extracts Django model definitions
func (p *DjangoParser) ParseModels(ctx context.Context, file *types.File) ([]types.Model, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)

  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }

  var models []types.Model

  // Find class definitions that inherit from models.Model
  p.findModels(tree.RootNode(), []byte(file.Content), &models, file.Path)

  return models, nil
}

// findModels finds Django model class definitions
func (p *DjangoParser) findModels(node *sit.Node, source []byte, models *[]types.Model, filePath string) {
  if node.Type() == "class_definition" {
    className := p.getNodeText(node.Child(0), source)
    
    // Check if class inherits from models.Model
    superClass := p.findSuperClass(node, source)
    if superClass == "models.Model" || superClass == "Model" {
      model := types.Model{
        Name:   className,
        Table:  p.snakeCase(className),
        Fields: p.extractFields(node, source),
        File:   filePath,
        Line:   int(node.StartPoint().Row) + 1,
      }
      *models = append(*models, model)
    }
  }

  // Recursively search
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    p.findModels(child, source, models, filePath)
  }
}

// extractFields extracts model fields
func (p *DjangoParser) extractFields(classNode *sit.Node, source []byte) []types.Field {
  var fields []types.Field

  // Find field assignments in class body
  body := classNode.ChildByFieldName("body")
  if body == nil {
    return fields
  }

  for i := 0; i < int(body.NamedChildCount()); i++ {
    stmt := body.NamedChild(i)
    
    if stmt.Type() == "expression_statement" {
      expr := stmt.Child(0)
      if expr.Type() == "assignment" {
        fieldName := p.getNodeText(expr.Child(0), source)
        fieldValue := p.getNodeText(expr.Child(1), source)

        fieldType := p.extractFieldType(fieldValue)
        
        field := types.Field{
          Name: fieldName,
          Type: fieldType,
        }
        fields = append(fields, field)
      }
    }
  }

  return fields
}

// extractFieldType extracts field type from field definition
func (p *DjangoParser) extractFieldType(fieldValue string) string {
  // Extract field type from expressions like:
  // models.CharField(max_length=100)
  // models.IntegerField()
  // models.ForeignKey(User, on_delete=models.CASCADE)
  
  if strings.Contains(fieldValue, "CharField") {
    return "string"
  } else if strings.Contains(fieldValue, "IntegerField") {
    return "integer"
  } else if strings.Contains(fieldValue, "BooleanField") {
    return "boolean"
  } else if strings.Contains(fieldValue, "DateTimeField") {
    return "datetime"
  } else if strings.Contains(fieldValue, "ForeignKey") {
    return "foreign_key"
  } else if strings.Contains(fieldValue, "TextField") {
    return "text"
  }
  
  return "unknown"
}

// ParseViews extracts Django view functions
func (p *DjangoParser) ParseViews(ctx context.Context, file *types.File) ([]types.Handler, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)

  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }

  var handlers []types.Handler

  // Find function definitions that are views
  p.findViews(tree.RootNode(), []byte(file.Content), &handlers, file.Path)

  return handlers, nil
}

// findViews finds Django view functions
func (p *DjangoParser) findViews(node *sit.Node, source []byte, handlers *[]types.Handler, filePath string) {
  if node.Type() == "function_definition" {
    funcName := p.getNodeText(node.Child(0), source)
    
    // Check if function has request parameter (Django view signature)
    params := node.ChildByFieldName("parameters")
    if params != nil && params.NamedChildCount() > 0 {
      firstParam := p.getNodeText(params.NamedChild(0), source)
      
      if firstParam == "request" || strings.Contains(firstParam, "request") {
        handler := types.Handler{
          Name:   funcName,
          File:   filePath,
          Line:   int(node.StartPoint().Row) + 1,
        }
        *handlers = append(*handlers, handler)
      }
    }
  }

  // Recursively search
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    p.findViews(child, source, handlers, filePath)
  }
}

// Helper methods
func (p *DjangoParser) getNodeText(node *sit.Node, source []byte) string {
  if node == nil {
    return ""
  }
  return string(source[node.StartByte():node.EndByte()])
}

func (p *DjangoParser) findSuperClass(node *sit.Node, source []byte) string {
  superclasses := node.ChildByFieldName("superclasses")
  if superclasses == nil || superclasses.NamedChildCount() == 0 {
    return ""
  }
  return p.getNodeText(superclasses.NamedChild(0), source)
}

func (p *DjangoParser) snakeCase(s string) string {
  // Convert CamelCase to snake_case
  result := ""
  for i, r := range s {
    if i > 0 && r >= 'A' && r <= 'Z' {
      result += "_"
    }
    result += strings.ToLower(string(r))
  }
  return result
}
```

**Acceptance Criteria:**
- Django URL patterns extracted correctly
- Django models parsed with fields
- Django views identified
- Parser handles common Django patterns

---

#### Task 17.2: Python Flask Parser

**Objective:** Parse Flask routes and models

**Implementation:**
```go
// internal/parser/python/flask.go
package python

import (
  "context"
  "regexp"
  "strings"

  "github.com/go-go-golems/gotest-agent/internal/parser/types"
  sit "github.com/smacker/go-tree-sitter"
)

type FlaskParser struct {
  language *sit.Language
}

func NewFlaskParser() *FlaskParser {
  return &FlaskParser{
    language: python.GetLanguage(),
  }
}

// ParseRoutes extracts Flask route decorators
func (p *FlaskParser) ParseRoutes(ctx context.Context, file *types.File) ([]types.Route, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)

  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }

  var routes []types.Route

  // Find @app.route() decorators
  p.findRoutes(tree.RootNode(), []byte(file.Content), &routes, file.Path)

  return routes, nil
}

// findRoutes finds Flask route decorators
func (p *FlaskParser) findRoutes(node *sit.Node, source []byte, routes *[]types.Route, filePath string) {
  if node.Type() == "decorated_definition" {
    // Find decorator and function
    decorators := node.ChildByFieldName("decorator")
    funcDef := node.ChildByFieldName("definition")
    
    if decorators != nil && funcDef != nil {
      decoratorText := p.getNodeText(decorators, source)
      
      // Check if it's a route decorator
      if strings.Contains(decoratorText, "@app.route") || 
         strings.Contains(decoratorText, "@blueprint.route") {
        
        // Extract route path from decorator
        path := p.extractRoutePath(decoratorText)
        methods := p.extractMethods(decoratorText)
        funcName := p.getNodeText(funcDef.Child(0), source)
        
        for _, method := range methods {
          route := types.Route{
            Method:  method,
            Path:    path,
            Handler: funcName,
            File:    filePath,
            Line:    int(decorators.StartPoint().Row) + 1,
          }
          *routes = append(*routes, route)
        }
      }
    }
  }

  // Recursively search
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    p.findRoutes(child, source, routes, filePath)
  }
}

// extractRoutePath extracts route path from decorator
func (p *FlaskParser) extractRoutePath(decoratorText string) string {
  // Extract path from @app.route('/path')
  re := regexp.MustCompile(`@app\.route\(['"]([^'"]+)['"]`)
  matches := re.FindStringSubmatch(decoratorText)
  
  if len(matches) >= 2 {
    return matches[1]
  }
  
  return "/"
}

// extractMethods extracts HTTP methods from decorator
func (p *FlaskParser) extractMethods(decoratorText string) []string {
  // Extract methods from @app.route('/path', methods=['GET', 'POST'])
  re := regexp.MustCompile(`methods\s*=\s*\[([^\]]+)\]`)
  matches := re.FindStringSubmatch(decoratorText)
  
  if len(matches) >= 2 {
    methodsStr := matches[1]
    // Parse methods list
    methods := []string{}
    for _, m := range strings.Split(methodsStr, ",") {
      method := strings.Trim(strings.TrimSpace(m), "'\"")
      methods = append(methods, method)
    }
    return methods
  }
  
  // Default to GET if no methods specified
  return []string{"GET"}
}

// ParseModels extracts Flask/SQLAlchemy models
func (p *FlaskParser) ParseModels(ctx context.Context, file *types.File) ([]types.Model, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)

  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }

  var models []types.Model

  // Find SQLAlchemy model classes
  p.findModels(tree.RootNode(), []byte(file.Content), &models, file.Path)

  return models, nil
}

// findModels finds SQLAlchemy model classes
func (p *FlaskParser) findModels(node *sit.Node, source []byte, models *[]types.Model, filePath string) {
  if node.Type() == "class_definition" {
    className := p.getNodeText(node.Child(0), source)
    
    // Check if class inherits from db.Model
    superClass := p.findSuperClass(node, source)
    if superClass == "db.Model" || superClass == "Model" {
      model := types.Model{
        Name:   className,
        Table:  p.snakeCase(className),
        Fields: p.extractFields(node, source),
        File:   filePath,
        Line:   int(node.StartPoint().Row) + 1,
      }
      *models = append(*models, model)
    }
  }

  // Recursively search
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    p.findModels(child, source, models, filePath)
  }
}

// extractFields extracts SQLAlchemy column definitions
func (p *FlaskParser) extractFields(classNode *sit.Node, source []byte) []types.Field {
  var fields []types.Field

  body := classNode.ChildByFieldName("body")
  if body == nil {
    return fields
  }

  for i := 0; i < int(body.NamedChildCount()); i++ {
    stmt := body.NamedChild(i)
    
    if stmt.Type() == "expression_statement" {
      expr := stmt.Child(0)
      if expr.Type() == "assignment" {
        fieldName := p.getNodeText(expr.Child(0), source)
        fieldValue := p.getNodeText(expr.Child(1), source)

        // Check if it's a Column definition
        if strings.Contains(fieldValue, "Column") {
          fieldType := p.extractColumnType(fieldValue)
          
          field := types.Field{
            Name: fieldName,
            Type: fieldType,
          }
          fields = append(fields, field)
        }
      }
    }
  }

  return fields
}

// extractColumnType extracts column type from Column definition
func (p *FlaskParser) extractColumnType(fieldValue string) string {
  if strings.Contains(fieldValue, "String") || strings.Contains(fieldValue, "Text") {
    return "string"
  } else if strings.Contains(fieldValue, "Integer") {
    return "integer"
  } else if strings.Contains(fieldValue, "Boolean") {
    return "boolean"
  } else if strings.Contains(fieldValue, "DateTime") {
    return "datetime"
  } else if strings.Contains(fieldValue, "ForeignKey") {
    return "foreign_key"
  }
  
  return "unknown"
}

// ParseViews extracts Flask view functions
func (p *FlaskParser) ParseViews(ctx context.Context, file *types.File) ([]types.Handler, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)

  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }

  var handlers []types.Handler

  // Find view functions (functions with @app.route decorators)
  p.findViews(tree.RootNode(), []byte(file.Content), &handlers, file.Path)

  return handlers, nil
}

// findViews finds Flask view functions
func (p *FlaskParser) findViews(node *sit.Node, source []byte, handlers *[]types.Handler, filePath string) {
  if node.Type() == "decorated_definition" {
    decorators := node.ChildByFieldName("decorator")
    funcDef := node.ChildByFieldName("definition")
    
    if decorators != nil && funcDef != nil {
      decoratorText := p.getNodeText(decorators, source)
      
      // Check if it's a route decorator
      if strings.Contains(decoratorText, "@app.route") || 
         strings.Contains(decoratorText, "@blueprint.route") {
        
        funcName := p.getNodeText(funcDef.Child(0), source)
        
        handler := types.Handler{
          Name: funcName,
          File: filePath,
          Line: int(funcDef.StartPoint().Row) + 1,
        }
        *handlers = append(*handlers, handler)
      }
    }
  }

  // Recursively search
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    p.findViews(child, source, handlers, filePath)
  }
}

// Helper methods
func (p *FlaskParser) getNodeText(node *sit.Node, source []byte) string {
  if node == nil {
    return ""
  }
  return string(source[node.StartByte():node.EndByte()])
}

func (p *FlaskParser) findSuperClass(node *sit.Node, source []byte) string {
  superclasses := node.ChildByFieldName("superclasses")
  if superclasses == nil || superclasses.NamedChildCount() == 0 {
    return ""
  }
  return p.getNodeText(superclasses.NamedChild(0), source)
}

func (p *FlaskParser) snakeCase(s string) string {
  result := ""
  for i, r := range s {
    if i > 0 && r >= 'A' && r <= 'Z' {
      result += "_"
    }
    result += strings.ToLower(string(r))
  }
  return result
}
```

**Acceptance Criteria:**
- Flask routes extracted with decorators
- HTTP methods extracted correctly
- SQLAlchemy models parsed
- Flask views identified

---

#### Task 17.3: Python FastAPI Parser

**Objective:** Parse FastAPI routes and Pydantic models

**Implementation:** Similar to Flask parser but with FastAPI-specific decorators and Pydantic models

---

#### Task 17.4: Ruby Rails Parser

**Objective:** Parse Rails routes, models, and controllers

**Implementation:**
```go
// internal/parser/ruby/rails.go
package ruby

import (
  "context"
  "regexp"
  "strings"

  "github.com/go-go-golems/gotest-agent/internal/parser/types"
  sit "github.com/smacker/go-tree-sitter"
)

type RailsParser struct {
  language *sit.Language
}

func NewRailsParser() *RailsParser {
  return &RailsParser{
    language: ruby.GetLanguage(),
  }
}

// ParseRoutes extracts Rails routes from config/routes.rb
func (p *RailsParser) ParseRoutes(ctx context.Context, file *types.File) ([]types.Route, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)

  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }

  var routes []types.Route

  // Find Rails route DSL
  p.findRoutes(tree.RootNode(), []byte(file.Content), &routes, file.Path)

  return routes, nil
}

// findRoutes finds Rails route definitions
func (p *RailsParser) findRoutes(node *sit.Node, source []byte, routes *[]types.Route, filePath string) {
  // Find Rails route methods: get, post, put, patch, delete, resources
  if node.Type() == "call" {
    methodName := p.getNodeText(node.Child(0), source)
    
    methodMap := map[string]string{
      "get":     "GET",
      "post":    "POST",
      "put":     "PUT",
      "patch":   "PATCH",
      "delete":  "DELETE",
    }
    
    if httpMethod, ok := methodMap[methodName]; ok {
      args := node.NamedChild(0)
      if args != nil && args.NamedChildCount() > 0 {
        pathNode := args.NamedChild(0)
        path := strings.Trim(p.getNodeText(pathNode, source), "\"'")
        
        // Extract controller#action
        controller := ""
        action := ""
        if args.NamedChildCount() > 1 {
          actionNode := args.NamedChild(1)
          actionText := p.getNodeText(actionNode, source)
          
          if strings.Contains(actionText, "#") {
            parts := strings.Split(actionText, "#")
            controller = parts[0]
            action = parts[1]
          }
        }
        
        route := types.Route{
          Method:  httpMethod,
          Path:    "/" + path,
          Handler: controller + "#" + action,
          File:    filePath,
          Line:    int(node.StartPoint().Row) + 1,
        }
        *routes = append(*routes, route)
      }
    }
    
    // Handle resources :name
    if methodName == "resources" {
      args := node.NamedChild(0)
      if args != nil && args.NamedChildCount() > 0 {
        resourceName := strings.Trim(p.getNodeText(args.NamedChild(0), source), "\"'")
        
        // Generate RESTful routes for resource
        resourceRoutes := p.generateResourceRoutes(resourceName, filePath, int(node.StartPoint().Row)+1)
        *routes = append(*routes, resourceRoutes...)
      }
    }
  }

  // Recursively search
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    p.findRoutes(child, source, routes, filePath)
  }
}

// generateResourceRoutes generates RESTful routes for a resource
func (p *RailsParser) generateResourceRoutes(resourceName, filePath string, line int) []types.Route {
  return []types.Route{
    {Method: "GET", Path: "/" + resourceName, Handler: resourceName + "#index", File: filePath, Line: line},
    {Method: "GET", Path: "/" + resourceName + "/new", Handler: resourceName + "#new", File: filePath, Line: line},
    {Method: "POST", Path: "/" + resourceName, Handler: resourceName + "#create", File: filePath, Line: line},
    {Method: "GET", Path: "/" + resourceName + "/:id", Handler: resourceName + "#show", File: filePath, Line: line},
    {Method: "GET", Path: "/" + resourceName + "/:id/edit", Handler: resourceName + "#edit", File: filePath, Line: line},
    {Method: "PATCH", Path: "/" + resourceName + "/:id", Handler: resourceName + "#update", File: filePath, Line: line},
    {Method: "DELETE", Path: "/" + resourceName + "/:id", Handler: resourceName + "#destroy", File: filePath, Line: line},
  }
}

// ParseModels extracts Rails ActiveRecord models
func (p *RailsParser) ParseModels(ctx context.Context, file *types.File) ([]types.Model, error) {
  // Rails models are defined in app/models/*.rb
  // Parse model class and extract associations
  
  parser := sit.NewParser()
  parser.SetLanguage(p.language)

  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }

  var models []types.Model

  // Find model class definitions
  p.findModels(tree.RootNode(), []byte(file.Content), &models, file.Path)

  return models, nil
}

// findModels finds Rails model classes
func (p *RailsParser) findModels(node *sit.Node, source []byte, models *[]types.Model, filePath string) {
  if node.Type() == "class" {
    className := p.getNodeText(node.Child(0), source)
    
    // Check if class inherits from ApplicationRecord or ActiveRecord::Base
    superClass := p.findSuperClass(node, source)
    if superClass == "ApplicationRecord" || superClass == "ActiveRecord::Base" {
      model := types.Model{
        Name:   className,
        Table:  p.tableize(className),
        Fields: p.extractAssociations(node, source),
        File:   filePath,
        Line:   int(node.StartPoint().Row) + 1,
      }
      *models = append(*models, model)
    }
  }

  // Recursively search
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    p.findModels(child, source, models, filePath)
  }
}

// extractAssociations extracts model associations
func (p *RailsParser) extractAssociations(classNode *sit.Node, source []byte) []types.Field {
  var fields []types.Field

  // Find belongs_to, has_many, has_one associations
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
          associationName := strings.Trim(p.getNodeText(args.NamedChild(0), source), "\"'")
          
          field := types.Field{
            Name: associationName,
            Type: methodName, // belongs_to, has_many, has_one
          }
          fields = append(fields, field)
        }
      }
    }
  }

  return fields
}

// ParseControllers extracts Rails controller actions
func (p *RailsParser) ParseControllers(ctx context.Context, file *types.File) ([]types.Handler, error) {
  parser := sit.NewParser()
  parser.SetLanguage(p.language)

  tree, err := parser.ParseCtx(ctx, nil, []byte(file.Content))
  if err != nil {
    return nil, err
  }

  var handlers []types.Handler

  // Find controller class and extract public methods
  p.findControllers(tree.RootNode(), []byte(file.Content), &handlers, file.Path)

  return handlers, nil
}

// findControllers finds Rails controller actions
func (p *RailsParser) findControllers(node *sit.Node, source []byte, handlers *[]types.Handler, filePath string) {
  if node.Type() == "class" {
    className := p.getNodeText(node.Child(0), source)
    
    // Check if it's a controller
    if strings.HasSuffix(className, "Controller") {
      // Extract public methods
      body := node.ChildByFieldName("body")
      if body != nil {
        for i := 0; i < int(body.NamedChildCount()); i++ {
          stmt := body.NamedChild(i)
          
          if stmt.Type() == "method" {
            methodName := p.getNodeText(stmt.Child(0), source)
            
            // Skip private methods and Rails callbacks
            if !strings.HasPrefix(methodName, "_") && 
               methodName != "initialize" &&
               !strings.HasPrefix(methodName, "before_") &&
               !strings.HasPrefix(methodName, "after_") {
              
              handler := types.Handler{
                Name: className + "#" + methodName,
                File: filePath,
                Line: int(stmt.StartPoint().Row) + 1,
              }
              *handlers = append(*handlers, handler)
            }
          }
        }
      }
    }
  }

  // Recursively search
  for i := 0; i < int(node.NamedChildCount()); i++ {
    child := node.NamedChild(i)
    p.findControllers(child, source, handlers, filePath)
  }
}

// Helper methods
func (p *RailsParser) getNodeText(node *sit.Node, source []byte) string {
  if node == nil {
    return ""
  }
  return string(source[node.StartByte():node.EndByte()])
}

func (p *RailsParser) findSuperClass(node *sit.Node, source []byte) string {
  superClass := node.ChildByFieldName("superclass")
  if superClass == nil {
    return ""
  }
  return p.getNodeText(superClass, source)
}

func (p *RailsParser) tableize(s string) string {
  // Convert ModelName to table_name (Rails convention)
  result := ""
  for i, r := range s {
    if i > 0 && r >= 'A' && r <= 'Z' {
      result += "_"
    }
    result += strings.ToLower(string(r))
  }
  return result + "s" // pluralize
}
```

**Acceptance Criteria:**
- Rails routes extracted (including resources)
- RESTful routes generated for resources
- Rails models parsed with associations
- Controller actions extracted

---

### Sprint 19-20: Java & C# Support (Weeks 37-40)

### Goal
Add support for Java (Spring Boot) and C# (ASP.NET Core)

### Tasks

#### Task 19.1: Java Spring Boot Parser

**Objective:** Parse Spring Boot controllers, models, and services

**Implementation:** Parse Java annotations (@RestController, @RequestMapping, @Entity, etc.)

---

#### Task 19.2: C# ASP.NET Core Parser

**Objective:** Parse ASP.NET Core controllers, models, and services

**Implementation:** Parse C# attributes ([ApiController], [Route], [HttpGet], etc.)

---

### Sprint 21-22: TypeScript & Rust Support (Weeks 41-44)

### Goal
Add support for TypeScript (Express, NestJS) and Rust (Actix, Rocket)

### Tasks

#### Task 21.1: TypeScript Express Parser

**Objective:** Parse Express routes in TypeScript

**Implementation:** Similar to JavaScript parser but with TypeScript type information

---

#### Task 21.2: TypeScript NestJS Parser

**Objective:** Parse NestJS decorators and controllers

**Implementation:** Parse NestJS decorators (@Controller, @Get, @Post, etc.)

---

#### Task 21.3: Rust Actix Parser

**Objective:** Parse Actix web routes

**Implementation:** Parse Rust route attributes and handlers

---

#### Task 21.4: Rust Rocket Parser

**Objective:** Parse Rocket routes

**Implementation:** Parse Rocket route macros and handlers

---

*End of Phase 4 Plan (Sprints 17-22)*

---

## Phase 5: Advanced AI Features (Weeks 45-50)

### Goal
Enhance AI capabilities with advanced test optimization and intelligent suggestions

### Sprint 23-24: AI-Driven Test Optimization (Weeks 45-48)

### Tasks

#### Task 23.1: Test Quality Analyzer

**Objective:** Analyze test quality and suggest improvements

**Implementation:**
- Analyze test coverage gaps
- Identify redundant tests
- Suggest test optimizations
- Detect flaky tests

---

#### Task 23.2: Intelligent Test Generation

**Objective:** Generate tests with higher quality using advanced AI

**Implementation:**
- Use GPT-4 for complex test scenarios
- Generate edge cases automatically
- Create negative test cases
- Generate performance tests

---

#### Task 23.3: Code Review Assistant

**Objective:** Provide AI-driven code review suggestions

**Implementation:**
- Analyze code changes
- Suggest test cases for new code
- Identify potential bugs
- Suggest code improvements

---

### Sprint 25-26: Advanced Analytics (Weeks 49-52)

### Tasks

#### Task 25.1: Test Coverage Insights

**Objective:** Provide detailed test coverage analytics

**Implementation:**
- Line coverage analysis
- Branch coverage analysis
- Mutation testing integration
- Coverage trend tracking

---

#### Task 25.2: Test Performance Analytics

**Objective:** Track and optimize test performance

**Implementation:**
- Test execution time tracking
- Identify slow tests
- Suggest test parallelization
- Performance trend analysis

---

#### Task 25.3: Test Flakiness Detection

**Objective:** Detect and fix flaky tests

**Implementation:**
- Run tests multiple times
- Identify inconsistent tests
- Suggest fixes for flaky tests
- Track flakiness trends

---

*End of Phase 5 Plan*

---

## Phase 6: Enterprise Features & Integrations (Weeks 51-56)

### Goal
Add enterprise features and integrate with external tools

### Sprint 27-28: Enterprise Features (Weeks 51-54)

### Tasks

#### Task 27.1: Team Collaboration

**Objective:** Add team collaboration features

**Implementation:**
- User roles and permissions
- Team dashboards
- Test assignment and review
- Comments and discussions

---

#### Task 27.2: Test Governance

**Objective:** Add test governance and compliance features

**Implementation:**
- Test approval workflows
- Compliance reporting
- Audit trails
- Test standards enforcement

---

#### Task 27.3: Advanced Reporting

**Objective:** Provide comprehensive test reports

**Implementation:**
- Custom report templates
- Scheduled report generation
- Report export (PDF, Excel, HTML)
- Report sharing and distribution

---

### Sprint 29-30: External Integrations (Weeks 55-58)

### Tasks

#### Task 29.1: CI/CD Integrations

**Objective:** Integrate with popular CI/CD platforms

**Implementation:**
- GitHub Actions integration
- GitLab CI integration
- Jenkins integration
- CircleCI integration

---

#### Task 29.2: Code Quality Tools

**Objective:** Integrate with code quality tools

**Implementation:**
- SonarQube integration
- ESLint integration
- Prettier integration
- Code coverage tools

---

#### Task 29.3: Monitoring & Alerting

**Objective:** Integrate with monitoring and alerting tools

**Implementation:**
- Slack notifications
- PagerDuty integration
- Email notifications
- Webhook support

---

#### Task 29.4: Project Management

**Objective:** Integrate with project management tools

**Implementation:**
- Jira integration
- Asana integration
- Trello integration
- Linear integration

---

*End of Phase 6 Plan*

---

## Summary

**Phase 4-6 Timeline:** 24 weeks (12 sprints)  
**Total Project Timeline:** 58 weeks (~14 months)

**Key Deliverables:**
1. Support for 8+ programming languages
2. Advanced AI-driven test optimization
3. Comprehensive analytics and insights
4. Enterprise collaboration features
5. External tool integrations
6. Test governance and compliance

**Success Criteria:**
- 8+ languages supported (JS, Go, PHP, Python, Ruby, Java, C#, TypeScript, Rust)
- 90% user satisfaction with AI-generated tests
- 95% test coverage maintained automatically
- 99% uptime for enterprise features
- Integration with 10+ external tools

**Next Steps:**
1. Implement Phase 4 (Python, Ruby, Java, C#, TypeScript, Rust support)
2. Implement Phase 5 (Advanced AI features and analytics)
3. Implement Phase 6 (Enterprise features and integrations)
4. Launch public beta
5. Iterate based on user feedback

---

*End of Complete Development Plan*
