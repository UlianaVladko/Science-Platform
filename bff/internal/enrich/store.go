// Package enrich keeps user-generated state that the consolidated API
// deliberately excludes (Q&A, likes, follows) plus discipline facts.
// State lives in memory and resets on restart, matching the upstream mock.
package enrich

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"scienceplatform/bff/internal/corpus"
	"scienceplatform/bff/internal/domain"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation")
)

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Answer struct {
	ID        int           `json:"id"`
	Body      string        `json:"body"`
	Author    Person        `json:"author"`
	Votes     int           `json:"votes"`
	CreatedAt string        `json:"created_at"`
	Source    domain.Source `json:"source"`
}

type Question struct {
	ID           int           `json:"id"`
	Title        string        `json:"title"`
	Body         string        `json:"body"`
	Author       Person        `json:"author"`
	Category     Category      `json:"category"`
	AnswersCount int           `json:"answers_count"`
	ViewsCount   int           `json:"views_count"`
	Votes        int           `json:"votes"`
	CreatedAt    string        `json:"created_at"`
	Answers      []Answer      `json:"answers,omitempty"`
	Source       domain.Source `json:"source"`
}

type Category struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Fact struct {
	DisciplineSlug string        `json:"discipline_slug"`
	FactType       string        `json:"fact_type"`
	Text           string        `json:"text"`
	Value          any           `json:"value"`
	Period         string        `json:"period"`
	Source         domain.Source `json:"source"`
}

// Store is the in-memory user-state container. All methods are safe for
// concurrent use.
type Store struct {
	mu sync.RWMutex

	questions map[int]*Question
	nextQ     int
	nextA     int
	votes     map[string]map[string]int // answer id -> voter -> vote
	likes     map[string]map[string]bool
	followers map[string]map[string]bool
	views     map[string]int
	corpus    *corpus.Corpus
	startedAt time.Time
}

func NewStore(c *corpus.Corpus) *Store {
	s := &Store{
		questions: map[int]*Question{},
		nextQ:     0,
		nextA:     0,
		votes:     map[string]map[string]int{},
		likes:     map[string]map[string]bool{},
		followers: map[string]map[string]bool{},
		views:     map[string]int{},
		corpus:    c,
		startedAt: corpus.AsOf,
	}
	s.seedQuestions()
	return s
}

func (s *Store) seedQuestions() {
	seeds := []struct {
		title, body, authorID, catSlug string
		answers                        []string
		votes                          int
		views                          int
	}{
		{"Как работает CRISPR-Cas9?", "Разбираю механизм направленного редактирования генома. Интересует роль tracrRNA и PAM-последовательности.", "demo-author-3", "genetics",
			[]string{"Cas9 — эндонуклеаза, разрезающая ДНК по гидовой РНК. Ключевой момент — PAM.", "Добавлю: точность зависит от длины спейсера."}, 7, 320},
		{"Почему графовые базы данных лучше для цепочек цитирования?", "В реляционной базе обход переменной глубины требует рекурсивного CTE. Какие ограничения это накладывает?", "demo-author-1", "databases",
			[]string{"На глубине больше 4–5 recursive CTE деградирует, а индексы перестают помогать."}, 12, 480},
		{"Как оценивать качество кластеризации научных статей?", "Кроме silhouette, что ещё имеет смысл смотреть применительно к текстовым корпусам?", "demo-author-7", "machine-learning",
			[]string{"Silhouette + согласованность ключевых слов кластера. Часто помогает ручная проверка выборки."}, 5, 210},
		{"Что даёт pgvector против отдельного векторного хранилища?", "Выбираем хранилище для эмбеддингов чанков. Стоит ли держать векторы рядом с метаданными?", "demo-author-11", "databases",
			[]string{"Плюс — транзакционная согласованность с метаданными. Минус — нагрузка на основной инстанс."}, 9, 380},
		{"Медицинская визуализация: где реально нужны нейросети?", "В каких задачах диагностики выигрыш от нейросетей уже доказан в публикациях?", "demo-author-15", "medical-imaging",
			[]string{"Сегментация опухолей и маммография — наиболее зрелые направления."}, 4, 190},
		{"Как воспроизводить эксперименты по NLP-извлечению понятий?", "Какой минимальный набор артефактов нужно сохранять, чтобы результат можно было повторить?", "demo-author-22", "nlp",
			[]string{"Версия модели, промпт, версия корпуса и сид генератора."}, 6, 260},
	}
	for _, seed := range seeds {
		s.nextQ++
		catName := seed.catSlug
		if cat, ok := s.corpus.CategoryBySlug[seed.catSlug]; ok {
			catName = cat.Name
		}
		q := &Question{
			ID: s.nextQ, Title: seed.title, Body: seed.body,
			Author:   Person{ID: seed.authorID, Name: s.authorName(seed.authorID)},
			Category: Category{Slug: seed.catSlug, Name: catName},
			Votes:    seed.votes, ViewsCount: seed.views,
			CreatedAt: corpus.AsOf.AddDate(0, 0, -30+s.nextQ*3).Format(time.RFC3339),
			Source:    domain.SourceBFF,
		}
		for i, body := range seed.answers {
			s.nextA++
			q.Answers = append(q.Answers, Answer{
				ID: s.nextA, Body: body,
				Author:    Person{ID: "demo-author-" + itoa(5+i*4), Name: s.authorName("demo-author-" + itoa(5+i*4))},
				Votes:     seed.votes - i*2,
				CreatedAt: corpus.AsOf.AddDate(0, 0, -20+s.nextQ*3+i).Format(time.RFC3339),
				Source:    domain.SourceBFF,
			})
		}
		q.AnswersCount = len(q.Answers)
		s.questions[q.ID] = q
	}
}

func (s *Store) authorName(id string) string {
	if a, ok := s.corpus.AuthorByID[id]; ok {
		return a.Name
	}
	return "Неизвестный автор"
}

type QuestionQuery struct {
	Category string
	Sort     string
	Page     int
	PageSize int
}

func (s *Store) ListQuestions(q QuestionQuery) ([]Question, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Question, 0, len(s.questions))
	for _, question := range s.questions {
		if q.Category != "" && question.Category.Slug != q.Category {
			continue
		}
		cp := *question
		cp.Answers = nil
		items = append(items, cp)
	}
	sort.Slice(items, func(i, j int) bool {
		if q.Sort == "top" {
			if items[i].Votes != items[j].Votes {
				return items[i].Votes > items[j].Votes
			}
			return items[i].AnswersCount > items[j].AnswersCount
		}
		return items[i].CreatedAt > items[j].CreatedAt
	})
	total := len(items)
	start := (q.Page - 1) * q.PageSize
	if start > total {
		start = total
	}
	end := start + q.PageSize
	if end > total {
		end = total
	}
	return items[start:end], total
}

func (s *Store) Question(id int) (*Question, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	question, ok := s.questions[id]
	if !ok {
		return nil, ErrNotFound
	}
	question.ViewsCount++
	cp := *question
	cp.Answers = append([]Answer{}, question.Answers...)
	return &cp, nil
}

func (s *Store) CreateQuestion(title, body, catSlug, authorName string) (*Question, error) {
	if len([]rune(strings.TrimSpace(title))) < 5 || len([]rune(title)) > 200 {
		return nil, ErrValidation
	}
	if len([]rune(strings.TrimSpace(body))) < 1 || len([]rune(body)) > 5000 {
		return nil, ErrValidation
	}
	cat, ok := s.corpus.CategoryBySlug[catSlug]
	if !ok {
		return nil, ErrNotFound
	}
	if authorName == "" {
		authorName = "Гость"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextQ++
	q := &Question{
		ID: s.nextQ, Title: strings.TrimSpace(title), Body: body,
		Author:    Person{ID: "guest", Name: authorName},
		Category:  Category{Slug: catSlug, Name: cat.Name},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Source:    domain.SourceBFF,
	}
	s.questions[q.ID] = q
	cp := *q
	return &cp, nil
}

func (s *Store) AddAnswer(questionID int, body, authorName string) (*Answer, error) {
	if len([]rune(strings.TrimSpace(body))) < 1 || len([]rune(body)) > 5000 {
		return nil, ErrValidation
	}
	if authorName == "" {
		authorName = "Гость"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	q, ok := s.questions[questionID]
	if !ok {
		return nil, ErrNotFound
	}
	s.nextA++
	a := Answer{
		ID: s.nextA, Body: body,
		Author:    Person{ID: "guest", Name: authorName},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Source:    domain.SourceBFF,
	}
	q.Answers = append(q.Answers, a)
	q.AnswersCount = len(q.Answers)
	return &a, nil
}

func (s *Store) Vote(answerID int, voter string, vote int) (int, error) {
	if vote != 1 && vote != -1 {
		return 0, ErrValidation
	}
	if voter == "" {
		voter = "guest"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, q := range s.questions {
		for i := range q.Answers {
			if q.Answers[i].ID != answerID {
				continue
			}
			key := itoa(answerID)
			if s.votes[key] == nil {
				s.votes[key] = map[string]int{}
			}
			previous := s.votes[key][voter]
			q.Answers[i].Votes += vote - previous
			s.votes[key][voter] = vote
			return q.Answers[i].Votes, nil
		}
	}
	return 0, ErrNotFound
}

func (s *Store) Like(publicationID, viewer string, liked bool) domain.LikeState {
	if viewer == "" {
		viewer = "guest"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.likes[publicationID] == nil {
		s.likes[publicationID] = map[string]bool{}
	}
	if liked {
		s.likes[publicationID][viewer] = true
	} else {
		delete(s.likes[publicationID], viewer)
	}
	return domain.LikeState{
		PublicationID: publicationID,
		Liked:         s.likes[publicationID][viewer],
		LikesCount:    len(s.likes[publicationID]),
	}
}

func (s *Store) LikeState(publicationID, viewer string) domain.LikeState {
	if viewer == "" {
		viewer = "guest"
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return domain.LikeState{
		PublicationID: publicationID,
		Liked:         s.likes[publicationID][viewer],
		LikesCount:    len(s.likes[publicationID]),
	}
}

func (s *Store) Follow(authorID, viewer string, follow bool) (bool, int) {
	if viewer == "" {
		viewer = "guest"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.followers[authorID] == nil {
		s.followers[authorID] = map[string]bool{}
	}
	if follow {
		s.followers[authorID][viewer] = true
	} else {
		delete(s.followers[authorID], viewer)
	}
	return s.followers[authorID][viewer], len(s.followers[authorID])
}

func (s *Store) FollowerState(authorID, viewer string) (bool, int) {
	if viewer == "" {
		viewer = "guest"
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.followers[authorID][viewer], len(s.followers[authorID])
}

func (s *Store) RegisterView(publicationID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.views[publicationID]++
	return s.views[publicationID]
}

func (s *Store) ViewDelta(publicationID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.views[publicationID]
}

// FunFact returns a deterministic-per-day fact for a discipline.
func (s *Store) FunFact(slug string) (Fact, bool) {
	seed, ok := findSeed(slug)
	if !ok {
		return Fact{}, false
	}
	dayIndex := int(corpus.AsOf.Unix() / 86400)
	pool := s.factPool(seed)
	if len(pool) == 0 {
		return Fact{
			DisciplineSlug: slug, FactType: "fallback",
			Text:   "Данные по направлению «" + seed.Name + "» ещё накапливаются",
			Period: "2026", Source: domain.SourceBFF,
		}, true
	}
	return pool[dayIndex%len(pool)], true
}

func (s *Store) factPool(seed corpus.DisciplineSeed) []Fact {
	var pubs []*domain.Publication
	for i := range s.corpus.Publications {
		p := &s.corpus.Publications[i]
		for _, cat := range p.Categories {
			if cat.Slug == seed.Slug || cat.ID == seed.ID {
				pubs = append(pubs, p)
				break
			}
		}
	}
	if len(pubs) == 0 {
		return nil
	}
	citations := 0
	keywordCount := map[string]int{}
	var mostCited *domain.Publication
	for _, p := range pubs {
		citations += p.Metrics.Citations
		for _, k := range p.Keywords {
			keywordCount[k]++
		}
		if mostCited == nil || p.Metrics.Citations > mostCited.Metrics.Citations {
			mostCited = p
		}
	}
	topKeyword := ""
	topCount := 0
	for k, v := range keywordCount {
		if v > topCount {
			topKeyword, topCount = k, v
		}
	}
	avgCoauthors := 0.0
	totalAuthors := 0
	for _, p := range pubs {
		totalAuthors += len(p.Authors)
	}
	if len(pubs) > 0 {
		avgCoauthors = float64(totalAuthors) / float64(len(pubs))
	}
	pool := []Fact{
		{DisciplineSlug: seed.Slug, FactType: "growth_yoy",
			Text:  "В направлении «" + seed.Name + "» в этом году на 23% больше публикаций, чем в прошлом",
			Value: 23, Period: "2025 → 2026", Source: domain.SourceBFF},
		{DisciplineSlug: seed.Slug, FactType: "top_keyword",
			Text:  "Чаще всего в «" + seed.Name + "» встречается понятие «" + topKeyword + "»",
			Value: topKeyword, Period: "2026", Source: domain.SourceBFF},
	}
	if mostCited != nil {
		pool = append(pool, Fact{
			DisciplineSlug: seed.Slug, FactType: "most_cited",
			Text:  "Самая цитируемая работа направления — «" + truncate(mostCited.Title, 70) + "» (" + itoa(mostCited.Metrics.Citations) + " цитирований)",
			Value: mostCited.ID, Period: "2026", Source: domain.SourceBFF,
		})
	}
	pool = append(pool,
		Fact{DisciplineSlug: seed.Slug, FactType: "new_authors",
			Text:  "За последний месяц к направлению присоединилось 7 новых авторов",
			Value: 7, Period: "октябрь 2026", Source: domain.SourceBFF},
		Fact{DisciplineSlug: seed.Slug, FactType: "avg_coauthors",
			Text:  "В среднем у публикации направления «" + seed.Name + "» " + formatFloat(avgCoauthors) + " соавтора",
			Value: roundTo(avgCoauthors, 1), Period: "2026", Source: domain.SourceBFF},
	)
	return pool
}

func findSeed(slug string) (corpus.DisciplineSeed, bool) {
	for _, seed := range corpus.DisciplineSeeds {
		if seed.Slug == slug || seed.ID == slug {
			return seed, true
		}
	}
	return corpus.DisciplineSeed{}, false
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func roundTo(v float64, digits int) float64 {
	mult := 1.0
	for i := 0; i < digits; i++ {
		mult *= 10
	}
	return float64(int(v*mult+0.5)) / mult
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
