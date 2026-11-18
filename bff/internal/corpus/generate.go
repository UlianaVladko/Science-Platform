package corpus

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"scienceplatform/bff/internal/domain"
)

// AsOf is the fixed snapshot date of the demo corpus. Deterministic generation
// keeps dashboards and graphs stable across restarts.
var AsOf = time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)

// DisciplineSeed describes one top-level discipline and its subcategories.
type DisciplineSeed struct {
	ID            string
	Slug          string
	Name          string
	Description   string
	Icon          string
	Accent        string
	Weight        float64
	Subcategories []struct {
		ID   string
		Slug string
		Name string
	}
}

// DisciplineSeeds is the fixed classification tree: 15 disciplines, 38 subcategories.
var DisciplineSeeds = []DisciplineSeed{
	{"cat-cs", "computer-science", "Информатика", "Алгоритмы, системы, вычислительные методы", "cpu", "#6d5efc", 1.35,
		[]struct{ ID, Slug, Name string }{
			{"cat-nlp", "nlp", "Обработка текста"},
			{"cat-ml", "machine-learning", "Машинное обучение"},
			{"cat-db", "databases", "Базы данных"},
			{"cat-sec", "cybersecurity", "Кибербезопасность"},
			{"cat-software", "software-engineering", "Программная инженерия"},
			{"cat-networks", "computer-networks", "Компьютерные сети"},
		}},
	{"cat-bio", "biology", "Биология", "Жизненные науки, генетика, экология", "dna", "#22b38a", 1.2,
		[]struct{ ID, Slug, Name string }{
			{"cat-genetics", "genetics", "Генетика"},
			{"cat-ecology", "ecology", "Экология"},
			{"cat-neuro", "neurobiology", "Нейробиология"},
		}},
	{"cat-phys", "physics", "Физика", "Квантовые явления, конденсированное состояние", "atom", "#3f8cff", 1.1,
		[]struct{ ID, Slug, Name string }{
			{"cat-quantum", "quantum-physics", "Квантовая физика"},
			{"cat-condensed", "condensed-matter", "Физика конденсированного состояния"},
			{"cat-astro", "astrophysics", "Астрофизика"},
			{"cat-optics", "optics", "Оптика"},
		}},
	{"cat-chem", "chemistry", "Химия", "Синтез, катализ, материалы", "flask", "#f0932b", 0.95,
		[]struct{ ID, Slug, Name string }{
			{"cat-catalysis", "catalysis", "Катализ"},
			{"cat-materials", "materials-chemistry", "Химия материалов"},
		}},
	{"cat-med", "medicine", "Медицина", "Клинические исследования, диагностика", "brain", "#e0516f", 1.25,
		[]struct{ ID, Slug, Name string }{
			{"cat-onco", "oncology", "Онкология"},
			{"cat-imaging", "medical-imaging", "Медицинская визуализация"},
			{"cat-epidem", "epidemiology", "Эпидемиология"},
		}},
	{"cat-math", "mathematics", "Математика", "Анализ, алгебра, дискретная математика", "math", "#8e44ad", 0.9,
		[]struct{ ID, Slug, Name string }{
			{"cat-probability", "probability", "Теория вероятностей"},
			{"cat-graphtheory", "graph-theory", "Теория графов"},
		}},
	{"cat-earth", "earth-sciences", "Науки о Земле", "Климат, геология, океанология", "globe", "#0aa3a3", 0.7,
		[]struct{ ID, Slug, Name string }{
			{"cat-climate", "climate", "Климатология"},
			{"cat-geology", "geology", "Геология"},
		}},
	{"cat-econ", "economics", "Экономика", "Рынки, эконометрика, поведение", "chart", "#d6a30c", 0.8,
		[]struct{ ID, Slug, Name string }{
			{"cat-econometrics", "econometrics", "Эконометрика"},
			{"cat-behaviour", "behavioural-economics", "Поведенческая экономика"},
		}},
	{"cat-eng", "engineering", "Инженерия", "Машиностроение, энергетика, строительство", "gear", "#5b6b7d", 1.05,
		[]struct{ ID, Slug, Name string }{
			{"cat-mech", "mechanical-engineering", "Машиностроение"},
			{"cat-energy", "energy-systems", "Энергетика"},
		}},
	{"cat-soc", "social-sciences", "Социальные науки", "Социология, психология, право", "building", "#b06fc9", 0.75,
		[]struct{ ID, Slug, Name string }{
			{"cat-sociology", "sociology", "Социология"},
			{"cat-psychology", "psychology", "Психология"},
		}},
	{"cat-agro", "agriculture", "Сельское хозяйство", "Агрономия, животноводство, почвы", "leaf", "#5aa832", 0.6,
		[]struct{ ID, Slug, Name string }{
			{"cat-agronomy", "agronomy", "Агрономия"},
			{"cat-soil", "soil-science", "Почвоведение"},
		}},
	{"cat-human", "humanities", "Гуманитарные науки", "История, филология, философия", "building", "#a1763f", 0.55,
		[]struct{ ID, Slug, Name string }{
			{"cat-history", "history", "История"},
			{"cat-linguistics", "linguistics", "Лингвистика"},
		}},
	{"cat-security", "information-security", "Информационная безопасность", "Защита данных, криптография", "shield", "#c0392b", 0.85,
		[]struct{ ID, Slug, Name string }{
			{"cat-crypto", "cryptography", "Криптография"},
			{"cat-detection", "intrusion-detection", "Обнаружение вторжений"},
		}},
	{"cat-data", "data-science", "Науки о данных", "Аналитика, визуализация, инженерия данных", "chart", "#2f80ed", 1.15,
		[]struct{ ID, Slug, Name string }{
			{"cat-datamining", "data-mining", "Интеллектуальный анализ данных"},
			{"cat-kg", "knowledge-graphs", "Графы знаний"},
		}},
	{"cat-materials", "materials-science", "Материаловедение", "Металлы, полимеры, композиты", "flask", "#7f8c8d", 0.65,
		[]struct{ ID, Slug, Name string }{
			{"cat-polymers", "polymers", "Полимеры"},
			{"cat-nanomaterials", "nanomaterials", "Наноматериалы"},
		}},
}

var organizations = []domain.Organization{
	{ID: "org-1", Name: "Институт научных вычислений"},
	{ID: "org-2", Name: "Московский государственный университет"},
	{ID: "org-3", Name: "Институт системного программирования"},
	{ID: "org-4", Name: "Национальный исследовательский центр"},
	{ID: "org-5", Name: "Институт биомедицинских технологий"},
	{ID: "org-6", Name: "Санкт-Петербургский политехнический университет"},
	{ID: "org-7", Name: "Институт проблем передачи информации"},
	{ID: "org-8", Name: "Институт катализа"},
	{ID: "org-9", Name: "Институт космических исследований"},
	{ID: "org-10", Name: "Высшая школа экономики"},
	{ID: "org-11", Name: "Институт физики полупроводников"},
	{ID: "org-12", Name: "Центр климатических исследований"},
}

var surnames = []string{
	"Петрова", "Смирнов", "Ким", "Иванов", "Соколов", "Кузнецов", "Волкова", "Морозов",
	"Лебедева", "Козлов", "Новиков", "Егорова", "Павлов", "Семёнова", "Голубев", "Виноградова",
	"Богданов", "Орлова", "Никитин", "Захарова", "Фёдоров", "Михайлова", "Тарасов", "Белова",
	"Комаров", "Акимова", "Крылов", "Гусева", "Титов", "Матвеева", "Киселёв", "Савельев",
}

var firstNames = []string{
	"Анна", "Иван", "Мария", "Дмитрий", "Ольга", "Сергей", "Елена", "Алексей", "Татьяна", "Павел",
	"Ирина", "Никита", "Екатерина", "Артём", "Наталья", "Максим", "Юлия", "Андрей", "Светлана", "Роман",
}

var journals = []string{
	"Вестник вычислительной науки", "Журнал экспериментальной физики", "Труды по геномике",
	"Журнал машинного обучения", "Клиническая медицина", "Известия Академии наук",
	"Математический сборник", "Обзор химических технологий", "Журнал системной инженерии",
	"Электронный научный журнал",
}

var titleTemplates = []string{
	"%s: методы и результаты", "О подходах к %s", "%s в задачах прогнозирования",
	"Опыт применения %s", "Анализ %s на реальных данных", "%s: систематический обзор",
	"Эффективные алгоритмы для %s", "Ограничения современных моделей %s",
	"Экспериментальная проверка %s", "Масштабирование %s: практический опыт",
	"%s в контексте больших данных", "Новый взгляд на %s",
}

var topicPools = map[string][]string{
	"computer-science":     {"графовых алгоритмов", "обработки потоков данных", "распределённого хранения", "компиляторов", "планирования задач"},
	"biology":              {"генного редактирования", "экосистемных моделей", "нейронных механизмов", "эволюции популяций"},
	"physics":              {"квантовых измерений", "сверхпроводимости", "оптики сверхкоротких импульсов", "космических лучей"},
	"chemistry":            {"гетерогенного катализа", "полимерных структур", "функциональных материалов"},
	"medicine":             {"ранней диагностики", "медицинской визуализации", "эпидемиологических моделей", "терапевтических схем"},
	"mathematics":          {"теории вероятностей", "структур графов", "численных методов", "асимптотического анализа"},
	"earth-sciences":       {"климатических моделей", "динамики ледников", "геологических разломов"},
	"economics":            {"рыночных механизмов", "эконометрических оценок", "поведенческих факторов"},
	"engineering":          {"турбинных установок", "энергосистем", "механической прочности"},
	"social-sciences":      {"социальных сетей", "поведенческих исследований", "правовых практик"},
	"agriculture":          {"почвенного плодородия", "урожайности культур", "агрономических практик"},
	"humanities":           {"исторических источников", "структуры языка", "философских оснований"},
	"information-security": {"криптографических протоколов", "обнаружения вторжений", "защиты данных"},
	"data-science":         {"визуальной аналитики", "графов знаний", "инженерии признаков"},
	"materials-science":    {"нанокомпозитных структур", "полимерных покрытий", "металлических сплавов"},
}

var conceptPool = map[string][]string{
	"computer-science":     {"Нейронные сети", "Графовые базы данных", "Распределённые транзакции", "Кэширование", "Семантический поиск"},
	"biology":              {"CRISPR", "Секвенирование", "Экосистемная устойчивость", "Генная экспрессия"},
	"physics":              {"Квантовая запутанность", "Сверхпроводимость", "Интерференция", "Фононы"},
	"chemistry":            {"Каталитические центры", "Полимеризация", "Адсорбция"},
	"medicine":             {"Диагностические биомаркеры", "МРТ", "Иммунотерапия", "Эпидемический порог"},
	"mathematics":          {"Стохастические процессы", "Спектральные методы", "Теория графов"},
	"earth-sciences":       {"Парниковые газы", "Циркуляция океана", "Сейсмическая активность"},
	"economics":            {"Эластичность спроса", "Поведенческие искажения", "Равновесие Нэша"},
	"engineering":          {"Термодинамический цикл", "Надёжность конструкций", "Энергетические потери"},
	"social-sciences":      {"Гомофилия", "Социальный капитал", "Институциональные нормы"},
	"agriculture":          {"Севооборот", "Гумус", "Микробиом почвы"},
	"humanities":           {"Текстология", "Морфология", "Историческая периодизация"},
	"information-security": {"Дифференциальная приватность", "Атака посредником", "Шифрование с открытым ключом"},
	"data-science":         {"Графы знаний", "Эмбеддинги", "Онтологии"},
	"materials-science":    {"Кристаллическая решётка", "Наноструктура", "Композиционные материалы"},
}

var keywordPool = []string{
	"нейросети", "граф", "кластеризация", "оптимизация", "данные", "модель", "алгоритм",
	"анализ", "эксперимент", "метрика", "визуализация", "индексирование", "поиск", "сжатие",
	"семантика", "онтология", "классификация", "регрессия", "корпус", "текст",
}

var lemmas = []string{
	"анализ", "метод", "данные", "модель", "результат", "алгоритм", "система", "исследование",
	"подход", "эксперимент", "задача", "выборка", "оценка", "прогноз", "признак", "структура",
	"решение", "метрика", "процесс", "пример",
}

var sourceTypes = []string{"crossref", "arxiv", "openlibrary", "manual"}

// Corpus is the in-memory demo dataset.
type Corpus struct {
	Publications   []domain.Publication
	Authors        []domain.Author
	Categories     []domain.Category
	References     []domain.Reference
	Concepts       []domain.Concept
	Organizations  []domain.Organization
	Journals       []string
	CreatedAt      time.Time
	ByID           map[string]*domain.Publication
	AuthorByID     map[string]*domain.Author
	CategoryBySlug map[string]*domain.Category
	CategoryByID   map[string]*domain.Category
	citationYears  map[string]int
	lemmasFreq     map[string]int
}

// uppercaseFirst uppercases the first Unicode code point without splitting
// a multi-byte UTF-8 character.
func uppercaseFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

// Generate builds the deterministic demo corpus.
func Generate() *Corpus {
	r := rand.New(rand.NewSource(20261118))
	c := &Corpus{
		CreatedAt:      AsOf,
		ByID:           map[string]*domain.Publication{},
		AuthorByID:     map[string]*domain.Author{},
		CategoryBySlug: map[string]*domain.Category{},
		CategoryByID:   map[string]*domain.Category{},
		citationYears:  map[string]int{},
		lemmasFreq:     map[string]int{},
		Organizations:  organizations,
		Journals:       journals,
	}

	c.buildCategories()
	c.buildAuthors(r)
	c.buildPublications(r)
	c.buildReferences(r)
	c.buildConcepts()
	c.indexAuthors()
	return c
}

func (c *Corpus) buildCategories() {
	for _, seed := range DisciplineSeeds {
		root := domain.Category{
			ID: seed.ID, Slug: seed.Slug, Name: seed.Name,
			Description: ptr(seed.Description), Icon: seed.Icon, Accent: seed.Accent,
			Children: []domain.Category{}, Source: domain.SourceDemo,
		}
		for _, sub := range seed.Subcategories {
			child := domain.Category{
				ID: sub.ID, Slug: sub.Slug, Name: sub.Name,
				ParentID: ptr(seed.ID), Children: []domain.Category{},
				Icon: seed.Icon, Accent: seed.Accent, Source: domain.SourceDemo,
			}
			root.Children = append(root.Children, child)
			c.CategoryBySlug[sub.Slug] = &child
		}
		c.CategoryBySlug[seed.Slug] = &root
		c.CategoryByID[seed.ID] = &root
		for i := range root.Children {
			c.CategoryByID[root.Children[i].ID] = &root.Children[i]
		}
		c.Categories = append(c.Categories, root)
	}
	// Re-point slug map at the slices actually stored in c.Categories.
	for i := range c.Categories {
		root := &c.Categories[i]
		c.CategoryByID[root.ID] = root
		c.CategoryBySlug[root.Slug] = root
		for j := range root.Children {
			child := &root.Children[j]
			c.CategoryByID[child.ID] = child
			c.CategoryBySlug[child.Slug] = child
		}
	}
}

func (c *Corpus) buildAuthors(r *rand.Rand) {
	const authorCount = 64
	for i := 0; i < authorCount; i++ {
		surname := surnames[i%len(surnames)]
		name := firstNames[(i*7)%len(firstNames)]
		patronymic := []string{"Александровна", "Иванович", "Петровна", "Сергеевич", "Николаевна", "Андреевич"}[i%6]
		org := organizations[r.Intn(len(organizations))]
		var orcid *string
		if i%3 != 0 {
			orcid = ptr(fmt.Sprintf("0000-0002-%04d-%04d", 1000+i, 7000+i*3%9999))
		}
		c.Authors = append(c.Authors, domain.Author{
			AuthorSummary: domain.AuthorSummary{
				ID:           fmt.Sprintf("demo-author-%d", i+1),
				Name:         fmt.Sprintf("%s %s %s", surname, name, patronymic),
				Orcid:        orcid,
				Affiliations: []domain.Organization{org},
			},
			Categories: []domain.Category{},
			HIndex:     ptr(r.Intn(30)),
			Source:     domain.SourceDemo,
		})
	}
}

func (c *Corpus) buildPublications(r *rand.Rand) {
	types := []string{"article", "article", "article", "journal_publication", "book", "monograph", "preprint", "other"}
	languages := []string{"ru", "ru", "ru", "en"}
	const pubCount = 218

	for i := 0; i < pubCount; i++ {
		seed := pickDiscipline(r)
		sub := seed.Subcategories[r.Intn(len(seed.Subcategories))]
		cat := c.CategoryBySlug[sub.Slug]
		root := c.CategoryBySlug[seed.Slug]

		year := weightedYear(r, int(AsOf.Year()))
		month := r.Intn(12) + 1
		day := r.Intn(28) + 1
		date := fmt.Sprintf("%04d-%02d-%02d", year, month, day)

		topics := topicPools[seed.Slug]
		topic := topics[r.Intn(len(topics))]
		title := fmt.Sprintf(titleTemplates[r.Intn(len(titleTemplates))], topic)
		title = uppercaseFirst(title)

		authorCount := 1 + r.Intn(4)
		authors := make([]domain.AuthorSummary, 0, authorCount)
		seen := map[string]bool{}
		for len(authors) < authorCount {
			a := c.Authors[r.Intn(len(c.Authors))]
			if seen[a.ID] {
				continue
			}
			seen[a.ID] = true
			authors = append(authors, a.AuthorSummary)
		}

		pubType := types[r.Intn(len(types))]
		lang := languages[r.Intn(len(languages))]

		concepts := conceptPool[seed.Slug]
		pubConcepts := []domain.Concept{}
		for _, cn := range concepts {
			if r.Float64() < 0.45 {
				pubConcepts = append(pubConcepts, domain.Concept{
					ID: "demo-concept-" + slugify(cn), Name: cn,
				})
			}
		}
		if len(pubConcepts) == 0 {
			pubConcepts = append(pubConcepts, domain.Concept{
				ID: "demo-concept-" + slugify(concepts[0]), Name: concepts[0],
			})
		}

		keywords := []string{}
		for r.Float64() < 0.7 && len(keywords) < 4 {
			kw := keywordPool[r.Intn(len(keywordPool))]
			if !contains(keywords, kw) {
				keywords = append(keywords, kw)
			}
		}

		age := int(AsOf.Year()) - year
		baseCitations := int(math.Max(0, 30.0-float64(age)*2.4)) + r.Intn(35)
		views := baseCitations*17 + r.Intn(900) + 40
		downloads := views/9 + r.Intn(40)

		abstract := fmt.Sprintf(
			"Работа посвящена %s. Рассматриваются методы обработки и анализа данных в области «%s». "+
				"Приведены результаты экспериментов, оценки качества и сравнение с существующими подходами.",
			strings.TrimPrefix(topic, ""), cat.Name)

		id := fmt.Sprintf("demo-%d", i+1)
		pub := domain.Publication{
			PublicationSummary: domain.PublicationSummary{
				ID:              id,
				Title:           title,
				Abstract:        ptr(abstract),
				PublicationType: pubType,
				PublicationDate: ptr(date),
				Language:        lang,
				Authors:         authors,
				Metrics:         domain.Metrics{Citations: baseCitations, Views: views, Downloads: downloads},
				Source:          domain.SourceDemo,
			},
			ExternalIdentifiers: []domain.ExternalIdentifier{},
			Journal:             nil,
			Publisher:           nil,
			Categories:          []domain.Category{*cat, *root},
			Keywords:            keywords,
			Concepts:            pubConcepts,
			Files:               []domain.PublicationFile{},
			ProcessingStatus:    "completed",
			DOI:                 ptr(fmt.Sprintf("10.9000/demo.%d", 1000+i)),
		}
		if r.Float64() < 0.4 {
			pub.ExternalIdentifiers = append(pub.ExternalIdentifiers,
				domain.ExternalIdentifier{Scheme: "doi", Value: *pub.DOI})
		}
		if pubType == "article" || pubType == "journal_publication" || pubType == "preprint" {
			if r.Float64() < 0.65 {
				pub.Journal = ptr(journals[r.Intn(len(journals))])
			}
		}
		if pubType == "book" || pubType == "monograph" {
			pub.Publisher = ptr("Научное издательство")
		}
		c.citationYears[date[:4]] += 0
		c.Publications = append(c.Publications, pub)
		c.ByID[id] = &c.Publications[len(c.Publications)-1]
	}
	// Fix pointers after slice growth.
	c.ByID = map[string]*domain.Publication{}
	for i := range c.Publications {
		c.ByID[c.Publications[i].ID] = &c.Publications[i]
	}
}

func (c *Corpus) buildReferences(r *rand.Rand) {
	for i := range c.Publications {
		pub := &c.Publications[i]
		refCount := r.Intn(4)
		if refCount == 0 {
			continue
		}
		for j := 0; j < refCount; j++ {
			idx := r.Intn(len(c.Publications))
			target := &c.Publications[idx]
			if target.ID == pub.ID {
				continue
			}
			c.References = append(c.References, domain.Reference{
				CitingPublicationID: pub.ID,
				CitedPublicationID:  ptr(target.ID),
				RawReference:        fmt.Sprintf("%s. %s.", target.Authors[0].Name, target.Title),
				Source:              domain.SourceDemo,
			})
			target.Metrics.Citations++
			year := "2026"
			if target.PublicationDate != nil {
				year = (*target.PublicationDate)[:4]
			}
			c.citationYears[year] += 1
		}
	}
}

func (c *Corpus) buildConcepts() {
	seen := map[string]bool{}
	for _, p := range c.Publications {
		for _, con := range p.Concepts {
			if !seen[con.ID] {
				seen[con.ID] = true
				c.Concepts = append(c.Concepts, con)
			}
		}
	}
	sort.Slice(c.Concepts, func(i, j int) bool { return c.Concepts[i].Name < c.Concepts[j].Name })

	for i := 0; i < 2600; i++ {
		l := lemmas[(i*37)%len(lemmas)]
		c.lemmasFreq[l] += 1 + (i % 7)
	}
}

func (c *Corpus) indexAuthors() {
	counters := map[string]struct {
		pubs      int
		citations int
	}{}
	authorCats := map[string]map[string]bool{}
	for _, p := range c.Publications {
		for _, a := range p.Authors {
			entry := counters[a.ID]
			entry.pubs++
			entry.citations += p.Metrics.Citations
			counters[a.ID] = entry
			if authorCats[a.ID] == nil {
				authorCats[a.ID] = map[string]bool{}
			}
			if len(p.Categories) > 0 {
				authorCats[a.ID][p.Categories[0].ID] = true
			}
		}
	}
	for i := range c.Authors {
		a := &c.Authors[i]
		entry := counters[a.ID]
		a.PublicationsCount = entry.pubs
		a.CitationsCount = entry.citations
		cats := []domain.Category{}
		for id := range authorCats[a.ID] {
			if cat, ok := c.CategoryByID[id]; ok {
				cats = append(cats, *cat)
			}
		}
		sort.Slice(cats, func(x, y int) bool { return cats[x].Name < cats[y].Name })
		a.Categories = cats
		c.AuthorByID[a.ID] = a
	}
}

// CitationYears returns aggregated citation observations per year.
func (c *Corpus) CitationYears() map[string]int { return c.citationYears }

// LemmaFrequencies returns the demo lemma frequency table.
func (c *Corpus) LemmaFrequencies() map[string]int { return c.lemmasFreq }

func pickDiscipline(r *rand.Rand) DisciplineSeed {
	total := 0.0
	for _, s := range DisciplineSeeds {
		total += s.Weight
	}
	roll := r.Float64() * total
	for _, s := range DisciplineSeeds {
		roll -= s.Weight
		if roll <= 0 {
			return s
		}
	}
	return DisciplineSeeds[0]
}

// weightedYear favours recent years to mimic a growing corpus.
func weightedYear(r *rand.Rand, maxYear int) int {
	span := maxYear - 2015
	for {
		offset := int(math.Floor(math.Abs(r.NormFloat64()) * float64(span) / 2.2))
		year := maxYear - offset
		if year >= 2015 && year <= maxYear {
			return year
		}
	}
}

func ptr[T any](v T) *T { return &v }

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func slugify(s string) string {
	var b strings.Builder
	for _, ch := range strings.ToLower(s) {
		switch {
		case ch >= 'a' && ch <= 'z', ch >= '0' && ch <= '9':
			b.WriteRune(ch)
		case ch == ' ':
			b.WriteRune('-')
		case ch >= 'а' && ch <= 'я':
			if m, ok := translit[ch]; ok {
				b.WriteString(m)
			}
		}
	}
	return b.String()
}

var translit = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh",
	'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o",
	'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "h", 'ц': "ts",
	'ч': "ch", 'ш': "sh", 'щ': "sch", 'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}
