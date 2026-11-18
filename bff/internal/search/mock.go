package search

import (
	"context"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"scienceplatform/bff/internal/corpus"
	"scienceplatform/bff/internal/domain"
)

// MockEngine is a deterministic in-memory engine over the demo corpus.
// It is intentionally simple: lexical matching plus concept/keyword boosting,
// which is enough to exercise every UI state. A real engine (C++ service) can
// replace it behind the same port.
type MockEngine struct {
	corpus *corpus.Corpus
}

func NewMockEngine(c *corpus.Corpus) *MockEngine { return &MockEngine{corpus: c} }

func (m *MockEngine) Name() string { return "mock-inmemory" }

func (m *MockEngine) Search(_ context.Context, q Query) (Result, error) {
	text := strings.ToLower(strings.TrimSpace(q.Text))
	if q.EntityType == "author" {
		return m.searchAuthors(q, text), nil
	}

	matched := make([]domain.Publication, 0, len(m.corpus.Publications))
	for _, p := range m.corpus.Publications {
		if !m.matchesFilters(p, q) {
			continue
		}
		matched = append(matched, p)
	}

	if q.Mode == "similar" && q.SimilarTo != "" {
		source, ok := m.corpus.ByID[q.SimilarTo]
		if ok {
			filtered := matched[:0]
			for _, p := range matched {
				if p.ID != source.ID {
					filtered = append(filtered, p)
				}
			}
			matched = filtered
		}
	}

	hits := make([]Hit, 0, len(matched))
	for i := range matched {
		pub := matched[i]
		score, snippets := m.score(pub, text, q)
		if score <= 0 {
			continue
		}
		cp := pub
		hits = append(hits, Hit{Publication: &cp, Score: score, Snippets: snippets})
	}

	switch q.Sort {
	case "date":
		sort.SliceStable(hits, func(i, j int) bool { return dateOf(hits[i]) > dateOf(hits[j]) })
	case "citations":
		sort.SliceStable(hits, func(i, j int) bool {
			return hits[i].Publication.Metrics.Citations > hits[j].Publication.Metrics.Citations
		})
	case "views":
		sort.SliceStable(hits, func(i, j int) bool { return hits[i].Publication.Metrics.Views > hits[j].Publication.Metrics.Views })
	case "downloads":
		sort.SliceStable(hits, func(i, j int) bool {
			return hits[i].Publication.Metrics.Downloads > hits[j].Publication.Metrics.Downloads
		})
	default:
		sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	}

	facets := []domain.SearchFacet{}
	if q.IncludeFacets {
		facets = m.facets(hits)
	}

	total := len(hits)
	start := (q.Page - 1) * q.PageSize
	if start > total {
		start = total
	}
	end := start + q.PageSize
	if end > total {
		end = total
	}
	return Result{Hits: hits[start:end], Facets: facets, Total: total}, nil
}

func (m *MockEngine) searchAuthors(q Query, text string) Result {
	eligible := map[string]bool{}
	for _, p := range m.corpus.Publications {
		if !m.matchesFilters(p, q) {
			continue
		}
		for _, a := range p.Authors {
			eligible[a.ID] = true
		}
	}
	hasFilters := len(q.CategoryIDs) > 0 || q.YearFrom != nil || q.YearTo != nil ||
		len(q.Types) > 0 || q.Language != "" || q.Author != "" || q.Organization != "" || q.Journal != ""

	hits := []Hit{}
	for i := range m.corpus.Authors {
		author := &m.corpus.Authors[i]
		if hasFilters && !eligible[author.ID] {
			continue
		}
		score := 1.0
		if text != "" {
			name := strings.ToLower(author.Name)
			if !strings.Contains(name, text) {
				ok := false
				for _, cat := range author.Categories {
					if strings.Contains(strings.ToLower(cat.Name), text) {
						ok = true
						break
					}
				}
				if !ok {
					continue
				}
				score = 0.6
			} else if strings.HasPrefix(name, text) {
				score = 1.0
			} else {
				score = 0.8
			}
		}
		cp := author.AuthorSummary
		hits = append(hits, Hit{Author: &cp, Score: score})
	}

	if q.Sort == "citations" {
		rank := map[string]int{}
		for _, a := range m.corpus.Authors {
			rank[a.ID] = a.CitationsCount
		}
		sort.SliceStable(hits, func(i, j int) bool { return rank[hits[i].Author.ID] > rank[hits[j].Author.ID] })
	} else {
		sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	}

	facets := []domain.SearchFacet{}
	if q.IncludeFacets {
		counts := map[string]int{}
		for _, h := range hits {
			for _, aff := range h.Author.Affiliations {
				counts[aff.Name]++
			}
		}
		buckets := make([]domain.FacetBucket, 0, len(counts))
		for k, v := range counts {
			buckets = append(buckets, domain.FacetBucket{Value: k, Count: v})
		}
		sort.Slice(buckets, func(i, j int) bool { return buckets[i].Count > buckets[j].Count })
		facets = append(facets, domain.SearchFacet{Field: "organization", Buckets: buckets})
	}

	total := len(hits)
	start := (q.Page - 1) * q.PageSize
	if start > total {
		start = total
	}
	end := start + q.PageSize
	if end > total {
		end = total
	}
	return Result{Hits: hits[start:end], Facets: facets, Total: total}
}

func (m *MockEngine) matchesFilters(p domain.Publication, q Query) bool {
	if len(q.CategoryIDs) > 0 {
		found := false
		for _, c := range p.Categories {
			for _, want := range q.CategoryIDs {
				if c.ID == want {
					found = true
				}
			}
		}
		if !found {
			return false
		}
	}
	year := 0
	if p.PublicationDate != nil && len(*p.PublicationDate) >= 4 {
		year = atoi((*p.PublicationDate)[:4])
	}
	if q.YearFrom != nil && (year == 0 || year < *q.YearFrom) {
		return false
	}
	if q.YearTo != nil && (year == 0 || year > *q.YearTo) {
		return false
	}
	if len(q.Types) > 0 && !containsFold(q.Types, p.PublicationType) {
		return false
	}
	if q.Language != "" && !strings.EqualFold(p.Language, q.Language) {
		return false
	}
	if q.Author != "" {
		ok := false
		for _, a := range p.Authors {
			if a.ID == q.Author {
				ok = true
			}
		}
		if !ok {
			return false
		}
	}
	if q.Organization != "" {
		ok := false
		for _, a := range p.Authors {
			for _, org := range a.Affiliations {
				if org.ID == q.Organization || strings.EqualFold(org.Name, q.Organization) {
					ok = true
				}
			}
		}
		if !ok {
			return false
		}
	}
	if q.Journal != "" && (p.Journal == nil || !strings.EqualFold(*p.Journal, q.Journal)) {
		return false
	}
	if len(q.Terms) > 0 {
		haystack := strings.ToLower(strings.Join(p.Keywords, " "))
		for _, c := range p.Concepts {
			haystack += " " + strings.ToLower(c.Name)
		}
		for _, term := range q.Terms {
			if !strings.Contains(haystack, strings.ToLower(term)) {
				return false
			}
		}
	}
	return true
}

func (m *MockEngine) score(p domain.Publication, text string, q Query) (float64, []string) {
	snippets := []string{}
	score := 0.15
	if text == "" {
		if q.Mode == "similar" {
			return 0.5, snippets
		}
		score = 1.0 / (1 + float64(p.Metrics.Views)/400.0)
		return score, snippets
	}

	title := strings.ToLower(p.Title)
	abstract := ""
	if p.Abstract != nil {
		abstract = strings.ToLower(*p.Abstract)
	}
	concepts := make([]string, 0, len(p.Concepts))
	for _, c := range p.Concepts {
		concepts = append(concepts, strings.ToLower(c.Name))
	}
	keywords := strings.ToLower(strings.Join(p.Keywords, " "))
	authors := ""
	for _, a := range p.Authors {
		authors += strings.ToLower(a.Name) + " "
	}

	tokens := strings.Fields(text)
	matchedTokens := 0
	for _, tok := range tokens {
		if len([]rune(tok)) < 2 {
			continue
		}
		hit := false
		if strings.Contains(title, tok) {
			score += 0.45
			hit = true
		}
		if strings.Contains(keywords, tok) {
			score += 0.22
			hit = true
		}
		for _, c := range concepts {
			if strings.Contains(c, tok) {
				score += 0.28
				hit = true
				break
			}
		}
		if strings.Contains(abstract, tok) {
			score += 0.12
			hit = true
			if s := makeSnippet(*p.Abstract, tok); s != "" {
				snippets = append(snippets, s)
			}
		}
		if strings.Contains(authors, tok) {
			score += 0.18
			hit = true
		}
		if hit {
			matchedTokens++
		}
	}
	if matchedTokens == 0 {
		return 0, nil
	}
	// Favour documents matching more distinct query tokens.
	score += 0.3 * float64(matchedTokens) / float64(len(tokens))
	if q.Mode == "similar" {
		score += 0.2
	}
	if len(snippets) > 2 {
		snippets = snippets[:2]
	}
	if len(snippets) == 0 && p.Abstract != nil {
		snippets = []string{truncate(*p.Abstract, 180)}
	}
	return math.Round(score*100) / 100, snippets
}

func (m *MockEngine) facets(hits []Hit) []domain.SearchFacet {
	langs := map[string]int{}
	cats := map[string]int{}
	labels := map[string]string{}
	years := map[string]int{}
	types := map[string]int{}
	for _, h := range hits {
		if h.Publication == nil {
			continue
		}
		langs[h.Publication.Language]++
		types[h.Publication.PublicationType]++
		if h.Publication.PublicationDate != nil {
			years[(*h.Publication.PublicationDate)[:4]]++
		}
		if len(h.Publication.Categories) > 0 {
			cat := h.Publication.Categories[0]
			cats[cat.ID]++
			labels[cat.ID] = cat.Name
		}
	}
	out := []domain.SearchFacet{
		{Field: "language", Buckets: sortedBuckets(langs)},
		{Field: "publication_type", Buckets: sortedBuckets(types)},
		{Field: "year", Buckets: sortedBuckets(years)},
		{Field: "category", Buckets: sortedBuckets(cats), Labels: labels},
	}
	return out
}

func (m *MockEngine) Suggest(_ context.Context, text string, limit int) ([]domain.Suggestion, error) {
	needle := strings.ToLower(strings.TrimSpace(text))
	out := []domain.Suggestion{}
	if needle == "" {
		return out, nil
	}
	add := func(kind, label, id string) {
		if len(out) >= limit {
			return
		}
		count := 1
		out = append(out, domain.Suggestion{Type: kind, Text: label, ID: id, MatchesCount: &count, Source: domain.SourceDemo})
	}
	for i := range m.corpus.Publications {
		p := &m.corpus.Publications[i]
		if strings.Contains(strings.ToLower(p.Title), needle) {
			add("publication", p.Title, p.ID)
		}
	}
	for i := range m.corpus.Authors {
		a := &m.corpus.Authors[i]
		if strings.Contains(strings.ToLower(a.Name), needle) {
			add("author", a.Name, a.ID)
		}
	}
	for _, c := range m.corpus.Concepts {
		if strings.Contains(strings.ToLower(c.Name), needle) {
			add("term", c.Name, c.ID)
		}
	}
	return out, nil
}

func sortedBuckets(counts map[string]int) []domain.FacetBucket {
	buckets := make([]domain.FacetBucket, 0, len(counts))
	for k, v := range counts {
		buckets = append(buckets, domain.FacetBucket{Value: k, Count: v})
	}
	sort.Slice(buckets, func(i, j int) bool {
		if buckets[i].Count != buckets[j].Count {
			return buckets[i].Count > buckets[j].Count
		}
		return buckets[i].Value < buckets[j].Value
	})
	return buckets
}

func dateOf(h Hit) string {
	if h.Publication == nil || h.Publication.PublicationDate == nil {
		return ""
	}
	return *h.Publication.PublicationDate
}

func containsFold(list []string, v string) bool {
	for _, x := range list {
		if strings.EqualFold(x, v) {
			return true
		}
	}
	return false
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return n
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func makeSnippet(text, needle string) string {
	lower := strings.ToLower(text)
	lowerNeedle := strings.ToLower(needle)
	byteIndex := strings.Index(lower, lowerNeedle)
	if byteIndex < 0 || lowerNeedle == "" {
		return ""
	}

	runes := []rune(text)
	matchIndex := utf8.RuneCountInString(lower[:byteIndex])
	needleLength := utf8.RuneCountInString(lowerNeedle)
	start := max(0, matchIndex-60)
	end := min(len(runes), matchIndex+needleLength+90)
	snippet := string(runes[start:end])
	if start > 0 {
		snippet = "…" + snippet
	}
	if end < len(runes) {
		snippet += "…"
	}
	return strings.TrimSpace(snippet)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}
