package validation

import (
	"fmt"
	"os"

	"github.com/theory/jsonpath"
	"gopkg.in/yaml.v3"
)

type parsedFile struct {
	TypeName string
	Multi    bool
	Items    []map[string]any
	Single   map[string]any
}

func parseDataFile(path string, expectedType string, selector string) (*parsedFile, error) {
	content, err := osReadFile(path)
	if err != nil {
		return nil, err
	}

	if selector == "" {
		var doc map[string]any
		if err := yaml.Unmarshal(content, &doc); err != nil {
			return nil, fmt.Errorf("unable to parse: %w", err)
		}

		typeName, hasType := getString(doc, "type")
		if hasType {
			delete(doc, "type")
		}

		if typeName == "" {
			typeName = expectedType
		}

		if expectedType != "" && typeName != "" && typeName != expectedType {
			return nil, fmt.Errorf("file declares type %q", typeName)
		}

		itemsRaw, itemsPresent := doc["items"]
		if itemsPresent {
			slice, err := toSliceMap(itemsRaw)
			if err != nil {
				return nil, err
			}
			return &parsedFile{TypeName: typeName, Multi: true, Items: slice}, nil
		}

		return &parsedFile{TypeName: typeName, Single: doc}, nil
	}

	var root any
	if err := yaml.Unmarshal(content, &root); err != nil {
		return nil, fmt.Errorf("unable to parse: %w", err)
	}

	normalizedRoot, err := normalizeYAMLValue(root)
	if err != nil {
		return nil, fmt.Errorf("unable to normalize: %w", err)
	}

	compiled, err := jsonpath.Parse(selector)
	if err != nil {
		return nil, fmt.Errorf("invalid selector %q: %w", selector, err)
	}

	located := compiled.SelectLocated(normalizedRoot)
	if len(located) == 0 {
		return nil, fmt.Errorf("selector %q matched no values", selector)
	}

	items := make([]map[string]any, 0, len(located))
	for _, node := range located {
		obj, err := normalizeObject(node.Node)
		if err != nil {
			return nil, fmt.Errorf("selector %q at %s: %w", selector, node.Path.String(), err)
		}

		if expectedType != "" {
			if typeName, ok := getString(obj, "type"); ok && typeName != "" {
				if typeName != expectedType {
					return nil, fmt.Errorf("selector %q at %s declares type %q", selector, node.Path.String(), typeName)
				}
			}
			if typeName, ok := getString(obj, "Type"); ok && typeName != "" {
				if typeName != expectedType {
					return nil, fmt.Errorf("selector %q at %s declares type %q", selector, node.Path.String(), typeName)
				}
			}
		}

		removeTypeKeys(obj)
		items = append(items, obj)
	}

	if len(items) == 1 {
		return &parsedFile{TypeName: expectedType, Single: items[0]}, nil
	}

	return &parsedFile{TypeName: expectedType, Multi: true, Items: items}, nil
}

// osReadFile is a test seam for injecting failures.
var osReadFile = func(path string) ([]byte, error) {
	return os.ReadFile(path)
}
