package fax

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestImages holds the image names to use for regular tests
var testImageNames = []string{"red"}

// BenchmarkImages holds the image names to use for benchmarks
var benchmarkImageNames = []string{"red"}

// testDataDir is the path to the test data directory
const testDataDir = "../../testdata/compress"

// getSampleImageData loads a TIFF image file and returns its data without the header
func getSampleImageData(name string) ([]byte, error) {
	path := filepath.Join(testDataDir, name+".tiff")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read sample image %s: %w", path, err)
	}

	// Strip TIFF header (8 bytes)
	if len(data) < 8 {
		return nil, fmt.Errorf("invalid TIFF file: %s (too small)", path)
	}
	return data[8:], nil
}

// getPrototypeImage loads a reference GIF image for comparison
func getPrototypeImage(name string) (image.Image, error) {
	path := filepath.Join(testDataDir, name+".gif")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open prototype image %s: %w", path, err)
	}
	defer file.Close()

	img, err := gif.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode prototype image %s: %w", path, err)
	}

	return img, nil
}

// imageEqual compares two images pixel by pixel and returns true if identical
func imageEqual(a, b image.Image) bool {
	if !a.Bounds().Eq(b.Bounds()) {
		return false
	}

	bounds := a.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			ra, ga, ba, aa := a.At(x, y).RGBA()
			rb, gb, bb, ab := b.At(x, y).RGBA()
			if ra != rb || ga != gb || ba != bb || aa != ab {
				return false
			}
		}
	}

	return true
}

// saveFailureImage saves a failed test image for debugging
func saveFailureImage(name string, img image.Image) error {
	failFile := name + "-fail.png"
	file, err := os.Create(failFile)
	if err != nil {
		return fmt.Errorf("failed to create failure image %s: %w", failFile, err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode failure image %s: %w", failFile, err)
	}

	return nil
}

// TestFullDecode tests the full G4 decoding process against reference images
func TestFullDecode(t *testing.T) {
	for _, name := range testImageNames {
		t.Run(name, func(t *testing.T) {
			// Clean up any previous failure files
			failFile := name + "-fail.png"
			_ = os.Remove(failFile)

			// Load reference image
			expected, err := getPrototypeImage(name)
			if err != nil {
				t.Fatalf("Failed to load prototype: %v", err)
			}
			bounds := expected.Bounds()

			// Load and decode the sample image
			sampleData, err := getSampleImageData(name)
			if err != nil {
				t.Fatalf("Failed to load sample: %v", err)
			}

			reader := bytes.NewBuffer(sampleData)
			result, err := DecodeG4(reader, bounds.Dx(), bounds.Dy())
			if err != nil {
				t.Fatalf("Failed to decode G4 image %s: %v", name, err)
			}

			// Validate dimensions
			if !bounds.Eq(result.Bounds()) {
				t.Fatalf("Image bounds mismatch: expected %v, got %v", bounds, result.Bounds())
			}

			// Compare images
			if !imageEqual(expected, result) {
				if err := saveFailureImage(name, result); err != nil {
					t.Errorf("Failed to save failure image: %v", err)
				}
				t.Errorf("Image content mismatch, see %s for details", failFile)
			}
		})
	}
}

// BenchmarkFullDecode benchmarks the G4 decoding performance
func BenchmarkFullDecode(b *testing.B) {
	// Calculate total pixels for benchmarking
	var pixelCount int64
	benchData := make(map[string]struct {
		data   []byte
		width  int
		height int
	})

	// Pre-load all benchmark images
	for _, name := range benchmarkImageNames {
		data, err := getSampleImageData(name)
		if err != nil {
			b.Fatalf("Failed to load benchmark data: %v", err)
		}

		img, err := getPrototypeImage(name)
		if err != nil {
			b.Fatalf("Failed to load prototype image: %v", err)
		}

		bounds := img.Bounds()
		width, height := bounds.Dx(), bounds.Dy()
		pixelCount += int64(width * height)

		benchData[name] = struct {
			data   []byte
			width  int
			height int
		}{
			data:   data,
			width:  width,
			height: height,
		}
	}

	b.ResetTimer()

	// Run benchmarks for each image
	for _, name := range benchmarkImageNames {
		data := benchData[name]
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(data.width * data.height))
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				reader := bytes.NewBuffer(data.data)
				img, err := DecodeG4(reader, data.width, data.height)
				if err != nil {
					b.Fatalf("Decode failed: %v", err)
				}

				// Force result to be used to prevent compiler optimizations
				if img.Bounds().Dx() != data.width {
					b.Fatalf("Unexpected image width")
				}
			}
		})
	}
}
