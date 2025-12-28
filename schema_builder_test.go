package openapi

import (
	"reflect"
	"testing"
)

func TestBuildSchemaPropertiesFromDTO(t *testing.T) {
	tests := []struct {
		name   string
		fields []structField
		want   map[string]interface{}
	}{
		{
			name: "basic types with JSON tags",
			fields: []structField{
				{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
				{Name: "Name", Type: "string", JSONTag: "name", IsPointer: false},
				{Name: "Email", Type: "string", JSONTag: "email", IsPointer: false},
			},
			want: map[string]interface{}{
				"id": map[string]interface{}{
					"type":     "integer",
					"format":   "int64",
					"nullable": false,
				},
				"name": map[string]interface{}{
					"type":     "string",
					"nullable": false,
				},
				"email": map[string]interface{}{
					"type":     "string",
					"nullable": false,
				},
			},
		},
		{
			name: "pointer types are nullable",
			fields: []structField{
				{Name: "Age", Type: "int", JSONTag: "age", IsPointer: true},
				{Name: "Bio", Type: "string", JSONTag: "bio", IsPointer: true},
			},
			want: map[string]interface{}{
				"age": map[string]interface{}{
					"type":     "integer",
					"format":   "int32",
					"nullable": true,
				},
				"bio": map[string]interface{}{
					"type":     "string",
					"nullable": true,
				},
			},
		},
		{
			name: "fields without JSON tags use lowercase field name",
			fields: []structField{
				{Name: "Username", Type: "string", JSONTag: "", IsPointer: false},
				{Name: "Active", Type: "bool", JSONTag: "", IsPointer: false},
			},
			want: map[string]interface{}{
				"username": map[string]interface{}{
					"type":     "string",
					"nullable": false,
				},
				"active": map[string]interface{}{
					"type":     "boolean",
					"nullable": false,
				},
			},
		},
		{
			name: "time.Time fields",
			fields: []structField{
				{Name: "CreatedAt", Type: "time.Time", JSONTag: "created_at", IsPointer: false},
				{Name: "UpdatedAt", Type: "time.Time", JSONTag: "updated_at", IsPointer: true},
			},
			want: map[string]interface{}{
				"created_at": map[string]interface{}{
					"type":     "string",
					"format":   "date-time",
					"nullable": false,
				},
				"updated_at": map[string]interface{}{
					"type":     "string",
					"format":   "date-time",
					"nullable": true,
				},
			},
		},
		{
			name: "mixed types",
			fields: []structField{
				{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
				{Name: "Price", Type: "float64", JSONTag: "price", IsPointer: false},
				{Name: "Available", Type: "bool", JSONTag: "available", IsPointer: false},
				{Name: "Metadata", Type: "interface{}", JSONTag: "metadata", IsPointer: false},
			},
			want: map[string]interface{}{
				"id": map[string]interface{}{
					"type":     "integer",
					"format":   "int64",
					"nullable": false,
				},
				"price": map[string]interface{}{
					"type":     "number",
					"format":   "double",
					"nullable": false,
				},
				"available": map[string]interface{}{
					"type":     "boolean",
					"nullable": false,
				},
				"metadata": map[string]interface{}{
					"type":     "object",
					"nullable": false,
				},
			},
		},
		{
			name:   "empty fields",
			fields: []structField{},
			want:   map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSchemaPropertiesFromDTO(tt.fields)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildSchemaPropertiesFromDTO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRequiredFieldsFromDTO(t *testing.T) {
	tests := []struct {
		name   string
		fields []structField
		want   []string
	}{
		{
			name: "non-pointer fields are required (except system fields)",
			fields: []structField{
				{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
				{Name: "Name", Type: "string", JSONTag: "name", IsPointer: false},
				{Name: "Email", Type: "string", JSONTag: "email", IsPointer: false},
			},
			want: []string{"name", "email"},
		},
		{
			name: "pointer fields are not required",
			fields: []structField{
				{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
				{Name: "Name", Type: "string", JSONTag: "name", IsPointer: false},
				{Name: "Bio", Type: "string", JSONTag: "bio", IsPointer: true},
				{Name: "Age", Type: "int", JSONTag: "age", IsPointer: true},
			},
			want: []string{"name"},
		},
		{
			name: "system fields excluded (id, created_at, updated_at)",
			fields: []structField{
				{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
				{Name: "CreatedAt", Type: "time.Time", JSONTag: "created_at", IsPointer: false},
				{Name: "UpdatedAt", Type: "time.Time", JSONTag: "updated_at", IsPointer: false},
				{Name: "Name", Type: "string", JSONTag: "name", IsPointer: false},
			},
			want: []string{"name"},
		},
		{
			name: "fields without JSON tags use lowercase name",
			fields: []structField{
				{Name: "Username", Type: "string", JSONTag: "", IsPointer: false},
				{Name: "Password", Type: "string", JSONTag: "", IsPointer: false},
			},
			want: []string{"username", "password"},
		},
		{
			name: "all pointer fields return empty slice",
			fields: []structField{
				{Name: "Bio", Type: "string", JSONTag: "bio", IsPointer: true},
				{Name: "Age", Type: "int", JSONTag: "age", IsPointer: true},
			},
			want: nil,
		},
		{
			name:   "empty fields return empty slice",
			fields: []structField{},
			want:   nil,
		},
		{
			name: "mixed required and optional fields",
			fields: []structField{
				{Name: "ID", Type: "int64", JSONTag: "id", IsPointer: false},
				{Name: "Title", Type: "string", JSONTag: "title", IsPointer: false},
				{Name: "Description", Type: "string", JSONTag: "description", IsPointer: true},
				{Name: "Price", Type: "float64", JSONTag: "price", IsPointer: false},
				{Name: "Discount", Type: "float64", JSONTag: "discount", IsPointer: true},
				{Name: "CreatedAt", Type: "time.Time", JSONTag: "created_at", IsPointer: false},
			},
			want: []string{"title", "price"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getRequiredFieldsFromDTO(tt.fields)
			// Handle nil vs empty slice comparison
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRequiredFieldsFromDTO() = %v, want %v", got, tt.want)
			}
		})
	}
}
