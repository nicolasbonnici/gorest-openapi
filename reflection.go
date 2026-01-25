package openapi

import (
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

func buildSchemaFromModel(model interface{}) map[string]interface{} {
	if model == nil {
		return map[string]interface{}{"type": "object"}
	}

	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return map[string]interface{}{"type": "object"}
	}

	properties := make(map[string]interface{})
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}

		isOmitEmpty := strings.Contains(jsonTag, "omitempty")
		fieldType := field.Type
		isPointer := fieldType.Kind() == reflect.Ptr

		if isPointer {
			fieldType = fieldType.Elem()
		}

		property := buildPropertySchema(fieldType, field.Tag)
		property["nullable"] = isPointer

		validateTag := field.Tag.Get("validate")
		applyValidationRules(property, validateTag)

		properties[jsonName] = property

		if !isPointer && !isOmitEmpty && jsonName != "id" && jsonName != "createdAt" && jsonName != "updatedAt" {
			if !strings.Contains(validateTag, "omitempty") {
				required = append(required, jsonName)
			}
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func buildPropertySchema(t reflect.Type, tag reflect.StructTag) map[string]interface{} {
	property := make(map[string]interface{})

	switch t.Kind() {
	case reflect.String:
		property["type"] = "string"
		if t == reflect.TypeOf(uuid.UUID{}) {
			property["format"] = "uuid"
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		property["type"] = "integer"
		property["format"] = "int32"
	case reflect.Int64:
		property["type"] = "integer"
		property["format"] = "int64"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		property["type"] = "integer"
	case reflect.Float32:
		property["type"] = "number"
		property["format"] = "float"
	case reflect.Float64:
		property["type"] = "number"
		property["format"] = "double"
	case reflect.Bool:
		property["type"] = "boolean"
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			property["type"] = "string"
			property["format"] = "date-time"
		} else if t == reflect.TypeOf(uuid.UUID{}) {
			property["type"] = "string"
			property["format"] = "uuid"
		} else {
			property["type"] = "object"
		}
	case reflect.Slice, reflect.Array:
		property["type"] = "array"
		elemType := t.Elem()
		property["items"] = buildPropertySchema(elemType, "")
	case reflect.Map:
		property["type"] = "object"
	default:
		property["type"] = "string"
	}

	return property
}

func applyValidationRules(property map[string]interface{}, validateTag string) {
	if validateTag == "" {
		return
	}

	rules := strings.Split(validateTag, ",")
	for _, rule := range rules {
		parts := strings.Split(rule, "=")
		ruleName := strings.TrimSpace(parts[0])
		var ruleValue string
		if len(parts) > 1 {
			ruleValue = strings.TrimSpace(parts[1])
		}

		switch ruleName {
		case "required":
		case "email":
			property["format"] = "email"
		case "uuid":
			property["format"] = "uuid"
		case "min":
			switch property["type"] {
			case "string":
				if minLength := parseIntOrZero(ruleValue); minLength > 0 {
					property["minLength"] = minLength
				}
			case "integer", "number":
				if min := parseIntOrZero(ruleValue); min > 0 {
					property["minimum"] = min
				}
			}
		case "max":
			switch property["type"] {
			case "string":
				if maxLength := parseIntOrZero(ruleValue); maxLength > 0 {
					property["maxLength"] = maxLength
				}
			case "integer", "number":
				if max := parseIntOrZero(ruleValue); max > 0 {
					property["maximum"] = max
				}
			}
		case "url":
			property["format"] = "uri"
		case "oneof":
			if property["type"] == "string" {
				values := strings.Split(ruleValue, " ")
				property["enum"] = values
			}
		}
	}
}

func parseIntOrZero(s string) int {
	var val int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		val = val*10 + int(c-'0')
	}
	return val
}
