package main

import "sync"

// Config holds configuration options for the program
type Config struct {
	Folder      string
	IsRecursive bool
	DryRun      bool
}

// DuplicatedImageJsonData holds data structure for duplicated images
type DuplicatedImageJsonData struct {
	Hash  string   `json:"hash"`
	Files []string `json:"files"`
}

// WebImageData holds data structure for web images
type WebImageData struct {
	Mime string `json:"mime"`
	Data string `json:"data"`
}

// Mutex for concurrency
var mu sync.Mutex

var imageHashes = make(map[uint64][]string)
