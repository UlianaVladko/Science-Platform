import type {
  Author,
  AuthorStats,
  CorpusSummary,
  Discipline,
  FunFact,
  Graph,
  Health,
  LikeState,
  LookupResult,
  Meta,
  Page,
  Publication,
  Question,
  Ranking,
  Reference,
  RelatedItem,
  SearchHit,
  SearchResponse,
  Subcategory,
  Suggestion,
  TimelinePoint,
  TopItem,
  UserProfile,
  Widget,
} from './types'

const BASE = '/api'

export class ApiError extends Error {
  status: number
  code: string
  constructor(status: number, code: string, message: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

const VIEWER_KEY = 'atlas.viewer'

export function viewerId(): string {
  let value = localStorage.getItem(VIEWER_KEY)
  if (!value) {
    value = 'visitor-' + Math.random().toString(36).slice(2, 10)
    localStorage.setItem(VIEWER_KEY, value)
  }
  return value
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers)
  headers.set('Accept', 'application/json')
  headers.set('X-Viewer', viewerId())
  if (init?.body && !(init.body instanceof Uint8Array) && !(init.body instanceof Blob)) {
    headers.set('Content-Type', 'application/json')
  }
  const response = await fetch(BASE + path, { ...init, headers })
  if (!response.ok) {
    let code = `HTTP_${response.status}`
    let message = response.statusText || 'Ошибка запроса'
    try {
      const payload = await response.json()
      if (payload?.error) {
        code = payload.error.code ?? code
        message = payload.error.message ?? message
      }
    } catch {
      /* non-JSON error body */
    }
    throw new ApiError(response.status, code, message)
  }
  if (response.status === 204) {
    return undefined as T
  }
  return (await response.json()) as T
}

function qs(params: Record<string, string | number | boolean | undefined | null>): string {
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue
    search.set(key, String(value))
  }
  const encoded = search.toString()
  return encoded ? `?${encoded}` : ''
}

export interface SearchPayload {
  query?: string
  entity_type?: 'publication' | 'author'
  mode?: 'lexical' | 'semantic' | 'hybrid' | 'similar'
  similar_to_publication_id?: string
  filters?: {
    category?: string
    year_from?: number
    year_to?: number
    publication_type?: string[]
    language?: string
    author?: string
    organization?: string
    journal?: string
    terms?: string[]
  }
  sort?: string
  page?: number
  page_size?: number
  include_facets?: boolean
  scope?: 'all' | 'api' | 'demo'
}

export const api = {
  health: () => request<Health>('/health'),
  meta: () => request<Meta>('/meta'),

  disciplines: () => request<{ items: Discipline[]; source: string }>('/disciplines'),
  subcategories: (slug: string) =>
    request<{ discipline: { id: string; slug: string; name: string }; items: Subcategory[] }>(
      `/disciplines/${encodeURIComponent(slug)}/subcategories`,
    ),
  funFact: (slug: string) => request<FunFact>(`/disciplines/${encodeURIComponent(slug)}/fun-fact`),

  search: (payload: SearchPayload) =>
    request<SearchResponse>('/search', { method: 'POST', body: JSON.stringify(payload) }),
  suggest: (query: string, limit = 8) => request<{ items: Suggestion[] }>(`/search/suggest${qs({ q: query, limit })}`),

  publication: (id: string) => request<Publication>(`/publications/${encodeURIComponent(id)}`),
  references: (id: string, direction: 'outgoing' | 'incoming', page = 1, pageSize = 20) =>
    request<Page<Reference>>(
      `/publications/${encodeURIComponent(id)}/references${qs({ direction, page, page_size: pageSize })}`,
    ),
  related: (id: string, limit = 5) =>
    request<{ items: RelatedItem[] }>(`/publications/${encodeURIComponent(id)}/related${qs({ limit })}`),
  view: (id: string) =>
    request<{ publication_id: string; views: number }>(`/publications/${encodeURIComponent(id)}/view`, {
      method: 'POST',
    }),
  like: (id: string) => request<LikeState>(`/publications/${encodeURIComponent(id)}/like`, { method: 'POST' }),
  unlike: (id: string) => request<LikeState>(`/publications/${encodeURIComponent(id)}/like`, { method: 'DELETE' }),
  lookup: (params: { doi?: string; title?: string }) =>
    request<LookupResult>(`/publications/lookup${qs(params)}`),
  createPublication: (payload: unknown) =>
    request<Publication>('/publications', { method: 'POST', body: JSON.stringify(payload) }),

  author: (id: string) => request<Author>(`/authors/${encodeURIComponent(id)}`),
  authorStats: (id: string) => request<AuthorStats>(`/authors/${encodeURIComponent(id)}/stats`),
  authorPublications: (id: string, page = 1, pageSize = 20) =>
    request<Page<Publication>>(
      `/authors/${encodeURIComponent(id)}/publications${qs({ page, page_size: pageSize })}`,
    ),
  coauthorGraph: (id: string, depth = 2) =>
    request<Graph>(`/authors/${encodeURIComponent(id)}/coauthor-graph${qs({ depth })}`),

  profile: (id: string) => request<UserProfile>(`/users/${encodeURIComponent(id)}/profile`),
  follow: (id: string) =>
    request<{ following: boolean; followers_count: number }>(`/users/${encodeURIComponent(id)}/follow`, {
      method: 'POST',
    }),
  unfollow: (id: string) =>
    request<{ following: boolean; followers_count: number }>(`/users/${encodeURIComponent(id)}/follow`, {
      method: 'DELETE',
    }),

  summary: () => request<CorpusSummary>('/analytics/summary'),
  top: (params: { entity: string; metric: string; category?: string; period?: string; limit?: number }) =>
    request<Ranking>(`/analytics/top${qs(params)}`),
  widget: <T,>(name: string, params: { category?: string; period?: string } = {}) =>
    request<Widget<T>>(`/dashboard/${encodeURIComponent(name)}${qs(params)}`),
  graph: (params: {
    type: string
    root_id?: string
    max_depth?: number
    max_nodes?: number
    category?: string
    period?: string
  }) => request<Graph>(`/graphs${qs(params)}`),

  questions: (params: { page?: number; page_size?: number; category?: string; sort?: string }) =>
    request<Page<Question>>(`/questions${qs(params)}`),
  question: (id: number) => request<Question>(`/questions/${id}`),
  createQuestion: (payload: { title: string; body: string; category_slug: string; author_name?: string }) =>
    request<Question>('/questions', { method: 'POST', body: JSON.stringify(payload) }),
  createAnswer: (id: number, payload: { body: string; author_name?: string }) =>
    request<{ id: number }>(`/questions/${id}/answers`, { method: 'POST', body: JSON.stringify(payload) }),
  vote: (answerId: number, vote: 1 | -1) =>
    request<{ id: number; votes: number }>(`/answers/${answerId}/vote`, {
      method: 'POST',
      body: JSON.stringify({ vote }),
    }),
}

export type { SearchHit, TimelinePoint, TopItem }
