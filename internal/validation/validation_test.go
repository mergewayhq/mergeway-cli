package validation

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mergewayhq/mergeway-cli/internal/config"
)

func TestValidateAllPhasesSuccess(t *testing.T) {
	root := fixturePath(t, "valid")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", res.Errors)
	}
}

func TestValidateJSONPathIncludes(t *testing.T) {
	root := fixturePath(t, "jsonpath")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", res.Errors)
	}
}

func TestValidateInheritanceFixture(t *testing.T) {
	root := fixturePath(t, "inheritance")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", res.Errors)
	}
}

func TestValidateFormatError(t *testing.T) {
	root := fixturePath(t, "format_error")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(res.Errors))
	}

	errItem := res.Errors[0]
	if errItem.Phase != PhaseFormat {
		t.Fatalf("expected format error, got %s", errItem.Phase)
	}
}

func TestValidateSchemaError(t *testing.T) {
	root := fixturePath(t, "schema_error")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) < 2 {
		t.Fatalf("expected at least 2 errors, got %d", len(res.Errors))
	}

	phases := collectPhases(res.Errors)
	if _, ok := phases[PhaseSchema]; !ok {
		t.Fatalf("expected schema errors, got phases %v", phases)
	}
}

func TestValidateReferenceError(t *testing.T) {
	root := fixturePath(t, "reference_error")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) == 0 {
		t.Fatalf("expected reference error, got none")
	}

	found := false
	for _, e := range res.Errors {
		if e.Phase == PhaseReferences {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected reference phase errors, got %v", res.Errors)
	}
}

func TestValidateReferenceUnion(t *testing.T) {
	root := writeReferenceUnionRepo(t, "user-1", "team-1", "user-1")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", res.Errors)
	}

	owner := cfg.Types["Activity"].Fields["owner"]
	if owner == nil {
		t.Fatalf("expected owner field")
		return
	}
	refTypes := owner.ReferenceTypes
	if !reflect.DeepEqual(refTypes, []string{"User", "Team"}) {
		t.Fatalf("expected parsed reference types, got %v", refTypes)
	}
}

func TestValidateReferenceUnionMissing(t *testing.T) {
	root := writeReferenceUnionRepo(t, "user-1", "team-1", "missing-owner")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %v", res.Errors)
	}
	if res.Errors[0].Phase != PhaseReferences {
		t.Fatalf("expected reference error, got %s", res.Errors[0].Phase)
	}
	if !strings.Contains(res.Errors[0].Message, "references missing User | Team") {
		t.Fatalf("expected missing union reference error, got %v", res.Errors[0])
	}
}

func TestValidateReferenceUnionAmbiguous(t *testing.T) {
	root := writeReferenceUnionRepo(t, "shared-id", "shared-id", "shared-id")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %v", res.Errors)
	}
	if res.Errors[0].Phase != PhaseReferences {
		t.Fatalf("expected reference error, got %s", res.Errors[0].Phase)
	}
	if !strings.Contains(res.Errors[0].Message, "ambiguous across User | Team") {
		t.Fatalf("expected ambiguous union reference error, got %v", res.Errors[0])
	}
}

func TestValidateInheritanceReferenceAcceptsDescendant(t *testing.T) {
	root := t.TempDir()
	writeValidationFile(t, filepath.Join(root, "mergeway.yaml"), `mergeway:
  version: 1

entities:
  Animal:
    identifier: id
    fields:
      id: string
      name:
        type: string
        required: true
  Dog:
    extends: Animal
    include:
      - data/dogs/*.yaml
    fields:
      breed: string
  Kennel:
    identifier: id
    include:
      - data/kennels/*.yaml
    fields:
      id: string
      resident:
        type: Animal
        required: true
`)
	writeValidationFile(t, filepath.Join(root, "data", "dogs", "dog-1.yaml"), "id: dog-1\nname: Fido\nbreed: collie\n")
	writeValidationFile(t, filepath.Join(root, "data", "kennels", "kennel-1.yaml"), "id: kennel-1\nresident: dog-1\n")

	cfg := loadConfig(t, root)
	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if len(res.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", res.Errors)
	}
}

func TestValidateInheritanceMissingInheritedRequiredField(t *testing.T) {
	root := t.TempDir()
	writeValidationFile(t, filepath.Join(root, "mergeway.yaml"), `mergeway:
  version: 1

entities:
  Animal:
    identifier: id
    fields:
      id: string
      name:
        type: string
        required: true
  Dog:
    extends: Animal
    include:
      - data/dogs/*.yaml
    fields:
      breed: string
`)
	writeValidationFile(t, filepath.Join(root, "data", "dogs", "dog-1.yaml"), "id: dog-1\nbreed: collie\n")

	cfg := loadConfig(t, root)
	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %v", res.Errors)
	}
	if res.Errors[0].Phase != PhaseSchema {
		t.Fatalf("expected schema error, got %s", res.Errors[0].Phase)
	}
	if !strings.Contains(res.Errors[0].Message, "missing required field \"name\"") {
		t.Fatalf("expected inherited required field error, got %v", res.Errors[0])
	}
}

func TestValidateRejectsDuplicateIdentifiersAcrossAssignableHierarchy(t *testing.T) {
	root := t.TempDir()
	writeValidationFile(t, filepath.Join(root, "mergeway.yaml"), `mergeway:
  version: 1

entities:
  Animal:
    identifier: id
    include:
      - data/animals/*.yaml
    fields:
      id: string
      name: string
  Dog:
    extends: Animal
    include:
      - data/dogs/*.yaml
    fields:
      breed: string
`)
	writeValidationFile(t, filepath.Join(root, "data", "animals", "animal-1.yaml"), "id: shared-1\nname: Generic\n")
	writeValidationFile(t, filepath.Join(root, "data", "dogs", "dog-1.yaml"), "id: shared-1\nname: Fido\nbreed: collie\n")

	cfg := loadConfig(t, root)
	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %v", res.Errors)
	}
	if res.Errors[0].Phase != PhaseSchema {
		t.Fatalf("expected schema error, got %s", res.Errors[0].Phase)
	}
	if !strings.Contains(res.Errors[0].Message, "duplicate identifier across assignable hierarchy") {
		t.Fatalf("expected assignable hierarchy duplicate error, got %v", res.Errors[0])
	}
}

func TestValidateReferenceUnionWithDescendant(t *testing.T) {
	root := t.TempDir()
	writeValidationFile(t, filepath.Join(root, "mergeway.yaml"), `mergeway:
  version: 1

entities:
  User:
    identifier: id
    include:
      - data/users/*.yaml
    fields:
      id: string
  Team:
    identifier: id
    fields:
      id: string
  DeliveryTeam:
    extends: Team
    include:
      - data/teams/*.yaml
    fields:
      service: string
  Activity:
    identifier: id
    include:
      - data/activities/*.yaml
    fields:
      id: string
      owner:
        type: User | Team
`)
	writeValidationFile(t, filepath.Join(root, "data", "users", "user-1.yaml"), "id: user-1\n")
	writeValidationFile(t, filepath.Join(root, "data", "teams", "team-1.yaml"), "id: team-1\nservice: delivery\n")
	writeValidationFile(t, filepath.Join(root, "data", "activities", "activity-1.yaml"), "id: activity-1\nowner: team-1\n")

	cfg := loadConfig(t, root)
	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if len(res.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", res.Errors)
	}
}

func TestValidateFailFast(t *testing.T) {
	root := fixturePath(t, "schema_error")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{FailFast: true})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error with fail-fast, got %d", len(res.Errors))
	}
}

func TestValidatePhaseSelection(t *testing.T) {
	root := fixturePath(t, "schema_error")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{Phases: []Phase{PhaseSchema}})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	for _, e := range res.Errors {
		if e.Phase != PhaseSchema {
			t.Fatalf("expected only schema errors, got %v", res.Errors)
		}
	}
}

func TestValidateAllowsNumericIdentifiers(t *testing.T) {
	root := fixturePath(t, "numeric_identifier")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", res.Errors)
	}
}

func TestValidateAllowsPathIdentifiers(t *testing.T) {
	root := fixturePath(t, "path_identifier")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", res.Errors)
	}
}

func TestValidateAllowsDeclaredPathDerivedFields(t *testing.T) {
	root := t.TempDir()
	cfgBody := `mergeway:
  version: 1

entities:
  Page:
    identifier: slug
    include:
      - data/library/*/*.yaml
    fields:
      slug:
        type: string
        required: true
      title:
        type: string
        required: true
      section:
        type: string
        required: true
        source:
          path_segment: 2
      filename:
        type: string
        required: true
        source:
          path_segment_rev: 0
`
	if err := os.WriteFile(filepath.Join(root, "mergeway.yaml"), []byte(cfgBody), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "data", "library", "guides"), 0o755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	body := "slug: guide-install\ntitle: Install Mergeway\n"
	if err := os.WriteFile(filepath.Join(root, "data", "library", "guides", "install.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write data: %v", err)
	}

	cfg := loadConfig(t, root)
	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if len(res.Errors) != 0 {
		t.Fatalf("expected no errors, got %v", res.Errors)
	}
}

func TestValidateRejectsPathIdentifierMultiObjectFiles(t *testing.T) {
	root := fixturePath(t, "path_identifier_multi")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %v", res.Errors)
	}
	if res.Errors[0].Phase != PhaseSchema {
		t.Fatalf("expected schema error, got %s", res.Errors[0].Phase)
	}
	if !strings.Contains(res.Errors[0].Message, "cannot be used with files containing multiple objects") {
		t.Fatalf("expected path identifier multi-object error, got %v", res.Errors[0])
	}
}

func TestValidateFieldConstraints(t *testing.T) {
	root := fixturePath(t, "field_constraints")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if len(res.Errors) < 4 {
		t.Fatalf("expected several schema errors for field constraints, got %v", res.Errors)
	}

	var foundPattern, foundFormat bool
	for _, e := range res.Errors {
		if strings.Contains(e.Message, "pattern") {
			foundPattern = true
		}
		if strings.Contains(e.Message, "format") {
			foundFormat = true
		}
	}
	if !foundPattern || !foundFormat {
		t.Fatalf("expected pattern/format errors, got %v", res.Errors)
	}
}

func TestValidateAllowsDefaultsForMissingFields(t *testing.T) {
	root := fixturePath(t, "defaults_valid")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if len(res.Errors) != 0 {
		t.Fatalf("expected defaults to satisfy required fields, got %v", res.Errors)
	}
}

func TestValidateUniqueComplexFields(t *testing.T) {
	root := fixturePath(t, "unique_structs")
	cfg := loadConfig(t, root)

	res, err := Validate(root, cfg, Options{})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if len(res.Errors) == 0 {
		t.Fatalf("expected uniqueness violation for duplicate attributes, got none")
	}
	found := false
	for _, e := range res.Errors {
		if strings.Contains(e.Message, "must be unique") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected unique field error, got %v", res.Errors)
	}
}

func fixturePath(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("testdata", name)
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}
	return abs
}

func loadConfig(t *testing.T, root string) *config.Config {
	t.Helper()
	cfgPath := filepath.Join(root, "mergeway.yaml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	return cfg
}

func collectPhases(errs []Error) map[Phase]struct{} {
	set := make(map[Phase]struct{})
	for _, e := range errs {
		set[e.Phase] = struct{}{}
	}
	return set
}

func writeReferenceUnionRepo(t *testing.T, userID, teamID, ownerID string) string {
	t.Helper()

	root := t.TempDir()
	cfgContent := []byte(`mergeway:
  version: 1

entities:
  User:
    identifier: id
    include:
      - data/users/*.yaml
    fields:
      id: string
  Team:
    identifier: id
    include:
      - data/teams/*.yaml
    fields:
      id: string
  Activity:
    identifier: id
    include:
      - data/activities/*.yaml
    fields:
      id: string
      owner:
        type: User | Team
`)
	if err := os.WriteFile(filepath.Join(root, "mergeway.yaml"), cfgContent, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	for _, dir := range []string{
		filepath.Join(root, "data", "users"),
		filepath.Join(root, "data", "teams"),
		filepath.Join(root, "data", "activities"),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	if err := os.WriteFile(filepath.Join(root, "data", "users", "user.yaml"), []byte("id: "+userID+"\n"), 0o644); err != nil {
		t.Fatalf("write user: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "teams", "team.yaml"), []byte("id: "+teamID+"\n"), 0o644); err != nil {
		t.Fatalf("write team: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "data", "activities", "activity.yaml"), []byte("id: activity-1\nowner: "+ownerID+"\n"), 0o644); err != nil {
		t.Fatalf("write activity: %v", err)
	}

	return root
}

func writeValidationFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
