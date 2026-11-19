package ui

import (
	"fmt"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui/editor"

	myui "github.com/egoist/mygo/ui"
)

// Editors holds one text editor per tool, plus the timestamp, PDF, and
// translation fields. JSON and Base64 each keep a single buffer, and
// operations replace that text. Translation keeps plain text boxes.
type Editors struct {
	JSON      *editor.State
	Base64    *editor.State
	Clock     *ClockState
	PDF       *PDFState
	Translate *TranslateState
	window    *mygo.Window
}

// NewEditors prepares the editors the main window keeps for its lifetime.
func NewEditors() *Editors {
	return &Editors{
		JSON:      newCodeEditor("JSON", "JSON 内容", "JSON 原始内容..."),
		Base64:    newCodeEditor("", "Base64 内容", "要编码或解码的文本..."),
		Clock:     &ClockState{},
		PDF:       newPDFState(),
		Translate: newTranslateState(),
	}
}

// UseWindow is the window background work publishes into.
// A nil window runs the work on the caller, which tests do.
func (e *Editors) UseWindow(w *mygo.Window) {
	if e != nil {
		e.window = w
	}
}

func newCodeEditor(language, label, placeholder string) *editor.State {
	st := &editor.State{
		Buffer:      editor.NewBuffer(""),
		Placeholder: placeholder,
		Label:       label,
		SoftWrap:    true,
	}
	st.Buffer.SetLanguage(editor.LanguageNamed(language))
	return st
}

// codeEditor fills the remaining workspace with one editable buffer.
func codeEditor(c *myui.Context, id string, st *editor.State) {
	t := c.Theme()
	lines := 0
	if st != nil && st.Buffer != nil {
		lines = st.Buffer.LineCount()
	}
	myui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Children(func() {
		myui.Text(c, fmt.Sprintf("内容：%d 行", lines)).FontSize(12).TextColor(t.TextMuted).Padding(8, 12, 4).Shrink(0)
		editor.View(c.Key(id), st).Grow(1).MinWidth(0).MinHeight(0)
	})
}
