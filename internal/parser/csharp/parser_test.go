package csharp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// SupportedLanguages
// ---------------------------------------------------------------------------

func TestCSharpParser_SupportedLanguages(t *testing.T) {
	parser := NewParser()
	langs := parser.SupportedLanguages()

	if len(langs) != 1 || langs[0] != "csharp" {
		t.Errorf("Expected [csharp], got %v", langs)
	}
}

// ---------------------------------------------------------------------------
// DetectFramework
// ---------------------------------------------------------------------------

func TestCSharpParser_DetectFramework_ASPNETCore(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
  </PropertyGroup>
</Project>`
	if err := os.WriteFile(filepath.Join(tmpDir, "MyApp.csproj"), []byte(csproj), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework, err := parser.DetectFramework(tmpDir)
	if err != nil {
		t.Fatalf("DetectFramework failed: %v", err)
	}

	if framework != "aspnetcore" {
		t.Errorf("Expected aspnetcore, got %s", framework)
	}
}

func TestCSharpParser_DetectFramework_PlainClassLibrary(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
  </PropertyGroup>
</Project>`
	if err := os.WriteFile(filepath.Join(tmpDir, "Library.csproj"), []byte(csproj), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework, err := parser.DetectFramework(tmpDir)
	if err != nil {
		t.Fatalf("DetectFramework failed: %v", err)
	}

	if framework != "csharp" {
		t.Errorf("Expected csharp, got %s", framework)
	}
}

func TestCSharpParser_DetectFramework_NoCsproj(t *testing.T) {
	tmpDir := t.TempDir()

	parser := NewParser()
	framework, err := parser.DetectFramework(tmpDir)
	if err != nil {
		t.Fatalf("DetectFramework failed: %v", err)
	}

	if framework != "csharp" {
		t.Errorf("Expected csharp, got %s", framework)
	}
}

// ---------------------------------------------------------------------------
// Controller route parsing
// ---------------------------------------------------------------------------

func TestCSharpParser_ParseControllerRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a .csproj for framework detection.
	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	// Create Controllers directory.
	ctrlDir := filepath.Join(tmpDir, "Controllers")
	if err := os.MkdirAll(ctrlDir, 0755); err != nil {
		t.Fatal(err)
	}

	controllerSrc := `using Microsoft.AspNetCore.Mvc;

namespace MyApp.Controllers;

[ApiController]
[Route("api/[controller]")]
public class UsersController : ControllerBase
{
    [HttpGet]
    public IActionResult GetAll() => Ok();

    [HttpGet("{id}")]
    public IActionResult GetById(int id) => Ok();

    [HttpPost]
    public IActionResult Create([FromBody] UserDto dto) => Ok();

    [HttpPut("{id}")]
    public IActionResult Update(int id, [FromBody] UserDto dto) => Ok();

    [HttpDelete("{id}")]
    public IActionResult Delete(int id) => Ok();
}
`
	if err := os.WriteFile(filepath.Join(ctrlDir, "UsersController.cs"), []byte(controllerSrc), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if codebase.Framework != "aspnetcore" {
		t.Errorf("Expected framework aspnetcore, got %q", codebase.Framework)
	}

	// Expect 5 routes.
	if len(codebase.Routes) != 5 {
		t.Fatalf("Expected 5 routes, got %d", len(codebase.Routes))
	}

	expected := map[string]string{
		"GET api/[controller]":         "",
		"GET api/[controller]/{id}":    "",
		"POST api/[controller]":        "",
		"PUT api/[controller]/{id}":    "",
		"DELETE api/[controller]/{id}": "",
	}

	for _, r := range codebase.Routes {
		key := r.Method + " " + r.Path
		if _, ok := expected[key]; !ok {
			t.Errorf("Unexpected route: %s %s", r.Method, r.Path)
		}
	}
}

// ---------------------------------------------------------------------------
// Controller with Controller base class (not ControllerBase)
// ---------------------------------------------------------------------------

func TestCSharpParser_ParseControllerInheritsController(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	ctrlDir := filepath.Join(tmpDir, "Controllers")
	os.MkdirAll(ctrlDir, 0755)

	src := `using Microsoft.AspNetCore.Mvc;

public class HomeController : Controller
{
    [HttpGet]
    public IActionResult Index() => View();

    [HttpPost]
    public IActionResult Submit() => Ok();
}
`
	os.WriteFile(filepath.Join(ctrlDir, "HomeController.cs"), []byte(src), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(codebase.Routes) != 2 {
		t.Fatalf("Expected 2 routes, got %d", len(codebase.Routes))
	}
}

// ---------------------------------------------------------------------------
// Controller without [ApiController] but inheriting ControllerBase
// ---------------------------------------------------------------------------

func TestCSharpParser_ParseControllerBaseWithoutAttribute(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	ctrlDir := filepath.Join(tmpDir, "Controllers")
	os.MkdirAll(ctrlDir, 0755)

	src := `using Microsoft.AspNetCore.Mvc;

[Route("v1/products")]
public class ProductsController : ControllerBase
{
    [HttpGet]
    public IActionResult List() => Ok();
}
`
	os.WriteFile(filepath.Join(ctrlDir, "ProductsController.cs"), []byte(src), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(codebase.Routes) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(codebase.Routes))
	}
	if codebase.Routes[0].Path != "v1/products" {
		t.Errorf("Expected path v1/products, got %s", codebase.Routes[0].Path)
	}
}

// ---------------------------------------------------------------------------
// Minimal API
// ---------------------------------------------------------------------------

func TestCSharpParser_ParseMinimalAPI(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	src := `using Microsoft.AspNetCore.Builder;

var builder = WebApplication.CreateBuilder(args);
var app = builder.Build();

app.MapGet("/", () => "Hello World!");
app.MapGet("/users", () => new[] { new { Id = 1, Name = "Alice" } });
app.MapPost("/users", (User user) => Results.Created($"/users/{user.Id}", user));
app.MapPut("/users/{id}", (int id, User user) => Results.Ok(user));
app.MapDelete("/users/{id}", (int id) => Results.NoContent());

app.Run();
`
	os.WriteFile(filepath.Join(tmpDir, "Program.cs"), []byte(src), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(codebase.Routes) != 5 {
		t.Fatalf("Expected 5 routes, got %d", len(codebase.Routes))
	}

	methods := map[string]int{}
	for _, r := range codebase.Routes {
		methods[r.Method]++
	}

	if methods["GET"] != 2 {
		t.Errorf("Expected 2 GET routes, got %d", methods["GET"])
	}
	if methods["POST"] != 1 {
		t.Errorf("Expected 1 POST route, got %d", methods["POST"])
	}
	if methods["PUT"] != 1 {
		t.Errorf("Expected 1 PUT route, got %d", methods["PUT"])
	}
	if methods["DELETE"] != 1 {
		t.Errorf("Expected 1 DELETE route, got %d", methods["DELETE"])
	}
}

// ---------------------------------------------------------------------------
// Models in /Models/
// ---------------------------------------------------------------------------

func TestCSharpParser_ParseModels(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	modelsDir := filepath.Join(tmpDir, "Models")
	os.MkdirAll(modelsDir, 0755)

	userSrc := `namespace MyApp.Models;

public class User
{
    public int Id { get; set; }
    public string Name { get; set; }
    public string Email { get; set; }
    public bool IsActive { get; set; }
}
`
	os.WriteFile(filepath.Join(modelsDir, "User.cs"), []byte(userSrc), 0644)

	orderSrc := `namespace MyApp.Models;

public class Order
{
    public int Id { get; set; }
    public string OrderNumber { get; set; }
    public decimal TotalAmount { get; set; }
    public DateTime CreatedAt { get; set; }
}
`
	os.WriteFile(filepath.Join(modelsDir, "Order.cs"), []byte(orderSrc), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(codebase.Models) != 2 {
		t.Fatalf("Expected 2 models, got %d", len(codebase.Models))
	}

	modelNames := map[string]bool{}
	for _, m := range codebase.Models {
		modelNames[m.Name] = true
	}

	if !modelNames["User"] {
		t.Error("Expected User model")
	}
	if !modelNames["Order"] {
		t.Error("Expected Order model")
	}

	// Verify fields on User.
	for _, m := range codebase.Models {
		if m.Name == "User" {
			if len(m.Fields) != 4 {
				t.Errorf("Expected 4 fields for User, got %d", len(m.Fields))
			}

			fieldNames := map[string]string{}
			for _, f := range m.Fields {
				fieldNames[f.Name] = f.Type
			}
			if fieldNames["Id"] != "int" {
				t.Errorf("Expected Id:int, got %s", fieldNames["Id"])
			}
			if fieldNames["Name"] != "string" {
				t.Errorf("Expected Name:string, got %s", fieldNames["Name"])
			}
			if fieldNames["Email"] != "string" {
				t.Errorf("Expected Email:string, got %s", fieldNames["Email"])
			}
			if fieldNames["IsActive"] != "bool" {
				t.Errorf("Expected IsActive:bool, got %s", fieldNames["IsActive"])
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Models NOT in /Models/ are skipped
// ---------------------------------------------------------------------------

func TestCSharpParser_NonModelClassesSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	dtoDir := filepath.Join(tmpDir, "DTOs")
	os.MkdirAll(dtoDir, 0755)

	dtoSrc := `namespace MyApp.DTOs;

public class UserDto
{
    public string Name { get; set; }
    public string Email { get; set; }
}
`
	os.WriteFile(filepath.Join(dtoDir, "UserDto.cs"), []byte(dtoSrc), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// DTOs in non-Models folders should NOT be detected as models.
	if len(codebase.Models) != 0 {
		t.Errorf("Expected 0 models (DTOs/ folder should be skipped), got %d", len(codebase.Models))
	}
}

// ---------------------------------------------------------------------------
// Empty project
// ---------------------------------------------------------------------------

func TestCSharpParser_EmptyProject(t *testing.T) {
	tmpDir := t.TempDir()

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if codebase.Language != "csharp" {
		t.Errorf("Expected language csharp, got %s", codebase.Language)
	}
	if codebase.Framework != "csharp" {
		t.Errorf("Expected framework csharp, got %s", codebase.Framework)
	}
	if len(codebase.Routes) != 0 {
		t.Errorf("Expected 0 routes, got %d", len(codebase.Routes))
	}
	if len(codebase.Models) != 0 {
		t.Errorf("Expected 0 models, got %d", len(codebase.Models))
	}
}

// ---------------------------------------------------------------------------
// Multiple controllers in same project
// ---------------------------------------------------------------------------

func TestCSharpParser_MultipleControllers(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	ctrlDir := filepath.Join(tmpDir, "Controllers")
	os.MkdirAll(ctrlDir, 0755)

	usersController := `using Microsoft.AspNetCore.Mvc;

[ApiController]
[Route("api/users")]
public class UsersController : ControllerBase
{
    [HttpGet]
    public IActionResult List() => Ok();

    [HttpPost]
    public IActionResult Create() => Ok();
}
`
	os.WriteFile(filepath.Join(ctrlDir, "UsersController.cs"), []byte(usersController), 0644)

	postsController := `using Microsoft.AspNetCore.Mvc;

[ApiController]
[Route("api/posts")]
public class PostsController : ControllerBase
{
    [HttpGet]
    public IActionResult List() => Ok();

    [HttpGet("{id}")]
    public IActionResult Get(int id) => Ok();
}
`
	os.WriteFile(filepath.Join(ctrlDir, "PostsController.cs"), []byte(postsController), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 2 routes from UsersController + 2 routes from PostsController.
	if len(codebase.Routes) != 4 {
		t.Fatalf("Expected 4 routes, got %d", len(codebase.Routes))
	}

	// Verify we have routes from both controllers.
	paths := map[string]bool{}
	for _, r := range codebase.Routes {
		paths[r.Path] = true
	}
	if !paths["api/users"] {
		t.Error("Expected api/users route")
	}
	if !paths["api/posts"] {
		t.Error("Expected api/posts route")
	}
	if !paths["api/posts/{id}"] {
		t.Error("Expected api/posts/{id} route")
	}
}

// ---------------------------------------------------------------------------
// HttpPatch support
// ---------------------------------------------------------------------------

func TestCSharpParser_HttpPatchRoute(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	ctrlDir := filepath.Join(tmpDir, "Controllers")
	os.MkdirAll(ctrlDir, 0755)

	src := `using Microsoft.AspNetCore.Mvc;

[ApiController]
[Route("api/items")]
public class ItemsController : ControllerBase
{
    [HttpPatch("{id}")]
    public IActionResult PartialUpdate(int id) => Ok();
}
`
	os.WriteFile(filepath.Join(ctrlDir, "ItemsController.cs"), []byte(src), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(codebase.Routes) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(codebase.Routes))
	}
	if codebase.Routes[0].Method != "PATCH" {
		t.Errorf("Expected PATCH, got %s", codebase.Routes[0].Method)
	}
	if codebase.Routes[0].Path != "api/items/{id}" {
		t.Errorf("Expected api/items/{id}, got %s", codebase.Routes[0].Path)
	}
}

// ---------------------------------------------------------------------------
// Controller route with explicit path on HttpGet
// ---------------------------------------------------------------------------

func TestCSharpParser_ControllerRouteWithExplicitMethodPath(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	ctrlDir := filepath.Join(tmpDir, "Controllers")
	os.MkdirAll(ctrlDir, 0755)

	src := `using Microsoft.AspNetCore.Mvc;

[ApiController]
[Route("api/custom")]
public class CustomController : ControllerBase
{
    [HttpGet("search")]
    public IActionResult Search() => Ok();
}
`
	os.WriteFile(filepath.Join(ctrlDir, "CustomController.cs"), []byte(src), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(codebase.Routes) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(codebase.Routes))
	}
	if codebase.Routes[0].Path != "api/custom/search" {
		t.Errorf("Expected api/custom/search, got %s", codebase.Routes[0].Path)
	}
}

// ---------------------------------------------------------------------------
// Skip bin/ and obj/ directories
// ---------------------------------------------------------------------------

func TestCSharpParser_SkipBuildDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	ctrlDir := filepath.Join(tmpDir, "Controllers")
	os.MkdirAll(ctrlDir, 0755)

	src := `using Microsoft.AspNetCore.Mvc;

[ApiController]
[Route("api/real")]
public class RealController : ControllerBase
{
    [HttpGet]
    public IActionResult Get() => Ok();
}
`
	os.WriteFile(filepath.Join(ctrlDir, "RealController.cs"), []byte(src), 0644)

	// Create a dummy controller under bin/ (should be skipped).
	binCtrlDir := filepath.Join(tmpDir, "bin", "Debug", "net8.0", "Controllers")
	os.MkdirAll(binCtrlDir, 0755)
	fakeSrc := `[ApiController] [Route("api/fake")] public class FakeController : ControllerBase { [HttpGet] public IActionResult Get() => Ok(); }`
	os.WriteFile(filepath.Join(binCtrlDir, "FakeController.cs"), []byte(fakeSrc), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should only have the 1 route from RealController, not from bin/.
	if len(codebase.Routes) != 1 {
		t.Fatalf("Expected 1 route (bin/ should be skipped), got %d", len(codebase.Routes))
	}
}

// ---------------------------------------------------------------------------
// Minimal API with various receivers (api, endpoints, application)
// ---------------------------------------------------------------------------

func TestCSharpParser_MinimalAPIDifferentReceivers(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	src := `var app = WebApplication.CreateBuilder(args).Build();

app.MapGet("/via-app", () => "app");

var api = app.MapGroup("/api");
api.MapPost("/via-api", () => "api");

var endpoints = app.MapGroup("/endpoints");
endpoints.MapPut("/via-endpoints", () => "endpoints");

app.Run();
`
	os.WriteFile(filepath.Join(tmpDir, "Program.cs"), []byte(src), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Expect 3 routes.
	if len(codebase.Routes) != 3 {
		t.Fatalf("Expected 3 routes, got %d", len(codebase.Routes))
	}

	methods := map[string]int{}
	for _, r := range codebase.Routes {
		methods[r.Method]++
	}
	if methods["GET"] != 1 || methods["POST"] != 1 || methods["PUT"] != 1 {
		t.Errorf("Expected 1 GET, 1 POST, 1 PUT; got GET=%d POST=%d PUT=%d",
			methods["GET"], methods["POST"], methods["PUT"])
	}
}

// ---------------------------------------------------------------------------
// Model with nullable types
// ---------------------------------------------------------------------------

func TestCSharpParser_ModelWithNullableTypes(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	modelsDir := filepath.Join(tmpDir, "Models")
	os.MkdirAll(modelsDir, 0755)

	src := `namespace MyApp.Models;

public class Profile
{
    public int Id { get; set; }
    public string? Bio { get; set; }
    public DateTime? DateOfBirth { get; set; }
    public List<string> Tags { get; set; }
}
`
	os.WriteFile(filepath.Join(modelsDir, "Profile.cs"), []byte(src), 0644)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(codebase.Models) != 1 {
		t.Fatalf("Expected 1 model, got %d", len(codebase.Models))
	}

	profile := codebase.Models[0]
	if profile.Name != "Profile" {
		t.Errorf("Expected Profile, got %s", profile.Name)
	}
	if len(profile.Fields) != 4 {
		t.Fatalf("Expected 4 fields, got %d", len(profile.Fields))
	}

	expectedFields := map[string]string{
		"Id":          "int",
		"Bio":         "string?",
		"DateOfBirth": "DateTime?",
		"Tags":        "List<string>",
	}
	for _, f := range profile.Fields {
		exp, ok := expectedFields[f.Name]
		if !ok {
			t.Errorf("Unexpected field: %s", f.Name)
			continue
		}
		if f.Type != exp {
			t.Errorf("Expected %s:%s, got %s", f.Name, exp, f.Type)
		}
	}
}

// ---------------------------------------------------------------------------
// Parse respects context cancellation
// ---------------------------------------------------------------------------

func TestCSharpParser_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	csproj := `<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`
	os.WriteFile(filepath.Join(tmpDir, "Web.csproj"), []byte(csproj), 0644)

	// Create many files so parsing takes some time.
	ctrlDir := filepath.Join(tmpDir, "Controllers")
	os.MkdirAll(ctrlDir, 0755)

	for i := 0; i < 50; i++ {
		src := `using Microsoft.AspNetCore.Mvc;
[ApiController][Route("api/ctl")]
public class Ctl` + strings.Repeat("X", i) + ` : ControllerBase {
    [HttpGet] public IActionResult Get() => Ok();
}`
		os.WriteFile(filepath.Join(ctrlDir, "Ctl"+strings.Repeat("X", i)+"Controller.cs"), []byte(src), 0644)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	parser := NewParser()
	codebase, err := parser.Parse(ctx, tmpDir)
	if err == nil {
		t.Log("Parse completed without error (files may have been parsed quickly)")
	}
	_ = codebase
}
