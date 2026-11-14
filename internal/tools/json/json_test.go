package json

import (
	"strings"
	"testing"
)

func TestFormatAndCompact(t *testing.T) {
	in := `{"name":"DevBox","enabled":true}`
	formatted, err := Format(in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(formatted, "\n    \"name\": \"DevBox\"") {
		t.Fatalf("formatted %q", formatted)
	}
	compact, err := Compact(formatted)
	if err != nil {
		t.Fatal(err)
	}
	if compact != in {
		t.Fatalf("compact %q", compact)
	}
}

func TestInvalidJSON(t *testing.T) {
	for _, input := range []string{"", "nope", "{"} {
		if _, err := Format(input); err == nil || !strings.Contains(err.Error(), "invalid JSON") {
			t.Fatalf("Format %q error %v", input, err)
		}
		if _, err := Compact(input); err == nil || !strings.Contains(err.Error(), "invalid JSON") {
			t.Fatalf("Compact %q error %v", input, err)
		}
		if err := Validate(input); err == nil {
			t.Fatalf("Validate %q succeeded", input)
		}
	}
	if err := Validate(`{"ok":true}`); err != nil {
		t.Fatal(err)
	}
}

func TestUnescape(t *testing.T) {
	cases := []struct {
		in, out string
	}{
		{`{\"name\":\"DevBox\"}`, `{"name":"DevBox"}`},
		{`"{\"name\":\"DevBox\"}"`, `{"name":"DevBox"}`},
		{`\u4e2d\u6587`, "中文"},
		{`"hello\nworld"`, "hello\nworld"},
		{`\\`, `\`},
	}
	for _, tc := range cases {
		got, err := Unescape(tc.in)
		if err != nil {
			t.Fatalf("Unescape %q: %v", tc.in, err)
		}
		if got != tc.out {
			t.Fatalf("Unescape %q = %q, want %q", tc.in, got, tc.out)
		}
	}
	if _, err := Unescape(`\q`); err == nil || !strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("Unescape invalid error %v", err)
	}
	if _, err := Unescape("  "); err == nil {
		t.Fatal("blank input succeeded")
	}
}
