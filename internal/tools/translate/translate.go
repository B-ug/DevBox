// Package translate calls an OpenAI-compatible chat completions endpoint.
package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	targetPlaceholder = "{{target}}"
	maxResponseBytes  = 2 << 20
)

// Settings is one translation request. The caller fills defaults.
type Settings struct {
	BaseURL string
	APIKey  string
	Model   string
	Prompt  string
	Target  string
}

// client does not follow redirects, so the API key stays on the host in BaseURL.
var client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// Endpoint joins base with the chat completions path. A base that already
// ends in that path is kept. An empty base returns an empty string.
func Endpoint(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return ""
	}
	if strings.HasSuffix(strings.ToLower(base), "/chat/completions") {
		return base
	}
	return base + "/chat/completions"
}

// SystemPrompt turns prompt into the system message. Every {{target}}
// becomes the chosen language. A prompt without that placeholder gets a
// final line naming the language, so the language choice still applies.
func SystemPrompt(prompt, target string) string {
	prompt = strings.TrimSpace(prompt)
	target = strings.TrimSpace(target)
	if strings.Contains(prompt, targetPlaceholder) {
		return strings.ReplaceAll(prompt, targetPlaceholder, target)
	}
	if target == "" {
		return prompt
	}
	if prompt == "" {
		return "目标语言：" + target
	}
	return prompt + "\n目标语言：" + target
}

// Check reports a missing source, key, model, or endpoint before any request.
func Check(settings Settings, text string) error {
	if strings.TrimSpace(text) == "" {
		return errors.New("请输入要翻译的文本")
	}
	if strings.TrimSpace(settings.APIKey) == "" {
		return errors.New("请先填写 API Key")
	}
	if strings.TrimSpace(settings.Model) == "" {
		return errors.New("请先填写模型名")
	}
	if Endpoint(settings.BaseURL) == "" {
		return errors.New("请先填写 Base URL")
	}
	return nil
}

// Complete sends text to the chat completions endpoint and returns the
// first message. A canceled context returns context.Canceled. A deadline
// returns a Chinese timeout error. Failures never include the API key.
func Complete(ctx context.Context, settings Settings, text string) (string, error) {
	if err := Check(settings, text); err != nil {
		return "", err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	system := SystemPrompt(settings.Prompt, settings.Target)
	messages := make([]chatMessage, 0, 2)
	if system != "" {
		messages = append(messages, chatMessage{Role: "system", Content: system})
	}
	messages = append(messages, chatMessage{Role: "user", Content: text})
	body, err := json.Marshal(chatRequest{Model: strings.TrimSpace(settings.Model), Messages: messages})
	if err != nil {
		return "", err
	}
	endpoint := Endpoint(settings.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", errors.New("Base URL 无效")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(settings.APIKey))

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return "", context.Canceled
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return "", errors.New("翻译超时")
		}
		return "", errors.New("翻译请求失败")
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return "", errors.New("无法读取翻译结果")
	}
	if len(payload) > maxResponseBytes {
		return "", errors.New("响应过大")
	}
	var parsed chatResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", errors.New(httpStatus(resp.StatusCode))
		}
		return "", errors.New("无法读取翻译结果")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := strings.TrimSpace(parsed.Error.Message); msg != "" {
			return "", errors.New(msg)
		}
		return "", errors.New(httpStatus(resp.StatusCode))
	}
	if msg := strings.TrimSpace(parsed.Error.Message); msg != "" {
		return "", errors.New(msg)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", errors.New("模型没有返回译文")
	}
	return parsed.Choices[0].Message.Content, nil
}

func httpStatus(code int) string {
	return "翻译失败（HTTP " + strconv.Itoa(code) + "）"
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}
