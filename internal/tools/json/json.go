package json

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"strings"
)

// Format pretty-prints input with four-space indentation.
func Format(input string) (string, error) {
	var output bytes.Buffer
	if err := stdjson.Indent(&output, []byte(input), "", "    "); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	return output.String(), nil
}

// Compact removes insignificant whitespace from input.
func Compact(input string) (string, error) {
	var output bytes.Buffer
	if err := stdjson.Compact(&output, []byte(input)); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	return output.String(), nil
}

// Unescape decodes a JSON string. The input may be a quoted JSON string,
// or the escaped text itself, such as {\"name\":\"DevBox\"}.
func Unescape(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", fmt.Errorf("invalid JSON")
	}
	if out, err := decodeJSONString(s); err == nil {
		return out, nil
	}
	out, err := decodeJSONString(`"` + s + `"`)
	if err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	return out, nil
}

func decodeJSONString(s string) (string, error) {
	var out string
	if err := stdjson.Unmarshal([]byte(s), &out); err != nil {
		return "", err
	}
	return out, nil
}

// Validate reports whether input is JSON.
func Validate(input string) error {
	if stdjson.Valid([]byte(input)) {
		return nil
	}
	return fmt.Errorf("invalid JSON")
}
