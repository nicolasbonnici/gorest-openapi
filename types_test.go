package openapi

import (
	"testing"
)

func TestResourceDTOs_getMainDTO(t *testing.T) {
	tests := []struct {
		name     string
		resource resourceDTOs
		want     *dtoSchema
	}{
		{
			name: "returns main DTO when present",
			resource: resourceDTOs{
				Name:       "user",
				PluralName: "users",
				DTOs: map[string]dtoSchema{
					"UserDTO": {
						Name:   "UserDTO",
						Fields: []structField{{Name: "ID", Type: "int64"}},
					},
					"CreateUserDTO": {
						Name:   "CreateUserDTO",
						Fields: []structField{{Name: "Name", Type: "string"}},
					},
					"UpdateUserDTO": {
						Name:   "UpdateUserDTO",
						Fields: []structField{{Name: "Name", Type: "string"}},
					},
				},
			},
			want: &dtoSchema{
				Name:   "UserDTO",
				Fields: []structField{{Name: "ID", Type: "int64"}},
			},
		},
		{
			name: "returns nil when only Create/Update DTOs exist",
			resource: resourceDTOs{
				Name:       "user",
				PluralName: "users",
				DTOs: map[string]dtoSchema{
					"CreateUserDTO": {
						Name:   "CreateUserDTO",
						Fields: []structField{{Name: "Name", Type: "string"}},
					},
					"UpdateUserDTO": {
						Name:   "UpdateUserDTO",
						Fields: []structField{{Name: "Name", Type: "string"}},
					},
				},
			},
			want: nil,
		},
		{
			name: "returns nil when DTOs map is empty",
			resource: resourceDTOs{
				Name:       "user",
				PluralName: "users",
				DTOs:       map[string]dtoSchema{},
			},
			want: nil,
		},
		{
			name: "ignores CreateDTO and UpdateDTO",
			resource: resourceDTOs{
				Name:       "product",
				PluralName: "products",
				DTOs: map[string]dtoSchema{
					"ProductDTO": {
						Name:   "ProductDTO",
						Fields: []structField{{Name: "ID", Type: "int64"}},
					},
					"CreateProductDTO": {
						Name: "CreateProductDTO",
					},
				},
			},
			want: &dtoSchema{
				Name:   "ProductDTO",
				Fields: []structField{{Name: "ID", Type: "int64"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource.getMainDTO()
			if (got == nil) != (tt.want == nil) {
				t.Errorf("getMainDTO() = %v, want %v", got, tt.want)
				return
			}
			if got != nil && tt.want != nil {
				if got.Name != tt.want.Name {
					t.Errorf("getMainDTO().Name = %v, want %v", got.Name, tt.want.Name)
				}
				if len(got.Fields) != len(tt.want.Fields) {
					t.Errorf("getMainDTO().Fields length = %v, want %v", len(got.Fields), len(tt.want.Fields))
				}
			}
		})
	}
}

func TestContainsSubstr(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{
			name:   "substring at beginning",
			s:      "CreateUserDTO",
			substr: "Create",
			want:   true,
		},
		{
			name:   "substring in middle",
			s:      "UserDTO",
			substr: "ser",
			want:   true,
		},
		{
			name:   "substring at end",
			s:      "UpdateUserDTO",
			substr: "DTO",
			want:   true,
		},
		{
			name:   "substring not found",
			s:      "UserDTO",
			substr: "Create",
			want:   false,
		},
		{
			name:   "empty substring",
			s:      "UserDTO",
			substr: "",
			want:   true,
		},
		{
			name:   "empty string",
			s:      "",
			substr: "test",
			want:   false,
		},
		{
			name:   "both empty",
			s:      "",
			substr: "",
			want:   true,
		},
		{
			name:   "exact match",
			s:      "test",
			substr: "test",
			want:   true,
		},
		{
			name:   "substring longer than string",
			s:      "abc",
			substr: "abcdef",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsSubstr(tt.s, tt.substr); got != tt.want {
				t.Errorf("containsSubstr(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}
