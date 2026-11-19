package model

import "strings"

const (
	// TranslateBaseURL is the OpenAI chat endpoint prefix.
	TranslateBaseURL = "https://api.openai.com/v1"
	// TranslateModel is the model name used when none is saved.
	TranslateModel = "gpt-4o-mini"
	// TranslateTarget is the language selected when none is saved.
	TranslateTarget = "简体中文"
	// TranslatePrompt is the system prompt used when none is saved.
	// {{target}} is replaced with the chosen language.
	TranslatePrompt = "你是翻译助手。把用户给出的文本翻译成{{target}}。只输出译文，不要解释，不要加引号。保留原文的段落、换行、代码和专有名词。"
)

// Translate is the translation tool's model settings.
type Translate struct {
	BaseURL string
	APIKey  string
	Model   string
	Prompt  string
	Target  string
}

// translateTargets is the language list the translation page offers.
// TranslateTarget is its first entry.
var translateTargets = []string{
	TranslateTarget,
	"繁体中文",
	"English",
	"日本語",
	"한국어",
	"Français",
	"Deutsch",
	"Español",
	"Русский",
}

// TranslateTargets returns the languages the translation page can select.
// The result is the shared list; callers must not modify it.
func TranslateTargets() []string {
	return translateTargets
}

// ApplyDefaults fills an empty endpoint, model, prompt, or unknown
// language. A blank API key is kept blank.
func (t *Translate) ApplyDefaults() {
	if strings.TrimSpace(t.BaseURL) == "" {
		t.BaseURL = TranslateBaseURL
	}
	if strings.TrimSpace(t.Model) == "" {
		t.Model = TranslateModel
	}
	if strings.TrimSpace(t.Prompt) == "" {
		t.Prompt = TranslatePrompt
	}
	if !knownTranslateTarget(t.Target) {
		t.Target = TranslateTarget
	}
}

// UseTranslate stores saved translation settings and fills whatever was
// left blank.
func (a *App) UseTranslate(saved Translate) {
	a.Translate = saved
	a.Translate.ApplyDefaults()
}

func knownTranslateTarget(target string) bool {
	for _, item := range translateTargets {
		if item == target {
			return true
		}
	}
	return false
}
