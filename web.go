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
		fmt.Printf("failed to parse html template: %v\n", err)
		return
	}

	err = tmpl.Execute(w, jsonData)
	if err != nil {
		fmt.Printf("error executing template: %v\n", err)
		return
	}
}

func deleteImage(path string, config Config) error {
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	// Extract the filename from the path
	filename := filepath.Base(path)

	if config.TrashBin == "" {
		// delete image
		err := os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to delete file: %v", err)
		}
	} else {
		// move image otherwise
		destinationPath := filepath.Join(config.TrashBin, filename)

		// Move the file to the destination folder
		err := os.Rename(path, destinationPath)
		if err != nil {
			return fmt.Errorf("failed to move file to %s: %v", config.TrashBin, err)
		}
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request, config Config) {
	if r.URL.Path == "/" {
		jsonData := generateDuplicatedImageJsonData()
		indexHTMLHandler(w, jsonData)
	} else if strings.HasPrefix(r.URL.Path, "/image") {
		path := r.URL.Query().Get("path")
		if path == "" {
			http.Error(w, "missing file path", http.StatusBadRequest)
			return
		}

		data, err := base64encode(path)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to encode file: %v", err), http.StatusInternalServerError)
			return
		}

		mimeType := mime.TypeByExtension(filepath.Ext(path))
		if mimeType == "" {
			http.Error(w, "unsupported file type", http.StatusBadRequest)
			return
		}

		imageData := WebImageData{
			Mime: mimeType,
			Data: data,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(imageData)
	} else if strings.HasPrefix(r.URL.Path, "/delete") {
		// Handle image deletion
		path := r.URL.Query().Get("path")
		if path == "" {
			http.Error(w, "missing file path", http.StatusBadRequest)
			return
		}

		// Perform deletion operation
		err := deleteImage(path, config)
		if err != nil {
			// Send error response as JSON
			jsonResponse := map[string]string{"error": fmt.Sprintf("failed to delete image: %v", err)}
			jsonResponseBytes, err := json.Marshal(jsonResponse)
			if err != nil {
				http.Error(w, "failed to marshal JSON response", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonResponseBytes)
			return
		}

		// Send success response as JSON
		jsonResponse := map[string]string{"message": fmt.Sprintf("image %s deleted successfully", path)}
		jsonResponseBytes, err := json.Marshal(jsonResponse)
		if err != nil {
			http.Error(w, "failed to marshal JSON response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponseBytes)
		fmt.Printf("image %s deleted successfully\n", path)
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

func webServer(config Config) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, config)
	})

	fmt.Println("Server listening on port 8888")
	return http.ListenAndServe(":8888", nil)
}
