package data

import (
	"encoding/json"
	"math"
	"testing"
)

func TestToFloat64Conversions(t *testing.T) {
	cases := []struct {
		name     string
		input    any
		expected float64
		ok       bool
	}{
		{name: "int", input: 42, expected: 42, ok: true},
		{name: "uint", input: uint32(7), expected: 7, ok: true},
		{name: "string", input: "3.14", expected: 3.14, ok: true},
		{name: "jsonNumber", input: json.Number("9.1"), expected: 9.1, ok: true},
		{name: "nan", input: math.NaN(), ok: false},
		{name: "invalid", input: struct{}{}, ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := toFloat64(tc.input)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}
			if ok && math.Abs(got-tc.expected) > 1e-9 {
				t.Fatalf("expected %f, got %f", tc.expected, got)
			}
		})
	}
}
