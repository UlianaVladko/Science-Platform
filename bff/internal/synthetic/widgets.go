// Package synthetic computes the dashboard widgets the consolidated API does
// not expose (its 20 operations replaced 20 widget endpoints with 4 universal
// ones). Aggregates are derived from the demo corpus and always labelled
// source=demo so the UI can badge them.
package synthetic

import (
	"math"
	"sort"
	"strings"

	"scienceplatform/bff/internal/corpus"
	"scienceplatform/bff/internal/domain"
)

type Widget struct {
	Widget   string        `json:"widget"`
	Source   domain.Source `json:"source"`
	DataAsOf string        `json:"data_as_of"`
	Ready    bool          `json:"ready"`
	Notice   string        `json:"notice,omitempty"`
	Payload  any           `json:"payload"`
}

type Service struct {
	corpus   *corpus.Corpus
	nlpReady bool
}

func New(c *corpus.Corpus, nlpReady bool) *Service {
	return &Service{corpus: c, nlpReady: nlpReady}
}

// WidgetNames is the closed set of dashboard widgets, mirroring ui-api-mapping §9.
var WidgetNames = []string{
	"publications-timeline", "intake-timeline", "by-category", "top-authors",
	"keywords-cloud", "top-concepts", "lemmas-frequency", "by-language", "by-type",
	"top-viewed", "top-cited", "top-downloaded", "by-status", "by-source",
	"coauthors-distribution", "citation-graph", "summary", "by-affiliation",
	"activity-heatmap", "citations-timeline", "top-publications",
}

func (s *Service) Ready() map[string]bool {
	out := map[string]bool{}
	for _, name := range WidgetNames {
		out[name] = true
	}
	if !s.nlpReady {
		out["keywords-cloud"] = false
		out["top-concepts"] = false
		out["lemmas-frequency"] = false
	}
	return out
}

func (s *Service) Build(widget, category, period string) (Widget, bool) {
	known := false
	for _, name := range WidgetNames {
		if name == widget {
			known = true
			break
		}
	}
	if !known {
		return Widget{}, false
	}
	pubs := s.selectPubs(category, period)
	w := Widget{
		Widget:   widget,
		Source:   domain.SourceDemo,
		DataAsOf: corpus.AsOf.Format("2006-01-02T15:04:05Z"),
		Ready:    true,
	}
	if !s.nlpReady {
		switch widget {
		case "keywords-cloud", "top-concepts", "lemmas-frequency":
			w.Ready = false
			w.Notice = "Данные готовятся: результаты NLP-пайплайна ещё не поступили в витрину"
			w.Payload = map[string]any{}
			return w, true
		}
	}
	w.Payload = s.payload(widget, pubs)
	return w, true
}

func (s *Service) selectPubs(category, period string) []*domain.Publication {
	ids := map[string]bool{}
	if category != "" {
		if cat, ok := s.corpus.CategoryByID[category]; ok {
			ids[cat.ID] = true
			for i := range cat.Children {
				ids[cat.Children[i].ID] = true
			}
		} else if cat, ok := s.corpus.CategoryBySlug[category]; ok {
			ids[cat.ID] = true
			for i := range cat.Children {
				ids[cat.Children[i].ID] = true
			}
		}
	}
	out := make([]*domain.Publication, 0, len(s.corpus.Publications))
	for i := range s.corpus.Publications {
		p := &s.corpus.Publications[i]
		if len(ids) > 0 {
			ok := false
			for _, c := range p.Categories {
				if ids[c.ID] {
					ok = true
				}
			}
			if !ok {
				continue
			}
		}
		if period != "" && period != "all" && p.PublicationDate != nil && (*p.PublicationDate)[:4] != period {
			continue
		}
		out = append(out, p)
	}
	return out
}

type seriesPoint struct {
	Year  int `json:"year"`
	Count int `json:"count"`
}

type slice struct {
	Slug  string `json:"slug,omitempty"`
	Value string `json:"value,omitempty"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Color string `json:"color,omitempty"`
}

type topItem struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Subtitle  string `json:"subtitle,omitempty"`
	Count     int    `json:"count"`
	Citations int    `json:"citations,omitempty"`
	Metric    string `json:"metric,omitempty"`
	Category  string `json:"category,omitempty"`
	Year      *int   `json:"year,omitempty"`
}

func (s *Service) payload(widget string, pubs []*domain.Publication) any {
	switch widget {
	case "publications-timeline":
		years := map[int]int{}
		for _, p := range pubs {
			if p.PublicationDate == nil {
				continue
			}
			years[atoi((*p.PublicationDate)[:4])]++
		}
		series := []seriesPoint{}
		for y := 2015; y <= 2026; y++ {
			if c, ok := years[y]; ok {
				series = append(series, seriesPoint{Year: y, Count: c})
			}
		}
		return map[string]any{"series": series}

	case "intake-timeline":
		months := map[string]int{}
		for i := 0; i < 12; i++ {
			key := corpus.AsOf.AddDate(0, -(11 - i), 0).Format("2006-01")
			months[key] = 6 + (i*13)%17
		}
		keys := make([]string, 0, len(months))
		for k := range months {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		series := make([]map[string]any, 0, len(keys))
		for _, k := range keys {
			series = append(series, map[string]any{"date": k, "count": months[k]})
		}
		return map[string]any{"series": series}

	case "by-category":
		counts := map[string]int{}
		for _, p := range pubs {
			if len(p.Categories) > 0 {
				counts[p.Categories[0].Slug]++
			}
		}
		slices := make([]slice, 0, len(counts))
		for _, seed := range corpus.DisciplineSeeds {
			if c, ok := counts[seed.Slug]; ok {
				slices = append(slices, slice{Slug: seed.Slug, Name: seed.Name, Count: c, Color: seed.Accent})
			}
		}
		sort.Slice(slices, func(i, j int) bool { return slices[i].Count > slices[j].Count })
		return map[string]any{"slices": slices}

	case "top-authors":
		type acc struct {
			name      string
			count     int
			citations int
			cat       string
		}
		board := map[string]*acc{}
		for _, p := range pubs {
			for _, a := range p.Authors {
				e, ok := board[a.ID]
				if !ok {
					e = &acc{name: a.Name}
					board[a.ID] = e
				}
				e.count++
				e.citations += p.Metrics.Citations
				if e.cat == "" && len(p.Categories) > 0 {
					e.cat = p.Categories[0].Name
				}
			}
		}
		items := make([]topItem, 0, len(board))
		for id, e := range board {
			items = append(items, topItem{ID: id, Label: e.name, Count: e.count, Citations: e.citations, Category: e.cat})
		}
		sort.Slice(items, func(i, j int) bool {
			if items[i].Count != items[j].Count {
				return items[i].Count > items[j].Count
			}
			return items[i].Citations > items[j].Citations
		})
		if len(items) > 10 {
			items = items[:10]
		}
		return map[string]any{"items": items}

	case "keywords-cloud":
		weighted := map[string]int{}
		for _, p := range pubs {
			for _, k := range p.Keywords {
				weighted[k] += 3
			}
			for _, c := range p.Concepts {
				weighted[strings.ToLower(c.Name)] += 5
			}
		}
		words := make([]map[string]any, 0, len(weighted))
		for text, weight := range weighted {
			words = append(words, map[string]any{"text": text, "weight": weight})
		}
		sort.Slice(words, func(i, j int) bool {
			return words[i]["weight"].(int) > words[j]["weight"].(int)
		})
		if len(words) > 60 {
			words = words[:60]
		}
		return map[string]any{"words": words}

	case "top-concepts":
		counts := map[string]int{}
		names := map[string]string{}
		for _, p := range pubs {
			for _, c := range p.Concepts {
				counts[c.ID]++
				names[c.ID] = c.Name
			}
		}
		items := make([]topItem, 0, len(counts))
		for id, c := range counts {
			items = append(items, topItem{ID: id, Label: names[id], Count: c})
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })
		if len(items) > 10 {
			items = items[:10]
		}
		return map[string]any{"items": items}

	case "lemmas-frequency":
		freq := s.corpus.LemmaFrequencies()
		items := make([]map[string]any, 0, len(freq))
		for lemma, count := range freq {
			items = append(items, map[string]any{"lemma": lemma, "count": count})
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i]["count"].(int) > items[j]["count"].(int)
		})
		if len(items) > 12 {
			items = items[:12]
		}
		return map[string]any{"items": items}

	case "by-language":
		counts := map[string]int{}
		for _, p := range pubs {
			counts[p.Language]++
		}
		names := map[string]string{"ru": "Русский", "en": "Английский"}
		slices := make([]slice, 0, len(counts))
		for v, c := range counts {
			slices = append(slices, slice{Value: v, Name: names[v], Count: c})
		}
		sort.Slice(slices, func(i, j int) bool { return slices[i].Count > slices[j].Count })
		return map[string]any{"slices": slices}

	case "by-type":
		counts := map[string]int{}
		for _, p := range pubs {
			counts[p.PublicationType]++
		}
		names := map[string]string{
			"article": "Статья", "journal_publication": "Журнальная публикация",
			"book": "Книга", "monograph": "Монография", "preprint": "Препринт", "other": "Другое",
		}
		slices := make([]slice, 0, len(counts))
		for v, c := range counts {
			slices = append(slices, slice{Value: v, Name: names[v], Count: c})
		}
		sort.Slice(slices, func(i, j int) bool { return slices[i].Count > slices[j].Count })
		return map[string]any{"slices": slices}

	case "top-viewed", "top-cited", "top-downloaded", "top-publications":
		metric := "views"
		switch widget {
		case "top-cited":
			metric = "citations"
		case "top-downloaded":
			metric = "downloads"
		}
		items := make([]topItem, 0, len(pubs))
		for _, p := range pubs {
			value := 0
			switch metric {
			case "citations":
				value = p.Metrics.Citations
			case "downloads":
				value = p.Metrics.Downloads
			default:
				value = p.Metrics.Views
			}
			var year *int
			if p.PublicationDate != nil {
				y := atoi((*p.PublicationDate)[:4])
				year = &y
			}
			cat := ""
			if len(p.Categories) > 0 {
				cat = p.Categories[0].Name
			}
			items = append(items, topItem{
				ID: p.ID, Label: p.Title, Count: value, Citations: p.Metrics.Citations,
				Metric: metric, Category: cat, Year: year,
				Subtitle: strings.Join(authorNames(p), ", "),
			})
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Count > items[j].Count })
		if len(items) > 10 {
			items = items[:10]
		}
		return map[string]any{"metric": metric, "items": items}

	case "by-status":
		counts := map[string]int{}
		for _, p := range pubs {
			counts[p.ProcessingStatus]++
		}
		slices := make([]slice, 0, len(counts))
		for v, c := range counts {
			slices = append(slices, slice{Value: v, Name: statusName(v), Count: c})
		}
		sort.Slice(slices, func(i, j int) bool { return slices[i].Count > slices[j].Count })
		return map[string]any{"slices": slices}

	case "by-source":
		slices := make([]slice, 0, 4)
		counts := map[string]int{}
		for i := range s.corpus.Publications {
			counts[sourceOf(i)]++
		}
		names := map[string]string{"crossref": "Crossref", "arxiv": "arXiv", "openlibrary": "OpenLibrary", "manual": "Ручной ввод"}
		for v, c := range counts {
			slices = append(slices, slice{Value: v, Name: names[v], Count: c})
		}
		sort.Slice(slices, func(i, j int) bool { return slices[i].Count > slices[j].Count })
		return map[string]any{"slices": slices}

	case "coauthors-distribution":
		buckets := map[int]int{}
		for _, p := range pubs {
			buckets[len(p.Authors)]++
		}
		out := make([]map[string]any, 0, len(buckets))
		keys := make([]int, 0, len(buckets))
		for k := range buckets {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			out = append(out, map[string]any{"bucket": itoa(k), "count": buckets[k]})
		}
		return map[string]any{"buckets": out}

	case "citation-graph":
		return s.citationGraph(pubs)

	case "summary":
		citations := 0
		downloads := 0
		thisPeriod := 0
		for _, p := range pubs {
			citations += p.Metrics.Citations
			downloads += p.Metrics.Downloads
			if p.PublicationDate != nil && (*p.PublicationDate)[:4] == "2026" {
				thisPeriod++
			}
		}
		avg := 0.0
		if len(pubs) > 0 {
			avg = math.Round(float64(citations)/float64(len(pubs))*100) / 100
		}
		return domain.CorpusSummary{
			TotalPublications:          len(pubs),
			TotalAuthors:               s.corpusAuthorCount(pubs),
			TotalCitations:             citations,
			TotalDownloads:             downloads,
			AvgCitationsPerPublication: avg,
			PublicationsThisPeriod:     thisPeriod,
			NewAuthorsThisPeriod:       9,
			Period:                     "2026",
			DataAsOf:                   corpus.AsOf.Format("2006-01-02T15:04:05Z"),
		}

	case "by-affiliation":
		counts := map[string]int{}
		for _, p := range pubs {
			seen := map[string]bool{}
			for _, a := range p.Authors {
				for _, org := range a.Affiliations {
					if !seen[org.Name] {
						seen[org.Name] = true
						counts[org.Name]++
					}
				}
			}
		}
		items := make([]map[string]any, 0, len(counts))
		for name, c := range counts {
			items = append(items, map[string]any{"name": name, "count": c})
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i]["count"].(int) > items[j]["count"].(int)
		})
		if len(items) > 12 {
			items = items[:12]
		}
		return map[string]any{"items": items}

	case "activity-heatmap":
		days := []string{}
		for i := 29; i >= 0; i-- {
			days = append(days, corpus.AsOf.AddDate(0, 0, -i).Format("2006-01-02"))
		}
		categories := make([]string, 0, 6)
		for i := 0; i < 6 && i < len(corpus.DisciplineSeeds); i++ {
			categories = append(categories, corpus.DisciplineSeeds[i].Name)
		}
		values := make([][3]int, 0, len(days)*len(categories))
		for di := range days {
			for ci := range categories {
				v := (di*7 + ci*13) % 9
				if (di+ci)%5 == 0 {
					v += 3
				}
				values = append(values, [3]int{di, ci, v})
			}
		}
		return map[string]any{"days": days, "categories": categories, "values": values}

	case "citations-timeline":
		years := s.corpus.CitationYears()
		series := []seriesPoint{}
		cumulative := 0
		for y := 2015; y <= 2026; y++ {
			cumulative += years[itoa(y)]
			series = append(series, seriesPoint{Year: y, Count: cumulative + (y-2015)*12})
		}
		return map[string]any{"series": series, "cumulative": true}
	}
	return map[string]any{}
}

func (s *Service) citationGraph(pubs []*domain.Publication) map[string]any {
	allowed := map[string]bool{}
	for _, p := range pubs {
		allowed[p.ID] = true
	}
	nodes := []domain.GraphNode{}
	for _, p := range pubs {
		cat := ""
		if len(p.Categories) > 0 {
			cat = p.Categories[0].Slug
		}
		nodes = append(nodes, domain.GraphNode{
			ID: p.ID, Type: "publication", Label: p.Title, Category: cat,
			Weight: float64(p.Metrics.Citations + 1),
		})
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Weight > nodes[j].Weight })
	truncated := false
	if len(nodes) > 60 {
		nodes = nodes[:60]
		truncated = true
	}
	keep := map[string]bool{}
	for _, n := range nodes {
		keep[n.ID] = true
	}
	edges := []domain.GraphEdge{}
	for _, ref := range s.corpus.References {
		if ref.CitedPublicationID == nil {
			continue
		}
		if keep[ref.CitingPublicationID] && keep[*ref.CitedPublicationID] && allowed[ref.CitingPublicationID] {
			edges = append(edges, domain.GraphEdge{
				Source: ref.CitingPublicationID, Target: *ref.CitedPublicationID, Type: "cites",
			})
		}
	}
	return map[string]any{"nodes": nodes, "edges": edges, "truncated": truncated}
}

func (s *Service) corpusAuthorCount(pubs []*domain.Publication) int {
	seen := map[string]bool{}
	for _, p := range pubs {
		for _, a := range p.Authors {
			seen[a.ID] = true
		}
	}
	return len(seen)
}

func authorNames(p *domain.Publication) []string {
	out := make([]string, 0, len(p.Authors))
	for _, a := range p.Authors {
		out = append(out, a.Name)
	}
	return out
}

func statusName(v string) string {
	switch v {
	case "published", "completed":
		return "Опубликовано"
	case "pending":
		return "Ожидает обработки"
	case "processing":
		return "Обрабатывается"
	case "failed":
		return "Ошибка"
	}
	return v
}

func sourceOf(index int) string {
	switch index % 10 {
	case 0, 1, 2, 3, 4:
		return "crossref"
	case 5, 6, 7:
		return "arxiv"
	case 8:
		return "openlibrary"
	}
	return "manual"
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

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [12]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
