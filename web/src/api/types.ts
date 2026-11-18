export type Source = 'api' | 'demo' | 'bff'

export interface Organization {
  id: string
  name: string
}

export interface Category {
  id: string
  slug: string
  name: string
  description: string | null
  parent_id: string | null
  publications_count: number | null
  children: Category[]
  icon?: string
  accent?: string
  source?: Source
}

export interface Discipline {
  id: string
  slug: string
  name: string
  icon: string
  accent: string
  description: string | null
  publications_count: number
  subcategories_count: number
  source: Source
}

export interface Subcategory {
  id: string
  slug: string
  name: string
  publications_count: number
  icon?: string
  accent?: string
  source?: Source
}

export interface AuthorSummary {
  id: string
  name: string
  orcid: string | null
  affiliations: Organization[]
}

export interface Author extends AuthorSummary {
  categories: Category[]
  publications_count: number
  citations_count: number
  h_index: number | null
  source?: Source
}

export interface Metrics {
  citations: number
  views: number
  downloads: number
}

export interface Concept {
  id: string
  name: string
}

export interface PublicationSummary {
  id: string
  title: string
  abstract: string | null
  publication_type: string
  publication_date: string | null
  language: string
  authors: AuthorSummary[]
  metrics: Metrics
  source?: Source
}

export interface Publication extends PublicationSummary {
  external_identifiers: { scheme: string; value: string }[]
  journal: string | null
  publisher: string | null
  categories: Category[]
  keywords: string[]
  concepts: Concept[]
  files: { kind: string; content_type: string; size_bytes: number }[]
  processing_status: string
  likes_count: number
  liked: boolean
  doi?: string | null
}

export interface Reference {
  citing_publication_id: string
  cited_publication_id: string | null
  raw_reference: string
  publication: PublicationSummary | null
  source?: Source
}

export interface RelatedItem {
  publication: PublicationSummary
  score: number
  shared: string[]
}

export interface Page<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface FacetBucket {
  value: string
  count: number
}

export interface SearchFacet {
  field: string
  buckets: FacetBucket[]
  labels?: Record<string, string>
}

export interface PublicationHit {
  entity_type: 'publication'
  score: number
  entity: PublicationSummary
  snippets: string[]
  source: Source
}

export interface AuthorHit {
  entity_type: 'author'
  score: number
  entity: AuthorSummary
  snippets: string[]
  source: Source
}

export type SearchHit = PublicationHit | AuthorHit

export interface SearchResponse {
  items: SearchHit[]
  facets: SearchFacet[]
  total: number
  page: number
  page_size: number
  took_ms: number
  degraded: boolean
}

export interface Suggestion {
  type: 'publication' | 'author' | 'term'
  text: string
  id?: string
  matches_count: number | null
  source: Source
}

export interface CorpusSummary {
  total_publications: number
  total_authors: number
  total_citations: number
  total_downloads: number
  avg_citations_per_publication: number
  publications_this_period: number
  new_authors_this_period: number
  period: string
  data_as_of: string
}

export interface RankingItem {
  position: number
  object: { entity_type: string; id: string; label: string }
  value: number
}

export interface Ranking {
  metric: string
  items: RankingItem[]
  source?: Source
}

export interface GraphNode {
  id: string
  type: string
  label: string
  category?: string
  weight?: number
}

export interface GraphEdge {
  source: string
  target: string
  type: string
  weight: number | null
}

export interface Graph {
  nodes: GraphNode[]
  edges: GraphEdge[]
  truncated: boolean
  data_as_of: string
  source?: Source
}

export interface FunFact {
  discipline_slug: string
  fact_type: string
  text: string
  value: string | number | null
  period: string
  source: Source
}

export interface Widget<T = unknown> {
  widget: string
  source: Source
  data_as_of: string
  ready: boolean
  notice?: string
  payload: T
}

export interface TimelinePoint {
  year: number
  count: number
}

export interface IntakePoint {
  date: string
  count: number
}

export interface Slice {
  slug?: string
  value?: string
  name: string
  count: number
  color?: string
}

export interface TopItem {
  id: string
  label: string
  subtitle?: string
  count: number
  citations?: number
  metric?: string
  category?: string
  year?: number | null
}

export interface AuthorStats {
  publications_by_year: { year: number; count: number }[]
  citations_by_year: { year: number; count: number }[]
  top_categories: { slug: string; name: string; count: number }[]
  coauthors: { id: string; name: string; count: number }[]
}

export interface UserProfile {
  id: string
  name: string
  orcid: string | null
  affiliation: string
  bio: string
  publications_count: number
  citations_total: number
  h_index: number | null
  followers_count: number
  following: boolean
  disciplines: string[]
  coauthors: AuthorStats['coauthors']
  top_categories: AuthorStats['top_categories']
  source: Source
}

export interface Answer {
  id: number
  body: string
  author: { id: string; name: string }
  votes: number
  created_at: string
  source: Source
}

export interface Question {
  id: number
  title: string
  body: string
  author: { id: string; name: string }
  category: { slug: string; name: string }
  answers_count: number
  views_count: number
  votes: number
  created_at: string
  answers?: Answer[]
  source: Source
}

export interface LikeState {
  publication_id: string
  liked: boolean
  likes_count: number
}

export interface Health {
  status: string
  upstream: 'up' | 'down'
  corpus: number
  version: string
}

export interface Meta {
  title: string
  version: string
  sources: Record<string, string>
  widgets_ready: Record<string, boolean>
  search_engine: string
  corpus_size: number
  authors: number
  as_of: string
}

export interface LookupResult {
  found: boolean
  query?: string
  source: string
  title?: string
  authors?: { name: string; affiliation: string }[]
  year?: number
  abstract?: string
  language?: string
  keywords?: string[]
  publication_type?: string
  category_slug?: string
}
