package pdf

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func TestScaledSize(t *testing.T) {
	w, h := scaledSize(1200, 900, 300)
	if w != 1200 || h != 900 {
		t.Fatalf("300 dpi %dx%d", w, h)
	}
	w, h = scaledSize(1200, 900, 100)
	if w != 400 || h != 300 {
		t.Fatalf("100 dpi %dx%d", w, h)
	}
	w, h = scaledSize(2, 2, 50)
	if w != 1 || h != 1 {
		t.Fatalf("tiny %dx%d", w, h)
	}
}

func TestOutputPath(t *testing.T) {
	got := OutputPath(filepath.Join("docs", "scan.pdf"), "")
	if filepath.Base(got) != "scan_compressed.pdf" || filepath.Dir(got) != "docs" {
		t.Fatalf("path %s", got)
	}
	got = OutputPath(filepath.Join("docs", "scan.pdf"), filepath.Join("out"))
	if got != filepath.Join("out", "scan_compressed.pdf") {
		t.Fatalf("path %s", got)
	}
}

func TestCompressRejectsBadOptions(t *testing.T) {
	_, err := Compress(context.Background(), "a.pdf", "b.pdf", Options{DPI: 10, Quality: 80})
	if err == nil || !strings.Contains(err.Error(), "DPI") {
		t.Fatalf("dpi error %v", err)
	}
	_, err = Compress(context.Background(), "a.pdf", "a.pdf", Options{DPI: 140, Quality: 80})
	if err == nil || !strings.Contains(err.Error(), "相同") {
		t.Fatalf("same path error %v", err)
	}
}

func TestCompressShrinksImageAndKeepsText(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "photo.jpg")
	writePhoto(t, imgPath)

	in := filepath.Join(dir, "scan.pdf")
	if err := api.ImportImagesFile(context.Background(), []string{imgPath}, in, nil, nil); err != nil {
		t.Fatal(err)
	}
	stamped := filepath.Join(dir, "stamped.pdf")
	if err := api.AddTextWatermarksFile(context.Background(), in, stamped, nil, true, "KeepThisSentence", "points:20, pos:c", nil); err != nil {
		t.Fatal(err)
	}
	beforeW, beforeH := imageSize(t, stamped)
	if beforeW != 1200 || beforeH != 900 {
		t.Fatalf("embedded image %dx%d", beforeW, beforeH)
	}

	out := filepath.Join(dir, "scan_compressed.pdf")
	var pages []int
	res, err := Compress(context.Background(), stamped, out, Options{
		DPI:     100,
		Quality: 40,
		OnPage:  func(page, total int) { pages = append(pages, page, total) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Pages != 1 || res.Replaced != 1 || res.Compressed >= res.Original {
		t.Fatalf("result %+v", res)
	}
	if len(pages) != 2 || pages[0] != 1 || pages[1] != 1 {
		t.Fatalf("progress %v", pages)
	}
	afterW, afterH := imageSize(t, out)
	if afterW != 400 || afterH != 300 {
		t.Fatalf("compressed image %dx%d", afterW, afterH)
	}
	count, err := api.PageCountFile(context.Background(), out)
	if err != nil || count != 1 {
		t.Fatalf("pages %d %v", count, err)
	}
	if !contentHas(t, out, "KeepThisSentence") {
		t.Fatal("text watermark was not kept")
	}
}

func writePhoto(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1200, 900))
	for y := 0; y < 900; y++ {
		for x := 0; x < 1200; x++ {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: uint8(x ^ y), A: 255})
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

func imageSize(t *testing.T, path string) (int, int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	groups, err := api.Images(context.Background(), f, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range groups {
		for _, img := range group {
			if img.Width > 0 && !img.Thumb && !img.IsImgMask {
				return img.Width, img.Height
			}
		}
	}
	t.Fatal("no image")
	return 0, 0
}

func contentHas(t *testing.T, path, text string) bool {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	pdf, err := api.ReadContext(context.Background(), f, model.NewDefaultConfiguration())
	if err != nil {
		t.Fatal(err)
	}
	needle := []byte(text)
	wide := utf16BE(text)
	for _, entry := range pdf.Table {
		if entry == nil || entry.Object == nil {
			continue
		}
		if strings.Contains(entry.Object.PDFString(), text) {
			return true
		}
		sd, ok := streamDict(entry.Object)
		if !ok {
			continue
		}
		if bytes.Contains(sd.Raw, needle) || bytes.Contains(sd.Raw, wide) {
			return true
		}
		if err := sd.Decode(); err == nil && (bytes.Contains(sd.Content, needle) || bytes.Contains(sd.Content, wide)) {
			return true
		}
	}
	return false
}

func streamDict(obj types.Object) (types.StreamDict, bool) {
	switch sd := obj.(type) {
	case types.StreamDict:
		return sd, true
	case *types.StreamDict:
		if sd == nil {
			return types.StreamDict{}, false
		}
		return *sd, true
	default:
		return types.StreamDict{}, false
	}
}

func utf16BE(s string) []byte {
	out := make([]byte, 0, len(s)*2)
	for _, r := range s {
		out = append(out, byte(r>>8), byte(r))
	}
	return out
}
