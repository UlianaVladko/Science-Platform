// Package domain holds transport-agnostic read models shared by every BFF layer.
package domain

// Source marks where an entity came from, so the UI can badge demo data honestly.
type Source string

const (
	SourceAPI  Source = "api"
	SourceDemo Source = "demo"
	SourceBFF  Source = "bff"
)

const (
	PublicationTypes = "article|journal_publication|book|monograph|preprint|other"
)

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Category struct {
	ID                string     `json:"id"`
	Slug              string     `json:"slug"`
	Name              string     `json:"name"`
	Description       *string    `json:"description"`
	ParentID          *string    `json:"parent_id"`
	PublicationsCount *int       `json:"publications_count"`
	Children          []Category `json:"children"`
	Icon              string     `json:"icon,omitempty"`
	Accent            string     `json:"accent,omitempty"`
	Source            Source     `json:"source,omitempty"`
}

type AuthorSummary struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Orcid        *string        `json:"orcid"`
	Affiliations []Organization `json:"affiliations"`
}

type Author struct {
	AuthorSummary
	Categories        []Category `json:"categories"`
	PublicationsCount int        `json:"publications_count"`
	CitationsCount    int        `json:"citations_count"`
	HIndex            *int       `json:"h_index"`
	Source            Source     `json:"source,omitempty"`
}

type Metrics struct {
	Citations int `json:"citations"`
	Views     int `json:"views"`
	Downloads int `json:"downloads"`
}

type ExternalIdentifier struct {
	Scheme string `json:"scheme"`
	Value  string `json:"value"`
}

type Concept struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PublicationFile struct {
	Kind        string `json:"kind"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
}

type PublicationSummary struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Abstract        *string         `json:"abstract"`
	PublicationType string          `json:"publication_type"`
	PublicationDate *string         `json:"publication_date"`
	Language        string          `json:"language"`
	Authors         []AuthorSummary `json:"authors"`
	Metrics         Metrics         `json:"metrics"`
	Source          Source          `json:"source,omitempty"`
}

type Publication struct {
	PublicationSummary
	ExternalIdentifiers []ExternalIdentifier `json:"external_identifiers"`
	Journal             *string              `json:"journal"`
	Publisher           *string              `json:"publisher"`
	Categories          []Category           `json:"categories"`
	Keywords            []string             `json:"keywords"`
	Concepts            []Concept            `json:"concepts"`
	Files               []PublicationFile    `json:"files"`
	ProcessingStatus    string               `json:"processing_status"`
	LikesCount          int                  `json:"likes_count"`
	Liked               bool                 `json:"liked"`
	DOI                 *string              `json:"doi,omitempty"`
}

type Reference struct {
	CitingPublicationID string              `json:"citing_publication_id"`
	CitedPublicationID  *string             `json:"cited_publication_id"`
	RawReference        string              `json:"raw_reference"`
	Publication         *PublicationSummary `json:"publication"`
	Source              Source              `json:"source,omitempty"`
}

type Page[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type FacetBucket struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type SearchFacet struct {
	Field   string            `json:"field"`
	Buckets []FacetBucket     `json:"buckets"`
	Labels  map[string]string `json:"labels,omitempty"`
}

type PublicationHit struct {
	EntityType string             `json:"entity_type"`
	Score      float64            `json:"score"`
	Entity     PublicationSummary `json:"entity"`
	Snippets   []string           `json:"snippets"`
	Source     Source             `json:"source"`
}

type AuthorHit struct {
	EntityType string        `json:"entity_type"`
	Score      float64       `json:"score"`
	Entity     AuthorSummary `json:"entity"`
	Snippets   []string      `json:"snippets"`
	Source     Source        `json:"source"`
}

type SearchResponse struct {
	Items    []any         `json:"items"`
	Facets   []SearchFacet `json:"facets"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	TookMs   int64         `json:"took_ms"`
	Degraded bool          `json:"degraded"`
}

type Suggestion struct {
	Type         string `json:"type"`
	Text         string `json:"text"`
	ID           string `json:"id,omitempty"`
	MatchesCount *int   `json:"matches_count"`
	Source       Source `json:"source"`
}

type CorpusSummary struct {
	TotalPublications          int     `json:"total_publications"`
	TotalAuthors               int     `json:"total_authors"`
	TotalCitations             int     `json:"total_citations"`
	TotalDownloads             int     `json:"total_downloads"`
	AvgCitationsPerPublication float64 `json:"avg_citations_per_publication"`
	PublicationsThisPeriod     int     `json:"publications_this_period"`
	NewAuthorsThisPeriod       int     `json:"new_authors_this_period"`
	Period                     string  `json:"period"`
	DataAsOf                   string  `json:"data_as_of"`
}

type RankingObject struct {
	EntityType string `json:"entity_type"`
	ID         string `json:"id"`
	Label      string `json:"label"`
}

type RankingItem struct {
	Position int           `json:"position"`
	Object   RankingObject `json:"object"`
	Value    float64       `json:"value"`
}

type Ranking struct {
	Metric string        `json:"metric"`
	Items  []RankingItem `json:"items"`
	Source Source        `json:"source,omitempty"`
}

type GraphNode struct {
	ID       string  `json:"id"`
	Type     string  `json:"type"`
	Label    string  `json:"label"`
	Category string  `json:"category,omitempty"`
	Weight   float64 `json:"weight,omitempty"`
}

type GraphEdge struct {
	Source string   `json:"source"`
	Target string   `json:"target"`
	Type   string   `json:"type"`
	Weight *float64 `json:"weight"`
}

type Graph struct {
	Nodes     []GraphNode `json:"nodes"`
	Edges     []GraphEdge `json:"edges"`
	Truncated bool        `json:"truncated"`
	DataAsOf  string      `json:"data_as_of"`
	Source    Source      `json:"source,omitempty"`
}

type ViewCount struct {
	PublicationID string `json:"publication_id"`
	Views         int    `json:"views"`
}

type LikeState struct {
	PublicationID string `json:"publication_id"`
	Liked         bool   `json:"liked"`
	LikesCount    int    `json:"likes_count"`
}

type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   any    `json:"details"`
	RequestID string `json:"request_id"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
