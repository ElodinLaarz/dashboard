package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"sort"
	"strings"
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

// PropertyGetter is a function that gets a property value from an Item
type PropertyGetter func(Item) string

// GroupedItems represents items grouped by a specific property
type GroupedItems struct {
	GroupName string
	Property  string
	Items     []Item
}

// getPropertyGetter returns the appropriate property getter function
func getPropertyGetter(property string) (PropertyGetter, bool) {
	switch property {
	case "color":
		return func(i Item) string { return i.Color }, true
	case "shape":
		return func(i Item) string { return i.Shape }, true
	case "category":
		return func(i Item) string { return i.Category }, true
	default:
		return nil, false
	}
}

// groupItems groups items by the specified property
func groupItems(items []Item, groupBy string) []GroupedItems {
	getter, ok := getPropertyGetter(groupBy)
	if !ok {
		// If property is invalid, return all items in a single group
		return []GroupedItems{{
			GroupName: "All Items",
			Property:  groupBy,
			Items:     items,
		}}
	}

	// Group items by the specified property
	groupMap := make(map[string][]Item)
	for _, item := range items {
		value := getter(item)
		groupMap[value] = append(groupMap[value], item)
	}

	// Convert to slice of GroupedItems
	var result []GroupedItems
	for value, items := range groupMap {
		result = append(result, GroupedItems{
			GroupName: value,
			Property:  groupBy,
			Items:     items,
		})
	}

	// Sort groups by group name for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return result[i].GroupName < result[j].GroupName
	})

	return result
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	// Get filter parameters
	filterBy := r.URL.Query().Get("filterBy")
	filterValue := r.URL.Query().Get("filterValue")
	groupBy := r.URL.Query().Get("groupBy")

	if groupBy == "" {
		groupBy = "shape" // Default grouping by shape
	}

	// Filter items
	var filteredItems []Item
	if filterBy != "" && filterValue != "" {
		getter, ok := getPropertyGetter(filterBy)
		if ok {
			for _, item := range items {
				if getter(item) == filterValue {
					filteredItems = append(filteredItems, item)
				}
			}
		}
	} else {
		filteredItems = items
	}

	// Group items by the specified property
	groupedItems := groupItems(filteredItems, groupBy)

	// Create template with custom functions
	funcMap := template.FuncMap{
		"title":   strings.Title,
		"multiply": func(a int, b float64) float64 { return float64(a) * b },
	}

	// Parse and execute template
	tmpl, err := template.New("items.html").Funcs(funcMap).ParseFS(embedFS, "templates/items.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Groups  []GroupedItems
		GroupBy string
	}{
		Groups:  groupedItems,
		GroupBy: groupBy,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
