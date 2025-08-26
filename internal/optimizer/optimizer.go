package optimizer

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// ImageOptimizer represents an image optimization tool
type ImageOptimizer struct {
	Quality int
}

// New creates a new ImageOptimizer with default settings
func New() *ImageOptimizer {
	return &ImageOptimizer{
		Quality: 80, // Default quality
	}
}

// Optimize takes an input image path and optimizes it
func (o *ImageOptimizer) Optimize(inputPath, outputPath string) error {
	// Open the input file
	file, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Optimize based on format
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		options := &jpeg.Options{Quality: o.Quality}
		err = jpeg.Encode(outFile, img, options)
	case "png":
		err = png.Encode(outFile, img)
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	// Get file info for reporting
	inputInfo, _ := os.Stat(inputPath)
	outputInfo, _ := os.Stat(outputPath)

	fmt.Printf("Optimized %s (%d bytes) -> %s (%d bytes)\n", 
		filepath.Base(inputPath), 
		inputInfo.Size(), 
		filepath.Base(outputPath), 
		outputInfo.Size())

	return nil
}