package python

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

func TestPythonParser_SupportedLanguages(t *testing.T) {
	parser := NewParser()
	langs := parser.SupportedLanguages()

	if len(langs) != 1 || langs[0] != "python" {
		t.Errorf("Expected [python], got %v", langs)
	}
}

func TestPythonParser_DetectFramework_Django(t *testing.T) {
	// Create a temporary Django project structure
	tmpDir := t.TempDir()

	// Create manage.py file (Django indicator)
	managePath := filepath.Join(tmpDir, "manage.py")
	if err := os.WriteFile(managePath, []byte("#!/usr/bin/env python\n"), 0755); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "django" {
		t.Errorf("Expected django, got %s", framework)
	}
}

func TestPythonParser_DetectFramework_FastAPI(t *testing.T) {
	// Create a temporary FastAPI project
	tmpDir := t.TempDir()

	// Create requirements.txt with fastapi
	reqPath := filepath.Join(tmpDir, "requirements.txt")
	reqContent := "fastapi==0.104.1\nuvicorn==0.24.0\n"
	if err := os.WriteFile(reqPath, []byte(reqContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "fastapi" {
		t.Errorf("Expected fastapi, got %s", framework)
	}
}

func TestPythonParser_DetectFramework_Flask(t *testing.T) {
	// Create a temporary Flask project
	tmpDir := t.TempDir()

	// Create requirements.txt with Flask
	reqPath := filepath.Join(tmpDir, "requirements.txt")
	reqContent := "Flask==3.0.0\nFlask-SQLAlchemy==3.1.1\n"
	if err := os.WriteFile(reqPath, []byte(reqContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "flask" {
		t.Errorf("Expected flask, got %s", framework)
	}
}

func TestPythonParser_DetectFramework_Plain(t *testing.T) {
	// Create a temporary plain Python project
	tmpDir := t.TempDir()

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "python" {
		t.Errorf("Expected python, got %s", framework)
	}
}

func TestPythonParser_ParseDjangoRoutes(t *testing.T) {
	// Create a temporary Django project
	tmpDir := t.TempDir()

	// Create manage.py
	managePath := filepath.Join(tmpDir, "manage.py")
	if err := os.WriteFile(managePath, []byte("#!/usr/bin/env python\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create app directory
	appDir := filepath.Join(tmpDir, "myapp")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create urls.py
	urlsContent := `from django.urls import path
from . import views

urlpatterns = [
    path('users/', views.user_list, name='user_list'),
    path('users/<int:id>/', views.user_detail, name='user_detail'),
    path('posts/', views.post_list, name='post_list'),
    path('posts/<int:id>/', views.post_detail, name='post_detail'),
]
`
	urlsPath := filepath.Join(appDir, "urls.py")
	if err := os.WriteFile(urlsPath, []byte(urlsContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 4 routes
	if len(codebase.Routes) != 4 {
		t.Errorf("Expected 4 routes, got %d", len(codebase.Routes))
	}

	// Check specific routes
	foundUsers := false
	foundUsersId := false
	foundPosts := false

	for _, route := range codebase.Routes {
		if route.Path == "users/" {
			foundUsers = true
			if route.Handler != "views.user_list" {
				t.Errorf("Expected handler views.user_list, got %s", route.Handler)
			}
		}
		if route.Path == "users/<int:id>/" {
			foundUsersId = true
		}
		if route.Path == "posts/" {
			foundPosts = true
		}
	}

	if !foundUsers {
		t.Error("Expected users/ route")
	}
	if !foundUsersId {
		t.Error("Expected users/<int:id>/ route")
	}
	if !foundPosts {
		t.Error("Expected posts/ route")
	}
}

func TestPythonParser_ParseDjangoModels(t *testing.T) {
	// Create a temporary Django project
	tmpDir := t.TempDir()

	// Create manage.py
	managePath := filepath.Join(tmpDir, "manage.py")
	if err := os.WriteFile(managePath, []byte("#!/usr/bin/env python\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create app directory
	appDir := filepath.Join(tmpDir, "myapp")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create models.py
	modelsContent := `from django.db import models

class User(models.Model):
    username = models.CharField(max_length=100)
    email = models.EmailField(unique=True)
    created_at = models.DateTimeField(auto_now_add=True)

class Post(models.Model):
    title = models.CharField(max_length=200)
    content = models.TextField()
    author = models.ForeignKey(User, on_delete=models.CASCADE)
    published = models.BooleanField(default=False)

class Comment(models.Model):
    post = models.ForeignKey(Post, on_delete=models.CASCADE)
    text = models.TextField()
`
	modelsPath := filepath.Join(appDir, "models.py")
	if err := os.WriteFile(modelsPath, []byte(modelsContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 3 models
	if len(codebase.Models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(codebase.Models))
	}

	// Check User model
	var userModel *types.Model
	for i := range codebase.Models {
		if codebase.Models[i].Name == "User" {
			userModel = &codebase.Models[i]
			break
		}
	}

	if userModel == nil {
		t.Fatal("Expected User model")
	}

	// Should have 3 fields (username, email, created_at)
	if len(userModel.Fields) < 3 {
		t.Errorf("Expected at least 3 fields, got %d", len(userModel.Fields))
	}

	// Check Post model
	var postModel *types.Model
	for i := range codebase.Models {
		if codebase.Models[i].Name == "Post" {
			postModel = &codebase.Models[i]
			break
		}
	}

	if postModel == nil {
		t.Fatal("Expected Post model")
	}

	// Should have 4 fields (title, content, author, published)
	if len(postModel.Fields) < 4 {
		t.Errorf("Expected at least 4 fields, got %d", len(postModel.Fields))
	}
}

func TestPythonParser_ParseDjangoViews(t *testing.T) {
	// Create a temporary Django project
	tmpDir := t.TempDir()

	// Create manage.py
	managePath := filepath.Join(tmpDir, "manage.py")
	if err := os.WriteFile(managePath, []byte("#!/usr/bin/env python\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create app directory
	appDir := filepath.Join(tmpDir, "myapp")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create views.py
	viewsContent := `from django.shortcuts import render
from django.http import JsonResponse
from .models import User, Post

def user_list(request):
    users = User.objects.all()
    return JsonResponse({'users': list(users.values())})

def user_detail(request, id):
    user = User.objects.get(id=id)
    return JsonResponse({'user': user})

def post_list(request):
    posts = Post.objects.all()
    return JsonResponse({'posts': list(posts.values())})

def post_detail(request, id):
    post = Post.objects.get(id=id)
    return JsonResponse({'post': post})

# This should not be extracted (no request parameter)
def helper_function():
    return "helper"

# This should not be extracted (no request parameter)
def another_helper(x, y):
    return x + y
`
	viewsPath := filepath.Join(appDir, "views.py")
	if err := os.WriteFile(viewsPath, []byte(viewsContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 4 views (not helper functions)
	if len(codebase.Handlers) != 4 {
		t.Errorf("Expected 4 handlers, got %d", len(codebase.Handlers))
	}

	// Check that all handlers have 'request' as first parameter
	expectedHandlers := map[string]bool{
		"user_list":   false,
		"user_detail": false,
		"post_list":   false,
		"post_detail": false,
	}

	for _, handler := range codebase.Handlers {
		if _, ok := expectedHandlers[handler.Name]; ok {
			expectedHandlers[handler.Name] = true
		} else {
			t.Errorf("Unexpected handler: %s", handler.Name)
		}
	}

	// Check all expected handlers were found
	for name, found := range expectedHandlers {
		if !found {
			t.Errorf("Expected handler %s not found", name)
		}
	}

	// Check that helper functions were not extracted
	for _, handler := range codebase.Handlers {
		if handler.Name == "helper_function" || handler.Name == "another_helper" {
			t.Errorf("Helper function %s should not be extracted", handler.Name)
		}
	}
}

func TestPythonParser_EmptyProject(t *testing.T) {
	// Create a temporary plain Python project
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

func TestPythonParser_DjangoProject_NoUrlsFile(t *testing.T) {
	// Create a Django project without urls.py
	tmpDir := t.TempDir()

	// Create manage.py
	managePath := filepath.Join(tmpDir, "manage.py")
	if err := os.WriteFile(managePath, []byte("#!/usr/bin/env python\n"), 0755); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have no routes
	if len(codebase.Routes) != 0 {
		t.Errorf("Expected 0 routes, got %d", len(codebase.Routes))
	}
}

func TestPythonParser_MultipleApps(t *testing.T) {
	// Create a Django project with multiple apps
	tmpDir := t.TempDir()

	// Create manage.py
	managePath := filepath.Join(tmpDir, "manage.py")
	if err := os.WriteFile(managePath, []byte("#!/usr/bin/env python\n"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create app1
	app1Dir := filepath.Join(tmpDir, "app1")
	if err := os.MkdirAll(app1Dir, 0755); err != nil {
		t.Fatal(err)
	}

	urls1 := `from django.urls import path
urlpatterns = [
    path('app1/', views.index, name='app1_index'),
]
`
	if err := os.WriteFile(filepath.Join(app1Dir, "urls.py"), []byte(urls1), 0644); err != nil {
		t.Fatal(err)
	}

	// Create app2
	app2Dir := filepath.Join(tmpDir, "app2")
	if err := os.MkdirAll(app2Dir, 0755); err != nil {
		t.Fatal(err)
	}

	urls2 := `from django.urls import path
urlpatterns = [
    path('app2/', views.index, name='app2_index'),
]
`
	if err := os.WriteFile(filepath.Join(app2Dir, "urls.py"), []byte(urls2), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 2 routes from 2 apps
	if len(codebase.Routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(codebase.Routes))
	}
}
