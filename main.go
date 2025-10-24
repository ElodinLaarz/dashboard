package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"sort"
	"strings"
	"unicode"
)

// Formatting utilities

// formatTitle converts a string to title case (e.g., "hello world" -> "Hello World")
func formatTitle(s string) string {
	if s == "" {
		return s
	}

	// Handle common acronyms and special cases
	specialCases := map[string]string{
		"id":  "ID",
		"url": "URL",
	}

	if val, ok := specialCases[strings.ToLower(s)]; ok {
		return val
	}

	// Convert to title case
	prev := ' '
	return strings.Map(
		func(r rune) rune {
			if unicode.IsSpace(prev) || prev == '-' || prev == '_' {
				prev = r
				return unicode.ToTitle(r)
			}
			prev = r
			return unicode.ToLower(r)
		},
		s,
	)
}

// formatPlural handles simple pluralization (adds 's' if count != 1)
func formatPlural(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	if plural == "" {
		plural = singular + "s"
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// formatColor formats a color string for display
func formatColor(color string) string {
	color = strings.TrimSpace(strings.ToLower(color))
	if color == "" {
		return "Unknown"
	}
	return formatTitle(color)
}

// formatShape formats a shape string for display
func formatShape(shape string) string {
	shape = strings.TrimSpace(strings.ToLower(shape))
	shapeMap := map[string]string{
		"square":   "Square",
		"circle":   "Circle",
		"triangle": "Triangle",
	}

	if formatted, ok := shapeMap[shape]; ok {
		return formatted
	}
	return formatTitle(shape)
}

//go:embed templates/* static/*
var embedFS embed.FS

// Item represents an item with multiple properties
type Item struct {
	ID       int    `json:"id"`
	Color    string `json:"color"`
	Shape    string `json:"shape"`
	Category string `json:"category"`
}

// Validate checks if the item has valid field values
func (i Item) Validate() error {
	if i.ID <= 0 {
		return fmt.Errorf("invalid item ID: %d", i.ID)
	}

	i.Color = strings.TrimSpace(i.Color)
	i.Shape = strings.TrimSpace(i.Shape)
	i.Category = strings.TrimSpace(i.Category)

	if i.Color == "" || i.Shape == "" || i.Category == "" {
		return fmt.Errorf("item %d has empty fields", i.ID)
	}

	return nil
}

// Format formats the item's fields for display
func (i Item) Format() Item {
	return Item{
		ID:       i.ID,
		Color:    formatColor(i.Color),
		Shape:    formatShape(i.Shape),
		Category: formatTitle(i.Category),
	}
}

// items is a collection of items
// Note: In a production environment, consider using a database
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
	// Initialize and validate items
	for i, item := range items {
		if err := item.Validate(); err != nil {
			log.Fatalf("Invalid item at index %d: %v", i, err)
		}
	}

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

	http.HandleFunc("/", logRequest(indexHandler))
	http.HandleFunc("/items", logRequest(itemsHandler))

	log.Println("Server starting on http://localhost:8080")
	// Start server
	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
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

// getUniqueValues returns a sorted slice of unique values for a given property
func getUniqueValues(items []Item, getter PropertyGetter) []string {
	valueMap := make(map[string]bool)
	var values []string

	for _, item := range items {
		value := getter(item)
		if !valueMap[value] {
			valueMap[value] = true
			values = append(values, value)
		}
	}

	sort.Strings(values)
	return values
}

// logRequest is a middleware that logs HTTP requests
func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	}
}

// writeJSON writes a JSON response with proper headers
func writeJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// writeError writes an error response in JSON format
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// itemsHandler handles requests to the /items endpoint
func itemsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
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

	// Get unique values for sidebar
	colorGetter, _ := getPropertyGetter("color")
	shapeGetter, _ := getPropertyGetter("shape")
	categoryGetter, _ := getPropertyGetter("category")

	allColors := getUniqueValues(items, colorGetter)
	allShapes := getUniqueValues(items, shapeGetter)
	allCategories := getUniqueValues(items, categoryGetter)

	// Create template with custom functions
	funcMap := template.FuncMap{
		"title":    strings.Title,
		"multiply": func(a int, b float64) float64 { return float64(a) * b },
	}

	// Parse and execute template
	tmpl, err := template.New("items.html").Funcs(funcMap).ParseFS(embedFS, "templates/items.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Groups       []GroupedItems
		GroupBy      string
		AllColors    []string
		AllShapes    []string
		AllCategories []string
	}{
		Groups:       groupedItems,
		GroupBy:      groupBy,
		AllColors:    allColors,
		AllShapes:    allShapes,
		AllCategories: allCategories,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
