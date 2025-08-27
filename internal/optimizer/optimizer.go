package optimizer

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"
)

// Optimizer interface (remote pipeline usage) - operates on in-memory bytes.
type Optimizer interface {
	OptimizeBytes(data []byte, format string, params Params) (out []byte, res Result, err error)
}

// Params holds format-specific optimization parameters.
type Params struct {
	JPEGQuality int
	MaxWidth    int // 0 = no width limit
	MaxHeight   int // 0 = no height limit
}

// Result describes optimization outcome.
type Result struct {
	OriginalSize  int64
	OptimizedSize int64
	Duration      time.Duration
	Skipped       bool
	Reason        string
}

// ImageOptimizer represents an image optimization tool (implements both legacy file API and new interface).
type ImageOptimizer struct {
	Quality int
}

// New creates a new ImageOptimizer with default settings
func New() *ImageOptimizer {
	return &ImageOptimizer{Quality: 80}
}

// resizeImage resizes an image while maintaining aspect ratio using CatmullRom interpolation
func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	if maxWidth <= 0 && maxHeight <= 0 {
		return img
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Calculate new dimensions while maintaining aspect ratio
	newWidth, newHeight := width, height

	if maxWidth > 0 && width > maxWidth {
		newWidth = maxWidth
		newHeight = height * maxWidth / width
	}

	if maxHeight > 0 && newHeight > maxHeight {
		newHeight = maxHeight
		newWidth = width * maxHeight / height
	}

	// Don't resize if dimensions are the same
	if newWidth == width && newHeight == height {
		return img
	}

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Resize with CatmullRom interpolation
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}

// OptimizeBytes implements Optimizer interface.
func (o *ImageOptimizer) OptimizeBytes(data []byte, format string, params Params) ([]byte, Result, error) {
	start := time.Now()
	r := Result{OriginalSize: int64(len(data))}
	if params.JPEGQuality <= 0 {
		params.JPEGQuality = o.Quality
	}

	// Decode
	img, decodeFormat, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		r.Skipped = true
		r.Reason = "decode-error"
		return nil, r, fmt.Errorf("decode: %w", err)
	}
	if format == "" {
		format = decodeFormat
	}

	// Resize if dimensions are specified
	if params.MaxWidth > 0 || params.MaxHeight > 0 {
		img = resizeImage(img, params.MaxWidth, params.MaxHeight)
	}
	buf := &bytes.Buffer{}
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: params.JPEGQuality}); err != nil {
			return nil, r, err
		}
	case "png":
		if err := png.Encode(buf, img); err != nil {
			return nil, r, err
		}
	default:
		r.Skipped = true
		r.Reason = "unsupported-format"
		return nil, r, fmt.Errorf("unsupported format: %s", format)
	}
	out := buf.Bytes()
	r.OptimizedSize = int64(len(out))
	r.Duration = time.Since(start)
	return out, r, nil
}

// Optimize (legacy) takes an input image path and optimizes it to outputPath.
func (o *ImageOptimizer) Optimize(inputPath, outputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(inputPath)), ".")
	out, res, err := o.OptimizeBytes(data, ext, Params{JPEGQuality: o.Quality})
	if err != nil && !res.Skipped {
		return err
	}
	if res.Skipped {
		return fmt.Errorf("skipped: %s", res.Reason)
	}
	if err := os.WriteFile(outputPath, out, 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	fmt.Printf("Optimized %s (%d bytes) -> %s (%d bytes)\n", filepath.Base(inputPath), res.OriginalSize, filepath.Base(outputPath), res.OptimizedSize)
	return nil
}
