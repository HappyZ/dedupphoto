package main

import (
	"encoding/json"
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
)

// Config holds configuration options for the program
type Config struct {
	Folder          string
	TemporaryOutput string
	IsRecursive     bool
	DryRun          bool
}

type DuplicatedImageData struct {
	Hash  string   `json:"hash"`
	Files []string `json:"files"`
}

func parseFlags() Config {
	var config Config

	flag.StringVar(&config.Folder, "folder", "", "folder to find all images")
	flag.StringVar(&config.TemporaryOutput, "tempoutput", "/tmp", "folder to store temporary outputs")
	flag.BoolVar(&config.IsRecursive, "recursive", true, "whether find images in nested folders")
	flag.BoolVar(&config.DryRun, "dryrun", true, "print the dryrun message")

	flag.Parse()

	return config
}

func dumpDuplicatestoJSON(duplicates map[uint64][]string, filepath string) error {
	var data []DuplicatedImageData

	for hash, paths := range duplicates {
		data = append(data, DuplicatedImageData{
			Hash:  fmt.Sprintf("%x", hash), // Convert uint64 to hex string for hash
			Files: paths,
		})
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func main() {
	config := parseFlags()

	if config.Folder == "" {
		log.Fatalf("input folder is empty!")
		return
	}

	fmt.Printf("search in folder path: %s\n", config.Folder)
	if config.IsRecursive {
		fmt.Println("search in nested folders as well")
	}

	if config.DryRun {
		fmt.Println("please run with --dryrun=false")
		return
	}

	imageHashes, imagePaths, err := getImageFilesAndEncode(config.Folder, config.IsRecursive)
	if err != nil {
		log.Fatalln(fmt.Errorf("fail to get images and compare: %v", err))
		return
	}

	fmt.Printf("went through %d images in total\n", len(imagePaths))
	duplicates := make(map[uint64][]string)
	for hash, paths := range imageHashes {
		if len(paths) > 1 {
			duplicates[hash] = paths
		}
	}

	duplicatesFilepath := filepath.Join(config.TemporaryOutput, "duplicates.json")
	err = dumpDuplicatestoJSON(duplicates, duplicatesFilepath)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to dump duplicates to JSON: %v", err))
	} else {
		fmt.Printf("dumpped duplicates to json at %s\n", duplicatesFilepath)
	}

	webserverHandler(duplicatesFilepath)
}
