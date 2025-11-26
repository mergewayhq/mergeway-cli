package scalar_test

import (
	"testing"

	"github.com/mergewayhq/mergeway-cli/pkg/scalar"
)

func TestAsStringReExport(t *testing.T) {
	got, ok := scalar.AsString(42)
	if !ok || got != "42" {
		t.Fatalf("expected numeric conversion, got %q ok=%v", got, ok)
	}
}
