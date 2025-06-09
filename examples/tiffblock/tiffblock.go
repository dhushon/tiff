// Copyright 2014 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tiff "github.com/dhushon/tiff"
)

var files = []string{
	"../testdata/video-001-tile-64x64.tiff",
	"../testdata/compress/compress_type_g4.tif",
	"../testdata/compress/red.tiff",
	"../testdata/lena512color.jpeg.tiff",
}

func main() {
	// Set up logging
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	// Process each file
	for _, name := range files {
		if err := processFile(name); err != nil {
			log.Printf("Error processing %s: %v\n", name, err)
			continue
		}
	}
}

// processFile handles the TIFF processing for a single file
func processFile(name string) error {
	fmt.Printf("Processing file: %s\n", name)

	// Load file data
	f, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Open tiff reader
	p, err := tiff.OpenReader(f)
	if err != nil {
		return fmt.Errorf("failed to open TIFF reader: %w", err)
	}
	defer p.Close()

	// Process each image and subimage
	for i := 0; i < p.ImageNum(); i++ {
		for j := 0; j < p.SubImageNum(i); j++ {
			if err := processImageBlocks(p, name, i, j); err != nil {
				log.Printf("Error processing image (%d,%d) in %s: %v\n", i, j, name, err)
				continue
			}
		}
	}

	return nil
}

// processImageBlocks extracts and saves blocks from a specific image/subimage
func processImageBlocks(p *tiff.Reader, filename string, imageIndex, subImageIndex int) error {
	// Check if image is tiled
	_, isTiled := p.Ifd[imageIndex][subImageIndex].TagGetter().GetTileWidth()
	fmt.Printf("%s(%02d,%02d) isTiled: %v\n", filename, imageIndex, subImageIndex, isTiled)

	// Get block dimensions
	blocksAcross := p.ImageBlocksAcross(imageIndex, subImageIndex)
	blocksDown := p.ImageBlocksDown(imageIndex, subImageIndex)

	fmt.Printf("Image has %d blocks across and %d blocks down\n", blocksAcross, blocksDown)

	// Process each block
	for col := 0; col < blocksAcross; col++ {
		for row := 0; row < blocksDown; row++ {
			if err := extractAndSaveBlock(p, filename, imageIndex, subImageIndex, col, row); err != nil {
				log.Printf("Error with block (%d,%d): %v\n", col, row, err)
				continue
			}
		}
	}

	return nil
}

// extractAndSaveBlock extracts a single block and saves it as a TIFF file
func extractAndSaveBlock(p *tiff.Reader, filename string, imageIndex, subImageIndex, col, row int) error {
	// Generate output filename
	baseFilename := filepath.Base(filename)
	outFilename := fmt.Sprintf("z_%s-%02d-%02d-%02d-%02d.tiff",
		baseFilename, imageIndex, subImageIndex, col, row)

	// Decode the image block
	m, err := p.DecodeImageBlock(imageIndex, subImageIndex, col, row)
	if err != nil {
		return fmt.Errorf("failed to decode block: %w", err)
	}

	// Encode as TIFF
	var buf bytes.Buffer
	if err = tiff.Encode(&buf, m, nil); err != nil {
		return fmt.Errorf("failed to encode TIFF: %w", err)
	}

	// Write to file
	if err = os.WriteFile(outFilename, buf.Bytes(), 0666); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Saved block to %s\n", outFilename)
	return nil
}
