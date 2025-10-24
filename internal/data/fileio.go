package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func (s *Store) loadFile(path string, expectedType string) (*fileContent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("data: parse %s: %w", path, err)
	}

	typeName, hasType := getString(doc, "type")
	if hasType {
		delete(doc, "type")
	}

	if typeName == "" {
		typeName = expectedType
	}
	if expectedType != "" && typeName != "" && typeName != expectedType {
		return nil, fmt.Errorf("data: file %s declares type %s; expected %s", path, typeName, expectedType)
	}

	itemsRaw, hasItems := doc["items"]

	fc := &fileContent{
		Path:     path,
		TypeName: typeName,
		Format:   detectFormat(path),
	}

	if hasItems {
		slice, err := toSliceMap(itemsRaw)
		if err != nil {
			return nil, fmt.Errorf("data: file %s items: %w", path, err)
		}
		fc.Multi = true
		fc.Items = slice
		return fc, nil
	}

	fc.Single = doc
	return fc, nil
}

func (s *Store) writeFile(path string, fc *fileContent) error {
	if fc == nil {
		return fmt.Errorf("data: writeFile nil content")
	}

	payload := make(map[string]any)
	if fc.TypeName != "" {
		payload["type"] = fc.TypeName
	}

	if fc.Multi {
		payload["items"] = fc.Items
	} else if fc.Single != nil {
		for k, v := range fc.Single {
			payload[k] = v
		}
	}

	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}

	if fc.Format == formatJSON {
		encoded, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Errorf("data: encode json %s: %w", path, err)
		}
		return os.WriteFile(path, append(encoded, '\n'), 0o644)
	}

	encoded, err := yaml.Marshal(payload)
	if err != nil {
		return fmt.Errorf("data: encode yaml %s: %w", path, err)
	}
	return os.WriteFile(path, encoded, 0o644)
}

func (s *Store) writeSingle(path string, format fileFormat, payload map[string]any) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}

	if format == formatJSON {
		encoded, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Errorf("data: encode json %s: %w", path, err)
		}
		return os.WriteFile(path, append(encoded, '\n'), 0o644)
	}

	encoded, err := yaml.Marshal(payload)
	if err != nil {
		return fmt.Errorf("data: encode yaml %s: %w", path, err)
	}
	return os.WriteFile(path, encoded, 0o644)
}

func (s *Store) removeFile(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("data: remove %s: %w", path, err)
	}
	return nil
}

func (s *Store) chooseCreateTarget(typeDef *config.TypeDefinition, id string) (*createTarget, error) {
	var multiTarget *createTarget

	for _, pattern := range typeDef.Include {
		absPattern := pattern
		if !filepath.IsAbs(absPattern) {
			absPattern = filepath.Join(s.root, filepath.Clean(pattern))
		}

		if strings.ContainsAny(absPattern, "*?[") {
			candidate := replaceGlob(absPattern, sanitizeFilename(id))
			return &createTarget{Path: candidate, Format: detectFormat(candidate), Multi: false}, nil
		}

		if multiTarget == nil {
			multiTarget = &createTarget{Path: absPattern, Format: detectFormat(absPattern), Multi: true}
		}
	}

	if multiTarget != nil {
		return multiTarget, nil
	}

	return nil, fmt.Errorf("data: unable to resolve create target for type %s", typeDef.Name)
}
