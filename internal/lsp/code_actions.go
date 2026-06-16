package lsp

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"gopkg.in/yaml.v3"
)

func (s *Server) codeActions(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	if !requestsQuickFixes(params.Context.Only) {
		return nil, nil
	}

	path := params.TextDocument.URI.Filename()
	var actions []protocol.CodeAction
	for _, diagnostic := range params.Context.Diagnostics {
		actions = append(actions, s.quickFixesForDiagnostic(path, diagnostic)...)
	}
	return uniqueCodeActions(actions), nil
}

func requestsQuickFixes(only []protocol.CodeActionKind) bool {
	if len(only) == 0 {
		return true
	}
	for _, kind := range only {
		if kind == protocol.QuickFix {
			return true
		}
	}
	return false
}

func (s *Server) quickFixesForDiagnostic(path string, diagnostic protocol.Diagnostic) []protocol.CodeAction {
	analysis, err := s.analyzePosition(path, diagnostic.Range.Start)
	if err != nil || analysis == nil || analysis.data == nil || analysis.data.typeDef == nil || analysis.data.objectNode == nil {
		return nil
	}

	switch {
	case strings.Contains(diagnostic.Message, "missing required field"):
		return s.quickFixesForMissingRequiredField(path, analysis, diagnostic)
	case strings.Contains(diagnostic.Message, "must be one of"):
		return s.quickFixesForEnum(path, analysis, diagnostic)
	case strings.Contains(diagnostic.Message, "references missing"):
		return s.quickFixesForReference(path, analysis, diagnostic)
	default:
		return nil
	}
}

func (s *Server) quickFixesForMissingRequiredField(path string, analysis *documentAnalysis, diagnostic protocol.Diagnostic) []protocol.CodeAction {
	fieldName := quotedFieldName(diagnostic.Message)
	if fieldName == "" {
		return nil
	}

	fieldDef := analysis.data.typeDef.Fields[fieldName]
	if fieldDef == nil {
		return nil
	}

	var actions []protocol.CodeAction
	if renameEdit, oldField, ok := renameFieldEdit(analysis.data.objectNode, analysis.data.typeDef, fieldName); ok {
		actions = append(actions, quickFixAction(
			fmt.Sprintf(`Rename field "%s" to "%s"`, oldField, fieldName),
			path,
			renameEdit,
			diagnostic,
			true,
		))
	}

	if insertEdit, ok := missingFieldInsertEdit(path, analysis, fieldName, fieldDef); ok {
		actions = append(actions, quickFixAction(
			fmt.Sprintf(`Insert missing field "%s"`, fieldName),
			path,
			insertEdit,
			diagnostic,
			false,
		))
	}

	return actions
}

func (s *Server) quickFixesForEnum(path string, analysis *documentAnalysis, diagnostic protocol.Diagnostic) []protocol.CodeAction {
	fieldName := quotedFieldName(diagnostic.Message)
	if fieldName == "" {
		return nil
	}

	fieldDef := analysis.data.typeDef.Fields[fieldName]
	if fieldDef == nil || len(fieldDef.Enum) == 0 {
		return nil
	}

	edit, replacement, ok := enumReplacementEdit(analysis.data.objectNode, fieldName, fieldDef.Enum)
	if !ok {
		return nil
	}

	return []protocol.CodeAction{quickFixAction(
		fmt.Sprintf(`Replace with "%s"`, replacement),
		path,
		edit,
		diagnostic,
		true,
	)}
}

func (s *Server) quickFixesForReference(path string, analysis *documentAnalysis, diagnostic protocol.Diagnostic) []protocol.CodeAction {
	fieldName := quotedFieldName(diagnostic.Message)
	if fieldName == "" {
		return nil
	}

	fieldDef := analysis.data.typeDef.Fields[fieldName]
	if fieldDef == nil || !fieldDef.IsReference() {
		return nil
	}

	candidates := referenceIDs(analysis.root, fieldDef)
	if len(candidates) == 0 {
		return nil
	}

	edit, replacement, ok := referenceReplacementEdit(analysis.data.objectNode, fieldName, candidates)
	if !ok {
		return nil
	}

	return []protocol.CodeAction{quickFixAction(
		fmt.Sprintf(`Replace with "%s"`, replacement),
		path,
		edit,
		diagnostic,
		true,
	)}
}

func quickFixAction(title, path string, edit protocol.TextEdit, diagnostic protocol.Diagnostic, preferred bool) protocol.CodeAction {
	return protocol.CodeAction{
		Title:       title,
		Kind:        protocol.QuickFix,
		Diagnostics: []protocol.Diagnostic{diagnostic},
		IsPreferred: preferred,
		Edit: &protocol.WorkspaceEdit{
			Changes: map[protocol.DocumentURI][]protocol.TextEdit{
				protocol.DocumentURI(uri.File(path)): {edit},
			},
		},
	}
}

func renameFieldEdit(objectNode *yaml.Node, typeDef *config.TypeDefinition, missingField string) (protocol.TextEdit, string, bool) {
	if objectNode == nil || objectNode.Kind != yaml.MappingNode || typeDef == nil {
		return protocol.TextEdit{}, "", false
	}

	unknownKeys := make(map[string]*yaml.Node)
	for idx := 0; idx+1 < len(objectNode.Content); idx += 2 {
		keyNode := objectNode.Content[idx]
		if typeDef.Fields[keyNode.Value] == nil {
			unknownKeys[keyNode.Value] = keyNode
		}
	}

	bestName, ok := bestStringCandidate(missingField, mapKeys(unknownKeys))
	if !ok {
		return protocol.TextEdit{}, "", false
	}

	return protocol.TextEdit{
		Range:   nodeRange(unknownKeys[bestName]),
		NewText: missingField,
	}, bestName, true
}

func missingFieldInsertEdit(path string, analysis *documentAnalysis, fieldName string, fieldDef *config.FieldDefinition) (protocol.TextEdit, bool) {
	if !isYAMLPath(path) {
		return protocol.TextEdit{}, false
	}

	insertRange, prefix, ok := insertionRangeForField(analysis.content, analysis.data.objectNode, analysis.data.typeDef, fieldName)
	if !ok {
		return protocol.TextEdit{}, false
	}

	indent := objectFieldIndent(analysis.data.objectNode)
	fieldLine := strings.Repeat(" ", indent) + fieldName + ": " + fieldPlaceholder(fieldDef)
	return protocol.TextEdit{
		Range:   insertRange,
		NewText: prefix + fieldLine + "\n",
	}, true
}

func enumReplacementEdit(objectNode *yaml.Node, fieldName string, candidates []string) (protocol.TextEdit, string, bool) {
	rng, current, ok := replacementTarget(objectNode, fieldName, func(value string) bool {
		return !containsString(candidates, value)
	})
	if !ok {
		return protocol.TextEdit{}, "", false
	}

	replacement, ok := bestStringCandidate(current, candidates)
	if !ok {
		return protocol.TextEdit{}, "", false
	}

	return protocol.TextEdit{Range: rng, NewText: replacement}, replacement, true
}

func referenceReplacementEdit(objectNode *yaml.Node, fieldName string, candidates []string) (protocol.TextEdit, string, bool) {
	rng, current, ok := replacementTarget(objectNode, fieldName, func(value string) bool {
		return !containsString(candidates, value)
	})
	if !ok {
		return protocol.TextEdit{}, "", false
	}

	replacement, ok := bestStringCandidate(current, candidates)
	if !ok {
		return protocol.TextEdit{}, "", false
	}

	return protocol.TextEdit{Range: rng, NewText: replacement}, replacement, true
}

func replacementTarget(objectNode *yaml.Node, fieldName string, invalid func(string) bool) (protocol.Range, string, bool) {
	if objectNode == nil || objectNode.Kind != yaml.MappingNode {
		return protocol.Range{}, "", false
	}

	_, valueNode := mappingEntry(objectNode, fieldName)
	if valueNode == nil {
		return protocol.Range{}, "", false
	}

	switch valueNode.Kind {
	case yaml.ScalarNode:
		if valueNode.Value == "" || !invalid(valueNode.Value) {
			return protocol.Range{}, "", false
		}
		return nodeRange(valueNode), valueNode.Value, true
	case yaml.SequenceNode:
		var current string
		var rng protocol.Range
		count := 0
		for _, child := range valueNode.Content {
			if child == nil || child.Kind != yaml.ScalarNode {
				continue
			}
			if !invalid(child.Value) {
				continue
			}
			current = child.Value
			rng = nodeRange(child)
			count++
		}
		if count == 1 {
			return rng, current, true
		}
	}

	return protocol.Range{}, "", false
}

func insertionRangeForField(content []byte, objectNode *yaml.Node, typeDef *config.TypeDefinition, fieldName string) (protocol.Range, string, bool) {
	lines := strings.Split(string(content), "\n")
	present := make(map[string]*yaml.Node)
	for idx := 0; idx+1 < len(objectNode.Content); idx += 2 {
		present[objectNode.Content[idx].Value] = objectNode.Content[idx]
	}

	fieldOrderIndex := fieldOrderIndex(typeDef, fieldName)
	if fieldOrderIndex >= 0 {
		for _, nextField := range typeDef.FieldOrder[fieldOrderIndex+1:] {
			keyNode := present[nextField]
			if keyNode == nil {
				continue
			}
			line := keyNode.Line - 1
			return protocol.Range{
				Start: protocol.Position{Line: uint32(line), Character: 0},
				End:   protocol.Position{Line: uint32(line), Character: 0},
			}, "", true
		}
	}

	endLine := int(structuralNodeRange(objectNode).End.Line) + 1
	if endLine < len(lines) {
		return protocol.Range{
			Start: protocol.Position{Line: uint32(endLine), Character: 0},
			End:   protocol.Position{Line: uint32(endLine), Character: 0},
		}, "", true
	}

	lastLine := maxInt(0, len(lines)-1)
	lastText := ""
	if len(lines) > 0 {
		lastText = lines[lastLine]
	}

	return protocol.Range{
		Start: protocol.Position{Line: uint32(lastLine), Character: uint32(len(lastText))},
		End:   protocol.Position{Line: uint32(lastLine), Character: uint32(len(lastText))},
	}, "\n", true
}

func fieldOrderIndex(typeDef *config.TypeDefinition, fieldName string) int {
	if typeDef == nil {
		return -1
	}
	for index, candidate := range typeDef.FieldOrder {
		if candidate == fieldName {
			return index
		}
	}
	return -1
}

func fieldPlaceholder(fieldDef *config.FieldDefinition) string {
	if fieldDef == nil {
		return `""`
	}

	switch {
	case fieldDef.Repeated:
		return "[]"
	case fieldDef.IsReference():
		return `""`
	case len(fieldDef.Enum) > 0:
		return fieldDef.Enum[0]
	}

	switch fieldDef.Type {
	case "string":
		return `""`
	case "integer", "number":
		return "0"
	case "boolean":
		return "false"
	case "object":
		return "{}"
	default:
		return `""`
	}
}

func bestStringCandidate(want string, candidates []string) (string, bool) {
	want = strings.TrimSpace(want)
	if want == "" || len(candidates) == 0 {
		return "", false
	}

	best := ""
	bestDistance := -1
	secondBest := -1
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		distance := damerauLevenshtein(strings.ToLower(want), strings.ToLower(candidate))
		if bestDistance == -1 || distance < bestDistance {
			secondBest = bestDistance
			bestDistance = distance
			best = candidate
			continue
		}
		if secondBest == -1 || distance < secondBest {
			secondBest = distance
		}
	}

	if best == "" || bestDistance > maxCandidateDistance(want) {
		return "", false
	}
	if secondBest != -1 && secondBest == bestDistance {
		return "", false
	}
	return best, true
}

func maxCandidateDistance(value string) int {
	switch {
	case len(value) >= 10:
		return 3
	case len(value) >= 5:
		return 2
	default:
		return 1
	}
}

func damerauLevenshtein(left, right string) int {
	a := []rune(left)
	b := []rune(right)
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = minInt(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				matrix[i][j] = minInt(matrix[i][j], matrix[i-2][j-2]+1)
			}
		}
	}

	return matrix[len(a)][len(b)]
}

func uniqueCodeActions(actions []protocol.CodeAction) []protocol.CodeAction {
	if len(actions) == 0 {
		return nil
	}

	sort.Slice(actions, func(i, j int) bool {
		return actions[i].Title < actions[j].Title
	})

	var result []protocol.CodeAction
	seen := make(map[string]struct{}, len(actions))
	for _, action := range actions {
		key := action.Title
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, action)
	}
	return result
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func mapKeys(values map[string]*yaml.Node) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func minInt(values ...int) int {
	best := values[0]
	for _, value := range values[1:] {
		if value < best {
			best = value
		}
	}
	return best
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func isYAMLPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
