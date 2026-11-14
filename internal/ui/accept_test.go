package ui

import (
	"strings"
	"testing"

	"devbox/internal/model"

	myui "github.com/egoist/mygo/ui"
)

func TestEditorPasteAndCopy(t *testing.T) {
	editors, tt := layoutTester(model.New(), 1200, 800)
	if err := tt.Click("JSON 内容"); err != nil {
		t.Fatal(err)
	}
	tt.SetClipboard(`{"pasted":true}`)
	tt.Key(myui.Cmd, myui.KeyV)
	tt.Frame()
	if !strings.Contains(editors.JSON.Buffer.String(), `"pasted"`) {
		t.Fatalf("pasted text %q", editors.JSON.Buffer.String())
	}
	tt.Key(myui.Cmd, myui.KeyA)
	tt.Key(myui.Cmd, myui.KeyC)
	if !strings.Contains(tt.Clipboard(), `"pasted"`) {
		t.Fatalf("copied %q", tt.Clipboard())
	}
}

func TestLayoutStaysInsideWhenResized(t *testing.T) {
	_, tt := layoutTester(model.New(), 1200, 800)
	for _, size := range [][2]int{{900, 600}, {1400, 900}, {1200, 800}} {
		tt.SetSize(size[0], size[1])
		for _, label := range []string{"DevBox", "JSON", "格式化", "状态：就绪", "设置", "收藏"} {
			r, ok := tt.Find(label)
			if !ok {
				t.Fatalf("%q missing at %dx%d in %q", label, size[0], size[1], tt.Texts())
			}
			if r.X < -1 || r.Y < -1 || r.X+r.W > float32(size[0])+1 || r.Y+r.H > float32(size[1])+1 {
				t.Fatalf("%q at %+v outside %dx%d", label, r, size[0], size[1])
			}
		}
	}
}
