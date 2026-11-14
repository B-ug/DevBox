package ui

import (
	"fmt"
	"time"

	"devbox/internal/model"
	jsontool "devbox/internal/tools/json"

	myui "github.com/egoist/mygo/ui"
)

func jsonWorkspace(a *model.App, ed *TextEditors, c *myui.Context) {
	myui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Children(func() {
		myui.Row(c).Shrink(0).Gap(8).Padding(8, 12).Children(func() {
			if myui.Button(c, "格式化").Clicked() {
				formatJSON(a, ed)
			}
			if myui.Button(c, "压缩").Clicked() {
				compactJSON(a, ed)
			}
			if myui.Button(c, "去除转义").Clicked() {
				unescapeJSON(a, ed)
			}
			if myui.Button(c, "校验").Clicked() {
				validateJSON(a, ed)
			}
			if myui.Button(c, "清空").Clicked() {
				clearJSON(a, ed)
			}
			if myui.Button(c, "复制").Clicked() {
				copyJSON(a, ed, c)
			}
		})
		editorSplit(c, "json", ed)
	})
}

func formatJSON(a *model.App, ed *TextEditors) {
	start := time.Now()
	output, err := jsontool.Format(sourceText(ed))
	finish(a, start, err, "已格式化")
	if err != nil {
		return
	}
	ed.Output.Buffer.Set(output)
}

func compactJSON(a *model.App, ed *TextEditors) {
	start := time.Now()
	output, err := jsontool.Compact(sourceText(ed))
	finish(a, start, err, "已压缩")
	if err != nil {
		return
	}
	ed.Output.Buffer.Set(output)
}

func unescapeJSON(a *model.App, ed *TextEditors) {
	start := time.Now()
	output, err := jsontool.Unescape(sourceText(ed))
	finish(a, start, err, "已去除转义")
	if err != nil {
		return
	}
	ed.Output.Buffer.Set(output)
}

func validateJSON(a *model.App, ed *TextEditors) {
	start := time.Now()
	err := jsontool.Validate(sourceText(ed))
	finish(a, start, err, "校验通过")
}

// sourceText is the text an operation transforms. Output is that text once
// a result is there and the input has not been edited since; otherwise the
// input is used, including when the output is still empty.
func sourceText(ed *TextEditors) string {
	if ed.Output.Buffer.String() != "" && ed.Input.Buffer.Version() == ed.inputVersion {
		return ed.Output.Buffer.String()
	}
	ed.inputVersion = ed.Input.Buffer.Version()
	return ed.Input.Buffer.String()
}

func clearJSON(a *model.App, ed *TextEditors) {
	ed.Input.Buffer.Set("")
	ed.Output.Buffer.Set("")
	a.Status = "已清空"
	a.Error = ""
	a.Elapsed = ""
}

func copyJSON(a *model.App, ed *TextEditors, c *myui.Context) {
	text := ed.Output.Buffer.String()
	if text == "" {
		a.Status = "输出为空"
		a.Error = ""
		a.Elapsed = ""
		return
	}
	c.WriteClipboard(text)
	a.Status = "已复制"
	a.Error = ""
	a.Elapsed = ""
}

func finish(a *model.App, start time.Time, err error, okStatus string) {
	a.Elapsed = fmt.Sprintf("%dms", time.Since(start).Milliseconds())
	if err != nil {
		a.Status = "失败"
		a.Error = err.Error()
		return
	}
	a.Status = okStatus
	a.Error = ""
}
