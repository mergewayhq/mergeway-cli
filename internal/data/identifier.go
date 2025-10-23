package data

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/mergewayhq/mergeway-cli/internal/config"
	"github.com/mergewayhq/mergeway-cli/internal/scalar"
)

func extractIdentifierValue(typeDef *config.TypeDefinition, fields map[string]any) (string, any, error) {
	if typeDef == nil {
		return "", nil, fmt.Errorf("data: type spec is required")
	}
	idField := typeDef.Identifier.Field
	if idField == "" {
		return "", nil, fmt.Errorf("data: identifier field is not configured")
	}

	if fields == nil {
		return "", nil, fmt.Errorf("missing field %q", idField)
	}

	raw, ok := fields[idField]
	if !ok {
		return "", nil, fmt.Errorf("missing field %q", idField)
	}

	str, ok := scalar.AsString(raw)
	if !ok {
		return "", nil, fmt.Errorf("field %q must be a non-empty string or number", idField)
	}

	fieldDef := typeDef.Fields[idField]
	if fieldDef == nil {
		return str, raw, nil
	}

	switch fieldDef.Type {
	case "string", "enum":
		return str, str, nil
	}

	normalized, err := coerceIdentifierValue(fieldDef.Type, idField, raw)
	if err != nil {
		return "", nil, err
	}

	return str, normalized, nil
}

func coerceIdentifierValue(fieldType, fieldName string, raw any) (any, error) {
	switch fieldType {
	case "integer":
		value, ok := toInt64(raw)
		if !ok {
			return nil, fmt.Errorf("field %q must be an integer", fieldName)
		}
		return value, nil
	case "number":
		value, ok := toFloat64(raw)
		if !ok {
			return nil, fmt.Errorf("field %q must be a number", fieldName)
		}
		return value, nil
	default:
		return raw, nil
	}
}

func toInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		if uint64(v) > math.MaxInt64 {
			return 0, false
		}
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		if v > math.MaxInt64 {
			return 0, false
		}
		return int64(v), true
	case float32:
		if math.Trunc(float64(v)) != float64(v) {
			return 0, false
		}
		return int64(v), true
	case float64:
		if math.Trunc(v) != v {
			return 0, false
		}
		return int64(v), true
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i, true
		}
		if f, err := v.Float64(); err == nil && math.Trunc(f) == f {
			return int64(f), true
		}
		return 0, false
	case string:
		if v == "" {
			return 0, false
		}
		iv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return iv, true
	default:
		return 0, false
	}
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case float32:
		return float64(v), true
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, false
		}
		return v, true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, false
		}
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, false
		}
		return f, true
	case string:
		if v == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}
