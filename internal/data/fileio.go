package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/theory/jsonpath"
	"gopkg.in/yaml.v3"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func (s *Store) loadFile(path string, expectedType string, selector string) (*fileContent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	format := detectFormat(path)

	if selector == "" {
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
			Format:   format,
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

	var root any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("data: parse %s: %w", path, err)
	}

	normalizedRoot, err := normalizeYAMLValue(root)
	if err != nil {
		return nil, fmt.Errorf("data: normalize %s: %w", path, err)
	}

	compiled, err := jsonpath.Parse(selector)
	if err != nil {
		return nil, fmt.Errorf("data: parse selector %q in %s: %w", selector, path, err)
	}

	located := compiled.SelectLocated(normalizedRoot)
	if len(located) == 0 {
		return nil, fmt.Errorf("data: selector %q in %s matched no values", selector, path)
	}

	items := make([]map[string]any, 0, len(located))
	for _, node := range located {
		obj, err := normalizeObject(node.Node)
		if err != nil {
			return nil, fmt.Errorf("data: selector %q in %s at %s: %w", selector, path, node.Path.String(), err)
		}

		if expectedType != "" {
			if typeName, ok := getString(obj, "type"); ok && typeName != "" {
				if typeName != expectedType {
					return nil, fmt.Errorf("data: selector %q in %s at %s declares type %s; expected %s", selector, path, node.Path.String(), typeName, expectedType)
				}
			}
			if typeName, ok := getString(obj, "Type"); ok && typeName != "" {
				if typeName != expectedType {
					return nil, fmt.Errorf("data: selector %q in %s at %s declares type %s; expected %s", selector, path, node.Path.String(), typeName, expectedType)
				}
			}
		}

		removeTypeKeys(obj)
		items = append(items, obj)
	}

	fc := &fileContent{
		Path:     path,
		TypeName: expectedType,
		Format:   format,
		Multi:    len(items) > 1,
		Selector: selector,
		ReadOnly: true,
	}

	if len(items) == 1 {
		fc.Single = items[0]
		fc.Multi = false
	} else {
		fc.Items = items
	}

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
	var foundWritable bool

	for _, include := range typeDef.Include {
		if include.Selector != "" {
			continue
		}
		foundWritable = true
		pattern := include.Path
		if pattern == "" {
			continue
		}
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

	if !foundWritable {
		return nil, fmt.Errorf("data: unable to resolve create target for type %s; includes use selector expressions", typeDef.Name)
	}

	return nil, fmt.Errorf("data: unable to resolve create target for type %s", typeDef.Name)
}
