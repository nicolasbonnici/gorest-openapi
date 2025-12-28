package openapi

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractTag(t *testing.T) {
	tests := []struct {
		name      string
		tagString string
		key       string
		want      string
	}{
		{
			name:      "extract json tag",
			tagString: "`json:\"user_id\" db:\"user_id\"`",
			key:       "json",
			want:      "user_id",
		},
		{
			name:      "extract db tag",
			tagString: "`json:\"user_id\" db:\"user_id\"`",
			key:       "db",
			want:      "user_id",
		},
		{
			name:      "extract dto tag",
			tagString: "`json:\"name\" dto:\"create,update\"`",
			key:       "dto",
			want:      "create,update",
		},
		{
			name:      "tag not found returns empty string",
			tagString: "`json:\"user_id\" db:\"user_id\"`",
			key:       "xml",
			want:      "",
		},
		{
			name:      "empty tag string",
			tagString: "``",
			key:       "json",
			want:      "",
		},
		{
			name:      "tag with options",
			tagString: "`json:\"email,omitempty\"`",
			key:       "json",
			want:      "email,omitempty",
		},
		{
			name:      "multiple tags",
			tagString: "`json:\"id\" db:\"id\" xml:\"id\" yaml:\"id\"`",
			key:       "xml",
			want:      "id",
		},
		{
			name:      "tag with dash (omit field)",
			tagString: "`json:\"-\"`",
			key:       "json",
			want:      "-",
		},
		{
			name:      "tag without backticks",
			tagString: "json:\"name\"",
			key:       "json",
			want:      "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractTag(tt.tagString, tt.key); got != tt.want {
				t.Errorf("extractTag(%q, %q) = %v, want %v", tt.tagString, tt.key, got, tt.want)
			}
		})
	}
}

func TestExtractDTOsFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		fileContent string
		fileName    string
		wantDTOs    map[string]dtoSchema
		wantErr     bool
	}{
		{
			name:     "extract single DTO",
			fileName: "user.go",
			fileContent: `package dto

type UserDTO struct {
	ID    int64  ` + "`json:\"id\"`" + `
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}`,
			wantDTOs: map[string]dtoSchema{
				"UserDTO": {
					Name: "UserDTO",
					Fields: []structField{
						{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
						{Name: "Name", Type: "string", JSONTag: "name", IsPointer: false},
						{Name: "Email", Type: "string", JSONTag: "email", IsPointer: false},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "extract multiple DTOs",
			fileName: "product.go",
			fileContent: `package dto

type ProductDTO struct {
	ID    int64   ` + "`json:\"id\"`" + `
	Name  string  ` + "`json:\"name\"`" + `
	Price float64 ` + "`json:\"price\"`" + `
}

type CreateProductDTO struct {
	Name  string  ` + "`json:\"name\"`" + `
	Price float64 ` + "`json:\"price\"`" + `
}`,
			wantDTOs: map[string]dtoSchema{
				"ProductDTO": {
					Name: "ProductDTO",
					Fields: []structField{
						{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
						{Name: "Name", Type: "string", JSONTag: "name", IsPointer: false},
						{Name: "Price", Type: "float64", JSONTag: "price", IsPointer: false},
					},
				},
				"CreateProductDTO": {
					Name: "CreateProductDTO",
					Fields: []structField{
						{Name: "Name", Type: "string", JSONTag: "name", IsPointer: false},
						{Name: "Price", Type: "float64", JSONTag: "price", IsPointer: false},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "DTO with pointer fields",
			fileName: "article.go",
			fileContent: `package dto

type ArticleDTO struct {
	ID          int64   ` + "`json:\"id\"`" + `
	Title       string  ` + "`json:\"title\"`" + `
	Description *string ` + "`json:\"description\"`" + `
}`,
			wantDTOs: map[string]dtoSchema{
				"ArticleDTO": {
					Name: "ArticleDTO",
					Fields: []structField{
						{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
						{Name: "Title", Type: "string", JSONTag: "title", IsPointer: false},
						{Name: "Description", Type: "string", JSONTag: "description", IsPointer: true},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "DTO with time.Time fields",
			fileName: "event.go",
			fileContent: `package dto

import "time"

type EventDTO struct {
	ID        int64     ` + "`json:\"id\"`" + `
	Name      string    ` + "`json:\"name\"`" + `
	StartDate time.Time ` + "`json:\"start_date\"`" + `
}`,
			wantDTOs: map[string]dtoSchema{
				"EventDTO": {
					Name: "EventDTO",
					Fields: []structField{
						{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
						{Name: "Name", Type: "string", JSONTag: "name", IsPointer: false},
						{Name: "StartDate", Type: "time.Time", JSONTag: "start_date", IsPointer: false},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "ignore non-DTO structs",
			fileName: "mixed.go",
			fileContent: `package dto

type User struct {
	ID   int64
	Name string
}

type UserDTO struct {
	ID   int64  ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}`,
			wantDTOs: map[string]dtoSchema{
				"UserDTO": {
					Name: "UserDTO",
					Fields: []structField{
						{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
						{Name: "Name", Type: "string", JSONTag: "name", IsPointer: false},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "DTO with interface{} field",
			fileName: "config.go",
			fileContent: `package dto

type ConfigDTO struct {
	Key   string      ` + "`json:\"key\"`" + `
	Value interface{} ` + "`json:\"value\"`" + `
}`,
			wantDTOs: map[string]dtoSchema{
				"ConfigDTO": {
					Name: "ConfigDTO",
					Fields: []structField{
						{Name: "Key", Type: "string", JSONTag: "key", IsPointer: false},
						{Name: "Value", Type: "interface{}", JSONTag: "value", IsPointer: false},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "invalid Go file returns error",
			fileName: "invalid.go",
			fileContent: `package dto

type InvalidDTO struct {
	// Missing closing brace
`,
			wantDTOs: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filePath := filepath.Join(tempDir, tt.fileName)
			err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			got, err := extractDTOsFromFile(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractDTOsFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.wantDTOs) {
					t.Errorf("extractDTOsFromFile() returned %d DTOs, want %d", len(got), len(tt.wantDTOs))
				}

				for name, wantDTO := range tt.wantDTOs {
					gotDTO, exists := got[name]
					if !exists {
						t.Errorf("extractDTOsFromFile() missing DTO %q", name)
						continue
					}

					if gotDTO.Name != wantDTO.Name {
						t.Errorf("extractDTOsFromFile() DTO name = %v, want %v", gotDTO.Name, wantDTO.Name)
					}

					if len(gotDTO.Fields) != len(wantDTO.Fields) {
						t.Errorf("extractDTOsFromFile() DTO %q has %d fields, want %d", name, len(gotDTO.Fields), len(wantDTO.Fields))
						continue
					}

					for i, wantField := range wantDTO.Fields {
						gotField := gotDTO.Fields[i]
						if !reflect.DeepEqual(gotField, wantField) {
							t.Errorf("extractDTOsFromFile() DTO %q field[%d] = %+v, want %+v", name, i, gotField, wantField)
						}
					}
				}
			}
		})
	}
}

func TestExtractStructFieldsFromAST(t *testing.T) {
	// This is tested indirectly through TestExtractDTOsFromFile
	// We can add specific edge case tests here if needed
	t.Run("embedded fields are skipped", func(t *testing.T) {
		tempDir := t.TempDir()
		fileContent := `package dto

type Base struct {
	ID int64
}

type EmbeddedDTO struct {
	Base
	Name string ` + "`json:\"name\"`" + `
}`
		filePath := filepath.Join(tempDir, "embedded.go")
		err := os.WriteFile(filePath, []byte(fileContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		got, err := extractDTOsFromFile(filePath)
		if err != nil {
			t.Fatalf("extractDTOsFromFile() error = %v", err)
		}

		dto, exists := got["EmbeddedDTO"]
		if !exists {
			t.Fatal("EmbeddedDTO not found")
		}

		// Should only have the Name field, Base is embedded (no field name)
		if len(dto.Fields) != 1 {
			t.Errorf("Expected 1 field (embedded fields skipped), got %d", len(dto.Fields))
		}

		if len(dto.Fields) > 0 && dto.Fields[0].Name != "Name" {
			t.Errorf("Expected field name 'Name', got %q", dto.Fields[0].Name)
		}
	})
}
