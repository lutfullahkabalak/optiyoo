package imagemin

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

const (
	// MaxEdgePixels uzun kenar bu değeri aşarsa oran korunarak küçültülür (disk + bant genişliği).
	MaxEdgePixels = 1920
	// JPEGQuality görünür kaliteyi koruyarak dosya boyutunu düşürür (1–100).
	JPEGQuality = 85
)

func decode(b []byte) (image.Image, string, error) {
	r := bytes.NewReader(b)
	img, err := webp.Decode(r)
	if err == nil {
		return img, "webp", nil
	}
	r.Seek(0, io.SeekStart)
	return image.Decode(r)
}

func isAnimatedGIF(b []byte) bool {
	g, err := gif.DecodeAll(bytes.NewReader(b))
	if err != nil {
		return false
	}
	return len(g.Image) > 1
}

func resizeIfNeeded(src image.Image) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= MaxEdgePixels && h <= MaxEdgePixels {
		return src
	}
	var nw, nh int
	if w >= h {
		nw = MaxEdgePixels
		nh = max(1, int(float64(h)*float64(MaxEdgePixels)/float64(w)))
	} else {
		nh = MaxEdgePixels
		nw = max(1, int(float64(w)*float64(MaxEdgePixels)/float64(h)))
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	return dst
}

// Compress yeniden kodlar ve gerekirse boyutlandırır. Çıktı orijinalden büyük veya eşitse orijinal baytları döner.
// inputMIME: "image/jpeg" gibi temel MIME (parametresiz).
// Dönüş: out, dosya uzantısı (.jpg / .png / .gif), çıktı MIME.
func Compress(in []byte, inputMIME string) (out []byte, ext string, mime string) {
	ext, mime = extAndMIME(inputMIME)
	origExt, origMIME := ext, mime

	if isAnimatedGIF(in) {
		return in, ".gif", "image/gif"
	}

	img, format, err := decode(in)
	if err != nil {
		return in, origExt, origMIME
	}

	img = resizeIfNeeded(img)

	var buf bytes.Buffer
	switch format {
	case "png", "gif":
		enc := &png.Encoder{CompressionLevel: png.BestCompression}
		if err := enc.Encode(&buf, img); err != nil {
			return in, origExt, origMIME
		}
		out = buf.Bytes()
		ext = ".png"
		mime = "image/png"
	default: // jpeg, webp ve bilinmeyen raster
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: JPEGQuality}); err != nil {
			return in, origExt, origMIME
		}
		out = buf.Bytes()
		ext = ".jpg"
		mime = "image/jpeg"
	}

	if len(out) >= len(in) {
		return in, origExt, origMIME
	}
	return out, ext, mime
}

func extAndMIME(inputMIME string) (ext, mime string) {
	switch inputMIME {
	case "image/jpeg":
		return ".jpg", "image/jpeg"
	case "image/png":
		return ".png", "image/png"
	case "image/gif":
		return ".gif", "image/gif"
	case "image/webp":
		return ".webp", "image/webp"
	default:
		return ".bin", inputMIME
	}
}
