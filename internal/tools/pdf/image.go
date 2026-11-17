package pdf

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"

	xdraw "golang.org/x/image/draw"
)

// scaledSize is the pixel size after treating the image as 300 DPI and
// targeting dpi. Dimensions only shrink, and each side stays at least 1.
func scaledSize(width, height, dpi int) (int, int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	if dpi >= 300 {
		return width, height
	}
	scale := float64(dpi) / 300
	return max(1, int(float64(width)*scale)), max(1, int(float64(height)*scale))
}

// prepare makes an image jpeg can store without an alpha channel.
// Gray stays gray. Everything else is drawn over white.
func prepare(src image.Image) image.Image {
	if _, ok := src.(*image.Gray); ok {
		return src
	}
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Over)
	return dst
}

func resize(src image.Image, width, height int) image.Image {
	src = prepare(src)
	b := src.Bounds()
	if b.Dx() == width && b.Dy() == height {
		return src
	}
	dr := image.Rect(0, 0, width, height)
	if _, ok := src.(*image.Gray); ok {
		dst := image.NewGray(dr)
		xdraw.CatmullRom.Scale(dst, dr, src, b, xdraw.Src, nil)
		return dst
	}
	dst := image.NewRGBA(dr)
	xdraw.CatmullRom.Scale(dst, dr, src, b, xdraw.Src, nil)
	return dst
}

func encodeJPEG(img image.Image, quality int) ([]byte, bool, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, false, err
	}
	_, gray := img.(*image.Gray)
	return buf.Bytes(), gray, nil
}
