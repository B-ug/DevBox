package ui

import (
	"strings"
	"testing"

	"devbox/internal/model"
)

func TestJSONFormatCompactCopyAndClear(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	input := `{"name":"DevBox","enabled":true}`
	editors.JSON.Buffer.Set(input)
	tt.Frame()

	if err := tt.Click("复制"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != input || a.Status != "已复制" {
		t.Fatalf("clipboard %q status %s", tt.Clipboard(), a.Status)
	}

	if err := tt.Click("格式化"); err != nil {
		t.Fatal(err)
	}
	formatted := editors.JSON.Buffer.String()
	if !strings.Contains(formatted, "\n    \"name\": \"DevBox\"") {
		t.Fatalf("text %q", formatted)
	}
	if a.Status != "已格式化" || a.Error != "" || !strings.HasSuffix(a.Elapsed, "ms") {
		t.Fatalf("status %s error %s elapsed %s", a.Status, a.Error, a.Elapsed)
	}
	if !tt.HasText("已格式化") || !tt.HasText("耗时："+a.Elapsed) {
		t.Fatalf("texts %q", tt.Texts())
	}

	if err := tt.Click("压缩"); err != nil {
		t.Fatal(err)
	}
	if editors.JSON.Buffer.String() != input || a.Status != "已压缩" {
		t.Fatalf("text %q status %s", editors.JSON.Buffer.String(), a.Status)
	}

	if err := tt.Click("复制"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != input || a.Status != "已复制" {
		t.Fatalf("clipboard %q status %s", tt.Clipboard(), a.Status)
	}

	if err := tt.Click("清空"); err != nil {
		t.Fatal(err)
	}
	if editors.JSON.Buffer.String() != "" || a.Status != "已清空" {
		t.Fatalf("text %q status %s", editors.JSON.Buffer.String(), a.Status)
	}
}

func TestJSONUnescape(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	editors.JSON.Buffer.Set(`{\"name\":\"DevBox\"}`)
	tt.Frame()

	if err := tt.Click("去除转义"); err != nil {
		t.Fatal(err)
	}
	if editors.JSON.Buffer.String() != `{"name":"DevBox"}` || a.Status != "已去除转义" || a.Error != "" {
		t.Fatalf("text %q status %s error %s", editors.JSON.Buffer.String(), a.Status, a.Error)
	}

	editors.JSON.Buffer.Set(`\q`)
	tt.Frame()
	if err := tt.Click("去除转义"); err != nil {
		t.Fatal(err)
	}
	if a.Status != "失败" || !strings.Contains(a.Error, "invalid JSON") || editors.JSON.Buffer.String() != `\q` {
		t.Fatalf("status %s error %s text %q", a.Status, a.Error, editors.JSON.Buffer.String())
	}
}

func TestJSONInvalidAndValidate(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	editors.JSON.Buffer.Set("{")
	tt.Frame()

	if err := tt.Click("格式化"); err != nil {
		t.Fatal(err)
	}
	if a.Status != "失败" || !strings.Contains(a.Error, "invalid JSON") || editors.JSON.Buffer.String() != "{" {
		t.Fatalf("status %s error %s text %q", a.Status, a.Error, editors.JSON.Buffer.String())
	}
	if !tt.HasText("错误：" + a.Error) {
		t.Fatalf("texts %q", tt.Texts())
	}

	editors.JSON.Buffer.Set(`{"ok":true}`)
	tt.Frame()
	if err := tt.Click("校验"); err != nil {
		t.Fatal(err)
	}
	if a.Status != "校验通过" || a.Error != "" || editors.JSON.Buffer.String() != `{"ok":true}` {
		t.Fatalf("status %s error %s text %q", a.Status, a.Error, editors.JSON.Buffer.String())
	}
}

func TestJSONTextSurvivesToolSwitch(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	editors.JSON.Buffer.Set(`{"ok":true}`)
	tt.Frame()
	if err := tt.Click("格式化"); err != nil {
		t.Fatal(err)
	}
	formatted := editors.JSON.Buffer.String()

	if err := tt.Click("Base64"); err != nil {
		t.Fatal(err)
	}
	if a.Status != model.StatusReady || a.Error != "" || tt.HasText("格式化") {
		t.Fatalf("status %s error %s texts %q", a.Status, a.Error, tt.Texts())
	}
	if err := tt.Click("JSON"); err != nil {
		t.Fatal(err)
	}
	if editors.JSON.Buffer.String() != formatted {
		t.Fatalf("text %q", editors.JSON.Buffer.String())
	}
}

func TestJSONCopyEmpty(t *testing.T) {
	a := model.New()
	_, tt := layoutTester(a, 1200, 800)
	tt.SetClipboard("previous")
	if err := tt.Click("复制"); err != nil {
		t.Fatal(err)
	}
	if a.Status != "内容为空" || tt.Clipboard() != "previous" {
		t.Fatalf("status %s clipboard %q", a.Status, tt.Clipboard())
	}
}

func TestJSONTyping(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("JSON 内容"); err != nil {
		t.Fatal(err)
	}
	tt.Type(`{"n":1}`)
	tt.Frame()
	if !strings.Contains(editors.JSON.Buffer.String(), `"n"`) {
		t.Fatalf("text %q", editors.JSON.Buffer.String())
	}
}

func TestJSONUnescapeThenFormat(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	editors.JSON.Buffer.Set(`"{\"name\":\"DevBox\"}"`)
	tt.Frame()

	if err := tt.Click("去除转义"); err != nil {
		t.Fatal(err)
	}
	if editors.JSON.Buffer.String() != `{"name":"DevBox"}` {
		t.Fatalf("text %q", editors.JSON.Buffer.String())
	}
	if err := tt.Click("格式化"); err != nil {
		t.Fatal(err)
	}
	formatted := editors.JSON.Buffer.String()
	if !strings.Contains(formatted, "\n    \"name\": \"DevBox\"") || a.Status != "已格式化" {
		t.Fatalf("text %q status %s", formatted, a.Status)
	}
}

func TestJSONEditedTextIsFormatted(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	editors.JSON.Buffer.Set(`{"a":1}`)
	tt.Frame()
	if err := tt.Click("格式化"); err != nil {
		t.Fatal(err)
	}

	editors.JSON.Buffer.Set(`{"b":2}`)
	tt.Frame()
	if err := tt.Click("格式化"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(editors.JSON.Buffer.String(), "\n    \"b\": 2") {
		t.Fatalf("text %q", editors.JSON.Buffer.String())
	}
}
