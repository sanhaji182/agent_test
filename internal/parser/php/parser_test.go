package php

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPHPParser_SupportedLanguages(t *testing.T) {
	parser := NewParser()
	langs := parser.SupportedLanguages()

	if len(langs) != 1 || langs[0] != "php" {
		t.Errorf("Expected [php], got %v", langs)
	}
}

func TestPHPParser_DetectFramework_Laravel(t *testing.T) {
	// Create a temporary Laravel project structure
	tmpDir := t.TempDir()

	// Create artisan file (Laravel indicator)
	artisanPath := filepath.Join(tmpDir, "artisan")
	if err := os.WriteFile(artisanPath, []byte("#!/usr/bin/env php\n"), 0755); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "laravel" {
		t.Errorf("Expected laravel, got %s", framework)
	}
}

func TestPHPParser_DetectFramework_Symfony(t *testing.T) {
	// Create a temporary Symfony project structure
	tmpDir := t.TempDir()

	// Create bin/console file (Symfony indicator)
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}

	consolePath := filepath.Join(binDir, "console")
	if err := os.WriteFile(consolePath, []byte("#!/usr/bin/env php\n"), 0755); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "symfony" {
		t.Errorf("Expected symfony, got %s", framework)
	}
}

func TestPHPParser_DetectFramework_Plain(t *testing.T) {
	// Create a temporary plain PHP project
	tmpDir := t.TempDir()

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "php" {
		t.Errorf("Expected php, got %s", framework)
	}
}

func TestPHPParser_ParseLaravelRoutes(t *testing.T) {
	// Create a temporary Laravel project
	tmpDir := t.TempDir()

	// Create artisan file
	artisanPath := filepath.Join(tmpDir, "artisan")
	if err := os.WriteFile(artisanPath, []byte("#!/usr/bin/env php\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create routes directory
	routesDir := filepath.Join(tmpDir, "routes")
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create routes/web.php with various route types
	webRoutes := `<?php

use App\Http\Controllers\UserController;
use App\Http\Controllers\PostController;

Route::get('/users', [UserController::class, 'index']);
Route::post('/users', [UserController::class, 'store']);
Route::get('/users/{id}', [UserController::class, 'show']);
Route::put('/users/{id}', [UserController::class, 'update']);
Route::delete('/users/{id}', [UserController::class, 'destroy']);

Route::get('/posts', [PostController::class, 'index'])->middleware('auth');
Route::post('/posts', [PostController::class, 'store'])->middleware(['auth', 'verified']);
`
	webPath := filepath.Join(routesDir, "web.php")
	if err := os.WriteFile(webPath, []byte(webRoutes), 0644); err != nil {
		t.Fatal(err)
	}

	// Create routes/api.php
	apiRoutes := `<?php

use App\Http\Controllers\Api\UserController;

Route::get('/api/users', [UserController::class, 'index']);
Route::post('/api/users', [UserController::class, 'store']);
`
	apiPath := filepath.Join(routesDir, "api.php")
	if err := os.WriteFile(apiPath, []byte(apiRoutes), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 9 routes total (7 from web.php + 2 from api.php)
	if len(codebase.Routes) != 9 {
		t.Errorf("Expected 9 routes, got %d", len(codebase.Routes))
	}

	// Check specific routes
	foundGetUsers := false
	foundPostUsers := false
	foundGetUsersId := false
	foundPostsWithAuth := false

	for _, route := range codebase.Routes {
		if route.Method == "GET" && route.Path == "/users" {
			foundGetUsers = true
		}
		if route.Method == "POST" && route.Path == "/users" {
			foundPostUsers = true
		}
		if route.Method == "GET" && route.Path == "/users/{id}" {
			foundGetUsersId = true
		}
		if route.Method == "GET" && route.Path == "/posts" && len(route.Middleware) == 1 && route.Middleware[0] == "auth" {
			foundPostsWithAuth = true
		}
	}

	if !foundGetUsers {
		t.Error("Expected GET /users route")
	}
	if !foundPostUsers {
		t.Error("Expected POST /users route")
	}
	if !foundGetUsersId {
		t.Error("Expected GET /users/{id} route")
	}
	if !foundPostsWithAuth {
		t.Error("Expected GET /posts route with auth middleware")
	}
}

func TestPHPParser_ParseLaravelModels(t *testing.T) {
	// Create a temporary Laravel project
	tmpDir := t.TempDir()

	// Create artisan file
	artisanPath := filepath.Join(tmpDir, "artisan")
	if err := os.WriteFile(artisanPath, []byte("#!/usr/bin/env php\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create app/Models directory
	modelsDir := filepath.Join(tmpDir, "app", "Models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create User model
	userModel := `<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class User extends Model
{
    protected $table = 'users';

    protected $fillable = [
        'name',
        'email',
        'password',
    ];

    protected $hidden = [
        'password',
        'remember_token',
    ];
}
`
	userPath := filepath.Join(modelsDir, "User.php")
	if err := os.WriteFile(userPath, []byte(userModel), 0644); err != nil {
		t.Fatal(err)
	}

	// Create Post model
	postModel := `<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Post extends Model
{
    protected $fillable = [
        'title',
        'content',
        'user_id',
    ];
}
`
	postPath := filepath.Join(modelsDir, "Post.php")
	if err := os.WriteFile(postPath, []byte(postModel), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 2 models
	if len(codebase.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(codebase.Models))
	}

	// Check User model
	var userModel *Model
	for i := range codebase.Models {
		if codebase.Models[i].Name == "User" {
			userModel = &codebase.Models[i]
			break
		}
	}

	if userModel == nil {
		t.Fatal("Expected User model")
	}

	if userModel.Table != "users" {
		t.Errorf("Expected table 'users', got %s", userModel.Table)
	}

	if len(userModel.Fields) != 3 {
		t.Errorf("Expected 3 fillable fields, got %d", len(userModel.Fields))
	}

	// Check Post model
	var postModel *Model
	for i := range codebase.Models {
		if codebase.Models[i].Name == "Post" {
			postModel = &codebase.Models[i]
			break
		}
	}

	if postModel == nil {
		t.Fatal("Expected Post model")
	}

	if len(postModel.Fields) != 3 {
		t.Errorf("Expected 3 fillable fields, got %d", len(postModel.Fields))
	}
}

func TestPHPParser_ParseLaravelControllers(t *testing.T) {
	// Create a temporary Laravel project
	tmpDir := t.TempDir()

	// Create artisan file
	artisanPath := filepath.Join(tmpDir, "artisan")
	if err := os.WriteFile(artisanPath, []byte("#!/usr/bin/env php\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create app/Http/Controllers directory
	controllersDir := filepath.Join(tmpDir, "app", "Http", "Controllers")
	if err := os.MkdirAll(controllersDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create UserController
	userController := `<?php

namespace App\Http\Controllers;

use App\Models\User;
use Illuminate\Http\Request;

class UserController extends Controller
{
    public function index()
    {
        return User::all();
    }

    public function store(Request $request)
    {
        return User::create($request->all());
    }

    public function show($id)
    {
        return User::findOrFail($id);
    }

    public function update(Request $request, $id)
    {
        $user = User::findOrFail($id);
        $user->update($request->all());
        return $user;
    }

    public function destroy($id)
    {
        User::destroy($id);
        return response()->noContent();
    }

    private function helperMethod()
    {
        // This should not be extracted
    }
}
`
	userPath := filepath.Join(controllersDir, "UserController.php")
	if err := os.WriteFile(userPath, []byte(userController), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 5 public methods (not private helperMethod)
	if len(codebase.Handlers) != 5 {
		t.Errorf("Expected 5 handlers, got %d", len(codebase.Handlers))
	}

	// Check that all handlers belong to UserController
	for _, handler := range codebase.Handlers {
		if handler.Controller != "UserController" {
			t.Errorf("Expected controller UserController, got %s", handler.Controller)
		}
	}

	// Check specific methods
	methods := make(map[string]bool)
	for _, handler := range codebase.Handlers {
		methods[handler.Name] = true
	}

	expectedMethods := []string{"index", "store", "show", "update", "destroy"}
	for _, method := range expectedMethods {
		if !methods[method] {
			t.Errorf("Expected method %s", method)
		}
	}

	// Check that helperMethod is not included
	if methods["helperMethod"] {
		t.Error("Private method helperMethod should not be extracted")
	}
}

func TestPHPParser_EmptyProject(t *testing.T) {
	// Create a temporary plain PHP project
	tmpDir := t.TempDir()

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have no routes, models, or handlers
	if len(codebase.Routes) != 0 {
		t.Errorf("Expected 0 routes, got %d", len(codebase.Routes))
	}
	if len(codebase.Models) != 0 {
		t.Errorf("Expected 0 models, got %d", len(codebase.Models))
	}
	if len(codebase.Handlers) != 0 {
		t.Errorf("Expected 0 handlers, got %d", len(codebase.Handlers))
	}
}
