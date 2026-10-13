package cli

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func TestPrepareERDData(t *testing.T) {
	cfg := &config.Config{
		Types: map[string]*config.TypeDefinition{
			"User": {
				Name: "User",
				Fields: map[string]*config.FieldDefinition{
					"ID":    {Type: "string"},
					"Name":  {Type: "string"},
					"Group": {Type: "Group"},
				},
				Include: []config.IncludeDefinition{
					{Path: "data/users/*.yaml"},
				},
			},
			"Group": {
				Name: "Group",
				Fields: map[string]*config.FieldDefinition{
					"ID": {Type: "string"},
				},
				Include: []config.IncludeDefinition{
					{Path: "data/groups.yaml"},
				},
			},
		},
	}

	data := prepareERDData(cfg)

	// Verify types
	if len(data.Types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(data.Types))
	}

	userType := data.Types[1] // Sorted alphabetically: Group, User -> User is at index 1? No, G < U. Group at 0, User at 1.
	if data.Types[0].Name != "Group" {
		t.Errorf("Expected Group at index 0, got %s", data.Types[0].Name)
	}
	if data.Types[1].Name != "User" {
		t.Errorf("Expected User at index 1, got %s", data.Types[1].Name)
	}

	userType = data.Types[1]
	if userType.Name != "User" {
		t.Errorf("Expected User type, got %s", userType.Name)
	}

	// Check paths
	if len(userType.Paths) != 1 || userType.Paths[0] != "data/users/*.yaml" {
		t.Errorf("Expected user path 'data/users/*.yaml', got %v", userType.Paths)
	}

	// Verify fields
	if len(userType.Fields) != 3 {
		t.Errorf("Expected 3 fields for User, got %d", len(userType.Fields))
	}

	// Verify edges
	if len(data.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(data.Edges))
	}

	edge := data.Edges[0]
	if edge.Source != "User" || edge.Target != "Group" || edge.Label != "Group" {
		t.Errorf("Incorrect edge: %+v", edge)
	}
}

func TestPrepareERDData_ExtensionsShowOnlyLocalFieldsAndRelations(t *testing.T) {
	cfg := &config.Config{
		Types: map[string]*config.TypeDefinition{
			"Animal": {
				Name: "Animal",
				Fields: map[string]*config.FieldDefinition{
					"ID":     {Type: "string"},
					"Keeper": {Type: "Person"},
				},
			},
			"Dog": {
				Name:    "Dog",
				Extends: "Animal",
				Fields: map[string]*config.FieldDefinition{
					// Normalized child types contain inherited fields as well as local ones.
					"ID":     {Type: "string"},
					"Keeper": {Type: "Person"},
					"Breed":  {Type: "Breed"},
				},
			},
			"Person": {Name: "Person"},
			"Breed":  {Name: "Breed"},
		},
	}

	data := prepareERDData(cfg)
	var dog erdType
	for _, typ := range data.Types {
		if typ.Name == "Dog" {
			dog = typ
			break
		}
	}
	if len(dog.Fields) != 1 || dog.Fields[0].Name != "Breed" {
		t.Fatalf("expected Dog to contain only its local field, got %+v", dog.Fields)
	}

	if len(data.Edges) != 3 {
		t.Fatalf("expected parent relation, child relation, and extension edge, got %+v", data.Edges)
	}
	for _, edge := range data.Edges {
		if edge.Source == "Dog" && edge.Target == "Person" {
			t.Fatalf("inherited relation should not be drawn on child: %+v", edge)
		}
	}

	var output bytes.Buffer
	if err := template.Must(template.New("erd").Parse(erdTemplate)).Execute(&output, data); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"Dog" -> "Animal" [arrowhead="empty"]`) {
		t.Fatalf("expected UML-style extension edge, got %s", output.String())
	}
}

func TestPrepareERDData_NoEdges(t *testing.T) {
	cfg := &config.Config{
		Types: map[string]*config.TypeDefinition{
			"A": {
				Name: "A",
				Fields: map[string]*config.FieldDefinition{
					"F": {Type: "string"},
				},
			},
			"B": {
				Name: "B",
				Fields: map[string]*config.FieldDefinition{
					"F": {Type: "string"},
				},
			},
		},
	}

	data := prepareERDData(cfg)
	if len(data.Edges) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(data.Edges))
	}
}

func TestPrepareERDData_ReferenceUnion(t *testing.T) {
	cfg := &config.Config{
		Types: map[string]*config.TypeDefinition{
			"Activity": {
				Name: "Activity",
				Fields: map[string]*config.FieldDefinition{
					"ID": {Type: "string"},
					"Owner": {
						Type:           "User | Team",
						ReferenceTypes: []string{"User", "Team"},
					},
				},
			},
			"Team": {
				Name: "Team",
				Fields: map[string]*config.FieldDefinition{
					"ID": {Type: "string"},
				},
			},
			"User": {
				Name: "User",
				Fields: map[string]*config.FieldDefinition{
					"ID": {Type: "string"},
				},
			},
		},
	}

	data := prepareERDData(cfg)
	if len(data.Edges) != 2 {
		t.Fatalf("Expected 2 edges, got %d", len(data.Edges))
	}

	if got := data.Edges[0]; got.Source != "Activity" || got.Target != "Team" || got.Label != "Owner" {
		t.Fatalf("Unexpected first edge: %+v", got)
	}
	if got := data.Edges[1]; got.Source != "Activity" || got.Target != "User" || got.Label != "Owner" {
		t.Fatalf("Unexpected second edge: %+v", got)
	}
}
