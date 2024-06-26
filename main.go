package main

import (
	"flag"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"
)

func parseFlags() Config {
	var config Config

	flag.StringVar(&config.Folder, "folder", "", "folder to find all images")
	flag.StringVar(&config.TrashBin, "trashbin", "", "when delete, we move duplicated images to the specified trash bin folder, or if empty we delete those")
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
	if config.TrashBin == "" {
		fmt.Println("deletion will actually delete the duplicated image")
	} else {
		fmt.Println("deletion will only move duplicated ones to", config.TrashBin)
	}

	if config.DryRun {
		fmt.Println("please run with --dryrun=false")
		return
	}

	errChan := make(chan error, 2)

	// run concurrent web server
	go func() {
		errChan <- webServer(config)
	}()

	// run concurrent image finder
	go func() {
		errChan <- duplicatedImageFinder(config)
	}()

	// run concurrent image monitor
	go func() {
		errChan <- folderMonitor(config)
	}()

	// Wait for both goroutines to finish
	for i := 0; i < 3; i++ {
		if err := <-errChan; err != nil {
			log.Fatalln("error:", err)
		}
	}
}
