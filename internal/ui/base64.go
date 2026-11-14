package ui

import (
	"time"

	"devbox/internal/model"
	b64 "devbox/internal/tools/base64"

	myui "github.com/egoist/mygo/ui"
)

func base64Workspace(a *model.App, ed *TextEditors, c *myui.Context) {
	myui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Children(func() {
		myui.Row(c).Shrink(0).Gap(8).Padding(8, 12).Children(func() {
			if myui.Button(c, "编码").Clicked() {
				encodeBase64(a, ed, false)
			}
			if myui.Button(c, "解码").Clicked() {
				decodeBase64(a, ed, false)
			}
			if myui.Button(c, "URL 编码").Clicked() {
				encodeBase64(a, ed, true)
			}
			if myui.Button(c, "URL 解码").Clicked() {
				decodeBase64(a, ed, true)
			}
			if myui.Button(c, "清空").Clicked() {
				clearJSON(a, ed)
			}
			if myui.Button(c, "复制").Clicked() {
				copyJSON(a, ed, c)
			}
		})
		editorSplit(c, "base64", ed)
	})
}

func encodeBase64(a *model.App, ed *TextEditors, url bool) {
	start := time.Now()
	text := sourceText(ed)
	output := b64.Encode(text)
	status := "已编码"
	if url {
		output = b64.EncodeURL(text)
		status = "URL 已编码"
	}
	finish(a, start, nil, status)
	ed.Output.Buffer.Set(output)
}

func decodeBase64(a *model.App, ed *TextEditors, url bool) {
	start := time.Now()
	decode := b64.Decode
	status := "已解码"
	if url {
		decode = b64.DecodeURL
		status = "URL 已解码"
	}
	output, err := decode(sourceText(ed))
	finish(a, start, err, status)
	if err != nil {
		return
	}
	ed.Output.Buffer.Set(output)
}
