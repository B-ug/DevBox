package ui

import (
	"fmt"

	"github.com/egoist/mygo/ui/editor"

	myui "github.com/egoist/mygo/ui"
)

// TextEditors is the shared input and output pair. JSON uses it now;
// Base64 can use another pair with the same layout.
type TextEditors struct {
	Input  *editor.State
	Output *editor.State
	Split  float32
	// inputVersion is the input buffer version last used as the source.
	// While the input is unchanged, the next operation transforms the output.
	inputVersion uint64
}

// Editors holds one text pair per tool that edits text.
type Editors struct {
	JSON   *TextEditors
	Base64 *TextEditors
}

// NewEditors prepares the editors the main window keeps for its lifetime.
func NewEditors() *Editors {
	return &Editors{
		JSON:   newTextEditors("JSON", "JSON 输入", "JSON 输出", "JSON 原始内容...", "格式化结果..."),
		Base64: newTextEditors("", "Base64 输入", "Base64 输出", "要编码或解码的文本...", "结果..."),
	}
}

func newTextEditors(language, inputLabel, outputLabel, inputPlaceholder, outputPlaceholder string) *TextEditors {
	lang := editor.LanguageNamed(language)
	input := &editor.State{
		Buffer:      editor.NewBuffer(""),
		Placeholder: inputPlaceholder,
		Label:       inputLabel,
		SoftWrap:    true,
	}
	output := &editor.State{
		Buffer:      editor.NewBuffer(""),
		Placeholder: outputPlaceholder,
		Label:       outputLabel,
		ReadOnly:    true,
		SoftWrap:    true,
	}
	input.Buffer.SetLanguage(lang)
	output.Buffer.SetLanguage(lang)
	return &TextEditors{Input: input, Output: output, Split: 50}
}

func editorSplit(c *myui.Context, id string, ed *TextEditors) {
	myui.SplitPercent(c.Key(id+"-split"), &ed.Split, func() {
		codePane(c, id+"-input", "输入", ed.Input)
	}, func() {
		codePane(c, id+"-output", "输出", ed.Output)
	}).Grow(1).MinWidth(0).MinHeight(0).SplitReset(50)
}

func codePane(c *myui.Context, key, title string, st *editor.State) {
	t := c.Theme()
	lines := 0
	if st != nil && st.Buffer != nil {
		lines = st.Buffer.LineCount()
	}
	myui.Column(c).Fill().MinWidth(0).MinHeight(0).Children(func() {
		myui.Text(c, fmt.Sprintf("%s：%d 行", title, lines)).FontSize(12).TextColor(t.TextMuted).Padding(8, 12, 4).Shrink(0)
		editor.View(c.Key(key), st).Grow(1).MinWidth(0).MinHeight(0)
	})
}
