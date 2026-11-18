import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { CorpusSummary, Discipline, Subcategory, TopItem } from '../api/types'
import { useDocumentTitle, useFetch } from '../hooks/useFetch'
import { SearchBox } from '../components/SearchBox'
import { DisciplineCard } from '../components/DisciplineCard'
import { CardSkeleton, EmptyState, ErrorState, SourceBadge, Skeleton } from '../components/ui'
import { Icon } from '../components/Icons'

function KpiCard({
  value,
  label,
  delta,
  accent,
}: {
  value: string
  label: string
  delta?: string
  accent: string
}) {
  return (
    <div className="kpi" style={{ ['--kpi-accent' as string]: accent }}>
      <div className="kpi-value">{value}</div>
      <div className="kpi-label">{label}</div>
      {delta && <div className="kpi-delta">{delta}</div>}
    </div>
  )
}

function TopPublications() {
  const [metric, setMetric] = useState<'citations' | 'views' | 'downloads'>('citations')
  const [discipline, setDiscipline] = useState<string>('all')

  const { data: disciplines } = useFetch(() => api.disciplines(), [])
  const { data, loading, error, reload } = useFetch(
    () =>
      api.widget<{ metric: string; items: TopItem[] }>('top-publications', {
        category: discipline === 'all' ? undefined : discipline,
      }),
    [discipline],
  )

  const items = useMemo(() => {
    if (!data?.payload.items) return []
    return [...data.payload.items].sort((a, b) => {
      if (metric === 'citations') return (b.citations ?? 0) - (a.citations ?? 0)
      return b.count - a.count
    })
  }, [data, metric])

  const metricLabel = { citations: 'цитирований', views: 'просмотров', downloads: 'скачиваний' }[metric]

  return (
    <section className="section">
      <div className="section-head">
        <div>
          <h2>Топ публикаций</h2>
          <p>Ранжирование по метрике внутри выбранного направления</p>
        </div>
        <SourceBadge source="demo" />
      </div>

      <div className="row wrap" style={{ marginBottom: 12 }}>
        <div className="chip" style={{ padding: 3, gap: 3 }}>
          <button
            type="button"
            className="chip"
            aria-pressed={discipline === 'all'}
            style={{ border: 'none', background: 'transparent' }}
            onClick={() => setDiscipline('all')}
          >
            Все
          </button>
          {(disciplines?.items ?? []).slice(0, 6).map((d) => (
            <button
              key={d.slug}
              type="button"
              className="chip"
              aria-pressed={discipline === d.slug}
              style={{ border: 'none', background: 'transparent' }}
              onClick={() => setDiscipline(d.slug)}
            >
              {d.name}
            </button>
          ))}
        </div>
        <div className="spacer" />
        <div className="chip" style={{ padding: 3, gap: 3 }}>
          {(['citations', 'views', 'downloads'] as const).map((m) => (
            <button
              key={m}
              type="button"
              className="chip"
              aria-pressed={metric === m}
              style={{ border: 'none', background: 'transparent' }}
              onClick={() => setMetric(m)}
            >
              {{ citations: 'Цитирования', views: 'Просмотры', downloads: 'Скачивания' }[m]}
            </button>
          ))}
        </div>
      </div>

      {loading && (
        <div className="stack">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} height={56} radius={14} />
          ))}
        </div>
      )}
      {error && <ErrorState message={error} onRetry={reload} />}
      {!loading && !error && items.length === 0 && (
        <EmptyState title="Нет данных по выбранному направлению" description="Попробуйте другую дисциплину или период" />
      )}
      {!loading && !error && items.length > 0 && (
        <ol className="list-plain" style={{ counterReset: 'top' }}>
          {items.slice(0, 7).map((item, index) => (
            <li key={item.id}>
              <Link to={`/publications/${encodeURIComponent(item.id)}`} className="pub-row">
                <div className="row" style={{ gap: 14 }}>
                  <span
                    className="mono"
                    style={{
                      fontSize: 13,
                      color: index < 3 ? 'var(--accent-soft)' : 'var(--text-500)',
                      minWidth: 22,
                    }}
                  >
                    {String(index + 1).padStart(2, '0')}
                  </span>
                  <span className="pub-title" style={{ flex: 1 }}>
                    {item.label}
                  </span>
                  <span className="chip chip-accent" style={{ flexShrink: 0 }}>
                    {item.count} {metricLabel}
                  </span>
                </div>
                <div className="pub-meta" style={{ paddingLeft: 36 }}>
                  {item.subtitle && <span>{item.subtitle}</span>}
                  {item.category && <span>· {item.category}</span>}
                  {item.year && <span>· {item.year}</span>}
                </div>
              </Link>
            </li>
          ))}
        </ol>
      )}
    </section>
  )
}

function DisciplineSection() {
  const navigate = useNavigate()
  const { data, loading, error, reload } = useFetch(() => api.disciplines(), [])
  const [active, setActive] = useState<Discipline | null>(null)
  const [subs, setSubs] = useState<Subcategory[] | null>(null)
  const [subsLoading, setSubsLoading] = useState(false)

  const openDiscipline = (slug: string) => {
    const found = data?.items.find((d) => d.slug === slug) ?? null
    setActive(found)
    if (!found) return
    setSubsLoading(true)
    api
      .subcategories(slug)
      .then((result) => setSubs(result.items))
      .catch(() => setSubs([]))
      .finally(() => setSubsLoading(false))
  }

  return (
    <section className="section" id="disciplines">
      <div className="section-head">
        <div>
          <h2>Классификация наук</h2>
          <p>Наведите на карточку — покажем факт по направлению</p>
        </div>
        {data && <SourceBadge source={data.source as 'api' | 'demo'} />}
      </div>

      {loading && (
        <div className="discipline-grid">
          {Array.from({ length: 8 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      )}
      {error && <ErrorState message={error} onRetry={reload} />}

      {data && (
        <div className="discipline-grid">
          {data.items.map((discipline) => (
            <DisciplineCard key={discipline.slug} discipline={discipline} onSelect={openDiscipline} />
          ))}
        </div>
      )}

      {active && (
        <div className="card card-pad rise" style={{ marginTop: 16 }}>
          <div className="row wrap" style={{ marginBottom: 12 }}>
            <Icon name={active.icon} size={20} style={{ color: active.accent }} />
            <h3 style={{ margin: 0 }}>{active.name}</h3>
            <span className="chip">{active.publications_count.toLocaleString('ru-RU')} публикаций</span>
            <div className="spacer" />
            <button type="button" className="btn btn-sm" onClick={() => navigate(`/search?category=${active.slug}`)}>
              Показать публикации <Icon name="arrowRight" size={15} />
            </button>
            <button type="button" className="btn btn-ghost btn-sm" onClick={() => setActive(null)}>
              Свернуть
            </button>
          </div>

          {subsLoading && <Skeleton height={48} radius={12} />}
          {subs && subs.length > 0 && (
            <div className="row wrap" style={{ gap: 8 }}>
              {subs.map((sub) => (
                <Link key={sub.slug} to={`/search?category=${sub.slug}`} className="chip">
                  {sub.name}
                  <span className="dim mono">{sub.publications_count}</span>
                </Link>
              ))}
            </div>
          )}
          {subs && subs.length === 0 && <p className="muted">У направления нет подкатегорий.</p>}
        </div>
      )}
    </section>
  )
}

export default function HomePage() {
  useDocumentTitle('Главная')
  const { data: summary } = useFetch<CorpusSummary>(() => api.summary(), [])
  const { data: meta } = useFetch(() => api.meta(), [])

  const kpis = summary
    ? [
        { value: summary.total_publications.toLocaleString('ru-RU'), label: 'публикаций в корпусе', accent: 'var(--accent)' },
        { value: summary.total_authors.toLocaleString('ru-RU'), label: 'авторов', accent: 'var(--teal)' },
        { value: summary.total_citations.toLocaleString('ru-RU'), label: 'цитирований', accent: 'var(--sky)' },
        { value: summary.total_downloads.toLocaleString('ru-RU'), label: 'скачиваний', accent: 'var(--amber)' },
        {
          value: summary.publications_this_period.toLocaleString('ru-RU'),
          label: `публикаций за ${summary.period} год`,
          delta: `+${summary.new_authors_this_period} новых авторов`,
          accent: 'var(--rose)',
        },
      ]
    : []

  return (
    <div className="page">
      <section className="hero">
        <div className="hero-inner">
          <h1>
            Научная литература как <em>связный граф</em>
          </h1>
          <p className="hero-lead">
            Поиск по метаданным и смыслу, карточки публикаций с библиографией, графы цитирования и
            соавторства, аналитика корпуса — в одной витрине поверх согласованного API.
          </p>
          <SearchBox />
          <div className="hero-stats">
            {kpis.slice(0, 4).map((kpi) => (
              <div key={kpi.label}>
                <div className="kpi-value" style={{ fontSize: 24 }}>
                  {kpi.value}
                </div>
                <div className="kpi-label">{kpi.label}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="section" style={{ marginTop: 0 }}>
        <div className="kpi-grid">
          {summary ? (
            kpis.map((kpi) => (
              <KpiCard key={kpi.label} value={kpi.value} label={kpi.label} delta={kpi.delta} accent={kpi.accent} />
            ))
          ) : (
            Array.from({ length: 5 }).map((_, i) => <Skeleton key={i} height={104} radius={14} />)
          )}
        </div>
        <div className="row wrap" style={{ marginTop: 12, gap: 10 }}>
          <SourceBadge source="demo" />
          <span className="dim" style={{ fontSize: 12.5 }}>
            KPI и витрины считаются по демо-корпусу BFF: согласованный API отдаёт сводку только по своим
            фикстурам ({meta ? `${meta.corpus_size} публикаций демо, ` : ''}20 операций).
          </span>
        </div>
      </section>

      <DisciplineSection />
      <TopPublications />
    </div>
  )
}
