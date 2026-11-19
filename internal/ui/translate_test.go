package ui

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"devbox/internal/config"
	"devbox/internal/model"
)

func TestTranslatePage(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 1800)
	if err := tt.Click("翻译"); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{
		"调用模型翻译文本",
		"目标语言",
		"简体中文",
		"开始翻译",
		"复制译文",
		"清空原文",
		"原文",
		"译文",
	} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}
	if tt.HasText("原文：") || tt.HasText("译文：") || tt.HasText("行") {
		t.Fatalf("line count still shown: %q", tt.Texts())
	}
	for _, text := range []string{"Base URL", "API Key", "Prompt", "恢复默认 Prompt"} {
		if tt.HasText(text) {
			t.Fatalf("settings field %q still on the translate page: %q", text, tt.Texts())
		}
	}
	if editors.Translate == nil {
		t.Fatal("missing translation state")
	}
	if a.Translate.BaseURL != model.TranslateBaseURL || a.Translate.APIKey != "" {
		t.Fatalf("%+v", a.Translate)
	}
}

func TestTranslateEmptyDoesNotRequest(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer server.Close()

	a := model.New()
	a.Translate.BaseURL = server.URL
	a.Translate.APIKey = "secret"
	_, tt := layoutTester(a, 1200, 1800)
	if err := tt.Click("翻译"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("开始翻译"); err != nil {
		t.Fatal(err)
	}
	if calls != 0 || a.Status != "失败" || a.Error != "请输入要翻译的文本" {
		t.Fatalf("calls %d status %s error %s", calls, a.Status, a.Error)
	}
	if !tt.HasText("错误：请输入要翻译的文本") {
		t.Fatalf("texts %q", tt.Texts())
	}
}

func TestTranslateWritesResult(t *testing.T) {
	const key = "sk-test"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+key {
			http.Error(w, `{"error":{"message":"bad key"}}`, http.StatusUnauthorized)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var sent struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		if json.Unmarshal(body, &sent) != nil || len(sent.Messages) != 2 || sent.Messages[1].Content != "你好" {
			t.Errorf("body %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"hello"}}]}`)
	}))
	defer server.Close()

	a := model.New()
	a.Translate.BaseURL = server.URL
	a.Translate.APIKey = key
	a.Translate.Model = "demo"
	a.Translate.Target = "English"
	editors, tt := layoutTester(a, 1200, 1800)
	if err := tt.Click("翻译"); err != nil {
		t.Fatal(err)
	}
	editors.Translate.Source = "你好"
	tt.Frame()
	if err := tt.Click("开始翻译"); err != nil {
		t.Fatal(err)
	}
	if editors.Translate.Running || editors.Translate.Result != "hello" || a.Status != "已翻译" || a.Error != "" {
		t.Fatalf("running %v result %q status %s error %s", editors.Translate.Running, editors.Translate.Result, a.Status, a.Error)
	}
	if !tt.HasText("已翻译") || !strings.HasSuffix(a.Elapsed, "ms") {
		t.Fatalf("status %s elapsed %s texts %q", a.Status, a.Elapsed, tt.Texts())
	}

	if err := tt.Click("复制译文"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "hello" || a.Status != "已复制" {
		t.Fatalf("clipboard %q status %s", tt.Clipboard(), a.Status)
	}

	if err := tt.Click("JSON"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("翻译"); err != nil {
		t.Fatal(err)
	}
	if editors.Translate.Result != "hello" || editors.Translate.Source != "你好" {
		t.Fatalf("source %q result %q", editors.Translate.Source, editors.Translate.Result)
	}
}

func TestTranslatePersistsPrompt(t *testing.T) {
	dir := t.TempDir()
	a := model.New()
	a.ConfigDir = dir
	a.Translate.APIKey = "sk-test"
	a.Translate.Prompt = "custom"
	_, tt := layoutTester(a, 1200, 1400)
	if err := tt.Click("设置"); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"Base URL", "API Key", "模型", "Prompt", "恢复默认 Prompt", "{{target}} 会替换成目标语言。原文作为用户消息发送。"} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}
	if err := tt.Click("恢复默认 Prompt"); err != nil {
		t.Fatal(err)
	}
	if a.Translate.Prompt != model.TranslatePrompt || a.Status != "已恢复默认 Prompt" {
		t.Fatalf("prompt %q status %s error %s", a.Translate.Prompt, a.Status, a.Error)
	}
	file, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if file.Translate.APIKey != "sk-test" || file.Translate.Prompt != model.TranslatePrompt || file.Translate.Model != model.TranslateModel {
		t.Fatalf("%+v", file.Translate)
	}
}

func TestTranslateIgnoresStaleResult(t *testing.T) {
	a := model.New()
	a.CurrentTool = model.ToolTranslate
	s := newTranslateState()
	s.generation = 2
	s.Running = true
	s.finish(a, time.Now(), "late", nil, 1)
	if !s.Running || s.Result != "" || a.Status != model.StatusReady {
		t.Fatalf("stale running %v result %q status %s", s.Running, s.Result, a.Status)
	}

	s.generation = 3
	s.Running = true
	a.CurrentTool = model.ToolJSON
	a.Status = model.StatusReady
	s.finish(a, time.Now(), "hello", nil, 3)
	if s.Running || s.Result != "hello" || a.Status != model.StatusReady {
		t.Fatalf("running %v result %q status %s", s.Running, s.Result, a.Status)
	}
}
