package ruby

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

func TestRubyParser_SupportedLanguages(t *testing.T) {
	parser := NewParser()
	langs := parser.SupportedLanguages()

	if len(langs) != 1 || langs[0] != "ruby" {
		t.Errorf("Expected [ruby], got %v", langs)
	}
}

func TestRubyParser_DetectFramework_RailsGemfile(t *testing.T) {
	tmpDir := t.TempDir()

	gemfile := `source "https://rubygems.org"

gem "rails", "~> 7.1"
gem "pg"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(gemfile), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "rails" {
		t.Errorf("Expected rails, got %s", framework)
	}
}

func TestRubyParser_DetectFramework_RailsConfigRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config/routes.rb (no Gemfile)
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "routes.rb"), []byte("# routes\n"), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "rails" {
		t.Errorf("Expected rails, got %s", framework)
	}
}

func TestRubyParser_DetectFramework_RailsConfigApplication(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config/application.rb (no Gemfile, no routes.rb)
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "application.rb"), []byte("# app\n"), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "rails" {
		t.Errorf("Expected rails, got %s", framework)
	}
}

func TestRubyParser_DetectFramework_PlainRuby(t *testing.T) {
	tmpDir := t.TempDir()

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "ruby" {
		t.Errorf("Expected ruby, got %s", framework)
	}
}

func TestRubyParser_ParseRailsRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config directory
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create routes.rb with various route types
	routesContent := `Rails.application.routes.draw do
  # Users resource with full CRUD
  get '/users', to: 'users#index'
  post '/users', to: 'users#create'
  get '/users/:id', to: 'users#show'
  put '/users/:id', to: 'users#update'
  patch '/users/:id', to: 'users#update'
  delete '/users/:id', to: 'users#destroy'

  # Posts resource (via resources)
  resources :posts

  # Comments resource with only
  resources :comments, only: [:index, :show]

  # Namespace for admin
  namespace :admin do
    get '/dashboard', to: 'dashboard#index'
    resources :users
  end

  # Scope for API versioning
  scope '/v1' do
    get '/health', to: 'health#check'
  end

  # Match route with via
  match '/legacy', to: 'legacy#handle', via: [:get, :post]

  # Root route
  root to: 'home#index'
end
`
	if err := os.WriteFile(filepath.Join(configDir, "routes.rb"), []byte(routesContent), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if codebase.Framework != "rails" {
		t.Errorf("Expected framework rails, got %q", codebase.Framework)
	}

	// Total route count:
	// 6 user routes
	// 8 posts resource routes
	// 2 comments routes (only: index, show)
	// 1 admin dashboard
	// 8 admin users resource routes
	// 1 health check
	// 2 match routes (GET + POST)
	// 1 root route
	// = 29
	expectedCount := 29
	if len(codebase.Routes) != expectedCount {
		t.Errorf("Expected %d routes, got %d. Routes:", expectedCount, len(codebase.Routes))
		for i, r := range codebase.Routes {
			t.Logf("  [%d] %s %s -> %s (line %d)", i, r.Method, r.Path, r.Handler, r.Line)
		}
	}

	// Verify key routes exist
	findRoute := func(method, path, handler string) bool {
		for _, r := range codebase.Routes {
			if r.Method == method && r.Path == path {
				if handler != "" && r.Handler != handler {
					continue
				}
				return true
			}
		}
		return false
	}

	tests := []struct{ method, path, handler string }{
		{"GET", "/users", "users#index"},
		{"POST", "/users", "users#create"},
		{"GET", "/users/:id", "users#show"},
		{"PUT", "/users/:id", "users#update"},
		{"PATCH", "/users/:id", "users#update"},
		{"DELETE", "/users/:id", "users#destroy"},
		// resources :posts
		{"GET", "/posts", "posts#index"},
		{"POST", "/posts", "posts#create"},
		{"GET", "/posts/:id", "posts#show"},
		{"DELETE", "/posts/:id", "posts#destroy"},
		// resources :comments, only: [:index, :show]
		{"GET", "/comments", "comments#index"},
		{"GET", "/comments/:id", "comments#show"},
		// namespace :admin
		{"GET", "/admin/dashboard", "dashboard#index"},
		{"GET", "/admin/users", "users#index"},
		// scope '/v1'
		{"GET", "/v1/health", "health#check"},
		// match with via
		{"GET", "/legacy", "legacy#handle"},
		{"POST", "/legacy", "legacy#handle"},
		// root
		{"GET", "/", "home#index"},
	}

	for _, tc := range tests {
		if !findRoute(tc.method, tc.path, tc.handler) {
			t.Errorf("Missing route: %s %s -> %s", tc.method, tc.path, tc.handler)
		}
	}
}

func TestRubyParser_ParseRailsModels(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config/routes.rb to trigger Rails framework detection
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "routes.rb"), []byte("\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create app/models directory
	modelsDir := filepath.Join(tmpDir, "app", "models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create User model with validations and associations
	userModel := `class User < ApplicationRecord
  has_many :posts
  has_one :profile
  validates :name, presence: true
  validates :email, presence: true, uniqueness: true
end
`
	if err := os.WriteFile(filepath.Join(modelsDir, "user.rb"), []byte(userModel), 0644); err != nil {
		t.Fatal(err)
	}

	// Create Post model with belongs_to
	postModel := `class Post < ApplicationRecord
  belongs_to :user
  has_many :comments
  validates :title, presence: true, length: { minimum: 3 }
  validates :content, presence: true
end
`
	if err := os.WriteFile(filepath.Join(modelsDir, "post.rb"), []byte(postModel), 0644); err != nil {
		t.Fatal(err)
	}

	// Create Comment model
	commentModel := `class Comment < ApplicationRecord
  belongs_to :post
  belongs_to :user
  validates :body, presence: true
end
`
	if err := os.WriteFile(filepath.Join(modelsDir, "comment.rb"), []byte(commentModel), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(codebase.Models) != 3 {
		t.Errorf("Expected 3 models, got %d", len(codebase.Models))
	}

	// Verify User model
	var user *types.Model
	for i := range codebase.Models {
		if codebase.Models[i].Name == "User" {
			user = &codebase.Models[i]
			break
		}
	}
	if user == nil {
		t.Fatal("Expected User model not found")
	}

	// Check relations
	if len(user.Relations) != 2 {
		t.Errorf("Expected 2 relations for User, got %d", len(user.Relations))
	}
	hasPosts := false
	hasProfile := false
	for _, rel := range user.Relations {
		if rel.Type == "hasMany" && rel.RelatedModel == "posts" {
			hasPosts = true
		}
		if rel.Type == "hasOne" && rel.RelatedModel == "profile" {
			hasProfile = true
		}
	}
	if !hasPosts {
		t.Error("Missing has_many :posts relation on User")
	}
	if !hasProfile {
		t.Error("Missing has_one :profile relation on User")
	}

	// Check validations
	if len(user.Validation) != 3 {
		t.Errorf("Expected 3 validations for User, got %d", len(user.Validation))
	}

	// Verify Post model
	var post *types.Model
	for i := range codebase.Models {
		if codebase.Models[i].Name == "Post" {
			post = &codebase.Models[i]
			break
		}
	}
	if post == nil {
		t.Fatal("Expected Post model not found")
	}

	if len(post.Relations) != 2 {
		t.Errorf("Expected 2 relations for Post, got %d", len(post.Relations))
	}

	belongsToUser := false
	for _, rel := range post.Relations {
		if rel.Type == "belongsTo" && rel.RelatedModel == "user" {
			belongsToUser = true
		}
	}
	if !belongsToUser {
		t.Error("Missing belongs_to :user relation on Post")
	}

	// Check Comment model
	var comment *types.Model
	for i := range codebase.Models {
		if codebase.Models[i].Name == "Comment" {
			comment = &codebase.Models[i]
			break
		}
	}
	if comment == nil {
		t.Fatal("Expected Comment model not found")
	}

	if len(comment.Relations) != 2 {
		t.Errorf("Expected 2 relations for Comment, got %d", len(comment.Relations))
	}
}

func TestRubyParser_ParseRailsControllers(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config/routes.rb for Rails detection
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "routes.rb"), []byte("\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create app/controllers directory
	controllersDir := filepath.Join(tmpDir, "app", "controllers")
	if err := os.MkdirAll(controllersDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create UsersController
	usersController := `class UsersController < ApplicationController
  def index
    @users = User.all
  end

  def show
    @user = User.find(params[:id])
  end

  def create
    @user = User.new(user_params)
    if @user.save
      render json: @user, status: :created
    else
      render json: @user.errors, status: :unprocessable_entity
    end
  end

  def update
    @user = User.find(params[:id])
    @user.update!(user_params)
    render json: @user
  end

  def destroy
    @user = User.find(params[:id])
    @user.destroy!
    head :no_content
  end

  private

  def user_params
    params.require(:user).permit(:name, :email)
  end
end
`
	if err := os.WriteFile(filepath.Join(controllersDir, "users_controller.rb"), []byte(usersController), 0644); err != nil {
		t.Fatal(err)
	}

	// Create PostsController
	postsController := `class PostsController < ApplicationController
  def index
    @posts = Post.all
  end

  def show
    @post = Post.find(params[:id])
  end

  protected

  def set_post
    @post = Post.find(params[:id])
  end
end
`
	if err := os.WriteFile(filepath.Join(controllersDir, "posts_controller.rb"), []byte(postsController), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// UsersController: 5 public methods (index, show, create, update, destroy)
	// PostsController: 2 public methods (index, show) — set_post is protected
	// Total: 7
	if len(codebase.Handlers) != 7 {
		t.Errorf("Expected 7 handlers, got %d", len(codebase.Handlers))
		for _, h := range codebase.Handlers {
			t.Logf("  %s#%s", h.Controller, h.Name)
		}
	}

	// Verify UsersController handlers
	usersMethods := map[string]bool{}
	postsMethods := map[string]bool{}
	for _, h := range codebase.Handlers {
		if h.Controller == "UsersController" {
			usersMethods[h.Name] = true
		} else if h.Controller == "PostsController" {
			postsMethods[h.Name] = true
		}
	}

	expectedUsers := []string{"index", "show", "create", "update", "destroy"}
	for _, m := range expectedUsers {
		if !usersMethods[m] {
			t.Errorf("Expected UsersController#%s", m)
		}
	}

	// user_params should NOT be in handlers (private)
	if usersMethods["user_params"] {
		t.Error("user_params should not be extracted (private method)")
	}

	expectedPosts := []string{"index", "show"}
	for _, m := range expectedPosts {
		if !postsMethods[m] {
			t.Errorf("Expected PostsController#%s", m)
		}
	}

	// set_post should NOT be in handlers (protected)
	if postsMethods["set_post"] {
		t.Error("set_post should not be extracted (protected method)")
	}
}

func TestRubyParser_EmptyProject(t *testing.T) {
	tmpDir := t.TempDir()

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if codebase.Framework != "ruby" {
		t.Errorf("Expected framework ruby, got %q", codebase.Framework)
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
}

func TestRubyParser_NoRoutesFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create config directory but no routes.rb
	// Just Gemfile for Rails detection
	gemfile := `gem 'rails'`
	if err := os.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(gemfile), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if codebase.Framework != "rails" {
		t.Errorf("Expected framework rails, got %q", codebase.Framework)
	}

	// No routes file, so no routes but no error
	if len(codebase.Routes) != 0 {
		t.Errorf("Expected 0 routes, got %d", len(codebase.Routes))
	}
}

func TestRubyParser_DetectFrameworkMethod(t *testing.T) {
	tmpDir := t.TempDir()

	gemfile := `gem 'rails'`
	if err := os.WriteFile(filepath.Join(tmpDir, "Gemfile"), []byte(gemfile), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework, err := parser.DetectFramework(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if framework != "rails" {
		t.Errorf("Expected rails, got %s", framework)
	}
}
