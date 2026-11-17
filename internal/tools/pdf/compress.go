package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// Options selects how images are reduced. DPI is 50–300 and Quality is
// 10–100. OnPage, when set, runs on the caller's goroutine before each page.
type Options struct {
	DPI     int
	Quality int
	OnPage  func(page, total int)
}

// Result is the file Compress wrote.
type Result struct {
	Output     string
	Original   int64
	Compressed int64
	Pages      int
	Replaced   int
}

// OutputPath is where a compressed copy of source is written.
// An empty dir uses the source file's directory.
func OutputPath(source, dir string) string {
	if strings.TrimSpace(dir) == "" {
		dir = filepath.Dir(source)
	}
	base := strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
	return filepath.Join(dir, base+"_compressed.pdf")
}

// Compress rewrites in to out. Images shrink toward DPI and are stored as
// JPEG at Quality when that stream is shorter. Text, vectors and annotations
// stay in place. JBIG2 and JPEG2000 images are left unchanged.
func Compress(ctx context.Context, in, out string, opt Options) (res Result, err error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if opt.DPI < 50 || opt.DPI > 300 {
		return Result{}, fmt.Errorf("DPI 必须在 50 到 300 之间")
	}
	if opt.Quality < 10 || opt.Quality > 100 {
		return Result{}, fmt.Errorf("图片质量必须在 10 到 100 之间")
	}
	if in == "" || out == "" {
		return Result{}, fmt.Errorf("请提供 PDF 文件")
	}
	if samePath(in, out) {
		return Result{}, fmt.Errorf("输出路径与源文件相同")
	}
	info, err := os.Stat(in)
	if err != nil {
		return Result{}, err
	}
	if info.IsDir() {
		return Result{}, fmt.Errorf("请提供 PDF 文件")
	}

	pdf, err := readPDF(ctx, in)
	if err != nil {
		return Result{}, err
	}
	replaced, err := shrinkImages(ctx, pdf, opt)
	if err != nil {
		return Result{}, err
	}
	written, err := writePDF(ctx, pdf, out)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Output:     out,
		Original:   info.Size(),
		Compressed: written,
		Pages:      pdf.PageCount,
		Replaced:   replaced,
	}, nil
}

func readPDF(ctx context.Context, in string) (*model.Context, error) {
	f, err := os.Open(in)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conf := model.NewDefaultConfiguration()
	conf.Cmd = model.OPTIMIZE
	pdf, err := api.ReadValidateAndOptimize(ctx, f, conf, nil)
	if err != nil {
		return nil, fmt.Errorf("读取 PDF 失败: %w", err)
	}
	return pdf, nil
}

func shrinkImages(ctx context.Context, pdf *model.Context, opt Options) (int, error) {
	masks := softMasks(pdf)
	thumbs := thumbnails(pdf)
	seen := map[int]struct{}{}
	replaced := 0
	for page := 1; page <= pdf.PageCount; page++ {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if opt.OnPage != nil {
			opt.OnPage(page, pdf.PageCount)
		}
		for _, objNr := range pdfcpu.ImageObjNrs(pdf, page) {
			if _, ok := seen[objNr]; ok {
				continue
			}
			seen[objNr] = struct{}{}
			if _, ok := masks[objNr]; ok {
				continue
			}
			if _, ok := thumbs[objNr]; ok {
				continue
			}
			ok, err := shrinkImage(ctx, pdf, objNr, opt.DPI, opt.Quality)
			if err != nil {
				return 0, err
			}
			if ok {
				replaced++
			}
		}
	}
	return replaced, nil
}

func shrinkImage(ctx context.Context, pdf *model.Context, objNr, dpi, quality int) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if pdf.Optimize == nil {
		return false, nil
	}
	obj := pdf.Optimize.ImageObjects[objNr]
	if obj == nil || obj.ImageDict == nil || imageMask(obj.ImageDict) {
		return false, nil
	}
	before := encodedSize(obj.ImageDict)

	extracted, err := pdfcpu.ExtractImage(ctx, pdf, obj.ImageDict, false, "", objNr, false)
	if err != nil {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		return false, nil
	}
	if extracted == nil {
		return false, nil
	}
	if extracted.FileType == "jpx" || extracted.FileType == "jbig2" {
		return false, nil
	}
	raw, err := io.ReadAll(extracted)
	if err != nil {
		return false, nil
	}
	decoded, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return false, nil
	}
	w, h := scaledSize(decoded.Bounds().Dx(), decoded.Bounds().Dy(), dpi)
	jpegBytes, gray, err := encodeJPEG(resize(decoded, w, h), quality)
	if err != nil || len(jpegBytes) == 0 || len(jpegBytes) >= before {
		return false, nil
	}
	if err := replaceStream(pdf, objNr, jpegBytes, w, h, gray); err != nil {
		return false, nil
	}
	return true, nil
}

// replaceStream swaps the image object for a JPEG of a new pixel size.
// The page still draws it into the same rectangle.
func replaceStream(pdf *model.Context, objNr int, jpegBytes []byte, w, h int, gray bool) error {
	cs := model.DeviceRGBCS
	if gray {
		cs = model.DeviceGrayCS
	}
	stream, err := model.CreateDCTImageStreamDict(pdf.XRefTable, jpegBytes, w, h, 8, cs)
	if err != nil {
		return err
	}
	entry, ok := pdf.FindTableEntry(objNr, 0)
	if !ok || entry == nil {
		return fmt.Errorf("image obj#%d: missing xref entry", objNr)
	}
	entry.Object = *stream
	entry.Free = false
	entry.Compressed = false
	entry.ObjectStream = nil
	entry.ObjectStreamInd = nil
	return nil
}

func softMasks(pdf *model.Context) map[int]struct{} {
	out := map[int]struct{}{}
	if pdf.Optimize == nil {
		return out
	}
	for _, obj := range pdf.Optimize.ImageObjects {
		if obj == nil || obj.ImageDict == nil {
			continue
		}
		for _, key := range []string{"SMask", "Mask"} {
			ref, ok := indirect(obj.ImageDict, key)
			if ok {
				out[ref] = struct{}{}
			}
		}
	}
	return out
}

func thumbnails(pdf *model.Context) map[int]struct{} {
	out := map[int]struct{}{}
	for _, ref := range pdf.PageThumbs {
		out[ref.ObjectNumber.Value()] = struct{}{}
	}
	return out
}

func indirect(sd *types.StreamDict, key string) (int, bool) {
	o, ok := sd.Find(key)
	if !ok {
		return 0, false
	}
	switch ref := o.(type) {
	case types.IndirectRef:
		return ref.ObjectNumber.Value(), true
	case *types.IndirectRef:
		if ref == nil {
			return 0, false
		}
		return ref.ObjectNumber.Value(), true
	default:
		return 0, false
	}
}

func imageMask(sd *types.StreamDict) bool {
	o, ok := sd.Find("ImageMask")
	if !ok {
		return false
	}
	switch v := o.(type) {
	case types.Boolean:
		return bool(v)
	case *types.Boolean:
		return v != nil && bool(*v)
	default:
		return false
	}
}

func encodedSize(sd *types.StreamDict) int {
	if sd.StreamLength != nil && *sd.StreamLength > 0 {
		return int(*sd.StreamLength)
	}
	if len(sd.Raw) > 0 {
		return len(sd.Raw)
	}
	return len(sd.Content)
}

func writePDF(ctx context.Context, pdf *model.Context, out string) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return 0, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(out), ".pdf-compress-*.pdf")
	if err != nil {
		return 0, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	writeErr := api.Write(ctx, pdf, tmp, pdf.Configuration)
	closeErr := tmp.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return 0, fmt.Errorf("写入 PDF 失败: %w", err)
	}
	if err := replaceFile(tmpName, out); err != nil {
		return 0, fmt.Errorf("写入 PDF 失败: %w", err)
	}
	info, err := os.Stat(out)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func replaceFile(tmp, out string) error {
	if _, err := os.Stat(out); err != nil {
		if os.IsNotExist(err) {
			return os.Rename(tmp, out)
		}
		return err
	}
	backup, err := os.CreateTemp(filepath.Dir(out), ".pdf-backup-*")
	if err != nil {
		return err
	}
	backupName := backup.Name()
	if err := backup.Close(); err != nil {
		return err
	}
	if err := os.Remove(backupName); err != nil {
		return err
	}
	if err := os.Rename(out, backupName); err != nil {
		return err
	}
	if err := os.Rename(tmp, out); err != nil {
		_ = os.Rename(backupName, out)
		return err
	}
	return os.Remove(backupName)
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		aa, bb = a, b
	}
	aa, bb = filepath.Clean(aa), filepath.Clean(bb)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(aa, bb)
	}
	return aa == bb
}
