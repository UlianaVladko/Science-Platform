package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"scienceplatform/bff/internal/corpus"
	"scienceplatform/bff/internal/upstream"
)

func testCatalogServer(t *testing.T) *Server {
	t.Helper()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("parent_id") == "cat-cs" {
			_, _ = w.Write([]byte(`[
				{"id":"cat-nlp","slug":"nlp","name":"Natural language processing","publications_count":3,"children":[]},
				{"id":"cat-ml","slug":"machine-learning","name":"Machine learning","publications_count":1,"children":[]}
			]`))
			return
		}
		_, _ = w.Write([]byte(`[{"id":"cat-cs","slug":"computer-science","name":"Computer science","publications_count":4,"children":[]}]`))
	}))
	t.Cleanup(api.Close)
	return &Server{Upstream: upstream.New(api.URL), Corpus: corpus.Generate()}
}

func TestDisciplinesIncludesFullDemoTreeAndAPIOverlap(t *testing.T) {
	s := testCatalogServer(t)
	r := httptest.NewRequest(http.MethodGet, "/api/disciplines", nil)
	w := httptest.NewRecorder()
	s.handleDisciplines(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var response struct {
		Items []struct {
			ID                 string `json:"id"`
			Source             string `json:"source"`
			SubcategoriesCount int    `json:"subcategories_count"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 15 {
		t.Fatalf("got %d disciplines, want 15", len(response.Items))
	}
	totalSubcategories := 0
	for _, item := range response.Items {
		totalSubcategories += item.SubcategoriesCount
	}
	if totalSubcategories != 38 {
		t.Fatalf("got %d subcategories, want 38", totalSubcategories)
	}
	if response.Items[0].ID != "cat-cs" || response.Items[0].Source != "api+demo" {
		t.Fatalf("unexpected merged root: %+v", response.Items[0])
	}
	if response.Items[0].SubcategoriesCount != 6 {
		t.Fatalf("computer science subcategory count = %d, want 6", response.Items[0].SubcategoriesCount)
	}
}

func TestSubcategoriesMergesAPICategoriesWithoutDroppingDemoOnlyItems(t *testing.T) {
	s := testCatalogServer(t)
	mux := s.Routes()
	r := httptest.NewRequest(http.MethodGet, "/api/disciplines/computer-science/subcategories", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	var response struct {
		Items []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Source string `json:"source"`
		} `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 6 {
		t.Fatalf("got %d subcategories, want 6", len(response.Items))
	}
	if response.Items[0].ID != "cat-nlp" || response.Items[0].Source != "api+demo" {
		t.Fatalf("expected API/demo overlap for NLP, got %+v", response.Items[0])
	}
	if response.Items[2].ID != "cat-db" || response.Items[2].Source != "demo" {
		t.Fatalf("expected demo-only databases category, got %+v", response.Items[2])
	}
}
