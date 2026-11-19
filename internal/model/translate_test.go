package model

import "testing"

func TestTranslateDefaults(t *testing.T) {
	a := New()
	if a.Translate.BaseURL != TranslateBaseURL || a.Translate.Model != TranslateModel || a.Translate.Prompt != TranslatePrompt || a.Translate.Target != TranslateTarget || a.Translate.APIKey != "" {
		t.Fatalf("%+v", a.Translate)
	}

	a.UseTranslate(Translate{APIKey: "sk-test", Prompt: "custom", Target: "nope"})
	if a.Translate.APIKey != "sk-test" || a.Translate.Prompt != "custom" || a.Translate.Target != TranslateTarget {
		t.Fatalf("%+v", a.Translate)
	}
	if a.Translate.BaseURL != TranslateBaseURL || a.Translate.Model != TranslateModel {
		t.Fatalf("defaults %+v", a.Translate)
	}

	a.UseTranslate(Translate{BaseURL: " https://example.test/v1 ", Model: "demo", Target: "English"})
	if a.Translate.BaseURL != " https://example.test/v1 " || a.Translate.Model != "demo" || a.Translate.Target != "English" || a.Translate.Prompt != TranslatePrompt {
		t.Fatalf("%+v", a.Translate)
	}
}

func TestLookupTranslate(t *testing.T) {
	tool, ok := Lookup(ToolTranslate)
	if !ok || tool.Name != "翻译" || tool.Summary != "调用模型翻译文本" {
		t.Fatalf("%+v %v", tool, ok)
	}
}
