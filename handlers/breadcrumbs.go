package handlers

type Breadcrumb struct {
	Label  string // Label displayed on Link
	URL    string // URL
	IsLast bool   // If it's the last(current page)
}

// Page breadcrumbs
type Breadcrumbs []Breadcrumb
