package openapi

type structField struct {
	Name      string
	Type      string
	JSONTag   string
	DBTag     string
	DTOTag    string
	IsPointer bool
}

type dtoSchema struct {
	Name   string
	Fields []structField
}

type resourceDTOs struct {
	Name       string
	PluralName string
	DTOs       map[string]dtoSchema
}

func (r *resourceDTOs) getMainDTO() *dtoSchema {
	for name, dto := range r.DTOs {
		if !containsSubstr(name, "Create") && !containsSubstr(name, "Update") {
			return &dto
		}
	}
	return nil
}

func containsSubstr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
