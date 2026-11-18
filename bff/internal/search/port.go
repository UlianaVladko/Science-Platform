// Package search defines the search port. The C++ "industrial search" team can
// implement Engine and be swapped in without touching handlers, mirroring the
// port/adapter style used by the upstream FastAPI service.
package search

import (
	"context"

	"scienceplatform/bff/internal/domain"
)

// Query is the normalized search request after validation.
type Query struct {
	Text          string
	EntityType    string // publication | author
	Mode          string // lexical | semantic | hybrid | similar
	SimilarTo     string
	CategoryIDs   []string
	YearFrom      *int
	YearTo        *int
	Types         []string
	Language      string
	Author        string
	Organization  string
	Journal       string
	Terms         []string
	Sort          string
	Page          int
	PageSize      int
	IncludeFacets bool
}

// Result is what an engine returns: ranked hits plus pre-pagination facets.
type Result struct {
	Hits   []Hit
	Facets []domain.SearchFacet
	Total  int
}

// Hit is a single ranked match.
type Hit struct {
	Publication *domain.Publication   `json:"publication,omitempty"`
	Author      *domain.AuthorSummary `json:"author,omitempty"`
	Score       float64               `json:"score"`
	Snippets    []string              `json:"snippets"`
}

// Engine is the port implemented by the in-BFF mock and, later, by the C++ service.
type Engine interface {
	Search(ctx context.Context, q Query) (Result, error)
	Suggest(ctx context.Context, text string, limit int) ([]domain.Suggestion, error)
	Name() string
}
