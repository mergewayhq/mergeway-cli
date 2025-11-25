package format

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Result captures the outcome of formatting a file.
type Result struct {
	Path    string
	Content []byte
	Changed bool
}

// FormatFile loads a file, applies formatting rules, and returns the new body.
func FormatFile(path string, schema *Schema) (*Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	formatted, err := FormatBytes(path, data, schema)
	if err != nil {
		return nil, err
	}

	changed := !bytes.Equal(data, formatted)
	return &Result{
		Path:    path,
		Content: formatted,
		Changed: changed,
	}, nil
}

// FormatBytes normalizes raw file contents according to the configured rules.
func FormatBytes(path string, data []byte, schema *Schema) ([]byte, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return data, nil
	}

	var root yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&root); err != nil {
		if errors.Is(err, io.EOF) {
			return data, nil
		}
		return nil, fmt.Errorf("format: decode %s: %w", path, err)
	}

	if len(root.Content) == 0 {
		return data, nil
	}

	applyOnDocument(&root)
	applySchemaOrdering(&root, schema)

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		encoded, err := encodeJSON(&root)
		if err != nil {
			return nil, fmt.Errorf("format: encode json %s: %w", path, err)
		}
		return encoded, nil
	default:
		var out bytes.Buffer
		enc := yaml.NewEncoder(&out)
		enc.SetIndent(2)
		if err := enc.Encode(&root); err != nil {
			return nil, fmt.Errorf("format: encode yaml %s: %w", path, err)
		}
		if err := enc.Close(); err != nil {
			return nil, fmt.Errorf("format: finalize yaml %s: %w", path, err)
		}
		return out.Bytes(), nil
	}
}

func applyOnDocument(node *yaml.Node) {
	if node == nil {
		return
	}
	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			applyOnDocument(child)
		}
	case yaml.MappingNode:
		rewriteMapping(node)
	case yaml.SequenceNode:
		for _, child := range node.Content {
			applyOnDocument(child)
		}
	default:
		// No-op on scalars or aliases.
	}
}

func rewriteMapping(node *yaml.Node) {
	if node == nil {
		return
	}

	for idx := 0; idx < len(node.Content); idx += 2 {
		if idx+1 >= len(node.Content) {
			break
		}
		key := node.Content[idx]
		val := node.Content[idx+1]

		if strings.EqualFold(key.Value, "items") && val != nil && val.Kind == yaml.SequenceNode {
			reorderItems(val)
			continue
		}

		applyOnDocument(val)
	}
}

type entityNode struct {
	node      *yaml.Node
	sortKey   string
	hasKey    bool
	line      int
	origIndex int
}

func reorderItems(seq *yaml.Node) {
	if seq == nil || len(seq.Content) < 2 {
		for _, child := range seq.Content {
			applyOnDocument(child)
		}
		return
	}

	entities := make([]entityNode, len(seq.Content))
	keyed := 0

	for i, child := range seq.Content {
		applyOnDocument(child)
		key, hasKey := entitySortKey(child)
		if hasKey {
			keyed++
		}
		entities[i] = entityNode{
			node:      child,
			sortKey:   key,
			hasKey:    hasKey,
			line:      child.Line,
			origIndex: i,
		}
	}

	if keyed == 0 {
		return
	}

	sort.SliceStable(entities, func(i, j int) bool {
		a, b := entities[i], entities[j]
		if a.hasKey && b.hasKey {
			if a.sortKey == b.sortKey {
				return a.line < b.line
			}
			return a.sortKey < b.sortKey
		}
		if a.hasKey != b.hasKey {
			return a.hasKey
		}
		if a.line == b.line {
			return a.origIndex < b.origIndex
		}
		return a.line < b.line
	})

	changed := false
	for i, ent := range entities {
		if seq.Content[i] != ent.node {
			changed = true
			break
		}
	}
	if !changed {
		return
	}

	for i, ent := range entities {
		seq.Content[i] = ent.node
	}
}

func entitySortKey(node *yaml.Node) (string, bool) {
	if node == nil {
		return "", false
	}

	switch node.Kind {
	case yaml.MappingNode:
		return extractMapSortKey(node)
	default:
		return "", false
	}
}

func extractMapSortKey(node *yaml.Node) (string, bool) {
	var fallback string
	for i := 0; i < len(node.Content); i += 2 {
		if i+1 >= len(node.Content) {
			break
		}
		key := node.Content[i]
		val := node.Content[i+1]
		if val == nil {
			continue
		}
		if val.Kind != yaml.ScalarNode || val.Value == "" {
			continue
		}

		field := strings.ToLower(key.Value)
		switch field {
		case "id", "name", "slug", "key":
			return strings.ToLower(val.Value), true
		default:
			if fallback == "" {
				fallback = strings.ToLower(val.Value)
			}
		}
	}
	if fallback != "" {
		return fallback, false
	}
	return "", false
}

func applySchemaOrdering(root *yaml.Node, schema *Schema) {
	if schema == nil || root == nil {
		return
	}

	var mapping *yaml.Node
	switch root.Kind {
	case yaml.DocumentNode:
		if len(root.Content) == 0 {
			return
		}
		mapping = root.Content[0]
	case yaml.MappingNode:
		mapping = root
	default:
		return
	}

	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return
	}

	if items := findMappingValue(mapping, "items"); items != nil && items.Kind == yaml.SequenceNode {
		for _, child := range items.Content {
			if child.Kind == yaml.MappingNode {
				reorderMappingWithSchema(child, schema, false)
			}
		}
		return
	}

	reorderMappingWithSchema(mapping, schema, true)
}

func findMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if i+1 >= len(node.Content) {
			break
		}
		k := node.Content[i]
		if k.Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

type mappingPair struct {
	name  string
	key   *yaml.Node
	value *yaml.Node
	used  bool
}

func reorderMappingWithSchema(node *yaml.Node, schema *Schema, keepType bool) {
	if node == nil || node.Kind != yaml.MappingNode || schema == nil {
		return
	}

	pairs := make([]*mappingPair, 0, len(node.Content)/2)
	for i := 0; i < len(node.Content); i += 2 {
		if i+1 >= len(node.Content) {
			break
		}
		pairs = append(pairs, &mappingPair{
			name:  node.Content[i].Value,
			key:   node.Content[i],
			value: node.Content[i+1],
		})
	}

	var reordered []*yaml.Node

	if keepType {
		if pair := takePair(pairs, "type"); pair != nil {
			reordered = append(reordered, pair.key, pair.value)
		}
	}

	for _, field := range schema.orderedFields() {
		if field == nil {
			continue
		}
		pair := takePair(pairs, field.Name)
		if pair == nil {
			continue
		}
		applyNestedSchema(field, pair.value)
		reordered = append(reordered, pair.key, pair.value)
	}

	for _, pair := range pairs {
		if pair.used {
			continue
		}
		reordered = append(reordered, pair.key, pair.value)
	}

	if len(reordered) > 0 {
		node.Content = reordered
	}
}

func takePair(pairs []*mappingPair, key string) *mappingPair {
	for _, pair := range pairs {
		if pair.used {
			continue
		}
		if pair.name == key {
			pair.used = true
			return pair
		}
	}
	return nil
}

func applyNestedSchema(field *SchemaField, node *yaml.Node) {
	if field == nil || field.Nested == nil || node == nil {
		return
	}

	if field.Repeated {
		if node.Kind == yaml.SequenceNode {
			for _, child := range node.Content {
				if child.Kind == yaml.MappingNode {
					reorderMappingWithSchema(child, field.Nested, false)
				}
			}
			return
		}
		if node.Kind == yaml.MappingNode {
			reorderMappingWithSchema(node, field.Nested, false)
		}
		return
	}

	if node.Kind == yaml.MappingNode {
		reorderMappingWithSchema(node, field.Nested, false)
	}
}

func encodeJSON(root *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeJSONNode(&buf, root, 0); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func writeJSONNode(buf *bytes.Buffer, node *yaml.Node, depth int) error {
	if node == nil {
		buf.WriteString("null")
		return nil
	}
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			buf.WriteString("null")
			return nil
		}
		return writeJSONNode(buf, node.Content[0], depth)
	case yaml.MappingNode:
		return writeJSONObject(buf, node, depth)
	case yaml.SequenceNode:
		return writeJSONArray(buf, node, depth)
	default:
		var value any
		if err := node.Decode(&value); err != nil {
			return err
		}
		bytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		buf.Write(bytes)
		return nil
	}
}

func writeJSONObject(buf *bytes.Buffer, node *yaml.Node, depth int) error {
	buf.WriteByte('{')
	if len(node.Content) == 0 {
		buf.WriteByte('}')
		return nil
	}
	buf.WriteByte('\n')
	for i := 0; i < len(node.Content); i += 2 {
		if i > 0 {
			buf.WriteByte(',')
			buf.WriteByte('\n')
		}
		writeIndent(buf, depth+1)
		keyBytes, err := json.Marshal(node.Content[i].Value)
		if err != nil {
			return err
		}
		buf.Write(keyBytes)
		buf.WriteString(": ")
		if err := writeJSONNode(buf, node.Content[i+1], depth+1); err != nil {
			return err
		}
	}
	buf.WriteByte('\n')
	writeIndent(buf, depth)
	buf.WriteByte('}')
	return nil
}

func writeJSONArray(buf *bytes.Buffer, node *yaml.Node, depth int) error {
	buf.WriteByte('[')
	if len(node.Content) == 0 {
		buf.WriteByte(']')
		return nil
	}
	buf.WriteByte('\n')
	for i, child := range node.Content {
		if i > 0 {
			buf.WriteByte(',')
			buf.WriteByte('\n')
		}
		writeIndent(buf, depth+1)
		if err := writeJSONNode(buf, child, depth+1); err != nil {
			return err
		}
	}
	buf.WriteByte('\n')
	writeIndent(buf, depth)
	buf.WriteByte(']')
	return nil
}

func writeIndent(buf *bytes.Buffer, depth int) {
	for i := 0; i < depth; i++ {
		buf.WriteString("  ")
	}
}
