package itemstore

import (
	"fmt"
	"sort"
	"strings"
)

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

// ItemStore handles storage and retrieval of items
type ItemStore struct {
	items []Item
}

// New creates a new ItemStore with the given items
func New(items []Item) (*ItemStore, error) {
	// Validate all items
	for i, item := range items {
		if err := item.Validate(); err != nil {
			return nil, fmt.Errorf("invalid item at index %d: %w", i, err)
		}
	}

	return &ItemStore{
		items: items,
	}, nil
}

// Filter applies the given filters to the items and returns the result
func (s *ItemStore) Filter(filters map[string]string) []Item {
	if len(filters) == 0 {
		// Return a copy of all items
		result := make([]Item, len(s.items))
		copy(result, s.items)
		return result
	}

	var result []Item

	// For each item, check if it matches all filters
ItemLoop:
	for _, item := range s.items {
		for key, value := range filters {
			switch key {
			case "color":
				if item.Color != value {
					continue ItemLoop
				}
			case "shape":
				if item.Shape != value {
					continue ItemLoop
				}
			case "category":
				if item.Category != value {
					continue ItemLoop
				}
			}
		}
		// If we get here, the item matches all filters
		result = append(result, item)
	}

	return result
}

// GetUniqueValues returns all unique values for a given property
func (s *ItemStore) GetUniqueValues(property string) []string {
	values := make(map[string]struct{})
	var result []string

	for _, item := range s.items {
		var value string
		switch property {
		case "color":
			value = item.Color
		case "shape":
			value = item.Shape
		case "category":
			value = item.Category
		default:
			continue
		}

		if _, exists := values[value]; !exists {
			values[value] = struct{}{}
			result = append(result, value)
		}
	}

	sort.Strings(result)
	return result
}

// formatTitle converts a string to title case
func formatTitle(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// formatColor formats a color string for display
func formatColor(color string) string {
	return strings.ToLower(color)
}

// formatShape formats a shape string for display
func formatShape(shape string) string {
	shape = strings.ToLower(shape)
	switch shape {
	case "triangle":
		return "triangle"
	case "circle":
		return "circle"
	case "square":
		return "square"
	default:
		return shape
	}
}
