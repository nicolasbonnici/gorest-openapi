package openapi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadResourceDTOs(dtosDir string) (map[string]resourceDTOs, error) {
	if _, err := os.Stat(dtosDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("DTOs directory not found: %s", dtosDir)
	}

	files, err := os.ReadDir(dtosDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dtos directory: %w", err)
	}

	resources := make(map[string]resourceDTOs)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".go") {
			continue
		}

		filePath := filepath.Join(dtosDir, file.Name())
		resourceName := strings.TrimSuffix(file.Name(), ".go")

		dtos, err := extractDTOsFromFile(filePath)
		if err != nil {
			continue
		}

		if len(dtos) > 0 {
			resources[resourceName] = resourceDTOs{
				Name:       resourceName,
				PluralName: pluralize(resourceName),
				DTOs:       dtos,
			}
		}
	}

	return resources, nil
}
