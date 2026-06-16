package lsp

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/data"
	"github.com/mergewayhq/mergeway-cli/internal/workspace"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
)

var primitiveFieldTypes = []string{
	"string",
	"integer",
	"number",
	"boolean",
	"enum",
	"object",
}

type analysisKind string

const (
	analysisKindUnknown analysisKind = "unknown"
	analysisKindData    analysisKind = "data"
	analysisKindConfig  analysisKind = "config"
)

type documentAnalysis struct {
	path     string
	content  []byte
	lines    []string
	position protocol.Position
	root     *workspace.RootRuntime
	kind     analysisKind
	doc      *yaml.Node
	cfg      *config.Config

	data   *dataAnalysis
	config *configAnalysis
}

type dataAnalysis struct {
	typeDef       *config.TypeDefinition
	objectNode    *yaml.Node
	objectIndex   int
	fieldName     string
	fieldDef      *config.FieldDefinition
	currentValue  string
	onFieldKey    bool
	onFieldValue  bool
	pendingField  bool
	keyPrefix     string
	valuePrefix   string
	positionRange protocol.Range
}

type configAnalysis struct {
	completionKinds []protocol.CompletionItem
}

func (s *Server) completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	analysis, err := s.analyzePosition(params.TextDocument.URI.Filename(), params.Position)
	if err != nil {
		return nil, err
	}

	items := completionItemsForAnalysis(analysis)
	return &protocol.CompletionList{Items: items}, nil
}

func (s *Server) hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	analysis, err := s.analyzePosition(params.TextDocument.URI.Filename(), params.Position)
	if err != nil {
		return nil, err
	}

	content := hoverContentForAnalysis(analysis)
	if content == "" {
		return nil, nil
	}

	result := &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: content,
		},
	}
	if analysis.data != nil {
		result.Range = &analysis.data.positionRange
	}
	return result, nil
}

func (s *Server) definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	analysis, err := s.analyzePosition(params.TextDocument.URI.Filename(), params.Position)
	if err != nil {
		return nil, err
	}
	if analysis == nil || analysis.data == nil || analysis.data.fieldDef == nil || !analysis.data.fieldDef.IsReference() || analysis.data.currentValue == "" {
		return nil, nil
	}

	targets := resolveReferenceObjects(analysis.root, analysis.data.fieldDef, analysis.data.currentValue)
	if len(targets) != 1 {
		return nil, nil
	}

	target := targets[0]
	targetRange, ok := s.objectIdentifierRange(analysis.root, target)
	if !ok {
		return nil, nil
	}

	return []protocol.Location{{
		URI:   protocol.DocumentURI(uri.File(target.File)),
		Range: targetRange,
	}}, nil
}

func (s *Server) analyzePosition(path string, position protocol.Position) (*documentAnalysis, error) {
	if s.runtime == nil {
		return nil, nil
	}

	root := s.runtime.RootByPath(path)
	if root == nil {
		return nil, nil
	}

	content, err := s.documentContent(path)
	if err != nil {
		return nil, err
	}

	analysis := &documentAnalysis{
		path:     path,
		content:  content,
		lines:    strings.Split(string(content), "\n"),
		position: position,
		root:     root,
		cfg:      validationConfig(root),
	}

	if doc, ok := parseDocumentNode(content); ok {
		analysis.doc = doc
	}

	if root.Index != nil {
		if _, ok := root.Index.ConfigFiles[path]; ok {
			analysis.kind = analysisKindConfig
			analysis.config = analyzeConfigPosition(analysis)
			return analysis, nil
		}
		if len(root.Index.TypesForFile(path)) > 0 {
			analysis.kind = analysisKindData
			analysis.data = analyzeDataPosition(analysis)
			return analysis, nil
		}
	}

	return analysis, nil
}

func (s *Server) documentContent(path string) ([]byte, error) {
	if s.runtime != nil {
		if doc := s.runtime.Document(path); doc != nil {
			return []byte(doc.Text), nil
		}
	}
	return os.ReadFile(path)
}

func validationConfig(root *workspace.RootRuntime) *config.Config {
	if root == nil || root.Validation == nil {
		return nil
	}
	return root.Validation.Config
}

func analyzeConfigPosition(analysis *documentAnalysis) *configAnalysis {
	line := lineAt(analysis.lines, int(analysis.position.Line))
	beforeCursor := sliceToCursor(line, analysis.position.Character)
	trimmed := strings.TrimSpace(beforeCursor)
	if !strings.Contains(trimmed, "type:") {
		return nil
	}

	idx := strings.Index(trimmed, "type:")
	if idx < 0 {
		return nil
	}

	valuePrefix := strings.TrimSpace(trimmed[idx+len("type:"):])
	typeCount := 0
	if analysis.cfg != nil {
		typeCount = len(analysis.cfg.Types)
	}
	items := make([]protocol.CompletionItem, 0, len(primitiveFieldTypes)+typeCount)
	for _, primitive := range primitiveFieldTypes {
		if !strings.HasPrefix(primitive, valuePrefix) {
			continue
		}
		items = append(items, protocol.CompletionItem{
			Label:         primitive,
			Kind:          protocol.CompletionItemKindTypeParameter,
			Documentation: primitiveTypeDoc(primitive),
		})
	}
	if analysis.cfg != nil {
		for _, name := range sortedTypeNames(analysis.cfg.Types) {
			if !strings.HasPrefix(name, valuePrefix) {
				continue
			}
			items = append(items, protocol.CompletionItem{
				Label:         name,
				Kind:          protocol.CompletionItemKindClass,
				Documentation: typeDocumentation(analysis.cfg.Types[name]),
			})
		}
	}
	return &configAnalysis{completionKinds: items}
}

func analyzeDataPosition(analysis *documentAnalysis) *dataAnalysis {
	typeDef := resolveDataTypeDefinition(analysis)
	if typeDef == nil {
		return nil
	}

	result := &dataAnalysis{
		typeDef:     typeDef,
		objectIndex: -1,
	}

	if analysis.doc != nil {
		result.objectNode, result.objectIndex = objectNodeAtLine(analysis.doc, int(analysis.position.Line))
		if result.objectNode != nil {
			populateDataNodeContext(analysis, result)
		}
	}

	populateDataLineContext(analysis, result)
	if result.fieldName != "" && typeDef.Fields != nil {
		result.fieldDef = typeDef.Fields[result.fieldName]
	}
	if result.positionRange == (protocol.Range{}) {
		result.positionRange = singlePointRange(analysis.content, int(analysis.position.Line), int(analysis.position.Character))
	}

	return result
}

func resolveDataTypeDefinition(analysis *documentAnalysis) *config.TypeDefinition {
	if analysis == nil || analysis.root == nil || analysis.root.Index == nil {
		return nil
	}
	typeNames := analysis.root.Index.TypesForFile(analysis.path)
	if analysis.cfg == nil {
		return nil
	}
	if len(typeNames) == 1 {
		return analysis.cfg.Types[typeNames[0]]
	}
	if analysis.doc != nil {
		root := documentRoot(analysis.doc)
		if _, valueNode := mappingEntry(root, "type"); valueNode != nil && valueNode.Kind == yaml.ScalarNode {
			if typeDef := analysis.cfg.Types[valueNode.Value]; typeDef != nil {
				return typeDef
			}
		}
	}
	for _, typeName := range typeNames {
		if typeDef := analysis.cfg.Types[typeName]; typeDef != nil {
			return typeDef
		}
	}
	return nil
}

func objectNodeAtLine(doc *yaml.Node, line int) (*yaml.Node, int) {
	root := documentRoot(doc)
	if root == nil || root.Kind != yaml.MappingNode {
		return nil, -1
	}

	if _, itemsNode := mappingEntry(root, "items"); itemsNode != nil && itemsNode.Kind == yaml.SequenceNode {
		index := 0
		for idx, itemNode := range itemsNode.Content {
			if itemNode == nil {
				continue
			}
			if idx+1 < len(itemsNode.Content) && itemsNode.Content[idx+1] != nil {
				nextLine := itemsNode.Content[idx+1].Line - 1
				if line < nextLine {
					return itemNode, idx
				}
			}
			index = idx
		}
		if len(itemsNode.Content) > 0 && line >= itemsNode.Content[0].Line-1 {
			return itemsNode.Content[index], index
		}
	}

	return root, -1
}

func populateDataNodeContext(analysis *documentAnalysis, result *dataAnalysis) {
	if result.objectNode == nil || result.objectNode.Kind != yaml.MappingNode {
		return
	}

	for idx := 0; idx+1 < len(result.objectNode.Content); idx += 2 {
		keyNode := result.objectNode.Content[idx]
		valueNode := result.objectNode.Content[idx+1]

		if positionWithinNode(analysis.position, keyNode) {
			result.fieldName = keyNode.Value
			result.onFieldKey = true
			result.positionRange = nodeRange(keyNode)
			return
		}

		if positionWithinNode(analysis.position, valueNode) {
			result.fieldName = keyNode.Value
			result.currentValue = scalarNodeValueAtPosition(analysis.position, valueNode)
			result.onFieldValue = true
			result.positionRange = scalarNodeRangeAtPosition(analysis.position, valueNode)
			if result.positionRange == (protocol.Range{}) {
				result.positionRange = nodeRange(valueNode)
			}
			return
		}
	}
}

func populateDataLineContext(analysis *documentAnalysis, result *dataAnalysis) {
	lineNo := int(analysis.position.Line)
	line := lineAt(analysis.lines, lineNo)
	beforeCursor := sliceToCursor(line, analysis.position.Character)
	trimmed := strings.TrimSpace(beforeCursor)

	if result.onFieldKey || result.onFieldValue {
		if !result.onFieldValue && strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			result.keyPrefix = strings.TrimSpace(parts[0])
		}
		if result.onFieldValue {
			result.valuePrefix = trailingScalarPrefix(beforeCursor)
		}
		return
	}

	if result.objectNode == nil {
		return
	}

	objectIndent := objectFieldIndent(result.objectNode)
	trimmedLine := strings.TrimSpace(line)
	if trimmedLine == "" || indentation(line) == objectIndent {
		result.pendingField = true
		result.keyPrefix = strings.TrimSpace(strings.TrimSuffix(trimmed, ":"))
		result.positionRange = singlePointRange(analysis.content, lineNo, int(analysis.position.Character))
	}

	if strings.HasPrefix(strings.TrimSpace(line), "-") {
		fieldName := enclosingSequenceField(result.objectNode, lineNo)
		if fieldName != "" {
			result.fieldName = fieldName
			result.fieldDef = result.typeDef.Fields[fieldName]
			result.onFieldValue = true
			result.valuePrefix = strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
			result.positionRange = singlePointRange(analysis.content, lineNo, int(analysis.position.Character))
		}
	}
}

func completionItemsForAnalysis(analysis *documentAnalysis) []protocol.CompletionItem {
	if analysis == nil {
		return nil
	}
	if analysis.kind == analysisKindConfig && analysis.config != nil {
		return analysis.config.completionKinds
	}
	if analysis.data == nil || analysis.data.typeDef == nil {
		return nil
	}

	if analysis.data.pendingField || analysis.data.onFieldKey {
		return fieldCompletionItems(analysis.data.typeDef, analysis.data.objectNode, analysis.data.keyPrefix)
	}
	if analysis.data.onFieldValue && analysis.data.fieldDef != nil {
		if len(analysis.data.fieldDef.Enum) > 0 {
			return enumCompletionItems(analysis.data.fieldDef, analysis.data.valuePrefix)
		}
		if analysis.data.fieldDef.IsReference() {
			return referenceCompletionItems(analysis.root, analysis.data.fieldDef, analysis.data.valuePrefix)
		}
	}
	return nil
}

func fieldCompletionItems(typeDef *config.TypeDefinition, objectNode *yaml.Node, prefix string) []protocol.CompletionItem {
	if typeDef == nil {
		return nil
	}

	present := make(map[string]struct{})
	if objectNode != nil && objectNode.Kind == yaml.MappingNode {
		for idx := 0; idx+1 < len(objectNode.Content); idx += 2 {
			present[objectNode.Content[idx].Value] = struct{}{}
		}
	}

	items := make([]protocol.CompletionItem, 0, len(typeDef.FieldOrder))
	for _, fieldName := range typeDef.FieldOrder {
		if _, ok := present[fieldName]; ok {
			continue
		}
		if prefix != "" && !strings.HasPrefix(fieldName, prefix) {
			continue
		}
		fieldDef := typeDef.Fields[fieldName]
		items = append(items, protocol.CompletionItem{
			Label:         fieldName,
			Kind:          protocol.CompletionItemKindField,
			Documentation: fieldDocumentation(fieldDef),
		})
	}
	return items
}

func enumCompletionItems(fieldDef *config.FieldDefinition, prefix string) []protocol.CompletionItem {
	items := make([]protocol.CompletionItem, 0, len(fieldDef.Enum))
	for _, value := range fieldDef.Enum {
		if prefix != "" && !strings.HasPrefix(value, prefix) {
			continue
		}
		items = append(items, protocol.CompletionItem{
			Label:         value,
			Kind:          protocol.CompletionItemKindEnumMember,
			Documentation: fieldDocumentation(fieldDef),
		})
	}
	return items
}

func referenceCompletionItems(root *workspace.RootRuntime, fieldDef *config.FieldDefinition, prefix string) []protocol.CompletionItem {
	ids := referenceIDs(root, fieldDef)
	items := make([]protocol.CompletionItem, 0, len(ids))
	for _, id := range ids {
		if prefix != "" && !strings.HasPrefix(id, prefix) {
			continue
		}
		items = append(items, protocol.CompletionItem{
			Label:         id,
			Kind:          protocol.CompletionItemKindReference,
			Documentation: referenceCompletionDocumentation(root, fieldDef, id),
		})
	}
	return items
}

func hoverContentForAnalysis(analysis *documentAnalysis) string {
	if analysis == nil || analysis.data == nil {
		return ""
	}

	if analysis.data.onFieldKey && analysis.data.fieldDef != nil {
		return fieldDocumentation(analysis.data.fieldDef)
	}

	if analysis.data.onFieldValue {
		if analysis.data.fieldName == analysis.data.typeDef.Identifier.Field && analysis.data.currentValue != "" {
			return objectHoverContent(analysis.root, analysis.data.typeDef.Name, analysis.data.currentValue)
		}
		if analysis.data.fieldDef != nil && analysis.data.fieldDef.IsReference() && analysis.data.currentValue != "" {
			targets := resolveReferenceObjects(analysis.root, analysis.data.fieldDef, analysis.data.currentValue)
			if len(targets) == 1 {
				return objectSummaryMarkdown(targets[0], analysis.root)
			}
		}
	}

	return ""
}

func objectHoverContent(root *workspace.RootRuntime, typeName, id string) string {
	if root == nil || root.Workspace == nil || id == "" {
		return ""
	}
	matches := root.Workspace.Find(typeName, id)
	if len(matches) != 1 {
		return ""
	}
	return objectSummaryMarkdown(matches[0], root)
}

func objectSummaryMarkdown(obj *data.Object, root *workspace.RootRuntime) string {
	if obj == nil {
		return ""
	}

	cfg := validationConfig(root)
	var typeDef *config.TypeDefinition
	if cfg != nil {
		typeDef = cfg.Types[obj.Type]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "**%s** `%s`", obj.Type, obj.ID)
	if typeDef != nil && typeDef.Description != "" {
		fmt.Fprintf(&b, "\n\n%s", typeDef.Description)
	}
	for _, key := range []string{"title", "name", "label", "email", "status"} {
		if value, ok := obj.Fields[key]; ok {
			fmt.Fprintf(&b, "\n\n`%s`: %v", key, value)
		}
	}
	return b.String()
}

func resolveReferenceObjects(root *workspace.RootRuntime, fieldDef *config.FieldDefinition, id string) []*data.Object {
	if root == nil || root.Workspace == nil || fieldDef == nil || id == "" {
		return nil
	}
	var matches []*data.Object
	for _, refType := range fieldDef.ReferenceTypes {
		matches = append(matches, root.Workspace.Find(refType, id)...)
	}
	return matches
}

func referenceIDs(root *workspace.RootRuntime, fieldDef *config.FieldDefinition) []string {
	if root == nil || root.Workspace == nil || fieldDef == nil {
		return nil
	}
	set := make(map[string]struct{})
	for _, refType := range fieldDef.ReferenceTypes {
		for _, obj := range root.Workspace.Objects(refType) {
			set[obj.ID] = struct{}{}
		}
	}
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func referenceCompletionDocumentation(root *workspace.RootRuntime, fieldDef *config.FieldDefinition, id string) string {
	matches := resolveReferenceObjects(root, fieldDef, id)
	if len(matches) != 1 {
		return fieldDocumentation(fieldDef)
	}
	return objectSummaryMarkdown(matches[0], root)
}

func fieldDocumentation(fieldDef *config.FieldDefinition) string {
	if fieldDef == nil {
		return ""
	}
	var tags []string
	if fieldDef.Required {
		tags = append(tags, "required")
	}
	if fieldDef.Repeated {
		tags = append(tags, "repeated")
	}

	typeLabel := fieldDef.Type
	if fieldDef.IsReference() {
		typeLabel = fieldDef.ReferenceLabel()
	}

	var b strings.Builder
	fmt.Fprintf(&b, "**%s**", fieldDef.Name)
	fmt.Fprintf(&b, "\n\nType: `%s`", typeLabel)
	if len(tags) > 0 {
		fmt.Fprintf(&b, " (%s)", strings.Join(tags, ", "))
	}
	if len(fieldDef.Enum) > 0 {
		fmt.Fprintf(&b, "\n\nEnum: `%s`", strings.Join(fieldDef.Enum, "`, `"))
	}
	if fieldDef.Description != "" {
		fmt.Fprintf(&b, "\n\n%s", fieldDef.Description)
	}
	return b.String()
}

func primitiveTypeDoc(name string) string {
	switch name {
	case "enum":
		return "String field limited to configured enum values."
	case "object":
		return "Nested object field with configured properties."
	default:
		return fmt.Sprintf("Primitive field type `%s`.", name)
	}
}

func typeDocumentation(typeDef *config.TypeDefinition) string {
	if typeDef == nil {
		return ""
	}
	if typeDef.Description == "" {
		return fmt.Sprintf("Entity type `%s`.", typeDef.Name)
	}
	return fmt.Sprintf("Entity type `%s`.\n\n%s", typeDef.Name, typeDef.Description)
}

func (s *Server) objectIdentifierRange(root *workspace.RootRuntime, obj *data.Object) (protocol.Range, bool) {
	if obj == nil {
		return protocol.Range{}, false
	}

	content, err := s.documentContent(obj.File)
	if err != nil {
		return protocol.Range{}, false
	}

	doc, ok := parseDocumentNode(content)
	if !ok {
		return protocol.Range{}, false
	}

	rootNode := documentRoot(doc)
	cfg := validationConfig(root)
	if cfg == nil {
		return protocol.Range{}, false
	}
	typeDef := cfg.Types[obj.Type]
	if typeDef == nil {
		return protocol.Range{}, false
	}

	if _, itemsNode := mappingEntry(rootNode, "items"); itemsNode != nil && itemsNode.Kind == yaml.SequenceNode {
		for _, itemNode := range itemsNode.Content {
			if itemMatchesIdentifier(itemNode, typeDef.Identifier.Field, obj.ID) {
				if _, valueNode := mappingEntry(itemNode, typeDef.Identifier.Field); valueNode != nil {
					return nodeRange(valueNode), true
				}
			}
		}
	}

	if _, valueNode := mappingEntry(rootNode, typeDef.Identifier.Field); valueNode != nil && valueNode.Value == obj.ID {
		return nodeRange(valueNode), true
	}

	return protocol.Range{}, false
}

func itemMatchesIdentifier(node *yaml.Node, idField, want string) bool {
	if node == nil || node.Kind != yaml.MappingNode {
		return false
	}
	_, valueNode := mappingEntry(node, idField)
	return valueNode != nil && valueNode.Value == want
}

func positionWithinNode(pos protocol.Position, node *yaml.Node) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case yaml.ScalarNode:
		return positionWithinRange(pos, nodeRange(node))
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if positionWithinNode(pos, child) {
				return true
			}
		}
	case yaml.MappingNode:
		for idx := 0; idx+1 < len(node.Content); idx += 2 {
			if positionWithinNode(pos, node.Content[idx]) || positionWithinNode(pos, node.Content[idx+1]) {
				return true
			}
		}
	}
	return false
}

func scalarNodeValueAtPosition(pos protocol.Position, node *yaml.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == yaml.ScalarNode && positionWithinRange(pos, nodeRange(node)) {
		return node.Value
	}
	if node.Kind == yaml.SequenceNode {
		for _, child := range node.Content {
			if child.Kind == yaml.ScalarNode && positionWithinRange(pos, nodeRange(child)) {
				return child.Value
			}
		}
	}
	return ""
}

func scalarNodeRangeAtPosition(pos protocol.Position, node *yaml.Node) protocol.Range {
	if node == nil {
		return protocol.Range{}
	}
	if node.Kind == yaml.ScalarNode && positionWithinRange(pos, nodeRange(node)) {
		return nodeRange(node)
	}
	if node.Kind == yaml.SequenceNode {
		for _, child := range node.Content {
			if child.Kind == yaml.ScalarNode && positionWithinRange(pos, nodeRange(child)) {
				return nodeRange(child)
			}
		}
	}
	return protocol.Range{}
}

func positionWithinRange(pos protocol.Position, r protocol.Range) bool {
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}
	if pos.Line == r.Start.Line && pos.Character < r.Start.Character {
		return false
	}
	if pos.Line == r.End.Line && pos.Character > r.End.Character {
		return false
	}
	return true
}

func objectFieldIndent(node *yaml.Node) int {
	if node == nil || node.Kind != yaml.MappingNode || len(node.Content) == 0 {
		return 0
	}
	return node.Content[0].Column - 1
}

func indentation(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

func lineAt(lines []string, index int) string {
	if index < 0 || index >= len(lines) {
		return ""
	}
	return lines[index]
}

func sliceToCursor(line string, char uint32) string {
	if int(char) > len(line) {
		return line
	}
	return line[:char]
}

func trailingScalarPrefix(line string) string {
	if idx := strings.Index(line, ":"); idx >= 0 {
		return strings.TrimSpace(line[idx+1:])
	}
	return strings.TrimSpace(strings.TrimPrefix(line, "-"))
}

func enclosingSequenceField(objectNode *yaml.Node, line int) string {
	if objectNode == nil || objectNode.Kind != yaml.MappingNode {
		return ""
	}
	for idx := 0; idx+1 < len(objectNode.Content); idx += 2 {
		keyNode := objectNode.Content[idx]
		valueNode := objectNode.Content[idx+1]
		if valueNode.Kind != yaml.SequenceNode || len(valueNode.Content) == 0 {
			continue
		}
		if line >= valueNode.Content[0].Line-1 {
			return keyNode.Value
		}
	}
	return ""
}

func sortedTypeNames(types map[string]*config.TypeDefinition) []string {
	names := make([]string, 0, len(types))
	for name := range types {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
