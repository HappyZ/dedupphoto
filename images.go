package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/corona10/goimagehash" // Import the goimagehash library
)

var extensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp"}

// isImageFile returns true if the given file has a valid image extension
func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range extensions {
		if e == ext {
			return true
		}
	}
	return false
}

// calculateHash calculates the perceptual hash of an image file
// calculateHash calculates the average hash of an image file
func calculateHash(filePath string) (uint64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("fail to open file %s: %v", filePath, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return 0, fmt.Errorf("fail to decode image %s: %v", filePath, err)
	}
	// TODO(happyz): replace below with more accurate model.
	// Here we use phash algorithm to detect image diff.
	// No accuracy data is found but a few tests show it works well.
	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return 0, err
	}

	return hash.GetHash(), nil
}

// getImageFilesAndEncode encodes images to a hash map and returns it
func getImageFilesAndEncode(folder string, recursive bool) (map[uint64][]string, []string, error) {
	var imagePaths []string
	var imageHashes = make(map[uint64][]string)

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && !recursive && path != folder {
			return filepath.SkipDir
		}
		if !info.IsDir() && isImageFile(path) {
			fullPath, err := filepath.Abs(path) // Get absolute path
			if err != nil {
				return err
			}
			hash, err := calculateHash(fullPath)
			if err != nil {
				fmt.Println(fmt.Errorf("skip image %s due to error %v", fullPath, err))
				return nil
			}
			imagePaths = append(imagePaths, fullPath)
			imageHashes[hash] = append(imageHashes[hash], fullPath)
		}
		return nil
	})

	if err != nil {
		return imageHashes, imagePaths, err
	}

	return imageHashes, imagePaths, nil
}
