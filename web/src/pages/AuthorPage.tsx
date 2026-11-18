import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import type { EChartsOption } from 'echarts'
import { api } from '../api/client'
import type { Author, AuthorStats, Graph, Page, Publication, UserProfile } from '../api/types'
import { useDocumentTitle, useFetch } from '../hooks/useFetch'
import { Icon } from '../components/Icons'
import { BASE_TEXT, AXIS_STYLE, TOOLTIP_STYLE, Chart, PALETTE } from '../components/charts/Chart'
import { ForceGraph } from '../components/charts/ForceGraph'
import { EmptyState, ErrorState, Skeleton, SourceBadge } from '../components/ui'

export default function AuthorPage() {
  const { id = '' } = useParams()
  const { data: author, loading, error, reload } = useFetch<Author>(() => api.author(id), [id])
  const { data: profile } = useFetch<UserProfile>(() => api.profile(id), [id])
  const { data: stats } = useFetch<AuthorStats>(() => api.authorStats(id), [id])
  const [page, setPage] = useState(1)
  const { data: publications } = useFetch<Page<Publication>>(
    () => api.authorPublications(id, page, 8),
    [id, page],
  )
  const { data: graph } = useFetch<Graph>(() => api.coauthorGraph(id, 2), [id])

  useDocumentTitle(author?.name ?? 'Автор')

  useEffect(() => {
    setPage(1)
  }, [id])

  const chartOption = useMemo<EChartsOption | null>(() => {
    if (!stats) return null
    const years = stats.publications_by_year.map((item) => item.year)
    return {
      grid: { left: 8, right: 8, top: 24, bottom: 8, containLabel: true },
      tooltip: { trigger: 'axis', ...TOOLTIP_STYLE },
      legend: { ...BASE_TEXT, top: 0, itemWidth: 10, itemHeight: 10, data: ['публикации', 'цитирования'] },
      xAxis: { type: 'category', data: years, ...AXIS_STYLE },
      yAxis: { type: 'value', ...AXIS_STYLE },
      series: [
        {
          name: 'публикации',
          type: 'bar',
          data: stats.publications_by_year.map((item) => item.count),
          itemStyle: { color: PALETTE[0], borderRadius: [4, 4, 0, 0] },
          barMaxWidth: 18,
        },
        {
          name: 'цитирования',
          type: 'line',
          smooth: true,
          symbolSize: 5,
          data: stats.citations_by_year.map((item) => item.count),
          lineStyle: { color: PALETTE[1], width: 2 },
          itemStyle: { color: PALETTE[1] },
        },
      ],
    }
  }, [stats])

  if (loading) {
    return (
      <div className="page" style={{ paddingTop: 40 }}>
        <div className="profile-hero">
          <Skeleton height={84} width={84} radius={24} />
          <div style={{ flex: 1 }}>
            <Skeleton height={30} width="45%" />
            <div style={{ height: 10 }} />
            <Skeleton height={16} width="30%" />
          </div>
        </div>
        <Skeleton height={240} radius={16} />
      </div>
    )
  }

  if (error || !author) {
    return (
      <div className="page" style={{ paddingTop: 40 }}>
        <ErrorState message={error ?? 'Автор не найден'} onRetry={reload} />
        <p style={{ marginTop: 16 }}>
          <Link to="/search" className="btn btn-sm">
            <Icon name="arrowLeft" size={14} /> К поиску
          </Link>
        </p>
      </div>
    )
  }

  const initials = author.name
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('')

  const totalPages = publications ? Math.max(1, Math.ceil(publications.total / publications.page_size)) : 1

  return (
    <div className="page">
      <nav className="breadcrumbs" aria-label="Хлебные крошки">
        <Link to="/">Главная</Link>
        <span>/</span>
        <Link to="/search">Авторы</Link>
        <span>/</span>
        <span>{author.name}</span>
      </nav>

      <header className="profile-hero">
        <div className="avatar" aria-hidden="true">
          {initials || 'А'}
        </div>
        <div style={{ flex: 1, minWidth: 240 }}>
          <div className="row wrap" style={{ gap: 8, alignItems: 'center' }}>
            <h1 style={{ margin: 0 }}>{author.name}</h1>
            {(profile?.source ?? author.source) && <SourceBadge source={(profile?.source ?? author.source)!} />}
          </div>
          <p className="muted" style={{ margin: '6px 0 0' }}>
            {author.affiliations.map((item) => item.name).join(' · ') || 'Аффилиация не указана'}
            {author.orcid && <span className="mono dim"> · ORCID {author.orcid}</span>}
          </p>
          <div className="stat-strip">
            <Stat label="публикаций" value={author.publications_count} />
            <Stat label="цитирований" value={author.citations_count} />
            <Stat label="h-индекс" value={author.h_index ?? '—'} />
            {profile && <Stat label="подписчиков" value={profile.followers_count} />}
          </div>
        </div>
        {profile && <FollowButton profile={profile} />}
      </header>

      {profile?.bio && (
        <section className="panel" style={{ marginBottom: 'var(--space-5)' }}>
          <div className="panel-title">О себе</div>
          <p style={{ margin: 0, color: 'var(--text-200)', lineHeight: 1.6 }}>{profile.bio}</p>
        </section>
      )}

      <div className="detail-layout">
        <div className="stack" style={{ gap: 'var(--space-5)' }}>
          <section>
            <div className="section-head" style={{ marginTop: 0 }}>
              <div>
                <h2>Публикации автора</h2>
                <p>{publications ? `Всего ${publications.total}` : 'Загружаем список'}</p>
              </div>
            </div>

            {!publications && (
              <div className="stack">
                {Array.from({ length: 3 }).map((_, i) => (
                  <Skeleton key={i} height={84} radius={14} />
                ))}
              </div>
            )}

            {publications && publications.items.length === 0 && (
              <EmptyState title="Публикаций нет" description="Автор пока не связан ни с одной публикацией корпуса." />
            )}

            {publications && publications.items.length > 0 && (
              <div className="stack">
                {publications.items.map((publication) => (
                  <Link
                    key={publication.id}
                    to={`/publications/${encodeURIComponent(publication.id)}`}
                    className="pub-row"
                  >
                    <span className="pub-title">{publication.title}</span>
                    <div className="pub-meta">
                      {publication.publication_date && <span>{publication.publication_date.slice(0, 4)}</span>}
                      <span>· {publication.publication_type}</span>
                      <span>· цитирований: {publication.metrics.citations}</span>
                    </div>
                  </Link>
                ))}
              </div>
            )}

            {publications && totalPages > 1 && (
              <div className="pager">
                <button
                  type="button"
                  className="btn btn-sm"
                  disabled={page <= 1}
                  onClick={() => setPage((value) => Math.max(1, value - 1))}
                >
                  <Icon name="arrowLeft" size={14} /> Назад
                </button>
                <span className="dim" style={{ fontSize: 13 }}>
                  Страница {page} из {totalPages}
                </span>
                <button
                  type="button"
                  className="btn btn-sm"
                  disabled={page >= totalPages}
                  onClick={() => setPage((value) => value + 1)}
                >
                  Вперёд <Icon name="arrowRight" size={14} />
                </button>
              </div>
            )}
          </section>

          {stats && (
            <section className="panel">
              <div className="panel-title">Динамика публикаций и цитирований</div>
              {chartOption ? (
                <Chart option={chartOption} height={280} ariaLabel="Публикации и цитирования автора по годам" />
              ) : (
                <Skeleton height={280} radius={12} />
              )}
            </section>
          )}

          <section>
            <div className="panel-title">Граф соавторства</div>
            {graph && graph.nodes.length > 0 ? (
              <ForceGraph graph={graph} height={420} focusId={id} />
            ) : (
              <EmptyState
                title="Соавторы не найдены"
                description="В демо-корпусе у этого автора нет совместных публикаций."
              />
            )}
          </section>
        </div>

        <aside className="sidebar">
          {stats && stats.top_categories.length > 0 && (
            <div className="panel">
              <div className="panel-title">Основные направления</div>
              <div className="stack" style={{ gap: 8 }}>
                {stats.top_categories.map((category) => (
                  <Link
                    key={category.slug}
                    to={`/search?category=${encodeURIComponent(category.slug)}`}
                    className="row"
                    style={{ gap: 8, fontSize: 13.5 }}
                  >
                    <span style={{ flex: 1 }}>{category.name}</span>
                    <span className="dim mono">{category.count}</span>
                  </Link>
                ))}
              </div>
            </div>
          )}

          {stats && stats.coauthors.length > 0 && (
            <div className="panel">
              <div className="panel-title">Частые соавторы</div>
              <div className="stack" style={{ gap: 8 }}>
                {stats.coauthors.slice(0, 8).map((coauthor) => (
                  <Link
                    key={coauthor.id}
                    to={`/authors/${encodeURIComponent(coauthor.id)}`}
                    className="row"
                    style={{ gap: 8, fontSize: 13.5 }}
                  >
                    <span style={{ flex: 1 }}>{coauthor.name}</span>
                    <span className="chip">{coauthor.count}</span>
                  </Link>
                ))}
              </div>
            </div>
          )}

          {author.categories.length > 0 && (
            <div className="panel">
              <div className="panel-title">Разделы каталога</div>
              <div className="row wrap" style={{ gap: 8 }}>
                {author.categories.map((category) => (
                  <span key={category.id} className="chip">
                    {category.name}
                  </span>
                ))}
              </div>
            </div>
          )}
        </aside>
      </div>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: number | string }) {
  return (
    <div>
      <div className="kpi-value" style={{ fontSize: 22 }}>
        {typeof value === 'number' ? value.toLocaleString('ru-RU') : value}
      </div>
      <div className="kpi-label">{label}</div>
    </div>
  )
}

function FollowButton({ profile }: { profile: UserProfile }) {
  const [following, setFollowing] = useState(profile.following)
  const [followers, setFollowers] = useState(profile.followers_count)
  const [pending, setPending] = useState(false)

  const toggle = async () => {
    if (pending) return
    setPending(true)
    try {
      const state = following ? await api.unfollow(profile.id) : await api.follow(profile.id)
      setFollowing(state.following)
      setFollowers(state.followers_count)
    } catch {
      /* подписка живёт в памяти BFF — при сбое бейдж просто не меняется */
    } finally {
      setPending(false)
    }
  }

  return (
    <button
      type="button"
      className={`btn ${following ? '' : 'btn-primary'}`}
      onClick={toggle}
      disabled={pending}
      aria-pressed={following}
      title="Подписки хранятся в состоянии BFF в памяти"
    >
      <Icon name={following ? 'check' : 'plus'} size={15} />
      {following ? 'Вы подписаны' : 'Подписаться'}
      <span className="mono">{followers}</span>
    </button>
  )
}
