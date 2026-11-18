package corpus

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"scienceplatform/bff/internal/domain"
)

// LikeSource is the part of the in-memory state store this package needs. The
// interface lives here, not in enrich, so enrich can depend on corpus (it seeds
// questions from the corpus) without the two packages importing each other.
type LikeSource interface {
	LikeState(publicationID, viewer string) domain.LikeState
}

func atoi(value string) int {
	n, _ := strconv.Atoi(value)
	return n
}

// LikeState fills like counters for a publication copy.
func ApplyLikeState(p *domain.Publication, store LikeSource, viewer string) {
	state := store.LikeState(p.ID, viewer)
	p.Liked = state.Liked
	p.LikesCount = state.LikesCount
}

// ReferencesFor returns references in the requested direction.
func (c *Corpus) ReferencesFor(id, direction string) []domain.Reference {
	out := []domain.Reference{}
	for _, ref := range c.References {
		switch direction {
		case "incoming":
			if ref.CitedPublicationID != nil && *ref.CitedPublicationID == id {
				cp := ref
				if pub, ok := c.ByID[ref.CitingPublicationID]; ok {
					s := pub.PublicationSummary
					cp.Publication = &s
				}
				out = append(out, cp)
			}
		default:
			if ref.CitingPublicationID == id {
				cp := ref
				if ref.CitedPublicationID != nil {
					if pub, ok := c.ByID[*ref.CitedPublicationID]; ok {
						s := pub.PublicationSummary
						cp.Publication = &s
					}
				}
				out = append(out, cp)
			}
		}
	}
	return out
}

type RelatedItem struct {
	Publication domain.PublicationSummary `json:"publication"`
	Score       float64                   `json:"score"`
	Shared      []string                  `json:"shared"`
}

// Related ranks publications sharing concepts, keywords and category.
func (c *Corpus) Related(id string, limit int) []RelatedItem {
	source, ok := c.ByID[id]
	if !ok {
		return []RelatedItem{}
	}
	sourceConcepts := map[string]string{}
	for _, con := range source.Concepts {
		sourceConcepts[con.ID] = con.Name
	}
	sourceKeywords := map[string]bool{}
	for _, k := range source.Keywords {
		sourceKeywords[k] = true
	}
	sourceCat := ""
	if len(source.Categories) > 0 {
		sourceCat = source.Categories[0].ID
	}

	out := []RelatedItem{}
	for i := range c.Publications {
		p := &c.Publications[i]
		if p.ID == id {
			continue
		}
		score := 0.0
		shared := []string{}
		for _, con := range p.Concepts {
			if name, ok := sourceConcepts[con.ID]; ok {
				score += 0.45
				shared = append(shared, name)
			}
		}
		for _, k := range p.Keywords {
			if sourceKeywords[k] {
				score += 0.2
			}
		}
		if len(p.Categories) > 0 && p.Categories[0].ID == sourceCat {
			score += 0.25
		}
		if score <= 0 {
			continue
		}
		out = append(out, RelatedItem{Publication: p.PublicationSummary, Score: math.Round(score*100) / 100, Shared: shared})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// CreateFromPayload appends a publication to the demo corpus when upstream is down.
func (c *Corpus) CreateFromPayload(body map[string]any) (*domain.Publication, error) {
	title, _ := body["title"].(string)
	if strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title обязателен")
	}
	pubType, _ := body["publication_type"].(string)
	if pubType == "" {
		pubType = "article"
	}
	lang, _ := body["language"].(string)
	if lang == "" {
		lang = "ru"
	}
	id := fmt.Sprintf("demo-%d", len(c.Publications)+1)

	authors := []domain.AuthorSummary{}
	if raw, ok := body["authors"].([]any); ok {
		for _, entry := range raw {
			m, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			if aid, ok := m["author_id"].(string); ok {
				if a, ok := c.AuthorByID[aid]; ok {
					authors = append(authors, a.AuthorSummary)
					continue
				}
			}
			if name, ok := m["name"].(string); ok && name != "" {
				authors = append(authors, domain.AuthorSummary{
					ID:           "guest-author-" + name,
					Name:         name,
					Affiliations: []domain.Organization{{ID: "org-1", Name: strOr(m["affiliation"], "Не указана")}},
				})
			}
		}
	}
	if len(authors) == 0 {
		if len(c.Authors) > 0 {
			authors = append(authors, c.Authors[0].AuthorSummary)
		}
	}

	cats := []domain.Category{}
	slug, _ := body["category_slug"].(string)
	if slug == "" {
		if raw, ok := body["category_ids"].([]any); ok && len(raw) > 0 {
			if s, ok := raw[0].(string); ok {
				slug = s
			}
		}
	}
	if cat, ok := c.CategoryBySlug[slug]; ok {
		cats = append(cats, *cat)
		if cat.ParentID != nil {
			if root, ok := c.CategoryByID[*cat.ParentID]; ok {
				cats = append(cats, *root)
			}
		}
	} else if len(c.Categories) > 0 {
		root := c.Categories[0]
		if len(root.Children) > 0 {
			cats = append(cats, root.Children[0], root)
		} else {
			cats = append(cats, root)
		}
	}

	keywords := []string{}
	if raw, ok := body["keywords"].([]any); ok {
		for _, k := range raw {
			if s, ok := k.(string); ok {
				keywords = append(keywords, s)
			}
		}
	}

	var date *string
	if s, ok := body["publication_date"].(string); ok && s != "" {
		date = &s
	} else if y, ok := body["year"].(float64); ok {
		d := fmt.Sprintf("%04d-01-01", int(y))
		date = &d
	}

	pub := domain.Publication{
		PublicationSummary: domain.PublicationSummary{
			ID: id, Title: strings.TrimSpace(title),
			Abstract:        optionalString(body["abstract"]),
			PublicationType: pubType, PublicationDate: date, Language: lang,
			Authors: authors, Metrics: domain.Metrics{}, Source: domain.SourceDemo,
		},
		ExternalIdentifiers: []domain.ExternalIdentifier{},
		Categories:          cats,
		Keywords:            keywords,
		Concepts:            []domain.Concept{},
		Files:               []domain.PublicationFile{},
		ProcessingStatus:    "pending",
	}
	if doi, ok := body["doi"].(string); ok && doi != "" {
		pub.DOI = &doi
		pub.ExternalIdentifiers = append(pub.ExternalIdentifiers, domain.ExternalIdentifier{Scheme: "doi", Value: doi})
	}
	c.Publications = append(c.Publications, pub)
	c.ByID = map[string]*domain.Publication{}
	for i := range c.Publications {
		c.ByID[c.Publications[i].ID] = &c.Publications[i]
	}
	return c.ByID[id], nil
}

func optionalString(v any) *string {
	if s, ok := v.(string); ok && s != "" {
		return &s
	}
	return nil
}

func strOr(v any, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

type YearCount struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

type CategoryCount struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Coauthor struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type AuthorStats struct {
	PublicationsByYear []YearCount     `json:"publications_by_year"`
	CitationsByYear    []YearCount     `json:"citations_by_year"`
	TopCategories      []CategoryCount `json:"top_categories"`
	Coauthors          []Coauthor      `json:"coauthors"`
}

// AuthorStats computes per-author statistics used by the profile screen.
func (c *Corpus) AuthorStats(id string) AuthorStats {
	pubYears := map[int]int{}
	citYears := map[int]int{}
	catCounts := map[string]int{}
	catNames := map[string]string{}
	coauthors := map[string]int{}

	for i := range c.Publications {
		p := &c.Publications[i]
		isAuthor := false
		for _, a := range p.Authors {
			if a.ID == id {
				isAuthor = true
			}
		}
		if !isAuthor {
			continue
		}
		year := 2026
		if p.PublicationDate != nil {
			year = atoi((*p.PublicationDate)[:4])
		}
		pubYears[year]++
		citYears[year] += p.Metrics.Citations
		if len(p.Categories) > 0 {
			catCounts[p.Categories[0].Slug]++
			catNames[p.Categories[0].Slug] = p.Categories[0].Name
		}
		for _, a := range p.Authors {
			if a.ID != id {
				coauthors[a.ID]++
			}
		}
	}

	years := make([]int, 0, len(pubYears))
	for y := range pubYears {
		years = append(years, y)
	}
	sort.Ints(years)
	stats := AuthorStats{
		PublicationsByYear: []YearCount{},
		CitationsByYear:    []YearCount{},
		TopCategories:      []CategoryCount{},
		Coauthors:          []Coauthor{},
	}
	for _, y := range years {
		stats.PublicationsByYear = append(stats.PublicationsByYear, YearCount{Year: y, Count: pubYears[y]})
		stats.CitationsByYear = append(stats.CitationsByYear, YearCount{Year: y, Count: citYears[y]})
	}
	for slug, count := range catCounts {
		stats.TopCategories = append(stats.TopCategories, CategoryCount{Slug: slug, Name: catNames[slug], Count: count})
	}
	sort.Slice(stats.TopCategories, func(i, j int) bool { return stats.TopCategories[i].Count > stats.TopCategories[j].Count })
	for coID, count := range coauthors {
		name := coID
		if a, ok := c.AuthorByID[coID]; ok {
			name = a.Name
		}
		stats.Coauthors = append(stats.Coauthors, Coauthor{ID: coID, Name: name, Count: count})
	}
	sort.Slice(stats.Coauthors, func(i, j int) bool { return stats.Coauthors[i].Count > stats.Coauthors[j].Count })
	if len(stats.Coauthors) > 8 {
		stats.Coauthors = stats.Coauthors[:8]
	}
	return stats
}

// PublicationsByAuthor lists summaries for one author, newest first.
func (c *Corpus) PublicationsByAuthor(id string) ([]domain.PublicationSummary, int) {
	out := []domain.PublicationSummary{}
	for i := range c.Publications {
		p := &c.Publications[i]
		for _, a := range p.Authors {
			if a.ID == id {
				out = append(out, p.PublicationSummary)
				break
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		li, lj := "", ""
		if out[i].PublicationDate != nil {
			li = *out[i].PublicationDate
		}
		if out[j].PublicationDate != nil {
			lj = *out[j].PublicationDate
		}
		return li > lj
	})
	return out, len(out)
}

// CoauthorGraph builds a BFS-limited coauthorship graph around one author.
func (c *Corpus) CoauthorGraph(id string, depth int) domain.Graph {
	nodes := map[string]domain.GraphNode{}
	edges := map[string]domain.GraphEdge{}
	dist := map[string]int{id: 0}
	frontier := []string{id}

	edgesOf := map[string]map[string]float64{}
	for i := range c.Publications {
		p := &c.Publications[i]
		for x := 0; x < len(p.Authors); x++ {
			for y := x + 1; y < len(p.Authors); y++ {
				a, b := p.Authors[x].ID, p.Authors[y].ID
				if edgesOf[a] == nil {
					edgesOf[a] = map[string]float64{}
				}
				if edgesOf[b] == nil {
					edgesOf[b] = map[string]float64{}
				}
				edgesOf[a][b]++
				edgesOf[b][a]++
			}
		}
	}

	addNode := func(authorID string) {
		if _, ok := nodes[authorID]; ok {
			return
		}
		label := authorID
		category := ""
		weight := 0.0
		if a, ok := c.AuthorByID[authorID]; ok {
			label = a.Name
			weight = float64(a.PublicationsCount)
			if len(a.Categories) > 0 {
				category = a.Categories[0].Slug
			}
		}
		nodes[authorID] = domain.GraphNode{ID: authorID, Type: "author", Label: label, Category: category, Weight: weight}
	}

	for level := 0; level < depth; level++ {
		next := []string{}
		for _, current := range frontier {
			addNode(current)
			for neighbor, w := range edgesOf[current] {
				addNode(neighbor)
				key := current + "|" + neighbor
				if neighbor < current {
					key = neighbor + "|" + current
				}
				if _, ok := edges[key]; !ok {
					weight := w
					edges[key] = domain.GraphEdge{Source: current, Target: neighbor, Type: "coauthor", Weight: &weight}
				}
				if _, seen := dist[neighbor]; !seen {
					dist[neighbor] = dist[current] + 1
					next = append(next, neighbor)
				}
			}
		}
		frontier = next
		if len(frontier) == 0 {
			break
		}
	}

	nodeList := make([]domain.GraphNode, 0, len(nodes))
	for _, n := range nodes {
		nodeList = append(nodeList, n)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		if dist[nodeList[i].ID] != dist[nodeList[j].ID] {
			return dist[nodeList[i].ID] < dist[nodeList[j].ID]
		}
		return nodeList[i].Weight > nodeList[j].Weight
	})
	truncated := false
	if len(nodeList) > 40 {
		nodeList = nodeList[:40]
		truncated = true
	}
	keep := map[string]bool{}
	for _, n := range nodeList {
		keep[n.ID] = true
	}
	edgeList := []domain.GraphEdge{}
	for _, e := range edges {
		if keep[e.Source] && keep[e.Target] {
			edgeList = append(edgeList, e)
		}
	}
	return domain.Graph{
		Nodes: nodeList, Edges: edgeList, Truncated: truncated,
		DataAsOf: AsOf.Format("2006-01-02T15:04:05Z"), Source: domain.SourceDemo,
	}
}

// Graph builds a citation, coauthor or concept graph over the demo corpus.
func (c *Corpus) Graph(kind, rootID string, maxDepth, maxNodes int, categoryIDs []string) domain.Graph {
	if kind == "" {
		kind = "citation"
	}
	allowed := map[string]bool{}
	if len(categoryIDs) > 0 {
		for _, id := range categoryIDs {
			allowed[id] = true
		}
	}
	include := func(p *domain.Publication) bool {
		if len(allowed) == 0 {
			return true
		}
		for _, cat := range p.Categories {
			if allowed[cat.ID] {
				return true
			}
		}
		return false
	}

	nodes := map[string]domain.GraphNode{}
	edges := []domain.GraphEdge{}

	switch kind {
	case "coauthor":
		for i := range c.Publications {
			p := &c.Publications[i]
			if !include(p) {
				continue
			}
			for _, a := range p.Authors {
				if _, ok := nodes[a.ID]; !ok {
					nodes[a.ID] = domain.GraphNode{ID: a.ID, Type: "author", Label: a.Name}
				}
			}
			for x := 0; x < len(p.Authors); x++ {
				for y := x + 1; y < len(p.Authors); y++ {
					edges = append(edges, domain.GraphEdge{Source: p.Authors[x].ID, Target: p.Authors[y].ID, Type: "coauthor"})
				}
			}
		}
	case "concept":
		for i := range c.Publications {
			p := &c.Publications[i]
			if !include(p) {
				continue
			}
			for _, con := range p.Concepts {
				nodes[con.ID] = domain.GraphNode{ID: con.ID, Type: "concept", Label: con.Name}
			}
			for x := 0; x < len(p.Concepts); x++ {
				for y := x + 1; y < len(p.Concepts); y++ {
					edges = append(edges, domain.GraphEdge{
						Source: p.Concepts[x].ID, Target: p.Concepts[y].ID, Type: "cooccurs",
					})
				}
			}
		}
	default:
		for i := range c.Publications {
			p := &c.Publications[i]
			if !include(p) {
				continue
			}
			cat := ""
			if len(p.Categories) > 0 {
				cat = p.Categories[0].Slug
			}
			nodes[p.ID] = domain.GraphNode{
				ID: p.ID, Type: "publication", Label: p.Title, Category: cat,
				Weight: float64(p.Metrics.Citations + 1),
			}
		}
		for _, ref := range c.References {
			if ref.CitedPublicationID == nil {
				continue
			}
			if _, ok := nodes[ref.CitingPublicationID]; !ok {
				continue
			}
			if _, ok := nodes[*ref.CitedPublicationID]; !ok {
				continue
			}
			edges = append(edges, domain.GraphEdge{Source: ref.CitingPublicationID, Target: *ref.CitedPublicationID, Type: "cites"})
		}
	}

	adj := map[string][]string{}
	for _, e := range edges {
		adj[e.Source] = append(adj[e.Source], e.Target)
		adj[e.Target] = append(adj[e.Target], e.Source)
	}

	selected := []string{}
	if rootID != "" {
		if _, ok := nodes[rootID]; !ok {
			// Root outside the filtered set: empty graph per upstream semantics.
			return domain.Graph{Nodes: []domain.GraphNode{}, Edges: []domain.GraphEdge{},
				DataAsOf: AsOf.Format("2006-01-02T15:04:05Z"), Source: domain.SourceDemo}
		}
		visited := map[string]int{rootID: 0}
		queue := []string{rootID}
		selected = append(selected, rootID)
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			if visited[current] >= maxDepth {
				continue
			}
			for _, next := range adj[current] {
				if _, seen := visited[next]; seen {
					continue
				}
				visited[next] = visited[current] + 1
				selected = append(selected, next)
				queue = append(queue, next)
			}
		}
	} else {
		for id := range nodes {
			selected = append(selected, id)
		}
		sort.Slice(selected, func(i, j int) bool { return nodes[selected[i]].Weight > nodes[selected[j]].Weight })
	}

	truncated := false
	if len(selected) > maxNodes {
		selected = selected[:maxNodes]
		truncated = true
	}
	keep := map[string]bool{}
	nodeList := make([]domain.GraphNode, 0, len(selected))
	for _, id := range selected {
		keep[id] = true
		nodeList = append(nodeList, nodes[id])
	}
	edgeList := []domain.GraphEdge{}
	edgeSeen := map[string]bool{}
	for _, e := range edges {
		if !keep[e.Source] || !keep[e.Target] {
			continue
		}
		key := e.Source + "|" + e.Target
		if edgeSeen[key] {
			continue
		}
		edgeSeen[key] = true
		edgeList = append(edgeList, e)
	}
	return domain.Graph{
		Nodes: nodeList, Edges: edgeList, Truncated: truncated,
		DataAsOf: AsOf.Format("2006-01-02T15:04:05Z"), Source: domain.SourceDemo,
	}
}

// CategoryIDsFor expands slugs and ids into a category id set, including children.
func (c *Corpus) CategoryIDsFor(value string) []string {
	if value == "" {
		return nil
	}
	out := []string{}
	for _, seg := range strings.Split(value, ",") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		if cat, ok := c.CategoryBySlug[seg]; ok {
			out = append(out, cat.ID)
			for _, child := range cat.Children {
				out = append(out, child.ID)
			}
			continue
		}
		out = append(out, seg)
	}
	return out
}

// Summary returns corpus-wide KPI counters.
func (c *Corpus) Summary() domain.CorpusSummary {
	citations, downloads, thisPeriod := 0, 0, 0
	authorSet := map[string]bool{}
	recentAuthors := map[string]bool{}
	for i := range c.Publications {
		p := &c.Publications[i]
		citations += p.Metrics.Citations
		downloads += p.Metrics.Downloads
		year := ""
		if p.PublicationDate != nil {
			year = (*p.PublicationDate)[:4]
		}
		if year == "2026" {
			thisPeriod++
			for _, a := range p.Authors {
				recentAuthors[a.ID] = true
			}
		}
		for _, a := range p.Authors {
			authorSet[a.ID] = true
		}
	}
	avg := 0.0
	if len(c.Publications) > 0 {
		avg = math.Round(float64(citations)/float64(len(c.Publications))*100) / 100
	}
	return domain.CorpusSummary{
		TotalPublications:          len(c.Publications),
		TotalAuthors:               len(authorSet),
		TotalCitations:             citations,
		TotalDownloads:             downloads,
		AvgCitationsPerPublication: avg,
		PublicationsThisPeriod:     thisPeriod,
		NewAuthorsThisPeriod:       len(recentAuthors),
		Period:                     "2026",
		DataAsOf:                   AsOf.Format("2006-01-02T15:04:05Z"),
	}
}

// Ranking mirrors GET /api/v1/analytics/top over the demo corpus.
func (c *Corpus) Ranking(query map[string][]string) domain.Ranking {
	get := func(key, fallback string) string {
		if v, ok := query[key]; ok && len(v) > 0 && v[0] != "" {
			return v[0]
		}
		return fallback
	}
	entity := get("entity", "publication")
	metric := get("metric", "citations")
	limit := atoi(get("limit", "10"))
	if limit <= 0 {
		limit = 10
	}

	type entry struct {
		id    string
		label string
		value float64
	}
	board := map[string]*entry{}

	for i := range c.Publications {
		p := &c.Publications[i]
		switch entity {
		case "publication":
			var v float64
			switch metric {
			case "views":
				v = float64(p.Metrics.Views)
			case "downloads":
				v = float64(p.Metrics.Downloads)
			default:
				v = float64(p.Metrics.Citations)
			}
			board[p.ID] = &entry{id: p.ID, label: p.Title, value: v}
		case "author":
			for _, a := range p.Authors {
				e, ok := board[a.ID]
				if !ok {
					e = &entry{id: a.ID, label: a.Name}
					board[a.ID] = e
				}
				if metric == "publications" {
					e.value++
				} else {
					e.value += float64(p.Metrics.Citations)
				}
			}
		case "keyword":
			for _, k := range p.Keywords {
				e, ok := board[k]
				if !ok {
					e = &entry{id: k, label: k}
					board[k] = e
				}
				e.value++
			}
		case "concept":
			for _, con := range p.Concepts {
				e, ok := board[con.ID]
				if !ok {
					e = &entry{id: con.ID, label: con.Name}
					board[con.ID] = e
				}
				e.value++
			}
		}
	}

	items := make([]domain.RankingItem, 0, len(board))
	for _, e := range board {
		items = append(items, domain.RankingItem{
			Object: domain.RankingObject{EntityType: entity, ID: e.id, Label: e.label},
			Value:  e.value,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Value != items[j].Value {
			return items[i].Value > items[j].Value
		}
		return items[i].Object.ID < items[j].Object.ID
	})
	if len(items) > limit {
		items = items[:limit]
	}
	for i := range items {
		items[i].Position = i + 1
	}
	return domain.Ranking{Metric: metric, Items: items, Source: domain.SourceDemo}
}

// Lookup answers DOI/title autocomplete from demo metadata.
func (c *Corpus) Lookup(doi, title string) map[string]any {
	if strings.TrimSpace(doi) == "" && strings.TrimSpace(title) == "" {
		return map[string]any{"found": false, "source": "demo-mock"}
	}
	if doi != "" {
		for i := range c.Publications {
			p := &c.Publications[i]
			if p.DOI != nil && strings.EqualFold(*p.DOI, doi) {
				return c.lookupPayload(p, doi)
			}
		}
	}
	if title != "" {
		needle := strings.ToLower(title)
		for i := range c.Publications {
			p := &c.Publications[i]
			if strings.Contains(strings.ToLower(p.Title), needle) {
				return c.lookupPayload(p, title)
			}
		}
	}
	// Deterministic mock record so the "по DOI" flow is demonstrable.
	return map[string]any{
		"found": true, "query": doi, "source": "demo-mock",
		"title":            "Нейросетевые методы анализа научных текстов",
		"authors":          []map[string]string{{"name": "Анна Александровна Петрова", "affiliation": "Институт научных вычислений"}},
		"year":             2024,
		"abstract":         "Предлагается подход к извлечению понятий предметной области на основе графовых эмбеддингов.",
		"language":         "ru",
		"keywords":         []string{"нейросети", "извлечение понятий", "графы знаний"},
		"publication_type": "article",
		"category_slug":    "nlp",
	}
}

func (c *Corpus) lookupPayload(p *domain.Publication, query string) map[string]any {
	authors := []map[string]string{}
	for _, a := range p.Authors {
		aff := ""
		if len(a.Affiliations) > 0 {
			aff = a.Affiliations[0].Name
		}
		authors = append(authors, map[string]string{"name": a.Name, "affiliation": aff})
	}
	year := 2024
	if p.PublicationDate != nil {
		year = atoi((*p.PublicationDate)[:4])
	}
	slug := ""
	if len(p.Categories) > 0 {
		slug = p.Categories[0].Slug
	}
	return map[string]any{
		"found": true, "query": query, "source": "demo-mock",
		"title": p.Title, "authors": authors, "year": year,
		"abstract": optionalString(p.Abstract), "language": p.Language,
		"keywords": p.Keywords, "publication_type": p.PublicationType, "category_slug": slug,
	}
}
