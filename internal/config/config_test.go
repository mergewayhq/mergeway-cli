package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	path := filepath.Join("testdata", "valid", "mergeway.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Version != 1 {
		t.Fatalf("expected version 1, got %d", cfg.Version)
	}

	if len(cfg.Types) != 3 {
		t.Fatalf("expected 3 types, got %d", len(cfg.Types))
	}

	post, ok := cfg.Types["Post"]
	if !ok {
		t.Fatalf("expected type 'Post' to be present")
	}

	if post.Description != "Primary blog post content type" {
		t.Fatalf("expected Post description to be present")
	}

	if post.Identifier.Field != "id" {
		t.Fatalf("expected Post identifier field 'id', got %q", post.Identifier.Field)
	}

	title, ok := post.Fields["title"]
	if !ok {
		t.Fatalf("expected Post field 'title' to be present")
	}

	if title.Description != "Human readable title shown in listings" {
		t.Fatalf("expected title description, got %q", title.Description)
	}

	tags, ok := post.Fields["tags"]
	if !ok {
		t.Fatalf("expected Post field 'tags' to be present")
	}

	if !tags.Repeated {
		t.Fatalf("expected tags to be marked repeated")
	}

	if tags.Type != "Tag" {
		t.Fatalf("expected tags to reference Tag, got %q", tags.Type)
	}

	author, ok := post.Fields["author"]
	if !ok {
		t.Fatalf("expected Post field 'author' to be present")
	}

	if author.Type != "User" {
		t.Fatalf("expected author to reference User, got %q", author.Type)
	}

	tag, ok := cfg.Types["Tag"]
	if !ok {
		t.Fatalf("expected type 'Tag' to be present")
	}

	if len(tag.InlineData) != 1 {
		t.Fatalf("expected Tag to have 1 inline data item, got %d", len(tag.InlineData))
	}

	if value := tag.InlineData[0]["label"]; value != "Inline Tag" {
		t.Fatalf("expected inline tag label 'Inline Tag', got %v", value)
	}
}

func TestLoadInheritanceFixture(t *testing.T) {
	path := filepath.Join("testdata", "inheritance", "mergeway.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	dog := cfg.Types["Dog"]
	if dog == nil {
		t.Fatalf("expected Dog type")
	}
	if dog.Extends != "Animal" {
		t.Fatalf("expected Dog to extend Animal, got %q", dog.Extends)
	}
	if !reflect.DeepEqual(dog.FieldOrder, []string{"id", "name", "breed"}) {
		t.Fatalf("expected flattened field order, got %v", dog.FieldOrder)
	}
	if !reflect.DeepEqual(dog.Ancestors, []string{"Animal"}) {
		t.Fatalf("expected Dog ancestors [Animal], got %v", dog.Ancestors)
	}
	if got := cfg.Types["Animal"].Descendants; !reflect.DeepEqual(got, []string{"Dog"}) {
		t.Fatalf("expected Animal descendants [Dog], got %v", got)
	}
}

func TestLoadRepeatedUniqueError(t *testing.T) {
	path := filepath.Join("testdata", "repeated_unique", "mergeway.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for repeated unique field")
	}

	if got := err.Error(); got == "" || !strings.Contains(got, "cannot declare unique") {
		t.Fatalf("expected unique/repeated error, got %q", got)
	}
}

func TestLoadRejectsObjectFieldWithoutProperties(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Project:
    identifier: id
    data:
      - id: project-1
        contacts:
          - email: owner@example.com
    fields:
      id: string
      contacts:
        type: object
        repeated: true
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for object field without properties")
	}
	if got := err.Error(); !strings.Contains(got, "type object must define properties") {
		t.Fatalf("expected object/properties error, got %q", got)
	}
}

func TestLoadRejectsNestedFieldsForObjectField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Project:
    identifier: id
    data:
      - id: project-1
        contacts:
          - email: owner@example.com
    fields:
      id: string
      contacts:
        type: object
        repeated: true
        fields:
          email:
            type: string
            required: true
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for nested fields on object field")
	}
	if got := err.Error(); !strings.Contains(got, "use properties for object fields") {
		t.Fatalf("expected nested fields error, got %q", got)
	}
}

func TestLoadInvalidIdentifier(t *testing.T) {
	path := filepath.Join("testdata", "invalid_identifier", "mergeway.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for invalid identifier")
	}

	if got := err.Error(); got == "" || !strings.Contains(got, "invalid type identifier") {
		t.Fatalf("expected invalid identifier error, got %q", got)
	}
}

func TestLoadFieldPathSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Page:
    identifier: slug
    include:
      - data/pages/*.yaml
    fields:
      slug: string
      section:
        type: string
        source:
          path_segment: 1
      filename:
        type: string
        source:
          path_segment_rev: 0
      relative_path:
        type: string
        source:
          path: true
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	page := cfg.Types["Page"]
	if page == nil {
		t.Fatalf("expected Page type")
	}
	if page.Fields["section"].Source == nil || page.Fields["section"].Source.PathSegment == nil || *page.Fields["section"].Source.PathSegment != 1 {
		t.Fatalf("expected section field path_segment source, got %#v", page.Fields["section"].Source)
	}
	if page.Fields["filename"].Source == nil || page.Fields["filename"].Source.PathSegmentRev == nil || *page.Fields["filename"].Source.PathSegmentRev != 0 {
		t.Fatalf("expected filename field reverse path source, got %#v", page.Fields["filename"].Source)
	}
	if page.Fields["relative_path"].Source == nil || !page.Fields["relative_path"].Source.Path {
		t.Fatalf("expected relative_path field full path source, got %#v", page.Fields["relative_path"].Source)
	}
}

func TestLoadRejectsInlineDataWithFieldPathSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Page:
    identifier: slug
    data:
      - slug: guide-install
    fields:
      slug: string
      section:
        type: string
        source:
          path_segment: 1
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for inline data with field path source")
	}
	if got := err.Error(); !strings.Contains(got, "cannot use field \"section\" source with inline data") {
		t.Fatalf("expected inline-data/path-source error, got %q", got)
	}
}

func TestLoadUnknownReference(t *testing.T) {
	path := filepath.Join("testdata", "unknown_reference", "mergeway.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for unknown reference type")
	}

	if got := err.Error(); got == "" || !strings.Contains(got, "references unknown type") {
		t.Fatalf("expected unknown type error, got %q", got)
	}
}

func TestLoadReferenceUnion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  User:
    identifier: id
    data:
      - id: user-1
    fields:
      id: string
  Team:
    identifier: id
    data:
      - id: team-1
    fields:
      id: string
  Activity:
    identifier: id
    data:
      - id: activity-1
        owner: user-1
    fields:
      id: string
      owner:
        type: User | Team
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	owner := cfg.Types["Activity"].Fields["owner"]
	if owner == nil {
		t.Fatalf("expected owner field")
		return
	}
	ownerType := owner.Type
	if ownerType != "User | Team" {
		t.Fatalf("expected raw union type to be preserved, got %q", ownerType)
	}
	refTypes := owner.ReferenceTypes
	if !reflect.DeepEqual(refTypes, []string{"User", "Team"}) {
		t.Fatalf("expected parsed reference targets, got %v", refTypes)
	}
	if !owner.IsReference() || !owner.HasReferenceUnion() {
		t.Fatalf("expected owner to be treated as a reference union")
	}
}

func TestLoadRejectsInvalidReferenceUnion(t *testing.T) {
	tests := []struct {
		name        string
		fieldType   string
		wantMessage string
	}{
		{name: "trailing separator", fieldType: "User | ", wantMessage: "invalid reference union"},
		{name: "duplicate member", fieldType: "User | User", wantMessage: "duplicate member"},
		{name: "primitive union", fieldType: "string | number", wantMessage: "unions are only supported for references"},
		{name: "unknown type", fieldType: "User | Missing", wantMessage: "references unknown type"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "mergeway.yaml")
			content := []byte(`mergeway:
  version: 1

entities:
  User:
    identifier: id
    data:
      - id: user-1
    fields:
      id: string
  Activity:
    identifier: id
    data:
      - id: activity-1
        owner: user-1
    fields:
      id: string
      owner:
        type: ` + tc.fieldType + `
`)
			if err := os.WriteFile(path, content, 0o644); err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			_, err := Load(path)
			if err == nil {
				t.Fatalf("expected error")
			}
			if got := err.Error(); !strings.Contains(got, tc.wantMessage) {
				t.Fatalf("expected error containing %q, got %q", tc.wantMessage, got)
			}
		})
	}
}

func TestLoadShorthandFieldDefinitions(t *testing.T) {
	path := filepath.Join("testdata", "shorthand", "mergeway.yaml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	person, ok := cfg.Types["Person"]
	if !ok {
		t.Fatalf("expected type 'Person' to be present")
	}

	idField, ok := person.Fields["id"]
	if !ok {
		t.Fatalf("expected field 'id'")
	}
	if idField.Type != "string" {
		t.Fatalf("expected id field type 'string', got %q", idField.Type)
	}
	if idField.Required {
		t.Fatalf("expected shorthand id field to default to optional")
	}

	ageField, ok := person.Fields["age"]
	if !ok {
		t.Fatalf("expected field 'age'")
	}
	if ageField.Type != "integer" {
		t.Fatalf("expected age field type 'integer', got %q", ageField.Type)
	}
	if ageField.Required {
		t.Fatalf("expected shorthand age field to default to optional")
	}

	nameField, ok := person.Fields["name"]
	if !ok {
		t.Fatalf("expected field 'name'")
	}
	if !nameField.Required {
		t.Fatalf("expected explicit mapping to preserve required=true")
	}
}

func TestLoadIncludeWithJSONPath(t *testing.T) {
	path := filepath.Join("testdata", "jsonpath", "mergeway.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	item, ok := cfg.Types["Item"]
	if !ok {
		t.Fatalf("expected type 'Item' to be present")
	}

	if len(item.Include) != 1 {
		t.Fatalf("expected one include, got %d", len(item.Include))
	}

	include := item.Include[0]
	if include.Path != "data/items.json" {
		t.Fatalf("expected include path 'data/items.json', got %q", include.Path)
	}
	if include.Selector != "$.items[*]" {
		t.Fatalf("expected selector '$.items[*]', got %q", include.Selector)
	}
}

func TestLoadAllowsPathIdentifier(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Note:
    identifier: $path
    include:
      - data/notes/*.yaml
    fields:
      title:
        type: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	note, ok := cfg.Types["Note"]
	if !ok {
		t.Fatalf("expected type 'Note' to be present")
	}
	if note.Identifier.Field != PathIdentifierField {
		t.Fatalf("expected path identifier, got %q", note.Identifier.Field)
	}
	if !note.Identifier.IsPath() {
		t.Fatalf("expected identifier to report path mode")
	}
}

func TestLoadRejectsPathIdentifierWithInlineData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Note:
    identifier: $path
    include:
      - data/notes/*.yaml
    fields:
      title:
        type: string
    data:
      - title: Inline
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for inline data with path identifier")
	}

	if got := err.Error(); !strings.Contains(got, "cannot use identifier \"$path\" with inline data") {
		t.Fatalf("expected path identifier inline-data error, got %q", got)
	}
}

func TestLoadMissingMergewayBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`entities:
  Foo:
    identifier: id
    fields:
      id: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for missing mergeway block")
	}

	if got := err.Error(); !strings.Contains(got, "mergeway block is required") {
		t.Fatalf("expected mergeway block error, got %q", got)
	}
}

func TestConfigStringAndSources(t *testing.T) {
	cfg := &Config{
		Version: 1,
		Types: map[string]*TypeDefinition{
			"User": {Source: "types/User.yaml"},
			"Post": {Source: "types/Post.yaml"},
		},
	}
	if got := cfg.String(); !strings.Contains(got, "version:1") || !strings.Contains(got, "types:2") {
		t.Fatalf("unexpected string output: %s", got)
	}
	sources := cfg.Sources()
	expected := []string{"types/Post.yaml", "types/User.yaml"}
	if !reflect.DeepEqual(sources, expected) {
		t.Fatalf("expected %v, got %v", expected, sources)
	}

	var nilCfg *Config
	if nilCfg.String() != "<nil>" {
		t.Fatalf("expected <nil> string, got %s", nilCfg.String())
	}
	if nilCfg.Sources() != nil {
		t.Fatalf("expected nil sources slice")
	}
}

func TestLoadStoresExtends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Animal:
    identifier: id
    data:
      - id: animal-1
    fields:
      id: string
  Dog:
    extends: Animal
    data:
      - id: dog-1
        breed: collie
    fields:
      breed: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	dog := cfg.Types["Dog"]
	if dog == nil {
		t.Fatalf("expected Dog type")
	}
	if dog.Extends != "Animal" {
		t.Fatalf("expected Dog to extend Animal, got %q", dog.Extends)
	}
	if dog.Identifier.Field != "id" {
		t.Fatalf("expected Dog to inherit identifier field id, got %q", dog.Identifier.Field)
	}
	if _, ok := dog.Fields["id"]; !ok {
		t.Fatalf("expected Dog to inherit field id")
	}
}

func TestConfigHierarchyHelpers(t *testing.T) {
	cfg := &Config{
		Types: map[string]*TypeDefinition{
			"Animal":     {Name: "Animal"},
			"Dog":        {Name: "Dog", Extends: "Animal"},
			"WorkingDog": {Name: "WorkingDog", Extends: "Dog"},
			"Cat":        {Name: "Cat", Extends: "Animal"},
		},
	}

	if !cfg.IsA("Animal", "Animal") {
		t.Fatalf("expected Animal to be assignable to itself")
	}
	if !cfg.IsA("Dog", "Animal") {
		t.Fatalf("expected Dog to inherit from Animal")
	}
	if !cfg.IsA("WorkingDog", "Animal") {
		t.Fatalf("expected WorkingDog to inherit from Animal transitively")
	}
	if cfg.IsA("Animal", "Dog") {
		t.Fatalf("did not expect Animal to inherit from Dog")
	}
	if cfg.IsA("Missing", "Animal") {
		t.Fatalf("did not expect unknown child type to match")
	}

	if got := cfg.DescendantTypes("Animal"); !reflect.DeepEqual(got, []string{"Cat", "Dog", "WorkingDog"}) {
		t.Fatalf("unexpected Animal descendants: %v", got)
	}
	if got := cfg.AssignableTypes("Dog"); !reflect.DeepEqual(got, []string{"Dog", "WorkingDog"}) {
		t.Fatalf("unexpected Dog assignable types: %v", got)
	}
	if got := cfg.AssignableTypes("Missing"); got != nil {
		t.Fatalf("expected nil assignable types for unknown type, got %v", got)
	}
}

func TestLoadFlattensInheritedFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  User:
    identifier: id
    data:
      - id: user-1
    fields:
      id: string
  Content:
    identifier: id
    fields:
      id: string
      owner:
        type: User
        required: true
  Post:
    extends: Content
    data:
      - id: post-1
        owner: user-1
        body: hello
    fields:
      body: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	post := cfg.Types["Post"]
	if post == nil {
		t.Fatalf("expected Post type")
	}
	if !reflect.DeepEqual(post.FieldOrder, []string{"id", "owner", "body"}) {
		t.Fatalf("unexpected field order: %v", post.FieldOrder)
	}
	if post.Identifier.Field != "id" {
		t.Fatalf("expected inherited identifier id, got %q", post.Identifier.Field)
	}
	if owner := post.Fields["owner"]; owner == nil {
		t.Fatalf("expected inherited owner field")
	} else if !reflect.DeepEqual(owner.ReferenceTypes, []string{"User"}) {
		t.Fatalf("expected inherited owner reference types, got %v", owner.ReferenceTypes)
	}
	if !reflect.DeepEqual(post.Ancestors, []string{"Content"}) {
		t.Fatalf("unexpected ancestors: %v", post.Ancestors)
	}
	contentType := cfg.Types["Content"]
	if contentType == nil {
		t.Fatalf("expected Content type")
	}
	if !reflect.DeepEqual(contentType.Descendants, []string{"Post"}) {
		t.Fatalf("unexpected descendants: %v", contentType.Descendants)
	}
}

func TestLoadRejectsUnknownParent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Dog:
    extends: Animal
    identifier: id
    data:
      - id: dog-1
    fields:
      id: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for unknown parent")
	}
	if got := err.Error(); !strings.Contains(got, "extends unknown type") {
		t.Fatalf("expected unknown parent error, got %q", got)
	}
}

func TestLoadRejectsDirectInheritanceCycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Animal:
    extends: Animal
    identifier: id
    fields:
      id: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected cycle error")
	}
	if got := err.Error(); !strings.Contains(got, "cyclic inheritance detected") {
		t.Fatalf("expected cycle error, got %q", got)
	}
}

func TestLoadRejectsMultiHopInheritanceCycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Animal:
    extends: Dog
    identifier: id
    fields:
      id: string
  Dog:
    extends: Mammal
    identifier: id
    fields:
      breed: string
  Mammal:
    extends: Animal
    identifier: id
    fields:
      warm_blooded: boolean
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected cycle error")
	}
	if got := err.Error(); !strings.Contains(got, "cyclic inheritance detected") {
		t.Fatalf("expected cycle error, got %q", got)
	}
}

func TestLoadRejectsInheritedFieldRedefinition(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Animal:
    identifier: id
    fields:
      id: string
      name: string
  Dog:
    extends: Animal
    data:
      - id: dog-1
        name: spot
    fields:
      name: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected inherited field redefinition error")
	}
	if got := err.Error(); !strings.Contains(got, "cannot redefine inherited field") {
		t.Fatalf("expected inherited field error, got %q", got)
	}
}

func TestLoadRejectsInheritedIdentifierOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Animal:
    identifier: id
    fields:
      id: string
  Dog:
    extends: Animal
    identifier: slug
    data:
      - id: dog-1
        slug: dog-1
    fields:
      slug: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected inherited identifier override error")
	}
	if got := err.Error(); !strings.Contains(got, "cannot override inherited identifier") {
		t.Fatalf("expected identifier override error, got %q", got)
	}
}

func TestLoadAllowsSchemaOnlyBaseEntity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  Animal:
    identifier: id
    fields:
      id: string
  Mammal:
    extends: Animal
    fields:
      warm_blooded: boolean
  Dog:
    extends: Mammal
    data:
      - id: dog-1
        warm_blooded: true
        breed: collie
    fields:
      breed: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got := cfg.Types["Animal"].Descendants; !reflect.DeepEqual(got, []string{"Dog", "Mammal"}) {
		t.Fatalf("unexpected Animal descendants: %v", got)
	}
	if got := cfg.Types["Mammal"].Descendants; !reflect.DeepEqual(got, []string{"Dog"}) {
		t.Fatalf("unexpected Mammal descendants: %v", got)
	}
}

func TestLoadRejectsJSONSchemaInheritance(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 1

entities:
  CatalogItem:
    identifier: id
    fields:
      id: string
  Product:
    extends: CatalogItem
    identifier: id
    json_schema: schemas/product.json
    include:
      - data/products/*.yaml
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected json_schema inheritance error")
	}
	if got := err.Error(); !strings.Contains(got, "cannot use json_schema with inheritance") {
		t.Fatalf("expected json_schema inheritance error, got %q", got)
	}
}

func TestLoadUnsupportedConfigVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mergeway.yaml")
	content := []byte(`mergeway:
  version: 2

entities:
  Foo:
    identifier: id
    fields:
      id: string
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for unsupported version")
	}

	if got := err.Error(); !strings.Contains(got, "mergeway.version must be") {
		t.Fatalf("expected unsupported version error, got %q", got)
	}
}

func TestLoadJSONSchemaEntity(t *testing.T) {
	path := filepath.Join("testdata", "jsonschema", "mergeway.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	product, ok := cfg.Types["Product"]
	if !ok {
		t.Fatalf("expected type 'Product' to be present")
	}

	if product.JSONSchema != "schemas/product.json" {
		t.Fatalf("expected Product json_schema to be stored, got %q", product.JSONSchema)
	}

	status, ok := product.Fields["status"]
	if !ok {
		t.Fatalf("expected Product field 'status'")
	}
	if status.Type != "enum" {
		t.Fatalf("expected status to map to enum, got %q", status.Type)
	}
	if len(status.Enum) != 3 {
		t.Fatalf("expected status enum values, got %v", status.Enum)
	}
	if !status.Required {
		t.Fatalf("expected status to be required")
	}

	tags, ok := product.Fields["tags"]
	if !ok {
		t.Fatalf("expected Product field 'tags'")
	}
	if !tags.Repeated || tags.Type != "string" {
		t.Fatalf("expected tags to become repeated string field, got repeated=%v type=%q", tags.Repeated, tags.Type)
	}

	owner, ok := product.Fields["owner"]
	if !ok {
		t.Fatalf("expected Product field 'owner'")
	}
	if owner.Type != "User" {
		t.Fatalf("expected owner to resolve x-reference-type, got %q", owner.Type)
	}

	attributes, ok := product.Fields["attributes"]
	if !ok {
		t.Fatalf("expected Product field 'attributes'")
	}
	if attributes.Type != "object" {
		t.Fatalf("expected attributes to be object, got %q", attributes.Type)
	}
	sku, ok := attributes.Properties["sku"]
	if !ok {
		t.Fatalf("expected attributes.sku property")
	}
	if !sku.Required {
		t.Fatalf("expected attributes.sku to be required")
	}
	metrics, ok := attributes.Properties["metrics"]
	if !ok {
		t.Fatalf("expected attributes.metrics property")
	}
	if metrics.Type != "object" {
		t.Fatalf("expected attributes.metrics to be object, got %q", metrics.Type)
	}
	weight, ok := metrics.Properties["weight"]
	if !ok {
		t.Fatalf("expected attributes.metrics.weight property")
	}
	if weight.Type != "number" {
		t.Fatalf("expected attributes.metrics.weight number type, got %q", weight.Type)
	}
	if weight.Description != "Weight in kilograms" {
		t.Fatalf("expected weight description to be preserved")
	}
}

func TestLoadJSONSchemaConflict(t *testing.T) {
	path := filepath.Join("testdata", "jsonschema_conflict", "mergeway.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for json_schema and fields defined together")
	}
	if got := err.Error(); !strings.Contains(got, "cannot define both fields and json_schema") {
		t.Fatalf("expected json_schema conflict error, got %q", got)
	}
}

func TestLoadJSONSchemaMissingDefinition(t *testing.T) {
	path := filepath.Join("testdata", "jsonschema_missing", "mergeway.yaml")
	_, err := Load(path)
	if err == nil {
		t.Fatalf("expected error for missing fields/json_schema")
	}
	if got := err.Error(); !strings.Contains(got, "must define fields or json_schema") {
		t.Fatalf("expected missing schema error, got %q", got)
	}
}

func TestLoadRejectsJSONSchemaReferenceUnion(t *testing.T) {
	root := t.TempDir()
	schemaDir := filepath.Join(root, "schemas")
	if err := os.MkdirAll(schemaDir, 0o755); err != nil {
		t.Fatalf("mkdir schema dir: %v", err)
	}

	cfgPath := filepath.Join(root, "mergeway.yaml")
	cfgContent := []byte(`mergeway:
  version: 1

entities:
  User:
    identifier: id
    data:
      - id: user-1
    fields:
      id: string
  Team:
    identifier: id
    data:
      - id: team-1
    fields:
      id: string
  Activity:
    identifier: id
    include:
      - data/activities/*.yaml
    json_schema: schemas/activity.json
`)
	if err := os.WriteFile(cfgPath, cfgContent, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	schemaContent := []byte(`{
  "type": "object",
  "properties": {
    "id": { "type": "string" },
    "owner": {
      "type": "string",
      "x-reference-type": "User | Team"
    }
  },
  "required": ["id", "owner"]
}`)
	if err := os.WriteFile(filepath.Join(schemaDir, "activity.json"), schemaContent, 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); !strings.Contains(got, "reference unions are only supported in fields definitions") {
		t.Fatalf("expected json schema union rejection, got %q", got)
	}
}
