package scalar

import (
	"encoding/json"
	"math"
	"testing"
)

func TestAsString(t *testing.T) {
	cases := []struct {
		name     string
		input    any
		expected string
		ok       bool
	}{
		{name: "string", input: "abc", expected: "abc", ok: true},
		{name: "int", input: 12, expected: "12", ok: true},
		{name: "uint", input: uint(7), expected: "7", ok: true},
		{name: "float", input: 3.14, expected: "3.14", ok: true},
		{name: "jsonNumber", input: json.Number("99"), expected: "99", ok: true},
		{name: "stringer", input: testStringer("value"), expected: "value", ok: true},
		{name: "nan", input: math.NaN(), expected: "", ok: false},
		{name: "empty", input: "", expected: "", ok: false},
		{name: "unsupported", input: []int{1}, expected: "", ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := AsString(tc.input)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}
			if !ok {
				return
			}
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

type testStringer string

func (t testStringer) String() string {
	return string(t)
}
