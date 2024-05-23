package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/*.html
var content embed.FS

func base64encode(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func indexHTMLHandler(w http.ResponseWriter, jsonData []DuplicatedImageJsonData) {
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

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		jsonData := generateDuplicatedImageJsonData()
		indexHTMLHandler(w, jsonData)
	} else if strings.HasPrefix(r.URL.Path, "/image") {
		path := r.URL.Query().Get("path")
		if path == "" {
			http.Error(w, "Missing file path", http.StatusBadRequest)
			return
		}

		data, err := base64encode(path)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode file: %v", err), http.StatusInternalServerError)
			return
		}

		mimeType := mime.TypeByExtension(filepath.Ext(path))
		if mimeType == "" {
			http.Error(w, "Unsupported file type", http.StatusBadRequest)
			return
		}

		imageData := WebImageData{
			Mime: mimeType,
			Data: data,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(imageData)
	} else {
		http.NotFound(w, r)
	}
}

func generateDuplicatedImageJsonData() []DuplicatedImageJsonData {
	mu.Lock()
	defer mu.Unlock()

	var jsonData []DuplicatedImageJsonData

	for hash, files := range imageHashes {
		// skip if no duplicates
		if len(files) < 2 {
			continue
		}
		imageData := DuplicatedImageJsonData{
			Hash:  fmt.Sprintf("%d", hash),
			Files: files,
		}
		jsonData = append(jsonData, imageData)
	}

	return jsonData
}

func webServer() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	})

	fmt.Println("Server listening on port 8888")
	return http.ListenAndServe(":8888", nil)
}
