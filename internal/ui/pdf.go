package ui

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devbox/internal/model"
	pdftool "devbox/internal/tools/pdf"

	"github.com/egoist/mygo"

	myui "github.com/egoist/mygo/ui"
)

// PDFState is the PDF tool's fields. DPI and Quality are what the sliders bind.
type PDFState struct {
	Source     string
	SourceSize int64
	OutputDir  string
	DPI        float64
	Quality    float64
	Preset     string
	Running    bool
	Page       int
	Pages      int
	Detail     string
	Failure    string
	Output     string
	generation int
}

func newPDFState() *PDFState {
	return &PDFState{DPI: 140, Quality: 80}
}

type pdfPreset struct {
	name    string
	dpi     float64
	quality float64
}

var pdfPresets = []pdfPreset{
	{"高质量", 200, 90},
	{"均衡", 140, 75},
	{"最小体积", 80, 50},
}

func pdfWorkspace(a *model.App, editors *Editors, c *myui.Context) {
	s := editors.PDF
	t := c.Theme()
	myui.Scroll(c).Grow(1).MinHeight(0).Children(func() {
		myui.Column(c).Padding(16, 24).Gap(14).Children(func() {
			myui.Text(c, "缩小内嵌图片并重新编码为 JPEG，文字保持不变").TextColor(t.TextMuted).MinWidth(0)

			zone := myui.Column(c.Key("pdf-drop")).MinHeight(88).Padding(16).Gap(8).Radius(8).AlignItems(myui.Center).Children(func() {
				myui.Text(c, "拖放 PDF 到此处").TextColor(t.TextMuted)
				if myui.Button(c, "选择文件").Disabled(s.Running).Clicked() {
					if path, ok := pickPath(a, "选择 PDF 文件", false, s.Source); ok {
						s.useFile(path)
					}
				}
			})
			border := t.Border
			if zone.FileDragOver() {
				border = t.Accent
			}
			zone.Border(1, border)
			for _, path := range zone.DroppedFiles() {
				if !s.Running && strings.EqualFold(filepath.Ext(path), ".pdf") {
					s.useFile(path)
					break
				}
			}

			if s.Source != "" {
				myui.Text(c, fmt.Sprintf("已选择 %s（%s）", filepath.Base(s.Source), formatBytes(s.SourceSize))).TextColor(t.TextMuted).MinWidth(0)
			}

			myui.Text(c, "输出目录").FontSize(13).Bold()
			myui.Row(c).Gap(8).AlignItems(myui.Center).Children(func() {
				myui.TextInput(c, &s.OutputDir).Placeholder("默认保存到源文件所在目录").Label("输出目录").Grow(1).MinWidth(0).Disabled(s.Running)
				if myui.Button(c, "浏览").Disabled(s.Running).Clicked() {
					if path, ok := pickPath(a, "选择输出目录", true, s.OutputDir); ok {
						s.OutputDir = path
					}
				}
			})

			myui.Text(c, "压缩参数").FontSize(13).Bold()
			myui.Row(c).Gap(8).Children(func() {
				for _, preset := range pdfPresets {
					var button myui.Element
					if s.Preset == preset.name {
						button = myui.PrimaryButton(c.Key(preset.name), preset.name)
					} else {
						button = myui.Button(c.Key(preset.name), preset.name)
					}
					if button.Disabled(s.Running).Clicked() {
						s.DPI = preset.dpi
						s.Quality = preset.quality
						s.Preset = preset.name
					}
				}
			})

			dpi := myui.Slider(c.Key("pdf-dpi"), &s.DPI, 50, 300).Label("分辨率").Disabled(s.Running)
			if dpi.Changed() {
				s.DPI = math.Round(s.DPI)
				s.Preset = s.matchedPreset()
			}
			myui.Text(c, fmt.Sprintf("分辨率 %d", int(math.Round(s.DPI)))).TextColor(t.Accent)
			myui.Text(c, "50 到 300").FontSize(12).TextColor(t.TextMuted)

			quality := myui.Slider(c.Key("pdf-quality"), &s.Quality, 10, 100).Label("图片质量").Disabled(s.Running)
			if quality.Changed() {
				s.Quality = math.Round(s.Quality)
				s.Preset = s.matchedPreset()
			}
			myui.Text(c, fmt.Sprintf("质量 %d", int(math.Round(s.Quality)))).TextColor(t.Accent)
			myui.Text(c, "10 到 100").FontSize(12).TextColor(t.TextMuted)

			if s.Running {
				value := -1.0
				if s.Pages > 0 {
					value = float64(s.Page) / float64(s.Pages)
				}
				myui.Progress(c, value).Label("压缩进度")
				if s.Pages > 0 {
					myui.Text(c, fmt.Sprintf("正在处理 %d / %d 页", s.Page, s.Pages)).TextColor(t.TextMuted)
				} else {
					myui.Text(c, "正在读取 PDF").TextColor(t.TextMuted)
				}
			}

			label := "开始压缩"
			if s.Running {
				label = "压缩中…"
			}
			if myui.PrimaryButton(c, label).Disabled(s.Running).Clicked() {
				startPDF(a, s, editors.window)
			}

			if s.Failure != "" {
				myui.Text(c, s.Failure).TextColor(t.Danger).MinWidth(0)
			}
			if s.Detail != "" {
				for _, line := range strings.Split(s.Detail, "\n") {
					myui.Text(c, line).MinWidth(0)
				}
				if myui.Button(c, "打开所在文件夹").Clicked() {
					openPDFFolder(s.Output)
				}
			}
		})
	})
}

func (s *PDFState) matchedPreset() string {
	for _, preset := range pdfPresets {
		if s.DPI == preset.dpi && s.Quality == preset.quality {
			return preset.name
		}
	}
	return ""
}

func (s *PDFState) useFile(path string) {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return
	}
	s.Source = path
	s.SourceSize = info.Size()
	if strings.TrimSpace(s.OutputDir) == "" {
		s.OutputDir = filepath.Dir(path)
	}
	s.Detail = ""
	s.Failure = ""
	s.Output = ""
}

func pickPath(a *model.App, title string, directory bool, current string) (string, bool) {
	if !mygo.App.IsReady() {
		return "", false
	}
	dir := current
	if !directory && current != "" {
		dir = filepath.Dir(current)
	}
	opts := mygo.OpenDialogOptions{
		Parent:      mygo.FocusedWindow(),
		Title:       title,
		DefaultPath: dir,
		Directory:   directory,
	}
	if !directory {
		opts.Filters = []mygo.FileFilter{{Name: "PDF", Extensions: []string{"pdf"}}}
	}
	paths, err := mygo.Dialog.Open(opts)
	if err != nil {
		a.Status = "失败"
		a.Error = err.Error()
		a.Elapsed = ""
		return "", false
	}
	if len(paths) == 0 {
		return "", false
	}
	return paths[0], true
}

func startPDF(a *model.App, s *PDFState, win *mygo.Window) {
	if s.Running {
		return
	}
	if strings.TrimSpace(s.Source) == "" {
		a.Status = "失败"
		a.Error = "请先选择要压缩的 PDF 文件"
		a.Elapsed = ""
		return
	}
	info, err := os.Stat(s.Source)
	if err != nil || info.IsDir() {
		a.Status = "失败"
		a.Error = "源文件不存在，请重新选择"
		a.Elapsed = ""
		return
	}
	out := pdftool.OutputPath(s.Source, s.OutputDir)
	s.Running = true
	s.Page = 0
	s.Pages = 0
	s.Detail = ""
	s.Failure = ""
	s.Output = ""
	s.generation++
	gen := s.generation
	dpi := int(math.Round(s.DPI))
	quality := int(math.Round(s.Quality))
	in := s.Source
	a.Status = "正在压缩"
	a.Error = ""
	a.Elapsed = ""
	start := time.Now()

	opt := pdftool.Options{
		DPI:     dpi,
		Quality: quality,
		OnPage: func(page, total int) {
			publishPDF(win, func() {
				if s.generation != gen {
					return
				}
				s.Page = page
				s.Pages = total
				if a.CurrentTool == model.ToolPDF {
					a.Status = fmt.Sprintf("正在处理 %d / %d 页", page, total)
					a.Error = ""
				}
			})
		},
	}
	run := func() {
		res, err := pdftool.Compress(context.Background(), in, out, opt)
		publishPDF(win, func() {
			if s.generation != gen {
				return
			}
			s.finish(a, start, res, err)
		})
	}
	if win == nil {
		run()
		return
	}
	go run()
}

func publishPDF(win *mygo.Window, fn func()) {
	if win == nil {
		fn()
		return
	}
	win.Update(fn)
}

func (s *PDFState) finish(a *model.App, start time.Time, res pdftool.Result, err error) {
	s.Running = false
	if err != nil {
		s.Failure = err.Error()
		s.Detail = ""
		s.Output = ""
		if a.CurrentTool == model.ToolPDF {
			finish(a, start, err, "")
		}
		return
	}
	s.Output = res.Output
	s.Failure = ""
	s.Detail = formatPDFResult(res)
	if a.CurrentTool != model.ToolPDF {
		return
	}
	status := "已压缩"
	if res.Original > 0 && res.Compressed >= res.Original {
		status = "体积没有减少"
	}
	finish(a, start, nil, status)
}

func formatPDFResult(res pdftool.Result) string {
	line := fmt.Sprintf("%s  →  %s", formatBytes(res.Original), formatBytes(res.Compressed))
	if res.Original > 0 && res.Compressed < res.Original {
		ratio := (1 - float64(res.Compressed)/float64(res.Original)) * 100
		line = fmt.Sprintf("体积减少 %.1f%%    %s", ratio, line)
	} else {
		line = "文件已是较优体积    " + line
	}
	return line + "\n保存至 " + res.Output
}

func formatBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
}

func openPDFFolder(path string) {
	if path == "" || !mygo.App.IsReady() {
		return
	}
	_ = mygo.Shell.OpenPath(filepath.Dir(path))
}
