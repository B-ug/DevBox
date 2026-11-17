package ui

import (
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"devbox/internal/model"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func TestPDFPage(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("PDF"); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{
		"缩小图片以减小体积",
		"缩小内嵌图片并重新编码为 JPEG，文字保持不变",
		"拖放 PDF 到此处",
		"选择文件",
		"输出目录",
		"高质量",
		"均衡",
		"最小体积",
		"分辨率 140",
		"质量 80",
		"开始压缩",
	} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q in %q", text, tt.Texts())
		}
	}

	if err := tt.Click("最小体积"); err != nil {
		t.Fatal(err)
	}
	if editors.PDF.DPI != 80 || editors.PDF.Quality != 50 || editors.PDF.Preset != "最小体积" {
		t.Fatalf("preset %+v", editors.PDF)
	}
	if !tt.HasText("分辨率 80") || !tt.HasText("质量 50") {
		t.Fatalf("preset labels %q", tt.Texts())
	}

	if err := tt.Click("开始压缩"); err != nil {
		t.Fatal(err)
	}
	if a.Error != "请先选择要压缩的 PDF 文件" || !tt.HasText("错误：请先选择要压缩的 PDF 文件") {
		t.Fatalf("status %q error %q texts %q", a.Status, a.Error, tt.Texts())
	}
}

func TestPDFRejectsMissingFile(t *testing.T) {
	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	editors.PDF.Source = filepath.Join(t.TempDir(), "missing.pdf")
	if err := tt.Click("PDF"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("开始压缩"); err != nil {
		t.Fatal(err)
	}
	if a.Error != "源文件不存在，请重新选择" || !tt.HasText("错误：源文件不存在，请重新选择") {
		t.Fatalf("status %q error %q", a.Status, a.Error)
	}
}

func TestPDFDrop(t *testing.T) {
	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "scan.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4"), 0o644); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(other, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}

	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	if err := tt.Click("PDF"); err != nil {
		t.Fatal(err)
	}
	r, ok := tt.Find("拖放 PDF 到此处")
	if !ok {
		t.Fatal("drop zone missing")
	}
	if !tt.DropFiles(r.X+8, r.Y+8, other) {
		t.Fatal("drop was not accepted")
	}
	if editors.PDF.Source != "" {
		t.Fatalf("accepted %s", editors.PDF.Source)
	}
	if !tt.DropFiles(r.X+8, r.Y+8, pdfPath) {
		t.Fatal("pdf drop was not accepted")
	}
	if editors.PDF.Source != pdfPath || !tt.HasText("已选择 scan.pdf（8 B）") {
		t.Fatalf("source %q texts %q", editors.PDF.Source, tt.Texts())
	}
}

func TestPDFCompressFromPage(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "photo.jpg")
	writeUIPhoto(t, imgPath)
	in := filepath.Join(dir, "scan.pdf")
	if err := api.ImportImagesFile(context.Background(), []string{imgPath}, in, nil, nil); err != nil {
		t.Fatal(err)
	}

	a := model.New()
	editors, tt := layoutTester(a, 1200, 800)
	editors.PDF.Source = in
	editors.PDF.SourceSize = 1
	editors.PDF.OutputDir = dir
	editors.PDF.DPI = 50
	editors.PDF.Quality = 20
	if err := tt.Click("PDF"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("开始压缩"); err != nil {
		t.Fatal(err)
	}
	if editors.PDF.Running || editors.PDF.Failure != "" || editors.PDF.Output == "" {
		t.Fatalf("running %v failure %q output %q status %q", editors.PDF.Running, editors.PDF.Failure, editors.PDF.Output, a.Status)
	}
	if !tt.HasText("打开所在文件夹") || a.Status == "失败" {
		t.Fatalf("status %q error %q texts %q", a.Status, a.Error, tt.Texts())
	}
}

func writeUIPhoto(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 80, 60))
	for y := 0; y < 60; y++ {
		for x := 0; x < 80; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x * 3), G: uint8(y * 4), B: uint8(x ^ y), A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
}
