package ui

import (
	"context"
	"errors"
	"time"

	"devbox/internal/model"
	"devbox/internal/tools/translate"

	"github.com/egoist/mygo"

	myui "github.com/egoist/mygo/ui"
)

const (
	translateTimeout = 60 * time.Second
	translateBoxMin  = 120
	translateBoxMax  = 240
)

// TranslateState is the translation tool's text and the request in flight.
type TranslateState struct {
	Source     string
	Result     string
	Running    bool
	generation int
	cancel     context.CancelFunc
}

func newTranslateState() *TranslateState {
	return &TranslateState{}
}

func translateWorkspace(a *model.App, editors *Editors, c *myui.Context) {
	s := editors.Translate
	t := c.Theme()
	myui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Children(func() {
		myui.Row(c).Shrink(0).Padding(12, 16).Gap(8).AlignItems(myui.Center).Children(func() {
			target := myui.Select(c.Key("translate-target"), &a.Translate.Target, model.TranslateTargets()).Label("目标语言")
			if target.Changed() {
				persist(a)
			}
			label := "翻译"
			if s.Running {
				label = "翻译中…"
			}
			// Label distinguishes the action from the sidebar item of the same name.
			if myui.PrimaryButton(c, label).Label("开始翻译").Disabled(s.Running).Shrink(0).Clicked() {
				startTranslate(a, s, editors.window)
			}
			if myui.Button(c, "复制译文").Shrink(0).Clicked() {
				copyTranslate(a, s, c)
			}
			if myui.Button(c, "清空原文").Shrink(0).Clicked() {
				clearTranslate(a, s)
			}
		})
		myui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
			myui.Column(c).Padding(4, 16, 16).Gap(8).MinWidth(0).Children(func() {
				translateField(c, t, "translate-source", "原文", "要翻译的文本...", &s.Source)
				translateField(c, t, "translate-result", "译文", "翻译结果...", &s.Result)
			})
		})
	})
}

func translateField(c *myui.Context, t *myui.Theme, id, caption, placeholder string, value *string) {
	myui.Text(c, caption).FontSize(13).Bold().TextColor(t.TextMuted).Shrink(0)
	myui.TextArea(c.Key(id), value).Placeholder(placeholder).Label(caption).MinHeight(translateBoxMin).MaxHeight(translateBoxMax).MinWidth(0)
}

func copyTranslate(a *model.App, s *TranslateState, c *myui.Context) {
	text := ""
	if s != nil {
		text = s.Result
	}
	if text == "" {
		a.Status = "译文为空"
		a.Error = ""
		a.Elapsed = ""
		return
	}
	c.WriteClipboard(text)
	a.Status = "已复制"
	a.Error = ""
	a.Elapsed = ""
}

func clearTranslate(a *model.App, s *TranslateState) {
	if s != nil {
		s.Source = ""
	}
	a.Status = "已清空"
	a.Error = ""
	a.Elapsed = ""
}

func startTranslate(a *model.App, s *TranslateState, win *mygo.Window) {
	if s == nil {
		return
	}
	text := s.Source
	settings := translate.Settings{
		BaseURL: a.Translate.BaseURL,
		APIKey:  a.Translate.APIKey,
		Model:   a.Translate.Model,
		Prompt:  a.Translate.Prompt,
		Target:  a.Translate.Target,
	}
	if err := translate.Check(settings, text); err != nil {
		a.Status = "失败"
		a.Error = err.Error()
		a.Elapsed = ""
		return
	}
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.generation++
	gen := s.generation
	ctx, cancel := context.WithTimeout(context.Background(), translateTimeout)
	s.cancel = cancel
	s.Running = true
	a.Status = "正在翻译"
	a.Error = ""
	a.Elapsed = ""
	start := time.Now()
	run := func() {
		out, err := translate.Complete(ctx, settings, text)
		publishTranslate(win, func() {
			s.finish(a, start, out, err, gen)
		})
	}
	if win == nil {
		run()
		return
	}
	go run()
}

func publishTranslate(win *mygo.Window, fn func()) {
	if win == nil {
		fn()
		return
	}
	win.Update(fn)
}

func (s *TranslateState) finish(a *model.App, start time.Time, text string, err error, gen int) {
	if s.generation != gen {
		return
	}
	s.Running = false
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	if err != nil {
		if a.CurrentTool != model.ToolTranslate {
			return
		}
		if errors.Is(err, context.Canceled) {
			a.Status = "已取消"
			a.Error = ""
			a.Elapsed = ""
			return
		}
		finish(a, start, err, "")
		return
	}
	s.Result = text
	if a.CurrentTool != model.ToolTranslate {
		return
	}
	finish(a, start, nil, "已翻译")
}
