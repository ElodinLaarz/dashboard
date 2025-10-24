package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

//go:embed templates/* static/*
var embedFS embed.FS

// Item represents an item with multiple properties
type Item struct {
	ID       int
	Color    string
	Shape    string
	Category string
}

// Items is a collection of items
var items = []Item{
	{ID: 1, Color: "blue", Shape: "square", Category: "A"},
	{ID: 2, Color: "red", Shape: "circle", Category: "B"},
	{ID: 3, Color: "green", Shape: "triangle", Category: "C"},
	{ID: 4, Color: "blue", Shape: "circle", Category: "B"},
	{ID: 5, Color: "red", Shape: "square", Category: "A"},
	{ID: 6, Color: "green", Shape: "circle", Category: "C"},
	{ID: 7, Color: "blue", Shape: "triangle", Category: "C"},
	{ID: 8, Color: "red", Shape: "triangle", Category: "A"},
	{ID: 9, Color: "green", Shape: "square", Category: "B"},
	{ID: 10, Color: "blue", Shape: "square", Category: "C"},
	{ID: 11, Color: "red", Shape: "circle", Category: "B"},
	{ID: 12, Color: "green", Shape: "triangle", Category: "A"},
}

func main() {
	// Serve static files
	staticFS, err := fs.Sub(embedFS, "static")
	if err != nil {
		log.Fatalf("Failed to get static directory from embedded filesystem: %v", err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/items", itemsHandler)

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(embedFS, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Items []Item
	}{
		Items: items,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	// Get filter parameters
	filterBy := r.URL.Query().Get("filterBy")
	filterValue := r.URL.Query().Get("filterValue")

	// Filter and sort items
	filteredItems := make([]Item, 0)
	
	if filterBy != "" && filterValue != "" {
		// Sort by matching items first
		for _, item := range items {
			matches := false
			switch filterBy {
			case "color":
				matches = item.Color == filterValue
			case "shape":
				matches = item.Shape == filterValue
			case "category":
				matches = item.Category == filterValue
			}
			
			if matches {
				filteredItems = append(filteredItems, item)
			}
		}
		
		// Then add non-matching items
		for _, item := range items {
			matches := false
			switch filterBy {
			case "color":
				matches = item.Color == filterValue
			case "shape":
				matches = item.Shape == filterValue
			case "category":
				matches = item.Category == filterValue
			}
			
			if !matches {
				filteredItems = append(filteredItems, item)
			}
		}
	} else {
		filteredItems = items
	}

	// Parse and execute template
	tmpl, err := template.ParseFS(embedFS, "templates/items.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Items []Item
	}{
		Items: filteredItems,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
