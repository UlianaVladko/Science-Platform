package api

import (
	"errors"
	"net/http"
	"sort"
	"strings"

	"scienceplatform/bff/internal/corpus"
	"scienceplatform/bff/internal/domain"
	"scienceplatform/bff/internal/enrich"
	"scienceplatform/bff/internal/search"
	"scienceplatform/bff/internal/synthetic"
	"scienceplatform/bff/internal/upstream"
)

// Server wires the BFF dependencies. Any port can be swapped for a real
// implementation (notably search.Engine for the C++ service).
type Server struct {
	Upstream    *upstream.Client
	Corpus      *corpus.Corpus
	Engine      search.Engine
	Widgets     *synthetic.Service
	State       *enrich.Store
	Version     string
	UpstreamURL string
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/meta", s.handleMeta)

	// Группа 1: каталог и дисциплины
	mux.HandleFunc("GET /api/categories", s.handleCategories)
	mux.HandleFunc("GET /api/disciplines", s.handleDisciplines)
	mux.HandleFunc("GET /api/disciplines/{slug}/subcategories", s.handleSubcategories)
	mux.HandleFunc("GET /api/disciplines/{slug}/fun-fact", s.handleFunFact)

	// Группа 2: поиск
	mux.HandleFunc("POST /api/search", s.handleSearch)
	mux.HandleFunc("GET /api/search/suggest", s.handleSuggest)

	// Группа 3: публикации
	mux.HandleFunc("GET /api/publications/lookup", s.handleLookup)
	mux.HandleFunc("GET /api/publications/{id}", s.handlePublication)
	mux.HandleFunc("POST /api/publications", s.handleCreatePublication)
	mux.HandleFunc("GET /api/publications/{id}/references", s.handleReferences)
	mux.HandleFunc("GET /api/publications/{id}/related", s.handleRelated)
	mux.HandleFunc("POST /api/publications/{id}/view", s.handleView)
	mux.HandleFunc("POST /api/publications/{id}/like", s.handleLike)
	mux.HandleFunc("DELETE /api/publications/{id}/like", s.handleUnlike)
	mux.HandleFunc("GET /api/publications/{id}/files/{kind}", s.handleGetFile)
	mux.HandleFunc("PUT /api/publications/{id}/files/{kind}", s.handlePutFile)

	// Группа 4: авторы
	mux.HandleFunc("GET /api/authors/{id}", s.handleAuthor)
	mux.HandleFunc("GET /api/authors/{id}/stats", s.handleAuthorStats)
	mux.HandleFunc("GET /api/authors/{id}/publications", s.handleAuthorPublications)
	mux.HandleFunc("GET /api/authors/{id}/coauthor-graph", s.handleCoauthorGraph)
	mux.HandleFunc("GET /api/users/{id}/profile", s.handleUserProfile)
	mux.HandleFunc("POST /api/users/{id}/follow", s.handleFollow)
	mux.HandleFunc("DELETE /api/users/{id}/follow", s.handleUnfollow)

	// Группа 5: аналитика и графы
	mux.HandleFunc("GET /api/analytics/summary", s.handleSummary)
	mux.HandleFunc("GET /api/analytics/top", s.handleTop)
	mux.HandleFunc("GET /api/dashboard/{widget}", s.handleWidget)
	mux.HandleFunc("GET /api/graphs", s.handleGraph)

	// Группа 6: Q&A
	mux.HandleFunc("GET /api/questions", s.handleListQuestions)
	mux.HandleFunc("POST /api/questions", s.handleCreateQuestion)
	mux.HandleFunc("GET /api/questions/{id}", s.handleQuestion)
	mux.HandleFunc("POST /api/questions/{id}/answers", s.handleCreateAnswer)
	mux.HandleFunc("POST /api/answers/{id}/vote", s.handleVote)

	return Middleware(mux)
}

// --- Служебные ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "down"
	if s.Upstream.Health(r.Context()) {
		status = "up"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"upstream": status,
		"corpus":   len(s.Corpus.Publications),
		"version":  s.Version,
	})
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"title":   "Научный атлас",
		"version": s.Version,
		"sources": map[string]string{
			"api":  s.UpstreamURL,
			"demo": "встроенный демо-корпус BFF",
			"bff":  "Q&A, лайки, подписки (in-memory)",
		},
		"widgets_ready": s.Widgets.Ready(),
		"search_engine": s.Engine.Name(),
		"corpus_size":   len(s.Corpus.Publications),
		"authors":       len(s.Corpus.Authors),
		"as_of":         corpus.AsOf.Format("2006-01-02T15:04:05Z"),
	})
}

// --- Дисциплины ---

func (s *Server) handleCategories(w http.ResponseWriter, r *http.Request) {
	includeCounts := r.URL.Query().Get("include_counts") != "false"
	parentID := r.URL.Query().Get("parent_id")
	if items, err := s.Upstream.Categories(r.Context(), parentID, includeCounts); err == nil {
		for i := range items {
			items[i].Source = domain.SourceAPI
			for j := range items[i].Children {
				items[i].Children[j].Source = domain.SourceAPI
			}
		}
		writeJSON(w, http.StatusOK, items)
		return
	}
	// Fallback: demo tree with computed counts.
	writeJSON(w, http.StatusOK, s.demoTree(parentID, includeCounts))
}

func (s *Server) demoTree(parentID string, includeCounts bool) []domain.Category {
	counts := map[string]int{}
	for i := range s.Corpus.Publications {
		p := &s.Corpus.Publications[i]
		for _, c := range p.Categories {
			counts[c.ID]++
		}
	}
	decorate := func(c domain.Category, root bool) domain.Category {
		out := c
		out.Children = []domain.Category{}
		if includeCounts {
			n := counts[c.ID]
			out.PublicationsCount = &n
		}
		return out
	}
	out := []domain.Category{}
	for i := range s.Corpus.Categories {
		root := s.Corpus.Categories[i]
		clone := decorate(root, true)
		for j := range root.Children {
			clone.Children = append(clone.Children, decorate(root.Children[j], false))
		}
		if parentID == "" || root.ID == parentID {
			out = append(out, clone)
		}
	}
	return out
}

func (s *Server) handleDisciplines(w http.ResponseWriter, r *http.Request) {
	// The demo classification is the complete navigation taxonomy. Enrich it
	// with upstream records where IDs overlap, but never let a partial API
	// fixture hide demo-only disciplines.
	demo := s.demoTree("", true)
	apiCats, apiErr := s.Upstream.Categories(r.Context(), "", true)
	apiByID := make(map[string]domain.Category, len(apiCats))
	if apiErr == nil {
		for _, c := range apiCats {
			apiByID[c.ID] = c
		}
	}

	items := make([]map[string]any, 0, len(demo)+len(apiCats))
	seen := make(map[string]bool, len(demo)+len(apiCats))
	for _, c := range demo {
		apiCat, hasAPI := apiByID[c.ID]
		count := 0
		if c.PublicationsCount != nil {
			count = *c.PublicationsCount
		}
		if hasAPI && apiCat.PublicationsCount != nil {
			count = *apiCat.PublicationsCount
		}
		source := string(domain.SourceDemo)
		if hasAPI {
			source = "api+demo"
		}
		items = append(items, map[string]any{
			"id": c.ID, "slug": c.Slug, "name": c.Name,
			"icon": c.Icon, "accent": c.Accent,
			"description":         c.Description,
			"publications_count":  count,
			"subcategories_count": len(c.Children),
			"source":              source,
		})
		seen[c.ID] = true
	}
	// Preserve API-only roots as well, should the upstream taxonomy grow.
	for _, c := range apiCats {
		if seen[c.ID] {
			continue
		}
		items = append(items, map[string]any{
			"id": c.ID, "slug": c.Slug, "name": c.Name,
			"icon": iconFor(c.Slug), "accent": accentFor(c.Slug),
			"description":         c.Description,
			"publications_count":  c.PublicationsCount,
			"subcategories_count": len(c.Children),
			"source":              domain.SourceAPI,
		})
	}
	source := string(domain.SourceDemo)
	if apiErr == nil && len(apiCats) > 0 {
		source = "api+demo"
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "source": source})
}

func (s *Server) handleSubcategories(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	cat, ok := s.Corpus.CategoryBySlug[slug]
	if !ok {
		writeError(w, r, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Дисциплина не найдена", nil)
		return
	}

	counts := map[string]int{}
	for i := range s.Corpus.Publications {
		for _, c := range s.Corpus.Publications[i].Categories {
			counts[c.ID]++
		}
	}

	apiChildren, apiErr := s.Upstream.Categories(r.Context(), cat.ID, true)
	apiByID := make(map[string]domain.Category, len(apiChildren))
	if apiErr == nil {
		for _, child := range apiChildren {
			apiByID[child.ID] = child
		}
	}

	items := make([]map[string]any, 0, len(cat.Children)+len(apiChildren))
	seen := make(map[string]bool, len(cat.Children)+len(apiChildren))
	for _, child := range cat.Children {
		count := counts[child.ID]
		source := string(domain.SourceDemo)
		if apiChild, exists := apiByID[child.ID]; exists {
			source = "api+demo"
			if apiChild.PublicationsCount != nil {
				count = *apiChild.PublicationsCount
			}
		}
		items = append(items, map[string]any{
			"id": child.ID, "slug": child.Slug, "name": child.Name,
			"publications_count": count,
			"icon":               cat.Icon, "accent": cat.Accent,
			"source": source,
		})
		seen[child.ID] = true
	}
	for _, child := range apiChildren {
		if seen[child.ID] {
			continue
		}
		count := 0
		if child.PublicationsCount != nil {
			count = *child.PublicationsCount
		}
		items = append(items, map[string]any{
			"id": child.ID, "slug": child.Slug, "name": child.Name,
			"publications_count": count,
			"icon":               cat.Icon, "accent": cat.Accent,
			"source": domain.SourceAPI,
		})
	}

	source := string(domain.SourceDemo)
	if apiErr == nil && len(apiChildren) > 0 {
		source = "api+demo"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"discipline": map[string]any{"id": cat.ID, "slug": cat.Slug, "name": cat.Name},
		"items":      items,
		"source":     source,
	})
}

func (s *Server) handleFunFact(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	fact, ok := s.State.FunFact(slug)
	if !ok {
		writeError(w, r, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Дисциплина не найдена", nil)
		return
	}
	writeJSON(w, http.StatusOK, fact)
}

func iconFor(slug string) string {
	for _, seed := range corpus.DisciplineSeeds {
		if seed.Slug == slug {
			return seed.Icon
		}
	}
	fallback := []string{"cpu", "atom", "flask", "globe", "brain", "chart", "math", "leaf"}
	sum := 0
	for _, ch := range slug {
		sum += int(ch)
	}
	return fallback[sum%len(fallback)]
}

func accentFor(slug string) string {
	for _, seed := range corpus.DisciplineSeeds {
		if seed.Slug == slug {
			return seed.Accent
		}
	}
	palette := []string{"#6d5efc", "#22b38a", "#3f8cff", "#f0932b", "#e0516f", "#8e44ad", "#0aa3a3", "#d6a30c"}
	sum := 0
	for _, ch := range slug {
		sum += int(ch)
	}
	return palette[sum%len(palette)]
}

// --- Поиск ---

type searchRequest struct {
	Query      *string `json:"query"`
	EntityType string  `json:"entity_type"`
	Mode       string  `json:"mode"`
	SimilarTo  *string `json:"similar_to_publication_id"`
	Filters    struct {
		Category        *string  `json:"category"`
		YearFrom        *int     `json:"year_from"`
		YearTo          *int     `json:"year_to"`
		PublicationType []string `json:"publication_type"`
		Language        *string  `json:"language"`
		Author          *string  `json:"author"`
		Organization    *string  `json:"organization"`
		Journal         *string  `json:"journal"`
		Terms           []string `json:"terms"`
	} `json:"filters"`
	Sort          string `json:"sort"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
	IncludeFacets bool   `json:"include_facets"`
	Scope         string `json:"scope"`
}

// SearchFilters is the public shape accepted by POST /api/search.
func normalizeFilterQuery(req *searchRequest) (search.Query, *string) {
	q := search.Query{
		EntityType: req.EntityType, Mode: req.Mode, Sort: req.Sort,
		Page: req.Page, PageSize: req.PageSize, IncludeFacets: req.IncludeFacets,
	}
	if q.EntityType == "" {
		q.EntityType = "publication"
	}
	if q.Mode == "" {
		q.Mode = "hybrid"
	}
	if q.Sort == "" {
		q.Sort = "relevance"
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	if req.Query != nil {
		q.Text = *req.Query
	}
	if req.SimilarTo != nil {
		q.SimilarTo = *req.SimilarTo
	}
	if req.Filters.Language != nil {
		q.Language = *req.Filters.Language
	}
	if req.Filters.Author != nil {
		q.Author = *req.Filters.Author
	}
	if req.Filters.Organization != nil {
		q.Organization = *req.Filters.Organization
	}
	if req.Filters.Journal != nil {
		q.Journal = *req.Filters.Journal
	}
	q.Types = req.Filters.PublicationType
	q.Terms = req.Filters.Terms
	q.YearFrom = req.Filters.YearFrom
	q.YearTo = req.Filters.YearTo
	return q, req.Filters.Category
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req searchRequest
	if err := decodeBody(w, r, &req); err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Некорректное тело запроса", err.Error())
		return
	}
	start := nowMillis()
	q, category := normalizeFilterQuery(&req)
	if category != nil && *category != "" {
		q.CategoryIDs = s.resolveCategoryIDs(*category)
	}

	scope := strings.ToLower(strings.TrimSpace(req.Scope))
	if scope == "" {
		scope = "all"
	}

	var upstreamItems []any
	var upstreamFacets []domain.SearchFacet
	upstreamTotal := 0
	degraded := false

	if scope == "all" || scope == "api" {
		upReq := upstream.SearchRequest{
			Query: req.Query, EntityType: q.EntityType, Mode: q.Mode, SimilarToPublicationID: req.SimilarTo,
			Filters: map[string]any{},
			Sort:    q.Sort, Page: q.Page, PageSize: q.PageSize, IncludeFacets: q.IncludeFacets,
		}
		filters := upReq.Filters
		if category != nil && *category != "" {
			filters["category"] = *category
		}
		if q.YearFrom != nil {
			filters["year_from"] = *q.YearFrom
		}
		if q.YearTo != nil {
			filters["year_to"] = *q.YearTo
		}
		if len(q.Types) > 0 {
			filters["publication_type"] = q.Types
		}
		if q.Language != "" {
			filters["language"] = q.Language
		}
		if q.Author != "" {
			filters["author"] = q.Author
		}
		if q.Organization != "" {
			filters["organization"] = q.Organization
		}
		if q.Journal != "" {
			filters["journal"] = q.Journal
		}
		if len(q.Terms) > 0 {
			filters["terms"] = q.Terms
		}
		items, facets, total, err := s.Upstream.Search(r.Context(), upReq)
		if err != nil {
			degraded = true
		} else {
			upstreamItems, upstreamFacets, upstreamTotal = items, facets, total
		}
	}

	demo, err := s.Engine.Search(r.Context(), q)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "SEARCH_FAILED", "Поиск недоступен", err.Error())
		return
	}

	items := make([]any, 0, len(upstreamItems)+len(demo.Hits))
	seen := map[string]bool{}
	for _, item := range upstreamItems {
		items = append(items, item)
		if m, ok := item.(map[string]any); ok {
			if entity, ok := m["entity"].(map[string]any); ok {
				if id, ok := entity["id"].(string); ok {
					seen[id] = true
				}
			}
		}
	}
	for _, hit := range demo.Hits {
		if hit.Publication != nil {
			if seen[hit.Publication.ID] {
				continue
			}
			items = append(items, map[string]any{
				"entity_type": "publication", "score": hit.Score,
				"entity": hit.Publication, "snippets": hit.Snippets,
				"source": domain.SourceDemo,
			})
		} else if hit.Author != nil {
			if seen[hit.Author.ID] {
				continue
			}
			items = append(items, map[string]any{
				"entity_type": "author", "score": hit.Score,
				"entity": hit.Author, "snippets": []string{},
				"source": domain.SourceDemo,
			})
		}
	}

	// Upstream pagination applies to its own window; demo hits extend the page
	// so a user always sees demo results even when upstream is empty.
	pageSize := q.PageSize
	if len(items) > pageSize && len(upstreamItems) < pageSize {
		items = items[:pageSize]
	}

	facets := append([]domain.SearchFacet{}, upstreamFacets...)
	seenFacet := map[string]int{}
	for i := range facets {
		seenFacet[facets[i].Field] = i + 1
	}
	for _, f := range demo.Facets {
		if seenFacet[f.Field] > 0 {
			existing := &facets[seenFacet[f.Field]-1]
			merged := map[string]int{}
			for _, b := range existing.Buckets {
				merged[b.Value] += b.Count
			}
			for _, b := range f.Buckets {
				merged[b.Value] += b.Count
			}
			buckets := make([]domain.FacetBucket, 0, len(merged))
			for v, c := range merged {
				buckets = append(buckets, domain.FacetBucket{Value: v, Count: c})
			}
			sort.Slice(buckets, func(i, j int) bool { return buckets[i].Count > buckets[j].Count })
			existing.Buckets = buckets
			if existing.Labels == nil {
				existing.Labels = f.Labels
			} else {
				for k, v := range f.Labels {
					existing.Labels[k] = v
				}
			}
			continue
		}
		facets = append(facets, f)
	}

	total := upstreamTotal + demo.Total
	if scope == "demo" {
		total = demo.Total
	} else if scope == "api" {
		total = upstreamTotal
	}

	writeJSON(w, http.StatusOK, domain.SearchResponse{
		Items: items, Facets: facets, Total: total,
		Page: q.Page, PageSize: q.PageSize,
		TookMs: nowMillis() - start, Degraded: degraded,
	})
}

func (s *Server) resolveCategoryIDs(value string) []string {
	ids := []string{}
	segments := strings.Split(value, ",")
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if cat, ok := s.Corpus.CategoryBySlug[seg]; ok {
			ids = append(ids, cat.ID)
			for _, child := range cat.Children {
				ids = append(ids, child.ID)
			}
			continue
		}
		if _, ok := s.Corpus.CategoryByID[seg]; ok {
			ids = append(ids, seg)
			continue
		}
		ids = append(ids, seg)
	}
	return ids
}

func (s *Server) handleSuggest(w http.ResponseWriter, r *http.Request) {
	text := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := queryInt(r, "limit", 10, 1, 50)
	if text == "" {
		writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
		return
	}
	items := []domain.Suggestion{}
	if up, err := s.Upstream.Suggest(r.Context(), text, limit); err == nil {
		items = append(items, up...)
	}
	demo, err := s.Engine.Suggest(r.Context(), text, limit)
	if err == nil {
		for _, d := range demo {
			if len(items) >= limit {
				break
			}
			dup := false
			for _, existing := range items {
				if strings.EqualFold(existing.Text, d.Text) && existing.Type == d.Type {
					dup = true
					break
				}
			}
			if !dup {
				items = append(items, d)
			}
		}
	}
	if len(items) > limit {
		items = items[:limit]
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// --- Публикации ---

func (s *Server) findPublication(r *http.Request, id string) (*domain.Publication, domain.Source, bool) {
	if strings.HasPrefix(id, "demo-") {
		if pub, ok := s.Corpus.ByID[id]; ok {
			cp := *pub
			corpus.ApplyLikeState(&cp, s.State, viewer(r))
			cp.Metrics.Views += s.State.ViewDelta(id)
			return &cp, domain.SourceDemo, true
		}
		return nil, "", false
	}
	if pub, err := s.Upstream.Publication(r.Context(), id); err == nil {
		return pub, domain.SourceAPI, true
	}
	if pub, ok := s.Corpus.ByID[id]; ok {
		cp := *pub
		corpus.ApplyLikeState(&cp, s.State, viewer(r))
		return &cp, domain.SourceDemo, true
	}
	return nil, "", false
}

func (s *Server) handlePublication(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	pub, source, ok := s.findPublication(r, id)
	if !ok {
		writeError(w, r, http.StatusNotFound, "PUBLICATION_NOT_FOUND", "Публикация не найдена", map[string]any{"id": id})
		return
	}
	pub.Source = source
	writeJSON(w, http.StatusOK, pub)
}

func (s *Server) handleCreatePublication(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := decodeBody(w, r, &body); err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Некорректное тело запроса", err.Error())
		return
	}
	if title, _ := body["title"].(string); len(strings.TrimSpace(title)) < 1 {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Поле title обязательно", nil)
		return
	}
	if pub, err := s.Upstream.CreatePublication(r.Context(), body); err == nil {
		writeJSON(w, http.StatusCreated, pub)
		return
	}
	// Degraded mode: accept into the demo corpus so the UI flow stays usable.
	pub, err := s.Corpus.CreateFromPayload(body)
	if err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusCreated, pub)
}

func (s *Server) handleReferences(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "outgoing"
	}
	page := queryInt(r, "page", 1, 1, 100000)
	pageSize := queryInt(r, "page_size", 20, 1, 100)

	if !strings.HasPrefix(id, "demo-") {
		if refs, err := s.Upstream.References(r.Context(), id, direction, page, pageSize); err == nil {
			writeJSON(w, http.StatusOK, refs)
			return
		}
	}
	refs := s.Corpus.ReferencesFor(id, direction)
	total := len(refs)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	writeJSON(w, http.StatusOK, domain.Page[domain.Reference]{
		Items: refs[start:end], Total: total, Page: page, PageSize: pageSize,
	})
}

func (s *Server) handleRelated(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	limit := queryInt(r, "limit", 5, 1, 20)
	items := s.Corpus.Related(id, limit)
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleView(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !strings.HasPrefix(id, "demo-") {
		if view, err := s.Upstream.RegisterView(r.Context(), id); err == nil {
			writeJSON(w, http.StatusOK, view)
			return
		}
	}
	if pub, ok := s.Corpus.ByID[id]; ok {
		s.State.RegisterView(id)
		writeJSON(w, http.StatusOK, domain.ViewCount{
			PublicationID: id,
			Views:         pub.Metrics.Views + s.State.ViewDelta(id),
		})
		return
	}
	writeError(w, r, http.StatusNotFound, "PUBLICATION_NOT_FOUND", "Публикация не найдена", nil)
}

func (s *Server) handleLike(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.State.Like(r.PathValue("id"), viewer(r), true))
}

func (s *Server) handleUnlike(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.State.Like(r.PathValue("id"), viewer(r), false))
}

func (s *Server) handleGetFile(w http.ResponseWriter, r *http.Request) {
	id, kind := r.PathValue("id"), r.PathValue("kind")
	if strings.HasPrefix(id, "demo-") {
		writeError(w, r, http.StatusNotFound, "FILE_NOT_FOUND", "У демонстрационной публикации нет файлов", nil)
		return
	}
	content, contentType, err := s.Upstream.GetFile(r.Context(), id, kind)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "FILE_NOT_FOUND", "Файл недоступен", nil)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\"publication."+map[string]string{"raw": "pdf", "markdown": "md"}[kind]+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (s *Server) handlePutFile(w http.ResponseWriter, r *http.Request) {
	id, kind := r.PathValue("id"), r.PathValue("kind")
	body := make([]byte, 0, 4096)
	buf := make([]byte, 32<<10)
	for {
		n, err := r.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if err != nil {
			break
		}
		if len(body) > 5<<20 {
			writeError(w, r, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", "Максимум 5 МиБ", nil)
			return
		}
	}
	result, err := s.Upstream.PutFile(r.Context(), id, kind, r.Header.Get("Content-Type"), body)
	if err != nil {
		var apiErr *upstream.APIError
		if errors.As(err, &apiErr) {
			writeError(w, r, apiErr.Status, apiErr.Code, "Не удалось сохранить файл", nil)
			return
		}
		writeError(w, r, http.StatusServiceUnavailable, "UPSTREAM_UNAVAILABLE", "Сервис хранения недоступен", nil)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// --- Авторы ---

func (s *Server) handleAuthor(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if strings.HasPrefix(id, "demo-author-") {
		if author, ok := s.Corpus.AuthorByID[id]; ok {
			writeJSON(w, http.StatusOK, author)
			return
		}
	}
	if author, err := s.Upstream.Author(r.Context(), id); err == nil {
		writeJSON(w, http.StatusOK, author)
		return
	}
	if author, ok := s.Corpus.AuthorByID[id]; ok {
		writeJSON(w, http.StatusOK, author)
		return
	}
	writeError(w, r, http.StatusNotFound, "AUTHOR_NOT_FOUND", "Автор не найден", nil)
}

func (s *Server) handleAuthorStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Corpus.AuthorStats(r.PathValue("id")))
}

func (s *Server) handleAuthorPublications(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1, 1, 10000)
	pageSize := queryInt(r, "page_size", 20, 1, 100)
	items, total := s.Corpus.PublicationsByAuthor(r.PathValue("id"))
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	writeJSON(w, http.StatusOK, domain.Page[domain.PublicationSummary]{
		Items: items[start:end], Total: total, Page: page, PageSize: pageSize,
	})
}

func (s *Server) handleCoauthorGraph(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	depth := queryInt(r, "depth", 2, 1, 4)
	graph := s.Corpus.CoauthorGraph(id, depth)
	writeJSON(w, http.StatusOK, graph)
}

func (s *Server) handleUserProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	author, ok := s.Corpus.AuthorByID[id]
	if !ok {
		if up, err := s.Upstream.Author(r.Context(), id); err == nil {
			author = up
			ok = true
		}
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "AUTHOR_NOT_FOUND", "Автор не найден", nil)
		return
	}
	following, followers := s.State.FollowerState(id, viewer(r))
	stats := s.Corpus.AuthorStats(id)
	disciplines := []string{}
	for _, c := range author.Categories {
		disciplines = append(disciplines, c.Slug)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id": author.ID, "name": author.Name, "orcid": author.Orcid,
		"avatar_url":         nil,
		"affiliation":        affiliationOf(author),
		"bio":                "Исследователь, публикуется в направлениях: " + strings.Join(names(author.Categories), ", "),
		"publications_count": author.PublicationsCount,
		"citations_total":    author.CitationsCount,
		"h_index":            author.HIndex,
		"followers_count":    followers,
		"following":          following,
		"disciplines":        disciplines,
		"coauthors":          stats.Coauthors,
		"top_categories":     stats.TopCategories,
		"source":             author.Source,
	})
}

func (s *Server) handleFollow(w http.ResponseWriter, r *http.Request) {
	following, count := s.State.Follow(r.PathValue("id"), viewer(r), true)
	writeJSON(w, http.StatusOK, map[string]any{"following": following, "followers_count": count})
}

func (s *Server) handleUnfollow(w http.ResponseWriter, r *http.Request) {
	following, count := s.State.Follow(r.PathValue("id"), viewer(r), false)
	writeJSON(w, http.StatusOK, map[string]any{"following": following, "followers_count": count})
}

// --- Аналитика ---

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	payload := s.Corpus.Summary()
	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) handleTop(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if ranking, err := s.Upstream.Top(r.Context(), query); err == nil && len(ranking.Items) > 0 {
		writeJSON(w, http.StatusOK, ranking)
		return
	}
	writeJSON(w, http.StatusOK, s.Corpus.Ranking(query))
}

func (s *Server) handleWidget(w http.ResponseWriter, r *http.Request) {
	widget := r.PathValue("widget")
	category := r.URL.Query().Get("category")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "all"
	}
	result, ok := s.Widgets.Build(widget, category, period)
	if !ok {
		writeError(w, r, http.StatusNotFound, "WIDGET_NOT_FOUND", "Виджет не найден", map[string]any{"widget": widget})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	rootID := r.URL.Query().Get("root_id")
	if rootID == "" || !strings.HasPrefix(rootID, "demo-") {
		if graph, err := s.Upstream.Graph(r.Context(), r.URL.Query()); err == nil && len(graph.Nodes) > 0 {
			writeJSON(w, http.StatusOK, graph)
			return
		}
	}
	graphType := r.URL.Query().Get("type")
	maxDepth := queryInt(r, "max_depth", 2, 0, 5)
	maxNodes := queryInt(r, "max_nodes", 50, 1, 200)
	graph := s.Corpus.Graph(graphType, rootID, maxDepth, maxNodes, s.Corpus.CategoryIDsFor(r.URL.Query().Get("category")))
	writeJSON(w, http.StatusOK, graph)
}

// --- Q&A ---

func (s *Server) handleListQuestions(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1, 1, 10000)
	pageSize := queryInt(r, "page_size", 10, 1, 100)
	sort := r.URL.Query().Get("sort")
	items, total := s.State.ListQuestions(enrich.QuestionQuery{
		Category: r.URL.Query().Get("category"), Sort: sort, Page: page, PageSize: pageSize,
	})
	writeJSON(w, http.StatusOK, domain.Page[enrich.Question]{
		Items: items, Total: total, Page: page, PageSize: pageSize,
	})
}

func (s *Server) handleCreateQuestion(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title        string `json:"title"`
		Body         string `json:"body"`
		CategorySlug string `json:"category_slug"`
		AuthorName   string `json:"author_name"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Некорректное тело запроса", nil)
		return
	}
	if body.AuthorName == "" {
		body.AuthorName = r.Header.Get("X-Author")
	}
	question, err := s.State.CreateQuestion(body.Title, body.Body, body.CategorySlug, body.AuthorName)
	if errors.Is(err, enrich.ErrValidation) {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR",
			"Заголовок 5–200 символов, текст 1–5000 символов", nil)
		return
	}
	if errors.Is(err, enrich.ErrNotFound) {
		writeError(w, r, http.StatusNotFound, "CATEGORY_NOT_FOUND", "Категория не найдена", nil)
		return
	}
	writeJSON(w, http.StatusCreated, question)
}

func (s *Server) handleQuestion(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(r.PathValue("id"))
	if !ok {
		writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Некорректный идентификатор", nil)
		return
	}
	question, err := s.State.Question(id)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "QUESTION_NOT_FOUND", "Вопрос не найден", nil)
		return
	}
	writeJSON(w, http.StatusOK, question)
}

func (s *Server) handleCreateAnswer(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(r.PathValue("id"))
	if !ok {
		writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Некорректный идентификатор", nil)
		return
	}
	var body struct {
		Body       string `json:"body"`
		AuthorName string `json:"author_name"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Некорректное тело запроса", nil)
		return
	}
	if body.AuthorName == "" {
		body.AuthorName = r.Header.Get("X-Author")
	}
	answer, err := s.State.AddAnswer(id, body.Body, body.AuthorName)
	if errors.Is(err, enrich.ErrValidation) {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Текст ответа 1–5000 символов", nil)
		return
	}
	if err != nil {
		writeError(w, r, http.StatusNotFound, "QUESTION_NOT_FOUND", "Вопрос не найден", nil)
		return
	}
	writeJSON(w, http.StatusCreated, answer)
}

func (s *Server) handleVote(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt(r.PathValue("id"))
	if !ok {
		writeError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Некорректный идентификатор", nil)
		return
	}
	var body struct {
		Vote int `json:"vote"`
	}
	if err := decodeBody(w, r, &body); err != nil {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Некорректное тело запроса", nil)
		return
	}
	voter := r.Header.Get("X-Voter")
	if voter == "" {
		voter = viewer(r)
	}
	votes, err := s.State.Vote(id, voter, body.Vote)
	if errors.Is(err, enrich.ErrValidation) {
		writeError(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "vote должен быть 1 или -1", nil)
		return
	}
	if err != nil {
		writeError(w, r, http.StatusNotFound, "ANSWER_NOT_FOUND", "Ответ не найден", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "votes": votes})
}

// --- Вспомогательное ---

func (s *Server) handleLookup(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.Corpus.Lookup(r.URL.Query().Get("doi"), r.URL.Query().Get("title")))
}

func pathInt(raw string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	n := 0
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0, false
		}
		n = n*10 + int(ch-'0')
	}
	return n, true
}

func affiliationOf(a *domain.Author) string {
	if len(a.Affiliations) > 0 {
		return a.Affiliations[0].Name
	}
	return ""
}

func names(cats []domain.Category) []string {
	out := make([]string, 0, len(cats))
	for _, c := range cats {
		out = append(out, c.Name)
	}
	return out
}
