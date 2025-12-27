package openapi

import "strings"

func buildSchemaPropertiesFromDTO(fields []structField) map[string]interface{} {
	properties := make(map[string]interface{})

	for _, field := range fields {
		typ, format := goTypeToOpenAPIType(field.Type)
		prop := map[string]interface{}{
			"type": typ,
		}

		if format != "" {
			prop["format"] = format
		}

		prop["nullable"] = field.IsPointer

		jsonName := field.JSONTag
		if jsonName == "" {
			jsonName = strings.ToLower(field.Name)
		}

		properties[jsonName] = prop
	}

	return properties
}

func getRequiredFieldsFromDTO(fields []structField) []string {
	var required []string

	for _, field := range fields {
		jsonName := field.JSONTag
		if jsonName == "" {
			jsonName = strings.ToLower(field.Name)
		}

		if jsonName == "id" || jsonName == "created_at" || jsonName == "updated_at" {
			continue
		}

		if !field.IsPointer {
			required = append(required, jsonName)
		}
	}

	return required
}
