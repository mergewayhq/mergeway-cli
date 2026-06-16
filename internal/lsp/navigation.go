package lsp

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/data"
	"github.com/mergewayhq/mergeway-cli/internal/workspace"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
)

type referenceTarget struct {
	root         *workspace.RootRuntime
	id           string
	typeNames    []string
	declarations []*data.Object
}

func (s *Server) references(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	analysis, err := s.analyzePosition(params.TextDocument.URI.Filename(), params.Position)
	if err != nil {
		return nil, err
	}

	target := resolveReferenceTarget(analysis)
	if target == nil {
		return nil, nil
	}

	var locations []protocol.Location
	if params.Context.IncludeDeclaration {
		locations = append(locations, s.referenceDeclarations(target)...)
	}
	locations = append(locations, s.referenceUsages(target)...)
	return sortUniqueLocations(locations), nil
}

func (s *Server) documentSymbols(ctx context.Context, params *protocol.DocumentSymbolParams) ([]protocol.DocumentSymbol, error) {
	path := params.TextDocument.URI.Filename()
	if s.runtime == nil {
		return nil, nil
	}

	root := s.runtime.RootByPath(path)
	if root == nil || root.Index == nil {
		return nil, nil
	}
	if len(root.Index.TypesForFile(path)) == 0 {
		return nil, nil
	}

	content, err := s.documentContent(path)
	if err != nil {
		return nil, err
	}
	doc, ok := parseDocumentNode(content)
	if !ok {
		return nil, nil
	}

	analysis := &documentAnalysis{
		path:    path,
		content: content,
		root:    root,
		doc:     doc,
		cfg:     validationConfig(root),
	}
	typeDef := resolveDataTypeDefinition(analysis)
	if typeDef == nil {
		return nil, nil
	}

	objectNodes := documentObjectNodes(doc)
	symbols := make([]protocol.DocumentSymbol, 0, len(objectNodes))
	for index, objectNode := range objectNodes {
		symbol := buildDocumentSymbol(root, typeDef, path, objectNode, index)
		if symbol.Name == "" {
			continue
		}
		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

func (s *Server) workspaceSymbols(ctx context.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	if s.runtime == nil {
		return nil, nil
	}

	query := strings.ToLower(strings.TrimSpace(params.Query))
	snapshot := s.runtime.Snapshot()
	rootPaths := sortedRootPaths(snapshot.Roots)

	var symbols []protocol.SymbolInformation
	for _, rootPath := range rootPaths {
		root := snapshot.Roots[rootPath]
		cfg := validationConfig(root)
		if root == nil || root.Workspace == nil || cfg == nil {
			continue
		}

		for _, typeName := range sortedTypeNames(cfg.Types) {
			for _, obj := range root.Workspace.Objects(typeName) {
				if !workspaceSymbolMatches(query, typeName, obj) {
					continue
				}

				targetRange, ok := s.objectIdentifierRange(root, obj)
				if !ok {
					continue
				}

				symbols = append(symbols, protocol.SymbolInformation{
					Name: workspaceSymbolName(obj),
					Kind: protocol.SymbolKindObject,
					Location: protocol.Location{
						URI:   protocol.DocumentURI(uri.File(obj.File)),
						Range: targetRange,
					},
					ContainerName: workspaceSymbolContainer(root, obj),
				})
			}
		}
	}

	sort.Slice(symbols, func(i, j int) bool {
		if symbols[i].Name != symbols[j].Name {
			return symbols[i].Name < symbols[j].Name
		}
		if symbols[i].ContainerName != symbols[j].ContainerName {
			return symbols[i].ContainerName < symbols[j].ContainerName
		}
		if symbols[i].Location.URI != symbols[j].Location.URI {
			return symbols[i].Location.URI < symbols[j].Location.URI
		}
		return compareRanges(symbols[i].Location.Range, symbols[j].Location.Range) < 0
	})

	return symbols, nil
}

func resolveReferenceTarget(analysis *documentAnalysis) *referenceTarget {
	if analysis == nil || analysis.root == nil || analysis.data == nil || analysis.data.typeDef == nil || analysis.data.currentValue == "" {
		return nil
	}

	if analysis.data.fieldName == analysis.data.typeDef.Identifier.Field {
		typeNames := []string{analysis.data.typeDef.Name}
		declarations := workspaceObjectsForFile(analysis.root, analysis.data.typeDef.Name, analysis.path, analysis.data.currentValue)
		if len(declarations) == 0 && analysis.root.Workspace != nil {
			declarations = analysis.root.Workspace.Find(analysis.data.typeDef.Name, analysis.data.currentValue)
		}
		return &referenceTarget{
			root:         analysis.root,
			id:           analysis.data.currentValue,
			typeNames:    typeNames,
			declarations: declarations,
		}
	}

	if analysis.data.fieldDef == nil || !analysis.data.fieldDef.IsReference() {
		return nil
	}

	declarations := resolveReferenceObjects(analysis.root, analysis.data.fieldDef, analysis.data.currentValue)
	typeNames := uniqueStrings(analysis.data.fieldDef.ReferenceTypes)
	if len(declarations) > 0 {
		typeNames = nil
		for _, obj := range declarations {
			typeNames = append(typeNames, obj.Type)
		}
		typeNames = uniqueStrings(typeNames)
	}

	return &referenceTarget{
		root:         analysis.root,
		id:           analysis.data.currentValue,
		typeNames:    typeNames,
		declarations: declarations,
	}
}

func (s *Server) referenceDeclarations(target *referenceTarget) []protocol.Location {
	if target == nil {
		return nil
	}

	locations := make([]protocol.Location, 0, len(target.declarations))
	for _, obj := range target.declarations {
		targetRange, ok := s.objectIdentifierRange(target.root, obj)
		if !ok {
			continue
		}
		locations = append(locations, protocol.Location{
			URI:   protocol.DocumentURI(uri.File(obj.File)),
			Range: targetRange,
		})
	}
	return locations
}

func (s *Server) referenceUsages(target *referenceTarget) []protocol.Location {
	if target == nil || target.root == nil || target.root.Workspace == nil {
		return nil
	}

	cfg := validationConfig(target.root)
	if cfg == nil {
		return nil
	}

	typeSet := make(map[string]struct{}, len(target.typeNames))
	for _, typeName := range target.typeNames {
		typeSet[typeName] = struct{}{}
	}

	var locations []protocol.Location
	for _, sourceType := range sortedTypeNames(cfg.Types) {
		typeDef := cfg.Types[sourceType]
		if typeDef == nil {
			continue
		}

		referenceFields := referenceFieldNames(typeDef, typeSet)
		if len(referenceFields) == 0 {
			continue
		}

		for _, obj := range target.root.Workspace.Objects(sourceType) {
			content, err := s.documentContent(obj.File)
			if err != nil {
				continue
			}
			doc, ok := parseDocumentNode(content)
			if !ok {
				continue
			}
			objectNode := objectNodeForObject(doc, typeDef, obj)
			if objectNode == nil {
				continue
			}

			for _, fieldName := range referenceFields {
				_, valueNode := mappingEntry(objectNode, fieldName)
				if valueNode == nil {
					continue
				}
				locations = append(locations, referenceNodeLocations(obj.File, valueNode, target.id)...)
			}
		}
	}

	return locations
}

func buildDocumentSymbol(root *workspace.RootRuntime, typeDef *config.TypeDefinition, path string, objectNode *yaml.Node, index int) protocol.DocumentSymbol {
	objectID, selectionRange := objectIdentifierSelection(objectNode, typeDef)
	obj := workspaceObjectForNode(root, typeDef.Name, path, objectID, index)

	name := objectID
	if name == "" {
		name = fallbackObjectName(typeDef.Name, index)
	}

	symbol := protocol.DocumentSymbol{
		Name:           name,
		Detail:         documentSymbolDetail(typeDef.Name, obj),
		Kind:           protocol.SymbolKindObject,
		Range:          structuralNodeRange(objectNode),
		SelectionRange: selectionRange,
	}
	if symbol.SelectionRange == (protocol.Range{}) {
		symbol.SelectionRange = symbol.Range
	}

	if objectNode != nil && objectNode.Kind == yaml.MappingNode {
		children := make([]protocol.DocumentSymbol, 0, len(objectNode.Content)/2)
		for idx := 0; idx+1 < len(objectNode.Content); idx += 2 {
			keyNode := objectNode.Content[idx]
			valueNode := objectNode.Content[idx+1]
			fieldDef := typeDef.Fields[keyNode.Value]
			children = append(children, protocol.DocumentSymbol{
				Name:           keyNode.Value,
				Detail:         documentFieldDetail(fieldDef),
				Kind:           documentFieldKind(fieldDef, valueNode),
				Range:          structuralRangeIncludingKey(keyNode, valueNode),
				SelectionRange: nodeRange(keyNode),
			})
		}
		symbol.Children = children
	}

	return symbol
}

func documentObjectNodes(doc *yaml.Node) []*yaml.Node {
	root := documentRoot(doc)
	if root == nil {
		return nil
	}

	if _, itemsNode := mappingEntry(root, "items"); itemsNode != nil && itemsNode.Kind == yaml.SequenceNode {
		return append([]*yaml.Node(nil), itemsNode.Content...)
	}

	return []*yaml.Node{root}
}

func objectIdentifierSelection(objectNode *yaml.Node, typeDef *config.TypeDefinition) (string, protocol.Range) {
	if objectNode == nil || typeDef == nil {
		return "", protocol.Range{}
	}
	if _, valueNode := mappingEntry(objectNode, typeDef.Identifier.Field); valueNode != nil && valueNode.Kind == yaml.ScalarNode {
		return valueNode.Value, nodeRange(valueNode)
	}
	return "", structuralNodeRange(objectNode)
}

func documentSymbolDetail(typeName string, obj *data.Object) string {
	detail := typeName
	if label := objectPrimaryText(obj); label != "" && label != obj.ID {
		detail = fmt.Sprintf("%s - %s", typeName, label)
	}
	return detail
}

func documentFieldDetail(fieldDef *config.FieldDefinition) string {
	if fieldDef == nil {
		return ""
	}
	typeLabel := fieldDef.Type
	if fieldDef.IsReference() {
		typeLabel = fieldDef.ReferenceLabel()
	}
	if fieldDef.Repeated {
		typeLabel = "[]" + typeLabel
	}
	return typeLabel
}

func documentFieldKind(fieldDef *config.FieldDefinition, valueNode *yaml.Node) protocol.SymbolKind {
	switch {
	case fieldDef != nil && fieldDef.IsReference():
		return protocol.SymbolKindField
	case fieldDef != nil && len(fieldDef.Enum) > 0:
		return protocol.SymbolKindEnumMember
	case valueNode != nil && valueNode.Kind == yaml.SequenceNode:
		return protocol.SymbolKindArray
	case valueNode != nil && valueNode.Kind == yaml.MappingNode:
		return protocol.SymbolKindObject
	default:
		return protocol.SymbolKindField
	}
}

func referenceFieldNames(typeDef *config.TypeDefinition, targetTypes map[string]struct{}) []string {
	if typeDef == nil {
		return nil
	}

	var fieldNames []string
	for _, fieldName := range typeDef.FieldOrder {
		fieldDef := typeDef.Fields[fieldName]
		if fieldDef == nil || !fieldDef.IsReference() {
			continue
		}
		if referenceFieldMatchesTypes(fieldDef, targetTypes) {
			fieldNames = append(fieldNames, fieldName)
		}
	}
	return fieldNames
}

func referenceFieldMatchesTypes(fieldDef *config.FieldDefinition, targetTypes map[string]struct{}) bool {
	if fieldDef == nil || len(targetTypes) == 0 {
		return false
	}
	for _, typeName := range fieldDef.ReferenceTypes {
		if _, ok := targetTypes[typeName]; ok {
			return true
		}
	}
	return false
}

func referenceNodeLocations(path string, valueNode *yaml.Node, want string) []protocol.Location {
	if valueNode == nil || want == "" {
		return nil
	}

	var locations []protocol.Location
	switch valueNode.Kind {
	case yaml.ScalarNode:
		if valueNode.Value == want {
			locations = append(locations, protocol.Location{
				URI:   protocol.DocumentURI(uri.File(path)),
				Range: nodeRange(valueNode),
			})
		}
	case yaml.SequenceNode:
		for _, child := range valueNode.Content {
			if child != nil && child.Kind == yaml.ScalarNode && child.Value == want {
				locations = append(locations, protocol.Location{
					URI:   protocol.DocumentURI(uri.File(path)),
					Range: nodeRange(child),
				})
			}
		}
	}
	return locations
}

func objectNodeForObject(doc *yaml.Node, typeDef *config.TypeDefinition, obj *data.Object) *yaml.Node {
	if doc == nil || typeDef == nil || obj == nil {
		return nil
	}

	root := documentRoot(doc)
	if root == nil {
		return nil
	}

	if _, itemsNode := mappingEntry(root, "items"); itemsNode != nil && itemsNode.Kind == yaml.SequenceNode {
		for _, itemNode := range itemsNode.Content {
			if itemMatchesIdentifier(itemNode, typeDef.Identifier.Field, obj.ID) {
				return itemNode
			}
		}
		return nil
	}

	return root
}

func workspaceObjectForNode(root *workspace.RootRuntime, typeName, path, id string, index int) *data.Object {
	if root == nil || root.Workspace == nil {
		return nil
	}

	matches := workspaceObjectsForFile(root, typeName, path, id)
	if len(matches) > 0 {
		return matches[0]
	}

	objects := root.Workspace.Objects(typeName)
	var byFile []*data.Object
	for _, obj := range objects {
		if obj.File == path {
			byFile = append(byFile, obj)
		}
	}
	if index >= 0 && index < len(byFile) {
		return byFile[index]
	}
	return nil
}

func workspaceObjectsForFile(root *workspace.RootRuntime, typeName, path, id string) []*data.Object {
	if root == nil || root.Workspace == nil {
		return nil
	}

	var matches []*data.Object
	for _, obj := range root.Workspace.Objects(typeName) {
		if obj.File != path {
			continue
		}
		if id != "" && obj.ID != id {
			continue
		}
		matches = append(matches, obj)
	}
	return matches
}

func structuralRangeIncludingKey(keyNode, valueNode *yaml.Node) protocol.Range {
	if keyNode == nil {
		return structuralNodeRange(valueNode)
	}
	return unionRange(nodeRange(keyNode), structuralNodeRange(valueNode))
}

func structuralNodeRange(node *yaml.Node) protocol.Range {
	if node == nil {
		return protocol.Range{}
	}

	switch node.Kind {
	case yaml.ScalarNode:
		return nodeRange(node)
	case yaml.SequenceNode:
		current := nodeRange(node)
		for _, child := range node.Content {
			current = unionRange(current, structuralNodeRange(child))
		}
		return current
	case yaml.MappingNode:
		current := nodeRange(node)
		for idx := 0; idx+1 < len(node.Content); idx += 2 {
			current = unionRange(current, structuralRangeIncludingKey(node.Content[idx], node.Content[idx+1]))
		}
		return current
	default:
		return nodeRange(node)
	}
}

func unionRange(left, right protocol.Range) protocol.Range {
	if left == (protocol.Range{}) {
		return right
	}
	if right == (protocol.Range{}) {
		return left
	}

	result := left
	if comparePositions(right.Start, result.Start) < 0 {
		result.Start = right.Start
	}
	if comparePositions(right.End, result.End) > 0 {
		result.End = right.End
	}
	return result
}

func comparePositions(left, right protocol.Position) int {
	if left.Line != right.Line {
		if left.Line < right.Line {
			return -1
		}
		return 1
	}
	if left.Character < right.Character {
		return -1
	}
	if left.Character > right.Character {
		return 1
	}
	return 0
}

func compareRanges(left, right protocol.Range) int {
	if diff := comparePositions(left.Start, right.Start); diff != 0 {
		return diff
	}
	return comparePositions(left.End, right.End)
}

func sortUniqueLocations(locations []protocol.Location) []protocol.Location {
	if len(locations) == 0 {
		return nil
	}

	sort.Slice(locations, func(i, j int) bool {
		if locations[i].URI != locations[j].URI {
			return locations[i].URI < locations[j].URI
		}
		return compareRanges(locations[i].Range, locations[j].Range) < 0
	})

	result := make([]protocol.Location, 0, len(locations))
	var last string
	for _, location := range locations {
		key := fmt.Sprintf("%s:%d:%d:%d:%d", location.URI, location.Range.Start.Line, location.Range.Start.Character, location.Range.End.Line, location.Range.End.Character)
		if key == last {
			continue
		}
		last = key
		result = append(result, location)
	}
	return result
}

func workspaceSymbolMatches(query, typeName string, obj *data.Object) bool {
	if query == "" {
		return true
	}

	candidates := []string{
		strings.ToLower(typeName),
	}
	if obj != nil {
		candidates = append(candidates,
			strings.ToLower(obj.ID),
			strings.ToLower(objectPrimaryText(obj)),
		)
		for _, key := range []string{"name", "title", "label", "email"} {
			if value, ok := obj.Fields[key]; ok {
				candidates = append(candidates, strings.ToLower(fmt.Sprint(value)))
			}
		}
	}

	for _, candidate := range candidates {
		if candidate != "" && strings.Contains(candidate, query) {
			return true
		}
	}
	return false
}

func workspaceSymbolName(obj *data.Object) string {
	if obj == nil || obj.ID == "" {
		return ""
	}
	return obj.ID
}

func workspaceSymbolContainer(root *workspace.RootRuntime, obj *data.Object) string {
	if root == nil || root.Index == nil || obj == nil {
		return ""
	}

	container := fmt.Sprintf("%s/%s", filepath.Base(root.Index.Root), obj.Type)
	if primary := objectPrimaryText(obj); primary != "" && primary != obj.ID {
		container = fmt.Sprintf("%s %s", container, primary)
	}
	return container
}

func objectPrimaryText(obj *data.Object) string {
	if obj == nil {
		return ""
	}
	for _, key := range []string{"title", "name", "label", "email"} {
		if value, ok := obj.Fields[key]; ok {
			text := strings.TrimSpace(fmt.Sprint(value))
			if text != "" {
				return text
			}
		}
	}
	return ""
}

func fallbackObjectName(typeName string, index int) string {
	return fmt.Sprintf("%s %d", typeName, index+1)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func sortedRootPaths(roots map[string]*workspace.RootRuntime) []string {
	paths := make([]string, 0, len(roots))
	for path := range roots {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
