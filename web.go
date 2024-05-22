package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/*.html
var content embed.FS

func mimeType(path string) string {
	return mime.TypeByExtension(filepath.Ext(path))
}

func base64encode(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func indexHTMLHandler(w http.ResponseWriter, jsonData []DuplicatedImageData) {
	// Parse and serve HTML template
	tmpl, err := template.ParseFS(content, "templates/index.html")
	if err != nil {
		fmt.Fprintf(w, "failed to parse html template: %v", err)
		return
	}

	err = tmpl.Execute(w, jsonData)
	if err != nil {
		fmt.Fprintf(w, "Error executing template: %v", err)
		return
	}
}

func handler(w http.ResponseWriter, r *http.Request, data map[uint64][]string) {
	if r.URL.Path == "/" {
		var jsonData []DuplicatedImageData

		for hash, files := range data {
			imageData := DuplicatedImageData{
				Hash:  fmt.Sprintf("%d", hash),
				Files: files,
			}
			jsonData = append(jsonData, imageData)
		}

		indexHTMLHandler(w, jsonData)
	} else if strings.HasPrefix(r.URL.Path, "/image") {
		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			http.Error(w, "Missing file path", http.StatusBadRequest)
			return
		}

		data, err := base64encode(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode file: %v", err), http.StatusInternalServerError)
			return
		}

		mimeType := mimeType(filePath)
		if mimeType == "" {
			http.Error(w, "Unsupported file type", http.StatusBadRequest)
			return
		}

		imageData := struct {
			Mime string `json:"mime"`
			Data string `json:"data"`
		}{
			Mime: mimeType,
			Data: data,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(imageData)
	} else {
		http.NotFound(w, r)
	}
}

func webserverHandler(duplicates map[uint64][]string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, duplicates)
	})

	fmt.Println("Server listening on port 8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
