package ui

import (
	"time"

	"devbox/internal/model"
	b64 "devbox/internal/tools/base64"

	"github.com/egoist/mygo/ui/editor"

	myui "github.com/egoist/mygo/ui"
)

func base64Workspace(a *model.App, ed *editor.State, c *myui.Context) {
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
				clearText(a, ed)
			}
			if myui.Button(c, "复制").Clicked() {
				copyText(a, ed, c)
			}
		})
		codeEditor(c, "base64", ed)
	})
}

func encodeBase64(a *model.App, ed *editor.State, url bool) {
	start := time.Now()
	text := ed.Buffer.String()
	output := b64.Encode(text)
	status := "已编码"
	if url {
		output = b64.EncodeURL(text)
		status = "URL 已编码"
	}
	applyText(a, ed, start, output, nil, status)
}

func decodeBase64(a *model.App, ed *editor.State, url bool) {
	start := time.Now()
	decode := b64.Decode
	status := "已解码"
	if url {
		decode = b64.DecodeURL
		status = "URL 已解码"
	}
	output, err := decode(ed.Buffer.String())
	applyText(a, ed, start, output, err, status)
}
