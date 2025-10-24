package itemstore

import (
	"testing"
)

var testItems = []Item{
	{ID: 1, Color: "red", Shape: "circle", Category: "A"},
	{ID: 2, Color: "blue", Shape: "square", Category: "A"},
	{ID: 3, Color: "red", Shape: "square", Category: "B"},
	{ID: 4, Color: "green", Shape: "circle", Category: "B"},
}

func TestItemStore_Filter(t *testing.T) {
	tests := []struct {
		name    string
		filters map[string]string
		wantLen int
	}{
		{
			name:    "no filters",
			filters: map[string]string{},
			wantLen: 4,
		},
		{
			name:    "filter by color",
			filters: map[string]string{"color": "red"},
			wantLen: 2,
		},
		{
			name:    "filter by shape",
			filters: map[string]string{"shape": "square"},
			wantLen: 2,
		},
		{
			name:    "filter by multiple properties",
			filters: map[string]string{"color": "red", "shape": "square"},
			wantLen: 1,
		},
		{
			name:    "filter by non-existent value",
			filters: map[string]string{"color": "purple"},
			wantLen: 0,
		},
	}

	store, err := New(testItems)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := store.Filter(tt.filters)
			if len(got) != tt.wantLen {
				t.Errorf("Filter() = %v items, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestItemStore_GetUniqueValues(t *testing.T) {
	tests := []struct {
		name     string
		property string
		want     []string
	}{
		{
			name:     "get unique colors",
			property: "color",
			want:     []string{"blue", "green", "red"},
		},
		{
			name:     "get unique shapes",
			property: "shape",
			want:     []string{"circle", "square"},
		},
		{
			name:     "get unique categories",
			property: "category",
			want:     []string{"A", "B"},
		},
		{
			name:     "invalid property",
			property: "invalid",
			want:     []string{},
		},
	}

	store, err := New(testItems)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := store.GetUniqueValues(tt.property)
			if len(got) != len(tt.want) {
				t.Fatalf("GetUniqueValues() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetUniqueValues()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestItem_Validate(t *testing.T) {
	tests := []struct {
		name    string
		item    Item
		wantErr bool
	}{
		{
			name:    "valid item",
			item:    Item{ID: 1, Color: "red", Shape: "circle", Category: "A"},
			wantErr: false,
		},
		{
			name:    "missing ID",
			item:    Item{ID: 0, Color: "red", Shape: "circle", Category: "A"},
			wantErr: true,
		},
		{
			name:    "missing color",
			item:    Item{ID: 1, Color: "", Shape: "circle", Category: "A"},
			wantErr: true,
		},
		{
			name:    "missing shape",
			item:    Item{ID: 1, Color: "red", Shape: "", Category: "A"},
			wantErr: true,
		},
		{
			name:    "missing category",
			item:    Item{ID: 1, Color: "red", Shape: "circle", Category: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Item.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
