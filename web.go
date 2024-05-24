package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

// deleteAllImages does best effort deletion of any duplicates and keep the largest image
func deleteAllImages(config Config) error {
	mu.Lock()
	defer mu.Unlock()
	var err error
	for hash, files := range imageHashes {
		// skip if no duplicates
		if len(files) < 2 {
			continue
		}

		var largestFile string
		var largestSize int64

		for _, file := range files {
			// retrieve file information
			fileInfo, err := os.Stat(file)
			if err != nil {
				fmt.Printf("failed to get file info for %s, skipped: %v\n", file, err)
				continue
			}
			fileSize := fileInfo.Size()

			if fileSize > largestSize {
				largestSize = fileSize
				largestFile = file
			}
		}

		var remainingFiles []string
		for _, file := range files {
			if file != largestFile {
				err = deleteImage(file, hash, config)
				if err != nil {
					fmt.Printf("failed to delete image %s: %v\n", file, err)
					remainingFiles = append(remainingFiles, file)
				}
			}
		}

		// update hash map here
		imageHashes[hash] = []string{largestFile}
		imageHashes[hash] = append(imageHashes[hash], remainingFiles...)
	}
	return err
}

func deleteImage(path string, hash uint64, config Config) error {
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	if _, ok := imageHashes[hash]; !ok {
		return fmt.Errorf("hash does not exist: %d", hash)
	}

	if config.TrashBin == "" {
		// delete image
		err := os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to delete file: %v", err)
		}
	} else {
		// move image otherwise
		filename := filepath.Base(path)
		destinationPath := filepath.Join(config.TrashBin, filename)

		// Check if source and destination are on the same device
		source := filepath.Dir(path)
		dest := filepath.Dir(destinationPath)
		srcStat, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf("failed to get source directory information: %v", err)
		}
		destStat, err := os.Stat(dest)
		if err != nil {
			return fmt.Errorf("failed to get destination directory information: %v", err)
		}
		if !os.SameFile(srcStat, destStat) {
			// Source and destination are on different devices, so copy and remove
			srcFile, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open source file: %v", err)
			}
			defer srcFile.Close()

			destFile, err := os.Create(destinationPath)
			if err != nil {
				return fmt.Errorf("failed to create destination file: %v", err)
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				return fmt.Errorf("failed to copy file: %v", err)
			}

			err = os.Remove(path)
			if err != nil {
				return fmt.Errorf("failed to remove original file: %v", err)
			}
		} else {
			// Source and destination are on the same device, so simply rename
			err := os.Rename(path, destinationPath)
			if err != nil {
				return fmt.Errorf("failed to move file to %s: %v", config.TrashBin, err)
			}
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

		// Retrieve file information
		fileInfo, err := os.Stat(path)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get file information: %v", err), http.StatusInternalServerError)
			return
		}

		size := fmt.Sprintf("%d bytes", fileInfo.Size())
		createTime := fileInfo.ModTime().Format("2006-01-02 15:04:05")

		imageData := WebImageData{
			Mime:       mimeType,
			Data:       data,
			Size:       size,
			CreateTime: createTime,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(imageData)
	} else if strings.HasPrefix(r.URL.Path, "/delete") {
		// Handle image deletion
		path := r.URL.Query().Get("path")
		hashString := r.URL.Query().Get("hash")
		if path == "" {
			http.Error(w, "missing file path", http.StatusBadRequest)
			return
		}
		if hashString == "" {
			http.Error(w, "missing hash value", http.StatusBadRequest)
			return
		}

		hash, err := strconv.ParseUint(hashString, 10, 64)
		if err != nil {
			http.Error(w, "not a valid hash value", http.StatusBadRequest)
			return
		}

		// Perform deletion operation
		err = deleteImage(path, hash, config)
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
			fmt.Printf("failed to delete image %s: %v\n", path, err)
			return
		}

		// update hash map
		mu.Lock()
		var updatedFiles []string
		for _, file := range imageHashes[hash] {
			if file != path {
				updatedFiles = append(updatedFiles, file)
			}
		}
		imageHashes[hash] = updatedFiles
		mu.Unlock()

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
	} else if strings.HasPrefix(r.URL.Path, "/autodelete") {
		// Handle deletion of all images (pick largest file size as always)

		// Perform best effort deletion operation
		err := deleteAllImages(config)
		if err != nil {
			// Send error response as JSON
			jsonResponse := map[string]string{"error": fmt.Sprintf("failed to delete all images: %v", err)}
			jsonResponseBytes, err := json.Marshal(jsonResponse)
			if err != nil {
				http.Error(w, "failed to marshal JSON response", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonResponseBytes)
			fmt.Printf("failed to delete all images: %v\n", err)
			return
		}

		// Send success response as JSON
		jsonResponse := map[string]string{"message": "images deletion operation done"}
		jsonResponseBytes, err := json.Marshal(jsonResponse)
		if err != nil {
			http.Error(w, "failed to marshal JSON response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponseBytes)
		fmt.Println("images deleted successfully")
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
