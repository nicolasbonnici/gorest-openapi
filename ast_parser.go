package openapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func extractDTOsFromFile(path string) (map[string]dtoSchema, error) {
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, path, nil, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	dtos := make(map[string]dtoSchema)

	for _, decl := range node.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if !strings.HasSuffix(ts.Name.Name, "DTO") {
				continue
			}

			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}

			fields := extractStructFieldsFromAST(st)
			dtos[ts.Name.Name] = dtoSchema{
				Name:   ts.Name.Name,
				Fields: fields,
			}
		}
	}

	return dtos, nil
}

func extractStructFieldsFromAST(st *ast.StructType) []structField {
	var fields []structField

	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		fieldName := field.Names[0].Name
		fieldType := ""
		isPointer := false

		switch t := field.Type.(type) {
		case *ast.Ident:
			fieldType = t.Name
		case *ast.StarExpr:
			isPointer = true
			if ident, ok := t.X.(*ast.Ident); ok {
				fieldType = ident.Name
			} else if sel, ok := t.X.(*ast.SelectorExpr); ok {
				if pkg, ok := sel.X.(*ast.Ident); ok {
					fieldType = pkg.Name + "." + sel.Sel.Name
				}
			}
		case *ast.SelectorExpr:
			if pkg, ok := t.X.(*ast.Ident); ok {
				fieldType = pkg.Name + "." + t.Sel.Name
			}
		case *ast.InterfaceType:
			fieldType = "interface{}"
		case *ast.MapType:
			fieldType = "map[string]interface{}"
		}

		jsonTag := ""
		dbTag := ""
		dtoTag := ""
		if field.Tag != nil {
			tag := field.Tag.Value
			jsonTag = extractTag(tag, "json")
			jsonTag = strings.Split(jsonTag, ",")[0]
			dbTag = extractTag(tag, "db")
			dtoTag = extractTag(tag, "dto")
		}

		fields = append(fields, structField{
			Name:      fieldName,
			Type:      fieldType,
			JSONTag:   jsonTag,
			DBTag:     dbTag,
			DTOTag:    dtoTag,
			IsPointer: isPointer,
		})
	}

	return fields
}

func extractTag(tagString, key string) string {
	tagString = strings.Trim(tagString, "`")
	for _, tag := range strings.Fields(tagString) {
		if strings.HasPrefix(tag, key+":") {
			value := strings.TrimPrefix(tag, key+":")
			value = strings.Trim(value, `"`)
			return value
		}
	}
	return ""
}
