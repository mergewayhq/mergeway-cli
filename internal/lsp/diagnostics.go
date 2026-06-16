package lsp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/validation"
	"github.com/mergewayhq/mergeway-cli/internal/workspace"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
)

const diagnosticSource = "mergeway"

type diagnosticPublisher func(context.Context, *protocol.PublishDiagnosticsParams) error

type diagnosticCollector struct {
	documents map[string]*workspace.OpenDocument
	content   map[string][]byte
}

type diagnosticLocation struct {
	path       string
	itemIndex  int
	inlineItem int
}

type fieldSegment struct {
	name     string
	index    int
	hasIndex bool
}

var (
	diagnosticItemPattern        = regexp.MustCompile(`^(.*) \(item (\d+)\)$`)
	diagnosticInlineItemPattern  = regexp.MustCompile(`^(.*) \(inline (\d+)\)$`)
	diagnosticLinePattern        = regexp.MustCompile(`line (\d+)(?::? column (\d+))?`)
	diagnosticQuotedFieldPattern = regexp.MustCompile(`field "([^"]+)"`)
	diagnosticConfigFieldPattern = regexp.MustCompile(`field ([A-Za-z0-9_.]+)`)
	diagnosticQuotedTypePattern  = regexp.MustCompile(`type "([^"]+)"`)
	diagnosticQuotedNamePattern  = regexp.MustCompile(`"([^"]+)"`)
)

func (s *Server) publishWorkspaceDiagnostics(ctx context.Context) error {
	if s.runtime == nil || s.publishDiagnostics == nil {
		return nil
	}

	snapshot := s.runtime.Snapshot()
	paramsByPath := collectSnapshotDiagnostics(snapshot)

	s.publishMu.Lock()
	defer s.publishMu.Unlock()

	for path, params := range paramsByPath {
		if err := s.publishDiagnostics(ctx, params); err != nil {
			return err
		}
		s.publishedPaths[path] = struct{}{}
	}

	for path := range s.publishedPaths {
		if _, ok := paramsByPath[path]; ok {
			continue
		}

		clearParams := &protocol.PublishDiagnosticsParams{
			URI:         protocol.DocumentURI(uri.File(path)),
			Version:     documentVersion(snapshot.Documents[path]),
			Diagnostics: []protocol.Diagnostic{},
		}
		if err := s.publishDiagnostics(ctx, clearParams); err != nil {
			return err
		}
		delete(s.publishedPaths, path)
	}

	return nil
}

func collectSnapshotDiagnostics(snapshot *workspace.Snapshot) map[string]*protocol.PublishDiagnosticsParams {
	if snapshot == nil {
		return nil
	}

	collector := &diagnosticCollector{
		documents: snapshot.Documents,
		content:   make(map[string][]byte),
	}

	grouped := make(map[string][]protocol.Diagnostic)
	for _, root := range snapshot.Roots {
		if root == nil || root.Index == nil {
			continue
		}
		collector.collectRootDiagnostics(grouped, root)
	}

	result := make(map[string]*protocol.PublishDiagnosticsParams, len(grouped))
	for path, diagnostics := range grouped {
		result[path] = &protocol.PublishDiagnosticsParams{
			URI:         protocol.DocumentURI(uri.File(path)),
			Version:     documentVersion(snapshot.Documents[path]),
			Diagnostics: diagnostics,
		}
	}

	return result
}

func (c *diagnosticCollector) collectRootDiagnostics(grouped map[string][]protocol.Diagnostic, root *workspace.RootRuntime) {
	if root.Validation != nil && root.Validation.Result != nil {
		for _, errItem := range root.Validation.Result.Errors {
			path, diagnostic, ok := c.validationDiagnostic(root, errItem)
			if !ok {
				continue
			}
			grouped[path] = append(grouped[path], diagnostic)
		}
		return
	}

	if root.LoadErr != nil {
		path, diagnostic, ok := c.loadErrorDiagnostic(root)
		if !ok {
			return
		}
		grouped[path] = append(grouped[path], diagnostic)
	}
}

func (c *diagnosticCollector) validationDiagnostic(root *workspace.RootRuntime, errItem validation.Error) (string, protocol.Diagnostic, bool) {
	loc, ok := resolveValidationLocation(root.Index.Root, errItem.File)
	if !ok {
		return "", protocol.Diagnostic{}, false
	}

	content, _ := c.readContent(loc.path)
	typeDef := validationType(root, errItem.Type)
	return loc.path, protocol.Diagnostic{
		Range:    diagnosticRangeForValidationError(content, loc, errItem, typeDef),
		Severity: protocol.DiagnosticSeverityError,
		Code:     string(errItem.Phase),
		Source:   diagnosticSource,
		Message:  errItem.Message,
	}, true
}

func (c *diagnosticCollector) loadErrorDiagnostic(root *workspace.RootRuntime) (string, protocol.Diagnostic, bool) {
	path := resolveLoadErrorPath(root.Index, root.LoadErr)
	if path == "" {
		return "", protocol.Diagnostic{}, false
	}

	content, _ := c.readContent(path)
	return path, protocol.Diagnostic{
		Range:    diagnosticRangeForLoadError(content, root.LoadErr),
		Severity: protocol.DiagnosticSeverityError,
		Code:     "config",
		Source:   diagnosticSource,
		Message:  root.LoadErr.Error(),
	}, true
}

func (c *diagnosticCollector) readContent(path string) ([]byte, bool) {
	if doc := c.documents[path]; doc != nil {
		return []byte(doc.Text), true
	}
	if cached, ok := c.content[path]; ok {
		return cached, true
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	c.content[path] = body
	return body, true
}

func validationType(root *workspace.RootRuntime, typeName string) *config.TypeDefinition {
	if root == nil || root.Validation == nil || root.Validation.Config == nil {
		return nil
	}
	return root.Validation.Config.Types[typeName]
}

func resolveValidationLocation(rootPath, file string) (diagnosticLocation, bool) {
	label := strings.TrimSpace(file)
	if label == "" {
		return diagnosticLocation{}, false
	}

	loc := diagnosticLocation{itemIndex: -1, inlineItem: -1}
	base := label

	if match := diagnosticItemPattern.FindStringSubmatch(label); len(match) == 3 {
		base = match[1]
		index, err := strconv.Atoi(match[2])
		if err == nil {
			loc.itemIndex = index - 1
		}
	} else if match := diagnosticInlineItemPattern.FindStringSubmatch(label); len(match) == 3 {
		base = match[1]
		index, err := strconv.Atoi(match[2])
		if err == nil {
			loc.inlineItem = index - 1
		}
	}

	if filepath.IsAbs(base) {
		loc.path = filepath.Clean(base)
		return loc, true
	}

	loc.path = filepath.Clean(filepath.Join(rootPath, base))
	return loc, true
}

func resolveLoadErrorPath(index *workspace.RootIndex, err error) string {
	if index == nil {
		return ""
	}

	message := ""
	if err != nil {
		message = err.Error()
	}

	var configFiles []string
	for path := range index.ConfigFiles {
		configFiles = append(configFiles, path)
	}
	slices.Sort(configFiles)
	for _, path := range configFiles {
		if strings.Contains(message, path) {
			return path
		}
	}

	if index.ConfigPath != "" {
		return index.ConfigPath
	}

	return ""
}

func diagnosticRangeForValidationError(content []byte, loc diagnosticLocation, errItem validation.Error, typeDef *config.TypeDefinition) protocol.Range {
	if lineRange, ok := rangeFromMessage(content, errItem.Message); ok {
		return lineRange
	}

	if fieldName := quotedFieldName(errItem.Message); fieldName != "" {
		if fieldRange, ok := locateFieldRange(content, fieldName, loc.itemIndex); ok {
			return fieldRange
		}
	}

	if strings.Contains(errItem.Message, "file declares type") {
		if typeRange, ok := locateFieldRange(content, "type", loc.itemIndex); ok {
			return typeRange
		}
	}

	if strings.Contains(errItem.Message, "identifier") || strings.Contains(errItem.Message, "missing required field") || errItem.ID != "" {
		if typeDef != nil && typeDef.Identifier.Field != "" {
			if idRange, ok := locateFieldRange(content, typeDef.Identifier.Field, loc.itemIndex); ok {
				return idRange
			}
		}
	}

	if objectRange, ok := locateObjectRange(content, loc.itemIndex); ok {
		return objectRange
	}

	return defaultDiagnosticRange(content)
}

func diagnosticRangeForLoadError(content []byte, err error) protocol.Range {
	message := ""
	if err != nil {
		message = err.Error()
	}

	if lineRange, ok := rangeFromMessage(content, message); ok {
		return lineRange
	}

	if configRange, ok := locateConfigErrorRange(content, message); ok {
		return configRange
	}

	return defaultDiagnosticRange(content)
}

func rangeFromMessage(content []byte, message string) (protocol.Range, bool) {
	match := diagnosticLinePattern.FindStringSubmatch(message)
	if len(match) < 2 {
		return protocol.Range{}, false
	}

	line, err := strconv.Atoi(match[1])
	if err != nil || line <= 0 {
		return protocol.Range{}, false
	}

	column := 1
	if len(match) > 2 && match[2] != "" {
		if parsed, err := strconv.Atoi(match[2]); err == nil && parsed > 0 {
			column = parsed
		}
	}

	return singlePointRange(content, line-1, column-1), true
}

func locateFieldRange(content []byte, fieldPath string, itemIndex int) (protocol.Range, bool) {
	node, ok := parseDocumentNode(content)
	if !ok {
		segments := parseFieldPath(fieldPath)
		if len(segments) == 0 {
			return protocol.Range{}, false
		}
		return locateLineContaining(content, segments[len(segments)-1].name)
	}

	object := selectObjectNode(node, itemIndex)
	if object == nil {
		return protocol.Range{}, false
	}

	current := object
	segments := parseFieldPath(fieldPath)
	for idx, segment := range segments {
		keyNode, valueNode := mappingEntry(current, segment.name)
		if keyNode == nil {
			return protocol.Range{}, false
		}

		if idx == len(segments)-1 {
			if segment.hasIndex && valueNode != nil && valueNode.Kind == yaml.SequenceNode && segment.index >= 0 && segment.index < len(valueNode.Content) {
				return nodeRange(valueNode.Content[segment.index]), true
			}
			return nodeRange(keyNode), true
		}

		if segment.hasIndex {
			if valueNode == nil || valueNode.Kind != yaml.SequenceNode || segment.index < 0 || segment.index >= len(valueNode.Content) {
				return nodeRange(keyNode), true
			}
			current = valueNode.Content[segment.index]
			continue
		}

		current = valueNode
		if current == nil {
			return nodeRange(keyNode), true
		}
	}

	return protocol.Range{}, false
}

func locateObjectRange(content []byte, itemIndex int) (protocol.Range, bool) {
	node, ok := parseDocumentNode(content)
	if !ok {
		return defaultDiagnosticRange(content), len(content) > 0
	}

	object := selectObjectNode(node, itemIndex)
	if object == nil {
		return protocol.Range{}, false
	}
	return nodeRange(object), true
}

func locateConfigErrorRange(content []byte, message string) (protocol.Range, bool) {
	if node, ok := parseDocumentNode(content); ok {
		switch {
		case strings.Contains(message, "mergeway.version"):
			if keyNode := findFirstKeyNode(node, "version"); keyNode != nil {
				return nodeRange(keyNode), true
			}
		case strings.Contains(message, "json_schema"):
			if keyNode := findFirstKeyNode(node, "json_schema"); keyNode != nil {
				return nodeRange(keyNode), true
			}
		case strings.Contains(message, "identifier"):
			if keyNode := findFirstKeyNode(node, "identifier"); keyNode != nil {
				return nodeRange(keyNode), true
			}
		case strings.Contains(message, "include"):
			if keyNode := findFirstKeyNode(node, "include"); keyNode != nil {
				return nodeRange(keyNode), true
			}
		}

		if match := diagnosticConfigFieldPattern.FindStringSubmatch(message); len(match) == 2 {
			parts := strings.Split(match[1], ".")
			if keyNode := findFirstKeyNode(node, parts[len(parts)-1]); keyNode != nil {
				return nodeRange(keyNode), true
			}
		}

		if match := diagnosticQuotedTypePattern.FindStringSubmatch(message); len(match) == 2 {
			if keyNode := findFirstKeyNode(node, match[1]); keyNode != nil {
				return nodeRange(keyNode), true
			}
		}

		if match := diagnosticQuotedNamePattern.FindStringSubmatch(message); len(match) == 2 {
			if keyNode := findFirstKeyNode(node, match[1]); keyNode != nil {
				return nodeRange(keyNode), true
			}
		}
	}

	if strings.Contains(message, "mergeway.version") {
		return locateLineContaining(content, "version")
	}
	if strings.Contains(message, "include") {
		return locateLineContaining(content, "include")
	}

	return protocol.Range{}, false
}

func quotedFieldName(message string) string {
	match := diagnosticQuotedFieldPattern.FindStringSubmatch(message)
	if len(match) != 2 {
		return ""
	}
	return match[1]
}

func parseFieldPath(path string) []fieldSegment {
	if strings.TrimSpace(path) == "" {
		return nil
	}

	parts := strings.Split(path, ".")
	segments := make([]fieldSegment, 0, len(parts))
	for _, part := range parts {
		segment := fieldSegment{name: part}
		if open := strings.Index(part, "["); open >= 0 && strings.HasSuffix(part, "]") {
			segment.name = part[:open]
			index, err := strconv.Atoi(part[open+1 : len(part)-1])
			if err == nil {
				segment.index = index
				segment.hasIndex = true
			}
		}
		segments = append(segments, segment)
	}
	return segments
}

func parseDocumentNode(content []byte) (*yaml.Node, bool) {
	if len(content) == 0 {
		return nil, false
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, false
	}
	if len(doc.Content) == 0 {
		return nil, false
	}
	return &doc, true
}

func selectObjectNode(doc *yaml.Node, itemIndex int) *yaml.Node {
	root := documentRoot(doc)
	if root == nil {
		return nil
	}

	if itemIndex >= 0 {
		_, itemsNode := mappingEntry(root, "items")
		if itemsNode != nil && itemsNode.Kind == yaml.SequenceNode && itemIndex < len(itemsNode.Content) {
			return itemsNode.Content[itemIndex]
		}
	}

	return root
}

func documentRoot(doc *yaml.Node) *yaml.Node {
	if doc == nil {
		return nil
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0]
	}
	return doc
}

func mappingEntry(node *yaml.Node, key string) (*yaml.Node, *yaml.Node) {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil, nil
	}

	for idx := 0; idx+1 < len(node.Content); idx += 2 {
		keyNode := node.Content[idx]
		valueNode := node.Content[idx+1]
		if keyNode.Value == key {
			return keyNode, valueNode
		}
	}

	return nil, nil
}

func findFirstKeyNode(node *yaml.Node, key string) *yaml.Node {
	node = documentRoot(node)
	if node == nil {
		return nil
	}

	switch node.Kind {
	case yaml.MappingNode:
		for idx := 0; idx+1 < len(node.Content); idx += 2 {
			keyNode := node.Content[idx]
			valueNode := node.Content[idx+1]
			if keyNode.Value == key {
				return keyNode
			}
			if nested := findFirstKeyNode(valueNode, key); nested != nil {
				return nested
			}
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if nested := findFirstKeyNode(child, key); nested != nil {
				return nested
			}
		}
	}

	return nil
}

func nodeRange(node *yaml.Node) protocol.Range {
	if node == nil {
		return protocol.Range{}
	}

	line := node.Line - 1
	if line < 0 {
		line = 0
	}
	column := node.Column - 1
	if column < 0 {
		column = 0
	}

	endColumn := column + 1
	if node.Value != "" {
		endColumn = column + len(node.Value)
	}

	return protocol.Range{
		Start: protocol.Position{Line: uint32(line), Character: uint32(column)},
		End:   protocol.Position{Line: uint32(line), Character: uint32(endColumn)},
	}
}

func locateLineContaining(content []byte, needle string) (protocol.Range, bool) {
	if strings.TrimSpace(needle) == "" {
		return protocol.Range{}, false
	}

	lines := strings.Split(string(content), "\n")
	for idx, line := range lines {
		column := strings.Index(line, needle)
		if column < 0 {
			continue
		}
		return protocol.Range{
			Start: protocol.Position{Line: uint32(idx), Character: uint32(column)},
			End:   protocol.Position{Line: uint32(idx), Character: uint32(column + max(1, len(needle)))},
		}, true
	}

	return protocol.Range{}, false
}

func singlePointRange(content []byte, line, column int) protocol.Range {
	lines := strings.Split(string(content), "\n")
	if line < 0 {
		line = 0
	}
	if line >= len(lines) {
		if len(lines) == 0 {
			line = 0
		} else {
			line = len(lines) - 1
		}
	}

	current := ""
	if line < len(lines) {
		current = lines[line]
	}

	if column < 0 {
		column = 0
	}
	if column > len(current) {
		column = len(current)
	}

	end := column + 1
	if len(current) == 0 {
		end = column
	}
	if end > len(current) {
		end = len(current)
	}

	return protocol.Range{
		Start: protocol.Position{Line: uint32(line), Character: uint32(column)},
		End:   protocol.Position{Line: uint32(line), Character: uint32(end)},
	}
}

func defaultDiagnosticRange(content []byte) protocol.Range {
	if len(content) == 0 {
		return protocol.Range{}
	}
	return singlePointRange(content, 0, 0)
}

func documentVersion(doc *workspace.OpenDocument) uint32 {
	if doc == nil || doc.Version <= 0 {
		return 0
	}
	return uint32(doc.Version)
}

func streamDiagnosticPublisher(stream jsonrpc2StreamWriter, writeMu locker) diagnosticPublisher {
	return func(ctx context.Context, params *protocol.PublishDiagnosticsParams) error {
		notification, err := newNotification(protocol.MethodTextDocumentPublishDiagnostics, params)
		if err != nil {
			return fmt.Errorf("publish diagnostics: %w", err)
		}
		return writeJSONRPCMessage(ctx, stream, writeMu, notification)
	}
}
