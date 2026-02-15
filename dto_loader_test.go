package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadResourceDTOs(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantCount int
		wantErr   bool
		validate  func(t *testing.T, resources map[string]resourceDTOs)
	}{
		{
			name:      "load DTOs from valid directory",
			setupFunc: setupValidDTOsDirectory,
			wantCount: 2,
			wantErr:   false,
			validate:  validateUserAndProductResources,
		},
		{
			name:      "skip non-Go files",
			setupFunc: setupDirectoryWithNonGoFiles,
			wantCount: 1,
			wantErr:   false,
			validate:  validateNonGoFilesSkipped,
		},
		{
			name:      "skip files without DTOs",
			setupFunc: setupDirectoryWithNonDTOFiles,
			wantCount: 1,
			wantErr:   false,
			validate:  validateFilesWithoutDTOsSkipped,
		},
		{
			name:      "handle directory with multiple DTOs per file",
			setupFunc: setupDirectoryWithMultipleDTOs,
			wantCount: 1,
			wantErr:   false,
			validate:  validateMultipleDTOsInFile,
		},
		{
			name:      "empty directory returns empty map",
			setupFunc: func(t *testing.T) string { return t.TempDir() },
			wantCount: 0,
			wantErr:   false,
			validate:  func(t *testing.T, resources map[string]resourceDTOs) {},
		},
		{
			name:      "non-existent directory returns error",
			setupFunc: func(t *testing.T) string { return "/non/existent/directory/path" },
			wantCount: 0,
			wantErr:   true,
			validate:  func(t *testing.T, resources map[string]resourceDTOs) {},
		},
		{
			name:      "pluralization works correctly",
			setupFunc: setupCategoryDTO,
			wantCount: 1,
			wantErr:   false,
			validate:  validatePluralization,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dtosDir := tt.setupFunc(t)

			got, err := loadResourceDTOs(dtosDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadResourceDTOs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != tt.wantCount {
					t.Errorf("loadResourceDTOs() returned %d resources, want %d", len(got), tt.wantCount)
				}

				if tt.validate != nil {
					tt.validate(t, got)
				}
			}
		})
	}
}

func setupValidDTOsDirectory(t *testing.T) string {
	tempDir := t.TempDir()

	userContent := `package dto

type UserDTO struct {
	ID    int64  ` + "`json:\"id\"`" + `
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}`
	err := os.WriteFile(filepath.Join(tempDir, "user.go"), []byte(userContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create user.go: %v", err)
	}

	productContent := `package dto

type ProductDTO struct {
	ID    int64   ` + "`json:\"id\"`" + `
	Name  string  ` + "`json:\"name\"`" + `
	Price float64 ` + "`json:\"price\"`" + `
}`
	err = os.WriteFile(filepath.Join(tempDir, "product.go"), []byte(productContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create product.go: %v", err)
	}

	return tempDir
}

func validateUserAndProductResources(t *testing.T, resources map[string]resourceDTOs) {
	validateUserResource(t, resources)
	validateProductResource(t, resources)
}

func validateUserResource(t *testing.T, resources map[string]resourceDTOs) {
	user, exists := resources["user"]
	if !exists {
		t.Error("Expected user resource not found")
		return
	}
	if user.Name != "user" {
		t.Errorf("user.Name = %v, want user", user.Name)
	}
	if user.PluralName != "users" {
		t.Errorf("user.PluralName = %v, want users", user.PluralName)
	}
	if _, exists := user.DTOs["UserDTO"]; !exists {
		t.Error("UserDTO not found in user resource")
	}
}

func validateProductResource(t *testing.T, resources map[string]resourceDTOs) {
	product, exists := resources["product"]
	if !exists {
		t.Error("Expected product resource not found")
		return
	}
	if product.Name != "product" {
		t.Errorf("product.Name = %v, want product", product.Name)
	}
	if product.PluralName != "products" {
		t.Errorf("product.PluralName = %v, want products", product.PluralName)
	}
}

func setupDirectoryWithNonGoFiles(t *testing.T) string {
	tempDir := t.TempDir()

	validContent := `package dto

type ValidDTO struct {
	ID int64 ` + "`json:\"id\"`" + `
}`
	err := os.WriteFile(filepath.Join(tempDir, "valid.go"), []byte(validContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid.go: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("not a go file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create readme.txt: %v", err)
	}

	err = os.WriteFile(filepath.Join(tempDir, "config.json"), []byte("{}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config.json: %v", err)
	}

	return tempDir
}

func validateNonGoFilesSkipped(t *testing.T, resources map[string]resourceDTOs) {
	if _, exists := resources["valid"]; !exists {
		t.Error("Expected valid resource not found")
	}
	if _, exists := resources["readme"]; exists {
		t.Error("readme resource should not exist")
	}
	if _, exists := resources["config"]; exists {
		t.Error("config resource should not exist")
	}
}

func setupDirectoryWithNonDTOFiles(t *testing.T) string {
	tempDir := t.TempDir()

	withDTOContent := `package dto

type ItemDTO struct {
	ID int64 ` + "`json:\"id\"`" + `
}`
	err := os.WriteFile(filepath.Join(tempDir, "item.go"), []byte(withDTOContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create item.go: %v", err)
	}

	withoutDTOContent := `package dto

type Helper struct {
	Value string
}`
	err = os.WriteFile(filepath.Join(tempDir, "helper.go"), []byte(withoutDTOContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create helper.go: %v", err)
	}

	return tempDir
}

func validateFilesWithoutDTOsSkipped(t *testing.T, resources map[string]resourceDTOs) {
	if _, exists := resources["item"]; !exists {
		t.Error("Expected item resource not found")
	}
	if _, exists := resources["helper"]; exists {
		t.Error("helper resource should not exist (no DTOs)")
	}
}

func setupDirectoryWithMultipleDTOs(t *testing.T) string {
	tempDir := t.TempDir()

	content := `package dto

type ArticleDTO struct {
	ID    int64  ` + "`json:\"id\"`" + `
	Title string ` + "`json:\"title\"`" + `
}

type CreateArticleDTO struct {
	Title string ` + "`json:\"title\"`" + `
}

type UpdateArticleDTO struct {
	Title string ` + "`json:\"title\"`" + `
}`
	err := os.WriteFile(filepath.Join(tempDir, "article.go"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create article.go: %v", err)
	}

	return tempDir
}

func validateMultipleDTOsInFile(t *testing.T, resources map[string]resourceDTOs) {
	article, exists := resources["article"]
	if !exists {
		t.Fatal("Expected article resource not found")
	}
	if len(article.DTOs) != 3 {
		t.Errorf("Expected 3 DTOs in article resource, got %d", len(article.DTOs))
	}
}

func setupCategoryDTO(t *testing.T) string {
	tempDir := t.TempDir()

	categoryContent := `package dto

type CategoryDTO struct {
	ID int64 ` + "`json:\"id\"`" + `
}`
	err := os.WriteFile(filepath.Join(tempDir, "category.go"), []byte(categoryContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create category.go: %v", err)
	}

	return tempDir
}

func validatePluralization(t *testing.T, resources map[string]resourceDTOs) {
	category, exists := resources["category"]
	if !exists {
		t.Fatal("Expected category resource not found")
	}
	if category.PluralName != "categories" {
		t.Errorf("category.PluralName = %v, want categories", category.PluralName)
	}
}
