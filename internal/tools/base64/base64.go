package base64

import (
	stdbase64 "encoding/base64"
	"fmt"
	"strings"
)

// Encode encodes input with standard Base64.
func Encode(input string) string {
	return stdbase64.StdEncoding.EncodeToString([]byte(input))
}

// Decode decodes standard Base64. Spaces and line breaks are ignored.
// Text without padding is accepted.
func Decode(input string) (string, error) {
	return decode(input, stdbase64.StdEncoding, stdbase64.RawStdEncoding)
}

// EncodeURL encodes input with URL-safe Base64.
func EncodeURL(input string) string {
	return stdbase64.URLEncoding.EncodeToString([]byte(input))
}

// DecodeURL decodes URL-safe Base64. Spaces and line breaks are ignored.
// Text without padding is accepted.
func DecodeURL(input string) (string, error) {
	return decode(input, stdbase64.URLEncoding, stdbase64.RawURLEncoding)
}

func decode(input string, padded, raw *stdbase64.Encoding) (string, error) {
	s := stripSpace(input)
	if s == "" {
		return "", fmt.Errorf("invalid base64")
	}
	b, err := padded.DecodeString(s)
	if err != nil {
		b, err = raw.DecodeString(s)
	}
	if err != nil {
		return "", fmt.Errorf("invalid base64: %w", err)
	}
	return string(b), nil
}

func stripSpace(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, s)
}
