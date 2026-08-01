package java

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJavaParser_SupportedLanguages(t *testing.T) {
	parser := NewParser()
	langs := parser.SupportedLanguages()

	if len(langs) != 1 || langs[0] != "java" {
		t.Errorf("Expected [java], got %v", langs)
	}
}

func TestJavaParser_DetectFramework_SpringBootMaven(t *testing.T) {
	tmpDir := t.TempDir()

	pomXML := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
    </dependencies>
</project>`
	if err := os.WriteFile(filepath.Join(tmpDir, "pom.xml"), []byte(pomXML), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "spring-boot" {
		t.Errorf("Expected spring-boot, got %s", framework)
	}
}

func TestJavaParser_DetectFramework_SpringBootGradle(t *testing.T) {
	tmpDir := t.TempDir()

	buildGradle := `plugins {
    id 'org.springframework.boot' version '3.2.0'
}

dependencies {
    implementation 'org.springframework.boot:spring-boot-starter-web'
}`
	if err := os.WriteFile(filepath.Join(tmpDir, "build.gradle"), []byte(buildGradle), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "spring-boot" {
		t.Errorf("Expected spring-boot, got %s", framework)
	}
}

func TestJavaParser_DetectFramework_Maven(t *testing.T) {
	tmpDir := t.TempDir()

	pomXML := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <dependencies>
        <dependency>
            <groupId>com.google.guava</groupId>
            <artifactId>guava</artifactId>
            <version>31.1-jre</version>
        </dependency>
    </dependencies>
</project>`
	if err := os.WriteFile(filepath.Join(tmpDir, "pom.xml"), []byte(pomXML), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "maven" {
		t.Errorf("Expected maven, got %s", framework)
	}
}

func TestJavaParser_DetectFramework_Gradle(t *testing.T) {
	tmpDir := t.TempDir()

	buildGradle := `plugins {
    id 'java'
}

dependencies {
    implementation 'com.google.guava:guava:31.1-jre'
}`
	if err := os.WriteFile(filepath.Join(tmpDir, "build.gradle"), []byte(buildGradle), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "gradle" {
		t.Errorf("Expected gradle, got %s", framework)
	}
}

func TestJavaParser_DetectFramework_SpringAnnotations(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a Java source file with @RestController
	srcDir := filepath.Join(tmpDir, "src", "main", "java", "com", "example")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	controllerJava := `package com.example;

import org.springframework.web.bind.annotation.*;

@RestController
public class HelloController {

    @GetMapping("/hello")
    public String hello() {
        return "Hello World";
    }
}`
	if err := os.WriteFile(filepath.Join(srcDir, "HelloController.java"), []byte(controllerJava), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "spring-boot" {
		t.Errorf("Expected spring-boot, got %s", framework)
	}
}

func TestJavaParser_DetectFramework_PlainJava(t *testing.T) {
	tmpDir := t.TempDir()

	parser := NewParser()
	framework := parser.detectFramework(tmpDir)

	if framework != "java" {
		t.Errorf("Expected java, got %s", framework)
	}
}

func TestJavaParser_DetectFramework_PublicAPI(t *testing.T) {
	tmpDir := t.TempDir()

	pomXML := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
</project>`
	if err := os.WriteFile(filepath.Join(tmpDir, "pom.xml"), []byte(pomXML), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	framework, err := parser.DetectFramework(tmpDir)
	if err != nil {
		t.Fatalf("DetectFramework failed: %v", err)
	}

	if framework != "spring-boot" {
		t.Errorf("Expected spring-boot, got %s", framework)
	}
}

func TestJavaParser_ParseSpringBootController(t *testing.T) {
	tmpDir := creatSpringBootProject(t)

	// Create a controller
	userController := `package com.example.controller;

import org.springframework.web.bind.annotation.*;
import java.util.*;

@RestController
@RequestMapping("/api/users")
public class UserController {

    @GetMapping
    public List<String> getAllUsers() {
        return Collections.emptyList();
    }

    @GetMapping("/{id}")
    public String getUserById(@PathVariable Long id) {
        return "user-" + id;
    }

    @PostMapping
    public String createUser(@RequestBody String body) {
        return "created";
    }

    @PutMapping("/{id}")
    public String updateUser(@PathVariable Long id, @RequestBody String body) {
        return "updated-" + id;
    }

    @DeleteMapping("/{id}")
    public void deleteUser(@PathVariable Long id) {
        // delete user
    }
}`
	writeJavaFile(t, tmpDir, "com/example/controller", "UserController.java", userController)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if codebase.Framework != "spring-boot" {
		t.Errorf("Expected framework spring-boot, got %s", codebase.Framework)
	}

	if codebase.Language != "java" {
		t.Errorf("Expected language java, got %s", codebase.Language)
	}

	// Should have 5 routes
	if len(codebase.Routes) != 5 {
		t.Errorf("Expected 5 routes, got %d", len(codebase.Routes))
	}

	// Verify specific routes
	routeMap := make(map[string]string) // method+path -> handler
	for _, route := range codebase.Routes {
		key := route.Method + " " + route.Path
		routeMap[key] = route.Handler
	}

	expectedRoutes := map[string]bool{
		"GET /api/users":         true,
		"GET /api/users/{id}":    true,
		"POST /api/users":        true,
		"PUT /api/users/{id}":    true,
		"DELETE /api/users/{id}": true,
	}

	for expected := range expectedRoutes {
		if _, ok := routeMap[expected]; !ok {
			t.Errorf("Expected route %s not found", expected)
		}
	}

	// Should have 5 handlers
	if len(codebase.Handlers) != 5 {
		t.Errorf("Expected 5 handlers, got %d", len(codebase.Handlers))
	}

	// Check handler naming
	handlerNames := make(map[string]bool)
	for _, h := range codebase.Handlers {
		if h.Controller != "UserController" {
			t.Errorf("Expected controller UserController, got %s", h.Controller)
		}
		handlerNames[h.Name] = true
	}

	expectedNames := []string{"getAllUsers", "getUserById", "createUser", "updateUser", "deleteUser"}
	for _, name := range expectedNames {
		if !handlerNames[name] {
			t.Errorf("Expected handler %s", name)
		}
	}
}

func TestJavaParser_ParseSpringBootControllerNoBasePath(t *testing.T) {
	tmpDir := creatSpringBootProject(t)

	// Controller without @RequestMapping base path
	healthController := `package com.example.controller;

import org.springframework.web.bind.annotation.*;

@RestController
public class HealthController {

    @GetMapping("/health")
    public String health() {
        return "OK";
    }

    @GetMapping("/ping")
    public String ping() {
        return "pong";
    }
}`
	writeJavaFile(t, tmpDir, "com/example/controller", "HealthController.java", healthController)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(codebase.Routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(codebase.Routes))
	}

	foundHealth := false
	foundPing := false
	for _, route := range codebase.Routes {
		if route.Method == "GET" && route.Path == "/health" {
			foundHealth = true
		}
		if route.Method == "GET" && route.Path == "/ping" {
			foundPing = true
		}
	}

	if !foundHealth {
		t.Error("Expected GET /health route")
	}
	if !foundPing {
		t.Error("Expected GET /ping route")
	}
}

func TestJavaParser_ParseMultipleControllers(t *testing.T) {
	tmpDir := creatSpringBootProject(t)

	userController := `package com.example.controller;

import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/users")
public class UserController {
    @GetMapping
    public String list() { return "[]"; }

    @PostMapping
    public String create() { return "ok"; }
}`
	writeJavaFile(t, tmpDir, "com/example/controller", "UserController.java", userController)

	productController := `package com.example.controller;

import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/products")
public class ProductController {
    @GetMapping
    public String list() { return "[]"; }

    @GetMapping("/{id}")
    public String get(@PathVariable Long id) { return "{}"; }
}`
	writeJavaFile(t, tmpDir, "com/example/controller", "ProductController.java", productController)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// 2 from UserController + 2 from ProductController = 4 routes
	if len(codebase.Routes) != 4 {
		t.Errorf("Expected 4 routes, got %d", len(codebase.Routes))
	}

	// Verify controllers are properly attributed
	controllers := make(map[string]int)
	for _, route := range codebase.Routes {
		parts := strings.Split(route.Handler, ".")
		if len(parts) == 2 {
			controllers[parts[0]]++
		}
	}

	if controllers["UserController"] != 2 {
		t.Errorf("Expected 2 routes from UserController, got %d", controllers["UserController"])
	}
	if controllers["ProductController"] != 2 {
		t.Errorf("Expected 2 routes from ProductController, got %d", controllers["ProductController"])
	}
}

func TestJavaParser_ParseSpringBootEntity(t *testing.T) {
	tmpDir := creatSpringBootProject(t)

	userEntity := `package com.example.entity;

import jakarta.persistence.*;

@Entity
@Table(name = "users")
public class User {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "username", nullable = false, unique = true)
    private String username;

    @Column(name = "email")
    private String email;

    @Column(name = "created_at")
    private java.time.LocalDateTime createdAt;

    private String nickname;
}`
	writeJavaFile(t, tmpDir, "com/example/entity", "User.java", userEntity)

	productEntity := `package com.example.entity;

import jakarta.persistence.*;

@Entity
@Table(name = "products")
public class Product {

    @Id
    private Long id;

    @Column(name = "product_name")
    private String name;

    @Column(name = "price")
    private java.math.BigDecimal price;
}`
	writeJavaFile(t, tmpDir, "com/example/entity", "Product.java", productEntity)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(codebase.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(codebase.Models))
	}

	// Find User model
	foundUser := false
	for _, model := range codebase.Models {
		if model.Name == "User" {
			foundUser = true
			if model.Table != "users" {
				t.Errorf("Expected User table 'users', got %s", model.Table)
			}
			if len(model.Fields) != 3 {
				t.Errorf("Expected 3 @Column fields for User, got %d", len(model.Fields))
			}
			break
		}
	}
	if !foundUser {
		t.Error("Expected User model")
	}

	// Find Product model
	foundProduct := false
	for _, model := range codebase.Models {
		if model.Name == "Product" {
			foundProduct = true
			if model.Table != "products" {
				t.Errorf("Expected Product table 'products', got %s", model.Table)
			}
			if len(model.Fields) != 2 {
				t.Errorf("Expected 2 @Column fields for Product, got %d", len(model.Fields))
			}
			break
		}
	}
	if !foundProduct {
		t.Error("Expected Product model")
	}
}

func TestJavaParser_ParseEntityPlainFields(t *testing.T) {
	tmpDir := creatSpringBootProject(t)

	// Entity without @Column annotations — should fall back to plain field scanning
	categoryEntity := `package com.example.entity;

import jakarta.persistence.*;

@Entity
@Table(name = "categories")
public class Category {

    @Id
    private Long id;

    private String name;

    private String description;

    private int sortOrder;

    // Skip static/logger
    private static final long serialVersionUID = 1L;
}`
	writeJavaFile(t, tmpDir, "com/example/entity", "Category.java", categoryEntity)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(codebase.Models) != 1 {
		t.Fatalf("Expected 1 model, got %d", len(codebase.Models))
	}

	if codebase.Models[0].Name != "Category" {
		t.Errorf("Expected Category model, got %s", codebase.Models[0].Name)
	}

	if codebase.Models[0].Table != "categories" {
		t.Errorf("Expected table 'categories', got %s", codebase.Models[0].Table)
	}

	// Should have 4 fields (id, name, description, sortOrder) — serialVersionUID skipped
	if len(codebase.Models[0].Fields) != 4 {
		t.Errorf("Expected 3 fields, got %d: %+v", len(codebase.Models[0].Fields), codebase.Models[0].Fields)
	}
}

func TestJavaParser_ParseRepository(t *testing.T) {
	tmpDir := creatSpringBootProject(t)

	userRepo := `package com.example.repository;

import com.example.entity.User;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface UserRepository extends JpaRepository<User, Long> {
}`
	writeJavaFile(t, tmpDir, "com/example/repository", "UserRepository.java", userRepo)

	productRepo := `package com.example.repository;

import com.example.entity.Product;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface ProductRepository extends JpaRepository<Product, Long> {
}`
	writeJavaFile(t, tmpDir, "com/example/repository", "ProductRepository.java", productRepo)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Repositories are stored as handlers (optional)
	// We should have 2 repository handlers
	hasUserRepo := false
	hasProductRepo := false
	for _, handler := range codebase.Handlers {
		if handler.Name == "UserRepository" {
			hasUserRepo = true
			if handler.Method != "REPOSITORY" {
				t.Errorf("Expected method REPOSITORY for UserRepository, got %s", handler.Method)
			}
		}
		if handler.Name == "ProductRepository" {
			hasProductRepo = true
		}
	}

	if !hasUserRepo {
		t.Error("Expected UserRepository handler")
	}
	if !hasProductRepo {
		t.Error("Expected ProductRepository handler")
	}
}

func TestJavaParser_EmptyProject(t *testing.T) {
	tmpDir := creatSpringBootProject(t)
	// No Java source files created — just the pom.xml

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
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

func TestJavaParser_SkipTestFiles(t *testing.T) {
	tmpDir := creatSpringBootProject(t)

	// Create a controller
	controller := `package com.example.controller;

import org.springframework.web.bind.annotation.*;

@RestController
public class HelloController {
    @GetMapping("/hello")
    public String hello() { return "Hello"; }
}`
	writeJavaFile(t, tmpDir, "com/example/controller", "HelloController.java", controller)

	// Create a test file (should be skipped)
	testFile := `package com.example.controller;

import org.junit.jupiter.api.Test;

class HelloControllerTest {
    @Test
    void testHello() {
        // test code
    }
}`
	writeJavaFile(t, tmpDir, "com/example/controller", "HelloControllerTest.java", testFile)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should only have 1 route from HelloController
	if len(codebase.Routes) != 1 {
		t.Errorf("Expected 1 route (test files skipped), got %d", len(codebase.Routes))
	}
}

func TestJavaParser_ComplexController(t *testing.T) {
	tmpDir := creatSpringBootProject(t)

	// Controller with @RequestMapping method-level usage
	complexController := `package com.example.controller;

import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api")
public class ComplexController {

    @GetMapping("/data")
    public String getData() { return "data"; }

    @RequestMapping(value = "/legacy", method = RequestMethod.POST)
    public String legacyEndpoint() { return "legacy"; }

    @RequestMapping(value = "/multi", method = {RequestMethod.GET, RequestMethod.POST})
    public String multiMethod() { return "multi"; }
}`
	writeJavaFile(t, tmpDir, "com/example/controller", "ComplexController.java", complexController)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(codebase.Routes) < 3 {
		t.Errorf("Expected at least 3 routes, got %d", len(codebase.Routes))
	}

	// Check GET /api/data
	foundGetData := false
	foundLegacy := false
	for _, route := range codebase.Routes {
		if route.Method == "GET" && route.Path == "/api/data" {
			foundGetData = true
		}
		if route.Path == "/api/legacy" {
			foundLegacy = true
		}
	}

	if !foundGetData {
		t.Error("Expected GET /api/data route")
	}
	if !foundLegacy {
		t.Error("Expected /api/legacy route")
	}
}

func TestJavaParser_CodebaseMetadata(t *testing.T) {
	tmpDir := creatSpringBootProject(t)

	controller := `package com.example;

import org.springframework.web.bind.annotation.*;

@RestController
public class AppController {
    @GetMapping("/app")
    public String app() { return "App"; }
}`
	writeJavaFile(t, tmpDir, "com/example", "AppController.java", controller)

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if codebase.Language != "java" {
		t.Errorf("Expected language java, got %s", codebase.Language)
	}
	if codebase.Framework != "spring-boot" {
		t.Errorf("Expected framework spring-boot, got %s", codebase.Framework)
	}
	if codebase.RootDir != tmpDir {
		t.Errorf("Expected RootDir %s, got %s", tmpDir, codebase.RootDir)
	}
	if codebase.FileCount != 1 {
		t.Errorf("Expected FileCount 1, got %d", codebase.FileCount)
	}
	if codebase.AnalyzedAt.IsZero() {
		t.Error("Expected AnalyzedAt to be set")
	}
}

// Helper: create a minimal Spring Boot project structure with pom.xml.
func creatSpringBootProject(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	pomXML := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
        </dependency>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-data-jpa</artifactId>
        </dependency>
    </dependencies>
</project>`
	if err := os.WriteFile(filepath.Join(tmpDir, "pom.xml"), []byte(pomXML), 0644); err != nil {
		t.Fatal(err)
	}

	return tmpDir
}

// Helper: write a Java source file in the project structure.
func writeJavaFile(t *testing.T, projectDir, pkg, filename, content string) {
	t.Helper()
	srcDir := filepath.Join(projectDir, "src", "main", "java", pkg)
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(srcDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
