package validation

import (
	"path/filepath"
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
