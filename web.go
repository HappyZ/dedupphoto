package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func mimeType(path string) string {
	return mime.TypeByExtension(filepath.Ext(path))
}

func base64encode(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return "" // Handle error
	}
	return base64.StdEncoding.EncodeToString(data)
}

func handler(w http.ResponseWriter, r *http.Request, jsonFilepath string) {
	data, err := os.ReadFile(jsonFilepath)
	if err != nil {
		fmt.Fprintf(w, "failed to read file %s: %v", jsonFilepath, err)
		return
	}

	var jsonData []DuplicatedImageData
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		fmt.Fprintf(w, "failed to parse json file %s: %v", jsonFilepath, err)
		return
	}

	// Parse and serve HTML template
	tmpl := template.New("index.html").Funcs(template.FuncMap{"mimeType": mimeType, "base64encode": base64encode})
	tmpl, err = tmpl.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Fprintf(w, "failed to parse html template: %v", err)
		return
	}

	err = tmpl.Execute(w, jsonData)
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "Error executing template: %v", err)
		return
	}
}

func webserverHandler(jsonFilepath string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, jsonFilepath)
	})
	fmt.Println("Server listening on port 8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
