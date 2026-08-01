package typescript_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/parser/typescript"
)

func TestExpressRoutes_TypeScript(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package.json with express
	packageJSON := `{"dependencies": {"express": "^4.18.0", "typescript": "^5.0.0"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create TypeScript Express app with typed handlers
	tsCode := `
import express, { Request, Response, NextFunction } from 'express';
const app = express();

type RequestHandler = (req: Request, res: Response, next: NextFunction) => void;

// Basic routes with typed handlers
app.get('/users', listUsers);
app.post('/users', authenticate, createUser);
app.get('/users/:id', (req: Request, res: Response) => {
  res.json({ id: req.params.id });
});
app.put('/users/:id', authenticate, updateUser);
app.delete('/users/:id', authenticate, authorize, deleteUser);

// Middleware mounting
app.use('/api', apiRouter);

// Route with type-annotated inline handler
app.post('/auth/login', validateInput, (req: Request, res: Response) => {
  res.json({ token: 'abc123' });
});
`
	if err := os.WriteFile(filepath.Join(tmpDir, "app.ts"), []byte(tsCode), 0644); err != nil {
		t.Fatalf("Failed to create app.ts: %v", err)
	}

	parser := typescript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if codebase.Framework != "express" {
		t.Errorf("Expected framework express, got %s", codebase.Framework)
	}

	if codebase.Language != "typescript" {
		t.Errorf("Expected language typescript, got %s", codebase.Language)
	}

	// We should find at least the routes with named handlers
	if len(codebase.Routes) < 4 {
		t.Errorf("Expected at least 4 routes, got %d", len(codebase.Routes))
	}

	// Verify some specific routes
	foundRoutes := make(map[string]string)
	for _, route := range codebase.Routes {
		key := route.Method + " " + route.Path
		foundRoutes[key] = route.Method
	}

	expectedPaths := []string{
		"GET /users",
		"POST /users",
		"GET /users/:id",
		"PUT /users/:id",
		"DELETE /users/:id",
		"USE /api",
		"POST /auth/login",
	}

	for _, expected := range expectedPaths {
		if _, ok := foundRoutes[expected]; !ok {
			t.Errorf("Expected to find route: %s", expected)
		}
	}
}

func TestExpressRoutes_RouterPattern(t *testing.T) {
	tmpDir := t.TempDir()

	packageJSON := `{"dependencies": {"express": "^4.18.0"}}`
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644)

	// TypeScript Express router
	routerCode := `
import { Router, Request, Response } from 'express';
const router = Router();

router.get('/', listItems);
router.post('/', createItem);
router.get('/:id', getItem);
router.put('/:id', updateItem);
router.delete('/:id', deleteItem);

export default router;
`
	os.WriteFile(filepath.Join(tmpDir, "routes.ts"), []byte(routerCode), 0644)

	parser := typescript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(codebase.Routes) < 5 {
		t.Errorf("Expected at least 5 routes, got %d", len(codebase.Routes))
	}

	// Check for GET / route
	found := false
	for _, route := range codebase.Routes {
		if route.Method == "GET" && route.Path == "/" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find GET / route")
	}
}

func TestNestJSController(t *testing.T) {
	tmpDir := t.TempDir()

	// Simulate NestJS project (no package.json with express)
	// isNestJS scans source files for @nestjs/* imports

	// NestJS controller with decorated routes
	controllerCode := `
import { Controller, Get, Post, Put, Delete, Patch, Param, Body, Query } from '@nestjs/common';

@Controller('users')
export class UsersController {
  @Get()
  findAll(@Query('page') page: number) {
    return 'All users';
  }

  @Get(':id')
  findOne(@Param('id') id: string) {
    return 'User ' + id;
  }

  @Post()
  create(@Body() createUserDto: CreateUserDto) {
    return 'Created';
  }

  @Put(':id')
  update(@Param('id') id: string, @Body() updateUserDto: UpdateUserDto) {
    return 'Updated ' + id;
  }

  @Delete(':id')
  remove(@Param('id') id: string) {
    return 'Deleted ' + id;
  }

  @Patch(':id')
  partialUpdate(@Param('id') id: string) {
    return 'Patched ' + id;
  }
}
`
	os.WriteFile(filepath.Join(tmpDir, "users.controller.ts"), []byte(controllerCode), 0644)

	parser := typescript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// With NestJS indicator present, framework should be nestjs
	if codebase.Framework != "nestjs" {
		t.Errorf("Expected framework nestjs, got %s", codebase.Framework)
	}

	if len(codebase.Routes) < 6 {
		t.Errorf("Expected at least 6 routes, got %d", len(codebase.Routes))
	}

	// Verify route paths
	foundPaths := make(map[string]bool)
	for _, route := range codebase.Routes {
		foundPaths[route.Path] = true
	}

	expectedPaths := []string{
		"/users",
		"/users/:id",
	}

	for _, expected := range expectedPaths {
		if !foundPaths[expected] {
			t.Errorf("Expected to find path: %s", expected)
		}
	}

	// Verify handlers were created
	if len(codebase.Handlers) < 6 {
		t.Errorf("Expected at least 6 handlers, got %d", len(codebase.Handlers))
	}

	// Check controller name in handlers
	for _, handler := range codebase.Handlers {
		if handler.Controller != "UsersController" {
			t.Errorf("Expected controller UsersController, got %s", handler.Controller)
		}
	}
}

func TestNestJSModule(t *testing.T) {
	tmpDir := t.TempDir()

	// NestJS module
	moduleCode := `
import { Module } from '@nestjs/common';
import { UsersController } from './users.controller';
import { UsersService } from './users.service';

@Module({
  imports: [],
  controllers: [UsersController],
  providers: [UsersService],
  exports: [UsersService],
})
export class UsersModule {}
`
	os.WriteFile(filepath.Join(tmpDir, "users.module.ts"), []byte(moduleCode), 0644)

	parser := typescript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should extract the @Module as a model
	foundModule := false
	for _, model := range codebase.Models {
		if model.Name == "UsersModule" && model.Table == "module" {
			foundModule = true
			// Check that fields include controllers, providers, imports, exports
			fieldNames := make(map[string]bool)
			fieldValues := make(map[string]string)
			for _, f := range model.Fields {
				fieldNames[f.Name] = true
				fieldValues[f.Name] = f.Type
			}
			if !fieldNames["controllers"] {
				t.Error("Expected controllers field in module")
			}
			if !fieldNames["providers"] {
				t.Error("Expected providers field in module")
			}
			if !fieldNames["imports"] {
				t.Error("Expected imports field in module")
			}
			if !fieldNames["exports"] {
				t.Error("Expected exports field in module")
			}
			if fieldValues["controllers"] != "UsersController" {
				t.Errorf("Expected controllers to contain UsersController, got %s", fieldValues["controllers"])
			}
			if fieldValues["imports"] != "" {
				t.Errorf("Expected imports to be empty, got %s", fieldValues["imports"])
			}
		}
	}

	if !foundModule {
		t.Error("Expected to find UsersModule model")
	}
}

func TestFrameworkDetection_Express(t *testing.T) {
	tmpDir := t.TempDir()

	packageJSON := `{"dependencies": {"express": "^4.18.0", "typescript": "^5.0.0"}}`
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644)

	parser := typescript.NewParser()
	framework, err := parser.DetectFramework(tmpDir)
	if err != nil {
		t.Fatalf("DetectFramework failed: %v", err)
	}

	if framework != "express" {
		t.Errorf("Expected express, got %s", framework)
	}
}

func TestFrameworkDetection_NestJS(t *testing.T) {
	tmpDir := t.TempDir()

	// NestJS project with @nestjs/core import in source file
	controllerCode := `
import { Controller, Get } from '@nestjs/common';

@Controller('app')
export class AppController {
  @Get()
  getHello() {
    return 'Hello';
  }
}
`
	os.WriteFile(filepath.Join(tmpDir, "app.controller.ts"), []byte(controllerCode), 0644)

	parser := typescript.NewParser()
	framework, err := parser.DetectFramework(tmpDir)
	if err != nil {
		t.Fatalf("DetectFramework failed: %v", err)
	}

	if framework != "nestjs" {
		t.Errorf("Expected nestjs, got %s", framework)
	}
}

func TestFrameworkDetection_TypescriptOnly(t *testing.T) {
	tmpDir := t.TempDir()

	// TypeScript project without express or nestjs
	tsCode := `
export interface User {
  id: number;
  name: string;
  email: string;
}
`
	os.WriteFile(filepath.Join(tmpDir, "types.ts"), []byte(tsCode), 0644)

	parser := typescript.NewParser()
	framework, err := parser.DetectFramework(tmpDir)
	if err != nil {
		t.Fatalf("DetectFramework failed: %v", err)
	}

	if framework != "typescript" {
		t.Errorf("Expected typescript, got %s", framework)
	}
}

func TestEmptyProject(t *testing.T) {
	tmpDir := t.TempDir()

	parser := typescript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if codebase == nil {
		t.Fatal("Parse returned nil codebase")
	}

	if len(codebase.Routes) != 0 {
		t.Errorf("Expected 0 routes, got %d", len(codebase.Routes))
	}
	if len(codebase.Models) != 0 {
		t.Errorf("Expected 0 models, got %d", len(codebase.Models))
	}
	if len(codebase.Handlers) != 0 {
		t.Errorf("Expected 0 handlers, got %d", len(codebase.Handlers))
	}
	if codebase.Language != "typescript" {
		t.Errorf("Expected language typescript, got %s", codebase.Language)
	}
}

func TestTypeScriptInterfaces(t *testing.T) {
	tmpDir := t.TempDir()

	tsCode := `
export interface User {
  id: number;
  name: string;
  email: string;
  age?: number;
  isActive: boolean;
}

export interface Product {
  id: number;
  name: string;
  price: number;
  description?: string;
}
`
	os.WriteFile(filepath.Join(tmpDir, "models.ts"), []byte(tsCode), 0644)

	parser := typescript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(codebase.Models) < 2 {
		t.Errorf("Expected at least 2 models, got %d", len(codebase.Models))
	}

	modelNames := make(map[string]bool)
	for _, model := range codebase.Models {
		modelNames[model.Name] = true

		if model.Name == "User" {
			if len(model.Fields) < 5 {
				t.Errorf("Expected at least 5 fields for User, got %d", len(model.Fields))
			}
			// Check required vs optional
			for _, field := range model.Fields {
				if field.Name == "name" && !field.Required {
					t.Errorf("Expected 'name' field to be required")
				}
				if field.Name == "age" && field.Required {
					t.Errorf("Expected 'age' field to be optional")
				}
			}
		}
	}

	if !modelNames["User"] {
		t.Error("Expected to find User model")
	}
	if !modelNames["Product"] {
		t.Error("Expected to find Product model")
	}
}

func TestSkipNodeModules(t *testing.T) {
	tmpDir := t.TempDir()

	packageJSON := `{"dependencies": {"express": "^4.18.0"}}`
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644)

	// Create app.ts with routes
	appCode := `
import express from 'express';
const app = express();
app.get('/test', handler);
app.post('/data', submitData);
`
	os.WriteFile(filepath.Join(tmpDir, "app.ts"), []byte(appCode), 0644)

	// Create node_modules with a TypeScript file
	nodeModulesDir := filepath.Join(tmpDir, "node_modules", "some-lib")
	os.MkdirAll(nodeModulesDir, 0755)
	libCode := `
import express from 'express';
const app = express();
app.get('/should-skip', theirHandler);
`
	os.WriteFile(filepath.Join(nodeModulesDir, "index.ts"), []byte(libCode), 0644)

	parser := typescript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Only routes from app.ts should be found
	if len(codebase.Routes) != 2 {
		t.Errorf("Expected 2 routes (node_modules skipped), got %d", len(codebase.Routes))
	}
}

func TestSupportedLanguages(t *testing.T) {
	parser := typescript.NewParser()
	languages := parser.SupportedLanguages()

	if len(languages) != 1 {
		t.Errorf("Expected 1 supported language, got %d", len(languages))
	}

	if languages[0] != "typescript" {
		t.Errorf("Expected 'typescript', got '%s'", languages[0])
	}
}

func TestNestJSController_BasePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Controller with base path and method-level path
	controllerCode := `
import { Controller, Get, Post } from '@nestjs/common';

@Controller('/api/v1/products')
export class ProductsController {
  @Get()
  findAll() {
    return [];
  }

  @Get('featured')
  findFeatured() {
    return [];
  }

  @Post('search')
  search() {
    return [];
  }
}
`
	os.WriteFile(filepath.Join(tmpDir, "products.controller.ts"), []byte(controllerCode), 0644)

	parser := typescript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	foundPaths := make(map[string]bool)
	for _, route := range codebase.Routes {
		foundPaths[route.Path] = true
	}

	expectedPaths := []string{
		"/api/v1/products",
		"/api/v1/products/featured",
		"/api/v1/products/search",
	}

	for _, expected := range expectedPaths {
		if !foundPaths[expected] {
			t.Errorf("Expected to find path: %s", expected)
		}
	}
}
