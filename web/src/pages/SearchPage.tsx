import { useEffect, useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { api, type SearchPayload } from '../api/client'
import type { Discipline, SearchFacet, SearchHit, SearchResponse } from '../api/types'
import { useDebounced, useDocumentTitle, useFetch } from '../hooks/useFetch'
import { Icon } from '../components/Icons'
import { SearchBox, type SearchMode } from '../components/SearchBox'
import { EmptyState, ErrorState, Skeleton, SourceBadge } from '../components/ui'

const PAGE_SIZE = 10

const SCOPES = [
  { value: 'all', label: 'Все источники' },
  { value: 'api', label: 'API' },
  { value: 'demo', label: 'Демо' },
] as const

type Scope = (typeof SCOPES)[number]['value']

const MODE_LABEL: Record<SearchMode, string> = {
  hybrid: 'гибридный',
  lexical: 'точный',
  semantic: 'по смыслу',
}

const SORT_OPTIONS = [
  { value: 'relevance', label: 'По релевантности' },
  { value: 'citations', label: 'По цитированиям' },
  { value: 'year', label: 'По году' },
] as const

type Sort = (typeof SORT_OPTIONS)[number]['value']

function facetLabel(facet: SearchFacet, value: string) {
  return facet.labels?.[value] ?? value
}

export default function SearchPage() {
  const [params, setParams] = useSearchParams()
  const query = params.get('q') ?? ''
  const mode = (params.get('mode') as SearchMode | null) ?? 'hybrid'
  const category = params.get('category') ?? ''
  const scopeParam = params.get('scope') as Scope | null
  const scope: Scope = SCOPES.some((item) => item.value === scopeParam) ? scopeParam! : 'all'
  const sortParam = params.get('sort') as Sort | null
  const sort: Sort = SORT_OPTIONS.some((item) => item.value === sortParam) ? sortParam! : 'relevance'
  const page = Math.max(1, Number(params.get('page') ?? 1) || 1)

  const [text, setText] = useState(query)
  const debouncedText = useDebounced(text, 320)
  const [language, setLanguage] = useState(params.get('language') ?? '')
  const [publicationType, setPublicationType] = useState(params.get('publication_type') ?? '')
  const [yearFrom, setYearFrom] = useState(params.get('year_from') ?? '')
  const [yearTo, setYearTo] = useState(params.get('year_to') ?? '')

  useDocumentTitle(query ? `Поиск: ${query}` : 'Поиск')

  useEffect(() => {
    setText(query)
  }, [query])

  const patch = (next: Record<string, string | number | undefined>) => {
    const merged = new URLSearchParams(params)
    Object.entries(next).forEach(([key, value]) => {
      if (value === undefined || value === '') merged.delete(key)
      else merged.set(key, String(value))
    })
    setParams(merged, { replace: true })
  }

  const payload: SearchPayload = useMemo(
    () => ({
      query: debouncedText.trim(),
      mode,
      page,
      page_size: PAGE_SIZE,
      scope,
      sort,
      include_facets: true,
      filters: {
        category: category || undefined,
        language: language || undefined,
        publication_type: publicationType ? [publicationType] : undefined,
        year_from: yearFrom ? Number(yearFrom) : undefined,
        year_to: yearTo ? Number(yearTo) : undefined,
      },
    }),
    [debouncedText, mode, page, scope, sort, category, language, publicationType, yearFrom, yearTo],
  )

  const skip = !payload.query && !category && !language && !publicationType && !yearFrom && !yearTo

  const { data, loading, error, reload } = useFetch<SearchResponse | null>(
    () => (skip ? Promise.resolve(null) : api.search(payload)),
    [JSON.stringify(payload)],
  )

  const { data: disciplines } = useFetch<{ items: Discipline[] }>(() => api.disciplines(), [])
  const facets = data?.facets ?? []

  const hits = useMemo(() => {
    const list = data?.items ?? []
    if (sort === 'citations') {
      return [...list].sort((a, b) => scoreCitations(b) - scoreCitations(a))
    }
    if (sort === 'year') {
      return [...list].sort((a, b) => scoreYear(b) - scoreYear(a))
    }
    return list
  }, [data, sort])

  const total = data?.total ?? 0
  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE))
  const activeCategory = disciplines?.items.find((item) => item.slug === category)

  const activeFilters: { key: string; label: string; clear: () => void }[] = []
  if (language) {
    const facet = facets.find((item) => item.field === 'language')
    activeFilters.push({
      key: 'language',
      label: facet ? facetLabel(facet, language) : language,
      clear: () => {
        setLanguage('')
        patch({ language: undefined, page: undefined })
      },
    })
  }
  if (publicationType) {
    const facet = facets.find((item) => item.field === 'publication_type')
    activeFilters.push({
      key: 'publication_type',
      label: facet ? facetLabel(facet, publicationType) : publicationType,
      clear: () => {
        setPublicationType('')
        patch({ publication_type: undefined, page: undefined })
      },
    })
  }
  if (yearFrom || yearTo) {
    activeFilters.push({
      key: 'years',
      label: `${yearFrom || '…'}–${yearTo || '…'}`,
      clear: () => {
        setYearFrom('')
        setYearTo('')
        patch({ year_from: undefined, year_to: undefined, page: undefined })
      },
    })
  }

  const toggleFacet = (field: string, value: string) => {
    if (field === 'language') {
      const next = language === value ? '' : value
      setLanguage(next)
      patch({ language: next || undefined, page: undefined })
      return
    }
    if (field === 'publication_type') {
      const next = publicationType === value ? '' : value
      setPublicationType(next)
      patch({ publication_type: next || undefined, page: undefined })
      return
    }
    if (field === 'category') {
      const next = category === value ? '' : value
      patch({ category: next || undefined, page: undefined })
      return
    }
    if (field === 'year') {
      setYearFrom(value)
      setYearTo(value)
      patch({ year_from: value, year_to: value, page: undefined })
    }
  }

  const isFacetActive = (field: string, value: string) => {
    if (field === 'language') return language === value
    if (field === 'publication_type') return publicationType === value
    if (field === 'category') return category === value
    if (field === 'year') return yearFrom === value && yearTo === value
    return false
  }

  const statusText = skip
    ? 'Введите запрос или выберите направление'
    : loading
      ? 'Ищем…'
      : `Найдено ${total.toLocaleString('ru-RU')} · режим ${MODE_LABEL[mode]}`

  return (
    <div className="page">
      <div style={{ padding: 'clamp(20px, 3vw, 34px) 0 18px', maxWidth: 760 }}>
        <SearchBox initialQuery={query} initialMode={mode} size="md" />
      </div>

      <div className="row wrap" style={{ gap: 10, marginBottom: 14 }}>
        <span className="dim" style={{ fontSize: 13 }}>
          {statusText}
        </span>
        {data && <SourceBadge source={data.degraded ? 'demo' : 'api'} />}
        {data?.degraded && (
          <span className="chip" title="Согласованный API недоступен, показан синтетический корпус BFF">
            <Icon name="alert" size={13} /> API недоступен — только демо-корпус
          </span>
        )}
        {activeCategory && (
          <span className="chip chip-accent">
            {activeCategory.name}
            <button
              type="button"
              aria-label="Сбросить направление"
              style={{ border: 'none', background: 'transparent', cursor: 'pointer', color: 'inherit', padding: 0 }}
              onClick={() => patch({ category: undefined, page: undefined })}
            >
              ×
            </button>
          </span>
        )}
        <div className="spacer" />
        <div className="chip" style={{ padding: 3, gap: 3 }}>
          {SCOPES.map((item) => (
            <button
              key={item.value}
              type="button"
              className="chip"
              aria-pressed={scope === item.value}
              style={{ border: 'none', background: 'transparent' }}
              onClick={() => patch({ scope: item.value, page: undefined })}
            >
              {item.label}
            </button>
          ))}
        </div>
      </div>

      <div className="filters-bar" style={{ marginBottom: 18 }}>
        <div className="row" style={{ gap: 8, flex: 1, minWidth: 220 }}>
          <Icon name="file" size={16} style={{ color: 'var(--text-500)' }} />
          <input
            className="input"
            value={text}
            placeholder="Уточнить запрос…"
            onChange={(event) => setText(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === 'Enter') patch({ q: text.trim(), page: undefined })
            }}
            aria-label="Поисковый запрос"
          />
        </div>

        <select
          className="input"
          style={{ maxWidth: 200 }}
          value={category}
          onChange={(event) => patch({ category: event.target.value || undefined, page: undefined })}
          aria-label="Направление"
        >
          <option value="">Все направления</option>
          {(disciplines?.items ?? []).map((item) => (
            <option key={item.slug} value={item.slug}>
              {item.name}
            </option>
          ))}
        </select>

        <select
          className="input"
          style={{ maxWidth: 190 }}
          value={sort}
          onChange={(event) => patch({ sort: event.target.value, page: undefined })}
          aria-label="Сортировка"
        >
          {SORT_OPTIONS.map((item) => (
            <option key={item.value} value={item.value}>
              {item.label}
            </option>
          ))}
        </select>

        {activeFilters.map((filter) => (
          <span key={filter.key} className="chip">
            {filter.label}
            <button
              type="button"
              aria-label={`Сбросить фильтр ${filter.key}`}
              style={{ border: 'none', background: 'transparent', cursor: 'pointer', color: 'inherit', padding: 0 }}
              onClick={filter.clear}
            >
              ×
            </button>
          </span>
        ))}
      </div>

      <div className="detail-layout">
        <div>
          {loading && (
            <div className="stack">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} height={96} radius={14} />
              ))}
            </div>
          )}

          {error && <ErrorState message={error} onRetry={reload} />}

          {!loading && !error && skip && (
            <EmptyState
              title="Что ищем?"
              description="Введите запрос, выберите направление или откройте подсказку в строке поиска."
            />
          )}

          {!loading && !error && !skip && hits.length === 0 && (
            <EmptyState
              title="Ничего не найдено"
              description="Попробуйте другой режим поиска, снимите фильтры или расширьте период."
            />
          )}

          {!loading && !error && !skip && data && (
            <p className="dim" style={{ fontSize: 12.5, marginBottom: 10 }}>
              {data.took_ms} мс · страница {data.page} из {pages}
            </p>
          )}

          {!loading && !error && hits.length > 0 && (
            <div className="stack">
              {hits.map((hit) => (
                <HitCard key={`${hit.entity_type}-${hit.entity.id}-${hit.source}`} hit={hit} query={data?.items.length ? query : ''} />
              ))}
            </div>
          )}

          {!skip && pages > 1 && (
            <div className="pager">
              <button
                type="button"
                className="btn btn-sm"
                disabled={page <= 1}
                onClick={() => patch({ page: page - 1 })}
              >
                <Icon name="arrowLeft" size={14} /> Назад
              </button>
              <span className="dim" style={{ fontSize: 13 }}>
                Страница {page} из {pages}
              </span>
              <button
                type="button"
                className="btn btn-sm"
                disabled={page >= pages}
                onClick={() => patch({ page: page + 1 })}
              >
                Вперёд <Icon name="arrowRight" size={14} />
              </button>
            </div>
          )}
        </div>

        <aside className="sidebar">
          {facets.length === 0 && (
            <div className="panel">
              <div className="panel-title">Фильтры</div>
              <p className="muted" style={{ fontSize: 13 }}>
                Фасеты появятся после первого запроса.
              </p>
            </div>
          )}

          {facets.map((facet) => (
            <div className="panel" key={facet.field}>
              <div className="panel-title">{facet.field}</div>
              <div className="facet-group">
                {facet.buckets.slice(0, 8).map((bucket) => {
                  const active = isFacetActive(facet.field, bucket.value)
                  return (
                    <button
                      key={bucket.value}
                      type="button"
                      className="chip"
                      aria-pressed={active}
                      style={{
                        justifyContent: 'space-between',
                        border: active ? '1px solid var(--accent-line)' : '1px solid transparent',
                        background: active ? 'var(--accent-soft)' : 'transparent',
                        color: active ? '#fff' : undefined,
                        paddingInline: 8,
                      }}
                      onClick={() => toggleFacet(facet.field, bucket.value)}
                    >
                      <span>{facetLabel(facet, bucket.value)}</span>
                      <span className="dim mono">{bucket.count}</span>
                    </button>
                  )
                })}
              </div>
            </div>
          ))}

          <div className="panel">
            <div className="panel-title">Год публикации</div>
            <div className="row" style={{ gap: 8 }}>
              <input
                className="input"
                inputMode="numeric"
                placeholder="с"
                value={yearFrom}
                onChange={(event) => {
                  setYearFrom(event.target.value)
                  patch({ year_from: event.target.value || undefined, page: undefined })
                }}
                aria-label="Год с"
              />
              <input
                className="input"
                inputMode="numeric"
                placeholder="по"
                value={yearTo}
                onChange={(event) => {
                  setYearTo(event.target.value)
                  patch({ year_to: event.target.value || undefined, page: undefined })
                }}
                aria-label="Год по"
              />
            </div>
          </div>
        </aside>
      </div>
    </div>
  )
}

function scoreCitations(hit: SearchHit) {
  if (hit.entity_type !== 'publication') return 0
  return hit.entity.metrics?.citations ?? 0
}

function scoreYear(hit: SearchHit) {
  if (hit.entity_type !== 'publication') return 0
  const date = hit.entity.publication_date
  return date ? Number(date.slice(0, 4)) || 0 : 0
}

function HitCard({ hit, query }: { hit: SearchHit; query: string }) {
  if (hit.entity_type === 'author') {
    const author = hit.entity
    return (
      <Link to={`/authors/${encodeURIComponent(author.id)}`} className="pub-row">
        <div className="row wrap" style={{ gap: 8, alignItems: 'center' }}>
          <span className="chip chip-accent">автор</span>
          <span className="pub-title" style={{ flex: 1 }}>
            {author.name}
          </span>
          <SourceBadge source={hit.source} />
        </div>
        <div className="pub-meta">
          {author.affiliations.map((item) => item.name).join(' · ')}
          {author.orcid && <span> · ORCID {author.orcid}</span>}
          <span className="dim"> · релевантность {hit.score.toFixed(2)}</span>
        </div>
      </Link>
    )
  }

  const pub = hit.entity
  const snippet = hit.snippets[0]
  return (
    <Link to={`/publications/${encodeURIComponent(pub.id)}`} className="pub-row">
      <div className="row wrap" style={{ gap: 8, alignItems: 'center' }}>
        <span className="pub-title" style={{ flex: 1 }}>
          {pub.title || 'Без названия'}
        </span>
        <SourceBadge source={hit.source} />
      </div>
      {pub.authors.length > 0 && (
        <div className="pub-authors">{pub.authors.map((author) => author.name).join(', ')}</div>
      )}
      {snippet && <p className="pub-snippet">{highlight(snippet, query)}</p>}
      <div className="pub-meta">
        {pub.publication_date && <span>{pub.publication_date.slice(0, 4)}</span>}
        {pub.publication_type && <span>· {pub.publication_type}</span>}
        {pub.language && <span>· {pub.language}</span>}
        <span>· цитирований: {pub.metrics?.citations ?? 0}</span>
        <span className="dim"> · релевантность {hit.score.toFixed(2)}</span>
      </div>
    </Link>
  )
}

function highlight(text: string, query: string) {
  const trimmed = query.trim()
  if (trimmed.length < 2) return text
  const tokens = trimmed
    .split(/\s+/)
    .filter((token) => token.length >= 2)
    .slice(0, 5)
  if (tokens.length === 0) return text
  const escaped = tokens.map((token) => token.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
  const lower = new Set(tokens.map((token) => token.toLowerCase()))
  const parts = text.split(new RegExp(`(${escaped.join('|')})`, 'gi'))
  return parts.map((part, index) =>
    lower.has(part.toLowerCase()) ? <mark key={index}>{part}</mark> : <span key={index}>{part}</span>,
  )
}
