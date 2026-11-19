package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEndpoint(t *testing.T) {
	if Endpoint("  ") != "" {
		t.Fatal("empty base")
	}
	if got := Endpoint("https://api.openai.com/v1/"); got != "https://api.openai.com/v1/chat/completions" {
		t.Fatalf("join %s", got)
	}
	if got := Endpoint("https://example.test/v1/chat/completions"); got != "https://example.test/v1/chat/completions" {
		t.Fatalf("kept %s", got)
	}
	if got := Endpoint("https://example.test/V1/CHAT/COMPLETIONS"); got != "https://example.test/V1/CHAT/COMPLETIONS" {
		t.Fatalf("suffix %s", got)
	}
}

func TestSystemPrompt(t *testing.T) {
	if got := SystemPrompt("译成{{target}}，再说一遍{{target}}", "English"); got != "译成English，再说一遍English" {
		t.Fatalf("replace %q", got)
	}
	if got := SystemPrompt("只翻译", "English"); got != "只翻译\n目标语言：English" {
		t.Fatalf("append %q", got)
	}
	if got := SystemPrompt("  ", "English"); got != "目标语言：English" {
		t.Fatalf("blank %q", got)
	}
	if got := SystemPrompt("只翻译", " "); got != "只翻译" {
		t.Fatalf("no target %q", got)
	}
}

func TestCompleteRejectsMissingInput(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer server.Close()
	settings := Settings{BaseURL: server.URL, APIKey: "secret", Model: "demo", Target: "English"}

	_, err := Complete(context.Background(), settings, " \n ")
	if err == nil || err.Error() != "请输入要翻译的文本" || calls != 0 {
		t.Fatalf("text err %v calls %d", err, calls)
	}
	_, err = Complete(context.Background(), Settings{BaseURL: server.URL, Model: "demo"}, "hello")
	if err == nil || err.Error() != "请先填写 API Key" || calls != 0 {
		t.Fatalf("key err %v calls %d", err, calls)
	}
	_, err = Complete(context.Background(), Settings{BaseURL: server.URL, APIKey: "secret"}, "hello")
	if err == nil || err.Error() != "请先填写模型名" {
		t.Fatalf("model err %v", err)
	}
	_, err = Complete(context.Background(), Settings{APIKey: "secret", Model: "demo"}, "hello")
	if err == nil || err.Error() != "请先填写 Base URL" {
		t.Fatalf("base err %v", err)
	}
}

func TestComplete(t *testing.T) {
	const key = "sk-test"
	var gotBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Errorf("request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer "+key {
			t.Errorf("authorization %q", r.Header.Get("Authorization"))
		}
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"hello"}}]}`)
	}))
	defer server.Close()

	out, err := Complete(context.Background(), Settings{
		BaseURL: server.URL + "/v1/",
		APIKey:  " " + key + " ",
		Model:   " demo ",
		Prompt:  "译成{{target}}",
		Target:  "English",
	}, "你好")
	if err != nil || out != "hello" {
		t.Fatalf("out %q err %v", out, err)
	}
	var sent chatRequest
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatal(err)
	}
	if sent.Model != "demo" || len(sent.Messages) != 2 || sent.Messages[0].Content != "译成English" || sent.Messages[1].Content != "你好" {
		t.Fatalf("body %+v", sent)
	}
	if strings.Contains(string(gotBody), key) {
		t.Fatal("key was written into the body")
	}
}

func TestCompleteError(t *testing.T) {
	const key = "sk-secret"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"invalid api key"}}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := Complete(context.Background(), Settings{BaseURL: server.URL, APIKey: key, Model: "demo"}, "你好")
	if err == nil || err.Error() != "invalid api key" || strings.Contains(err.Error(), key) {
		t.Fatalf("err %v", err)
	}
}

func TestCompleteEmptyChoice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"  "}}]}`)
	}))
	defer server.Close()
	_, err := Complete(context.Background(), Settings{BaseURL: server.URL, APIKey: "k", Model: "demo"}, "你好")
	if err == nil || err.Error() != "模型没有返回译文" {
		t.Fatalf("err %v", err)
	}
}

func TestCompleteTooLarge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(bytes.Repeat([]byte("x"), maxResponseBytes+8))
	}))
	defer server.Close()
	_, err := Complete(context.Background(), Settings{BaseURL: server.URL, APIKey: "k", Model: "demo"}, "你好")
	if err == nil || err.Error() != "响应过大" {
		t.Fatalf("err %v", err)
	}
}

func TestCompleteCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Complete(ctx, Settings{BaseURL: server.URL, APIKey: "k", Model: "demo"}, "你好")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err %v", err)
	}
}

func TestCompleteTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := Complete(ctx, Settings{BaseURL: server.URL, APIKey: "k", Model: "demo"}, "你好")
	if time.Since(start) > time.Second {
		t.Fatalf("took %s", time.Since(start))
	}
	if err == nil || err.Error() != "翻译超时" || strings.Contains(err.Error(), "k") {
		t.Fatalf("err %v", err)
	}
}
