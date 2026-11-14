package ui

import (
	"strings"
	"testing"

	"devbox/internal/model"
)

func TestBase64EncodeDecodeAndCopy(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("Base64"); err != nil {
		t.Fatal(err)
	}
	editors.Base64.Buffer.Set("DevBox")
	tt.Frame()

	if err := tt.Click("编码"); err != nil {
		t.Fatal(err)
	}
	if editors.Base64.Buffer.String() != "RGV2Qm94" || a.Status != "已编码" {
		t.Fatalf("text %q status %s", editors.Base64.Buffer.String(), a.Status)
	}

	if err := tt.Click("解码"); err != nil {
		t.Fatal(err)
	}
	if editors.Base64.Buffer.String() != "DevBox" || a.Status != "已解码" {
		t.Fatalf("text %q status %s", editors.Base64.Buffer.String(), a.Status)
	}

	if err := tt.Click("复制"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "DevBox" || a.Status != "已复制" {
		t.Fatalf("clipboard %q status %s", tt.Clipboard(), a.Status)
	}
}

func TestBase64URLAndInvalid(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("Base64"); err != nil {
		t.Fatal(err)
	}
	input := string([]byte{0xfb, 0xff, 0xbf})
	editors.Base64.Buffer.Set(input)
	tt.Frame()

	if err := tt.Click("URL 编码"); err != nil {
		t.Fatal(err)
	}
	encoded := editors.Base64.Buffer.String()
	if encoded == "" || strings.ContainsAny(encoded, "+/") || a.Status != "URL 已编码" {
		t.Fatalf("text %q status %s", encoded, a.Status)
	}

	if err := tt.Click("URL 解码"); err != nil {
		t.Fatal(err)
	}
	if editors.Base64.Buffer.String() != input || a.Status != "URL 已解码" {
		t.Fatalf("text %q status %s", editors.Base64.Buffer.String(), a.Status)
	}

	editors.Base64.Buffer.Set("!!!!")
	tt.Frame()
	if err := tt.Click("解码"); err != nil {
		t.Fatal(err)
	}
	if a.Status != "失败" || !strings.Contains(a.Error, "invalid base64") || editors.Base64.Buffer.String() != "!!!!" {
		t.Fatalf("status %s error %s text %q", a.Status, a.Error, editors.Base64.Buffer.String())
	}
	if !tt.HasText("错误：" + a.Error) {
		t.Fatalf("texts %q", tt.Texts())
	}
}
