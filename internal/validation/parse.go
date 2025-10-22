package validation

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type parsedFile struct {
	TypeName string
	Multi    bool
	Items    []map[string]any
	Single   map[string]any
}

func parseDataFile(path string, expectedType string) (*parsedFile, error) {
	content, err := osReadFile(path)
	if err != nil {
		return nil, err
	}

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

// osReadFile is a test seam for injecting failures.
var osReadFile = func(path string) ([]byte, error) {
	return os.ReadFile(path)
}
