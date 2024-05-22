package main

import (
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"
)

// Config holds configuration options for the program
type Config struct {
	Folder      string
	IsRecursive bool
	DryRun      bool
}

type DuplicatedImageData struct {
	Hash  string   `json:"hash"`
	Files []string `json:"files"`
}

func parseFlags() Config {
	var config Config

	flag.StringVar(&config.Folder, "folder", "", "folder to find all images")
	flag.BoolVar(&config.IsRecursive, "recursive", true, "whether find images in nested folders")
	flag.BoolVar(&config.DryRun, "dryrun", true, "print the dryrun message")

	flag.Parse()

	return config
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
		log.Fatalln(fmt.Errorf("fail to get images and encode: %v", err))
		return
	}

	fmt.Printf("went through %d images in total\n", len(imagePaths))
	duplicates := make(map[uint64][]string)
	for hash, paths := range imageHashes {
		if len(paths) > 1 {
			duplicates[hash] = paths
		}
	}

	webserverHandler(duplicates)
}
