package base64

import (
	"strings"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	encoded := Encode("DevBox")
	if encoded != "RGV2Qm94" {
		t.Fatalf("encoded %q", encoded)
	}
	decoded, err := Decode(encoded)
	if err != nil || decoded != "DevBox" {
		t.Fatalf("decoded %q %v", decoded, err)
	}
	decoded, err = Decode("RGV2\nQm94")
	if err != nil || decoded != "DevBox" {
		t.Fatalf("wrapped %q %v", decoded, err)
	}
	decoded, err = Decode("QQ")
	if err != nil || decoded != "A" {
		t.Fatalf("unpadded %q %v", decoded, err)
	}
}

func TestURLEncodeDecode(t *testing.T) {
	input := string([]byte{0xfb, 0xff, 0xbf})
	std := Encode(input)
	url := EncodeURL(input)
	if std == url || !strings.ContainsAny(std, "+/") || !strings.ContainsAny(url, "-_") {
		t.Fatalf("std %q url %q", std, url)
	}
	decoded, err := DecodeURL(url)
	if err != nil || decoded != input {
		t.Fatalf("decoded %q %v", decoded, err)
	}
	if _, err := DecodeURL(std); err == nil {
		t.Fatal("standard text was accepted as URL-safe")
	}
}

func TestInvalidBase64(t *testing.T) {
	for _, input := range []string{"", "   ", "!!!!"} {
		if _, err := Decode(input); err == nil || !strings.Contains(err.Error(), "invalid base64") {
			t.Fatalf("Decode %q error %v", input, err)
		}
		if _, err := DecodeURL(input); err == nil || !strings.Contains(err.Error(), "invalid base64") {
			t.Fatalf("DecodeURL %q error %v", input, err)
		}
	}
}
