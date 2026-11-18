package upstream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"scienceplatform/bff/internal/domain"
)

// ErrUpstreamUnavailable signals that the FastAPI service could not be reached;
// callers degrade to the demo corpus instead of failing the request.
var ErrUpstreamUnavailable = errors.New("upstream unavailable")

type APIError struct {
	Status int
	Code   string
	Body   domain.ErrorResponse
}

func (e *APIError) Error() string {
	return fmt.Sprintf("upstream %d %s", e.Status, e.Code)
}

// Client wraps the 20 public operations of the FastAPI literature service.
type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	trimmed := strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: trimmed + "/api/v1",
		http: &http.Client{
			Timeout: 6 * time.Second,
			Transport: &http.Transport{
				DialContext:         (&net.Dialer{Timeout: 2 * time.Second}).DialContext,
				MaxIdleConnsPerHost: 16,
			},
		},
	}
}

func (c *Client) BaseURL() string { return c.baseURL }

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	target := c.baseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}

	if resp.StatusCode >= 400 {
		var envelope domain.ErrorResponse
		if json.Unmarshal(payload, &envelope) == nil && envelope.Error.Code != "" {
			return &APIError{Status: resp.StatusCode, Code: envelope.Error.Code, Body: envelope}
		}
		return &APIError{Status: resp.StatusCode, Code: fmt.Sprintf("HTTP_%d", resp.StatusCode),
			Body: domain.ErrorResponse{Error: domain.ErrorDetail{
				Code:    fmt.Sprintf("HTTP_%d", resp.StatusCode),
				Message: strings.TrimSpace(string(payload)),
			}}}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("%w: decode: %v", ErrUpstreamUnavailable, err)
	}
	return nil
}

func (c *Client) Health(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/categories", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	return resp.StatusCode == http.StatusOK
}

// --- Classification ---

func (c *Client) Categories(ctx context.Context, parentID string, includeCounts bool) ([]domain.Category, error) {
	q := url.Values{}
	q.Set("include_counts", fmt.Sprintf("%t", includeCounts))
	if parentID != "" {
		q.Set("parent_id", parentID)
	}
	var out []domain.Category
	if err := c.do(ctx, http.MethodGet, "/categories", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// --- Search ---

type SearchRequest struct {
	Query                  *string        `json:"query"`
	EntityType             string         `json:"entity_type"`
	Mode                   string         `json:"mode"`
	SimilarToPublicationID *string        `json:"similar_to_publication_id,omitempty"`
	Filters                map[string]any `json:"filters"`
	Sort                   string         `json:"sort"`
	Page                   int            `json:"page"`
	PageSize               int            `json:"page_size"`
	IncludeFacets          bool           `json:"include_facets"`
}

type upstreamPage[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (c *Client) Search(ctx context.Context, req SearchRequest) ([]any, []domain.SearchFacet, int, error) {
	var raw struct {
		Items    []json.RawMessage    `json:"items"`
		Total    int                  `json:"total"`
		Page     int                  `json:"page"`
		PageSize int                  `json:"page_size"`
		Facets   []domain.SearchFacet `json:"facets"`
	}
	if err := c.do(ctx, http.MethodPost, "/search", nil, req, &raw); err != nil {
		return nil, nil, 0, err
	}
	items := make([]any, 0, len(raw.Items))
	for _, r := range raw.Items {
		var hit map[string]any
		if err := json.Unmarshal(r, &hit); err != nil {
			continue
		}
		hit["source"] = string(domain.SourceAPI)
		items = append(items, hit)
	}
	return items, raw.Facets, raw.Total, nil
}

func (c *Client) Suggest(ctx context.Context, query string, limit int) ([]domain.Suggestion, error) {
	q := url.Values{}
	q.Set("q", query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	var raw struct {
		Items []struct {
			Type         string `json:"type"`
			Text         string `json:"text"`
			MatchesCount *int   `json:"matches_count"`
		} `json:"items"`
	}
	if err := c.do(ctx, http.MethodGet, "/search/suggest", q, nil, &raw); err != nil {
		return nil, err
	}
	out := make([]domain.Suggestion, 0, len(raw.Items))
	for _, item := range raw.Items {
		out = append(out, domain.Suggestion{
			Type: item.Type, Text: item.Text, MatchesCount: item.MatchesCount, Source: domain.SourceAPI,
		})
	}
	return out, nil
}

// --- Publications ---

func (c *Client) Publication(ctx context.Context, id string) (*domain.Publication, error) {
	var out domain.Publication
	if err := c.do(ctx, http.MethodGet, "/publications/"+url.PathEscape(id), nil, nil, &out); err != nil {
		return nil, err
	}
	out.Source = domain.SourceAPI
	for i := range out.Categories {
		out.Categories[i].Source = domain.SourceAPI
	}
	return &out, nil
}

func (c *Client) CreatePublication(ctx context.Context, body any) (*domain.Publication, error) {
	var out domain.Publication
	if err := c.do(ctx, http.MethodPost, "/publications", nil, body, &out); err != nil {
		return nil, err
	}
	out.Source = domain.SourceAPI
	return &out, nil
}

func (c *Client) PatchPublication(ctx context.Context, id string, body any) (*domain.Publication, error) {
	var out domain.Publication
	if err := c.do(ctx, http.MethodPatch, "/publications/"+url.PathEscape(id), nil, body, &out); err != nil {
		return nil, err
	}
	out.Source = domain.SourceAPI
	return &out, nil
}

func (c *Client) DeletePublication(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/publications/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) References(ctx context.Context, id, direction string, page, pageSize int) (*domain.Page[domain.Reference], error) {
	q := url.Values{}
	q.Set("direction", direction)
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("page_size", fmt.Sprintf("%d", pageSize))
	var out domain.Page[domain.Reference]
	if err := c.do(ctx, http.MethodGet, "/publications/"+url.PathEscape(id)+"/references", q, nil, &out); err != nil {
		return nil, err
	}
	for i := range out.Items {
		out.Items[i].Source = domain.SourceAPI
	}
	return &out, nil
}

func (c *Client) RegisterView(ctx context.Context, id string) (*domain.ViewCount, error) {
	var out domain.ViewCount
	if err := c.do(ctx, http.MethodPost, "/publications/"+url.PathEscape(id)+"/view", nil, map[string]any{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetFile returns the raw bytes of a publication file.
func (c *Client) GetFile(ctx context.Context, id, kind string) ([]byte, string, error) {
	target := c.baseURL + "/publications/" + url.PathEscape(id) + "/files/" + url.PathEscape(kind)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode >= 400 {
		return nil, "", &APIError{Status: resp.StatusCode, Code: "FILE_ERROR"}
	}
	return payload, resp.Header.Get("Content-Type"), nil
}

func (c *Client) PutFile(ctx context.Context, id, kind, contentType string, content []byte) (map[string]any, error) {
	target := c.baseURL + "/publications/" + url.PathEscape(id) + "/files/" + url.PathEscape(kind)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var out map[string]any
	if err := json.Unmarshal(payload, &out); err != nil {
		return nil, &APIError{Status: resp.StatusCode, Code: "FILE_UPLOAD_FAILED"}
	}
	if resp.StatusCode >= 400 {
		return nil, &APIError{Status: resp.StatusCode, Code: "FILE_UPLOAD_FAILED"}
	}
	return out, nil
}

// --- Authors ---

func (c *Client) Author(ctx context.Context, id string) (*domain.Author, error) {
	var out domain.Author
	if err := c.do(ctx, http.MethodGet, "/authors/"+url.PathEscape(id), nil, nil, &out); err != nil {
		return nil, err
	}
	out.Source = domain.SourceAPI
	return &out, nil
}

// --- Analytics ---

func (c *Client) Summary(ctx context.Context) (*domain.CorpusSummary, error) {
	var out struct {
		TotalPublications          int     `json:"total_publications"`
		TotalAuthors               int     `json:"total_authors"`
		TotalCitations             int     `json:"total_citations"`
		TotalDownloads             int     `json:"total_downloads"`
		AvgCitationsPerPublication float64 `json:"avg_citations_per_publication"`
		PublicationsThisPeriod     int     `json:"publications_this_period"`
		NewAuthorsThisPeriod       int     `json:"new_authors_this_period"`
	}
	if err := c.do(ctx, http.MethodGet, "/analytics/summary", nil, nil, &out); err != nil {
		return nil, err
	}
	return &domain.CorpusSummary{
		TotalPublications:          out.TotalPublications,
		TotalAuthors:               out.TotalAuthors,
		TotalCitations:             out.TotalCitations,
		TotalDownloads:             out.TotalDownloads,
		AvgCitationsPerPublication: out.AvgCitationsPerPublication,
		PublicationsThisPeriod:     out.PublicationsThisPeriod,
		NewAuthorsThisPeriod:       out.NewAuthorsThisPeriod,
		Period:                     "2026",
	}, nil
}

func (c *Client) Top(ctx context.Context, query url.Values) (*domain.Ranking, error) {
	var out domain.Ranking
	if err := c.do(ctx, http.MethodGet, "/analytics/top", query, nil, &out); err != nil {
		return nil, err
	}
	out.Source = domain.SourceAPI
	return &out, nil
}

func (c *Client) Graph(ctx context.Context, query url.Values) (*domain.Graph, error) {
	var out struct {
		Nodes     []domain.GraphNode `json:"nodes"`
		Edges     []domain.GraphEdge `json:"edges"`
		Truncated bool               `json:"truncated"`
		DataAsOf  string             `json:"data_as_of"`
	}
	if err := c.do(ctx, http.MethodGet, "/graphs", query, nil, &out); err != nil {
		return nil, err
	}
	out2 := domain.Graph{Nodes: out.Nodes, Edges: out.Edges, Truncated: out.Truncated, DataAsOf: out.DataAsOf, Source: domain.SourceAPI}
	return &out2, nil
}

// --- Analysis jobs ---

func (c *Client) StartJob(ctx context.Context, body any) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodPost, "/analysis/jobs", nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Job(ctx context.Context, id string) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/jobs/"+url.PathEscape(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Runs(ctx context.Context, page, pageSize int) (map[string]any, error) {
	q := url.Values{}
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("page_size", fmt.Sprintf("%d", pageSize))
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/analysis/clustering/runs", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Run(ctx context.Context, id string) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/analysis/clustering/runs/"+url.PathEscape(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
