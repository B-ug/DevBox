package ui

import (
	"fmt"
	"time"

	"devbox/internal/model"
	jsontool "devbox/internal/tools/json"

	"github.com/egoist/mygo/ui/editor"

	myui "github.com/egoist/mygo/ui"
)

func jsonWorkspace(a *model.App, ed *editor.State, c *myui.Context) {
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
				clearText(a, ed)
			}
			if myui.Button(c, "复制").Clicked() {
				copyText(a, ed, c)
			}
		})
		codeEditor(c, "json", ed)
	})
}

func formatJSON(a *model.App, ed *editor.State) {
	start := time.Now()
	output, err := jsontool.Format(ed.Buffer.String())
	applyText(a, ed, start, output, err, "已格式化")
}

func compactJSON(a *model.App, ed *editor.State) {
	start := time.Now()
	output, err := jsontool.Compact(ed.Buffer.String())
	applyText(a, ed, start, output, err, "已压缩")
}

func unescapeJSON(a *model.App, ed *editor.State) {
	start := time.Now()
	output, err := jsontool.Unescape(ed.Buffer.String())
	applyText(a, ed, start, output, err, "已去除转义")
}

func validateJSON(a *model.App, ed *editor.State) {
	start := time.Now()
	err := jsontool.Validate(ed.Buffer.String())
	finish(a, start, err, "校验通过")
}

// applyText records the outcome and, when the operation succeeded,
// replaces the editor text. A failure leaves that text in place.
func applyText(a *model.App, ed *editor.State, start time.Time, output string, err error, okStatus string) {
	finish(a, start, err, okStatus)
	if err != nil {
		return
	}
	ed.Buffer.Set(output)
}

func clearText(a *model.App, ed *editor.State) {
	ed.Buffer.Set("")
	a.Status = "已清空"
	a.Error = ""
	a.Elapsed = ""
}

func copyText(a *model.App, ed *editor.State, c *myui.Context) {
	text := ed.Buffer.String()
	if text == "" {
		a.Status = "内容为空"
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
