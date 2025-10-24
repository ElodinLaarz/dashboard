package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"unicode"

	"github.com/ElodinLaarz/dashboard/pkg/itemstore"
)

// formatTitle converts a string to title case (e.g., "hello world" -> "Hello World")
func formatTitle(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	for i := 1; i < len(r); i++ {
		r[i] = unicode.ToLower(r[i])
	}
	return string(r)
}

// logRequest logs HTTP requests
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

//go:embed templates/* static/*
var embedFS embed.FS

var store *itemstore.ItemStore

func init() {
	// Initialize the item store with sample data
	sampleItems := []itemstore.Item{
		{ID: 1, Color: "red", Shape: "circle", Category: "A"},
		{ID: 2, Color: "blue", Shape: "square", Category: "A"},
		{ID: 3, Color: "green", Shape: "triangle", Category: "B"},
		{ID: 4, Color: "red", Shape: "square", Category: "B"},
		{ID: 5, Color: "blue", Shape: "circle", Category: "C"},
		{ID: 6, Color: "green", Shape: "square", Category: "C"},
	}

	var err error
	store, err = itemstore.New(sampleItems)
	if err != nil {
		log.Fatalf("Failed to initialize item store: %v", err)
	}
}

// groupItems groups items by the specified property
func groupItems(items []itemstore.Item, groupBy string) map[string][]itemstore.Item {
	grouped := make(map[string][]itemstore.Item)

	switch groupBy {
	case "color":
		for _, item := range items {
			grouped[item.Color] = append(grouped[item.Color], item)
		}
	case "shape":
		for _, item := range items {
			grouped[item.Shape] = append(grouped[item.Shape], item)
		}
	case "category":
		for _, item := range items {
			grouped[item.Category] = append(grouped[item.Category], item)
		}
	default:
		// If the groupBy property is invalid, return all items in a single group
		grouped["All"] = items
	}

	return grouped
}

func main() {
	// Serve static files
	staticFS, err := fs.Sub(embedFS, "static")
	if err != nil {
		log.Fatalf("Failed to get static directory from embedded filesystem: %v", err)
	}

	// Set up HTTP handlers
	http.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.FS(staticFS)),
		),
	)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/items", itemsHandler)

	// Start the server
	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, logRequest(http.DefaultServeMux)))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/items", http.StatusFound)
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
	// Get filter parameters
	groupBy := r.URL.Query().Get("groupBy")
	if groupBy == "" {
		groupBy = "shape" // Default grouping
	}

	// Parse filters from URL
	filters := make(map[string]string)
	for _, filter := range r.URL.Query()["filter"] {
		parts := strings.SplitN(filter, ":", 2)
		if len(parts) == 2 {
			filters[parts[0]] = parts[1]
		}
	}

	// For backward compatibility with old format
	if filterBy := r.URL.Query().Get("filterBy"); filterBy != "" {
		if filterValue := r.URL.Query().Get("filterValue"); filterValue != "" {
			filters[filterBy] = filterValue
		}
	}

	log.Printf("Processing filters: %v", filters)

	// Apply filters
	filteredItems := store.Filter(filters)
	log.Printf("Filtered items count: %d", len(filteredItems))

	// Group items by the specified property
	groupedItems := groupItems(filteredItems, groupBy)

	// Get unique values for sidebar
	uniqueColors := make(map[string]int)
	uniqueShapes := make(map[string]int)
	uniqueCategories := make(map[string]int)

	for _, color := range store.GetUniqueValues("color") {
		uniqueColors[color] = len(store.Filter(map[string]string{"color": color}))
	}

	for _, shape := range store.GetUniqueValues("shape") {
		uniqueShapes[shape] = len(store.Filter(map[string]string{"shape": shape}))
	}

	for _, category := range store.GetUniqueValues("category") {
		uniqueCategories[category] = len(store.Filter(map[string]string{"category": category}))
	}

	// Get all items for animation delays
	allItems := store.Filter(nil)

	// Prepare template data
	data := struct {
		Title           string
		GroupedItems    map[string][]itemstore.Item
		GroupBy         string
		UniqueColors    map[string]int
		UniqueShapes    map[string]int
		UniqueCategories map[string]int
		ActiveFilters   map[string]string
		AllItems       []itemstore.Item
	}{
		Title:           "Dashboard",
		GroupedItems:    groupedItems,
		GroupBy:         groupBy,
		UniqueColors:    uniqueColors,
		UniqueShapes:    uniqueShapes,
		UniqueCategories: uniqueCategories,
		ActiveFilters:   filters,
		AllItems:       allItems,
	}

	// Create a new template with the formatTitle function
	tmpl := template.New("items.html").Funcs(template.FuncMap{
		"title": formatTitle,
	})

	// Parse the template
	tmpl, parseErr := tmpl.ParseFS(embedFS, "templates/items.html")
	if parseErr != nil {
		http.Error(w, parseErr.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}
