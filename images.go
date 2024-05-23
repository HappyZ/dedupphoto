package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/corona10/goimagehash" // Import the goimagehash library
	"github.com/fsnotify/fsnotify"
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

// duplicatedImageFinder walks through a given folder in Config and creates a hash map for each image
func duplicatedImageFinder(config Config) error {
	count := uint64(0)
	countSkipped := uint64(0)

	err := filepath.Walk(config.Folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip "@eaDir" directories in Synology NAS
		if info.IsDir() && info.Name() == "@eaDir" {
			return filepath.SkipDir
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		if info.IsDir() && !config.IsRecursive && path != config.Folder {
			return filepath.SkipDir
		}

		if info.IsDir() {
			fmt.Printf("searching in folder: %s\n", info.Name())
		}

		if !info.IsDir() && isImageFile(path) {
			fullPath, err := filepath.Abs(path) // Get absolute path
			if err != nil {
				return err
			}
			hash, err := calculateHash(fullPath)
			count += 1

			if err != nil {
				countSkipped += 1
				fmt.Println(fmt.Errorf("skip image %s due to error: %v", fullPath, err))
				return nil
			}

			mu.Lock()
			imageHashes[hash] = append(imageHashes[hash], fullPath)
			mu.Unlock()

			// print if already identified as a duplicate
			if len(imageHashes[hash]) > 1 {
				fmt.Printf("%dth image seems to be a new duplicate: %s\n", count, fullPath)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Println("went through in total", count, "images, skipped", countSkipped)

	return nil
}

func folderMonitor(config Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(config.Folder)
	if err != nil {
		return err
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("folderMonitor() watcher closed unexpectedly")
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				time.Sleep(time.Second) // wait for the file to be fully written

				fullPath, err := filepath.Abs(event.Name) // Get absolute path
				if err != nil {
					return err
				}
				fmt.Println("found a new file added", fullPath)

				hash, err := calculateHash(fullPath)

				if err != nil {
					fmt.Println(fmt.Errorf("skip image %s due to error: %v", fullPath, err))
					return nil
				}

				mu.Lock()
				imageHashes[hash] = append(imageHashes[hash], fullPath)
				mu.Unlock()

				// print if already identified as a duplicate
				if len(imageHashes[hash]) > 1 {
					fmt.Println("the image seems to be a new duplicate")
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("folderMonitor() watcher closed unexpectedly")
			}
			fmt.Println("folderMonitor() error:", err)
		}
	}
}
