package scalar

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// AsString converts scalar identifier values into their string representation while
// enforcing that the resulting string is non-empty. Supported inputs include
// strings, numbers, and json.Number values produced by the JSON decoder.
func AsString(value any) (string, bool) {
	switch v := value.(type) {
	case string:
		if v == "" {
			return "", false
		}
		return v, true
	case fmt.Stringer:
		str := v.String()
		if str == "" {
			return "", false
		}
		return str, true
	case json.Number:
		str := v.String()
		if str == "" {
			return "", false
		}
		return str, true
	case int:
		return strconv.FormatInt(int64(v), 10), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case uint:
		return strconv.FormatUint(uint64(v), 10), true
	case uint8:
		return strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true
	case uint32:
		return strconv.FormatUint(uint64(v), 10), true
	case uint64:
		return strconv.FormatUint(v, 10), true
	case float32:
		return formatFloat(float64(v), 32)
	case float64:
		return formatFloat(v, 64)
	default:
		return "", false
	}
}

func formatFloat(value float64, bitSize int) (string, bool) {
	if value != value {
		return "", false
	}
	return strconv.FormatFloat(value, 'f', -1, bitSize), true
}
