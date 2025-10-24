# Interactive Item Dashboard

An interactive htmx-powered dashboard built with Go that allows dynamic organization of items by their properties.

## Features

- **Interactive Reorganization**: Hover over any item to reorganize the entire grid based on that item's properties
- **Multiple Properties**: Items have three properties:
  - Color: blue, red, green
  - Shape: square, circle, triangle  
  - Category: A, B, C
- **htmx Integration**: Lightweight, server-side rendering with minimal JavaScript
- **Responsive Grid Layout**: Clean, modern UI with visual feedback

## How It Works

When you hover over an item, the server filters and sorts all items to prioritize those matching the hovered item's color. Items with the same color are moved to the front of the grid, creating an interactive and dynamic organization experience.

## Running the Application

### Prerequisites

- Go 1.16 or higher

### Installation

1. Clone the repository
2. Navigate to the project directory
3. Run the server:

```bash
go run main.go
```

4. Open your browser to `http://localhost:8080`

## Project Structure

```
dashboard/
├── main.go                 # Go server with htmx handlers
├── templates/
│   ├── index.html         # Main page template
│   └── items.html         # Partial template for items
├── static/
│   └── htmx.min.js        # Minimal htmx implementation
└── README.md
```

## Technical Details

- **Backend**: Go with html/template for server-side rendering
- **Frontend**: htmx for dynamic interactions without full page reloads
- **Styling**: Modern CSS with gradients and transitions
- **Data**: In-memory data structure with 12 sample items

## API Endpoints

- `GET /` - Main dashboard page
- `GET /items?filterBy=color&filterValue={value}` - Get items filtered by property
- `GET /static/htmx.min.js` - htmx JavaScript library

## Screenshots

![Interactive Dashboard](https://github.com/user-attachments/assets/f4adf052-a031-4d4f-b37d-89818ad8b0b0)

*Items reorganized by color when hovering over a red item*

## License

MIT
