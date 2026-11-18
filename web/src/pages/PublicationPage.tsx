import { Fragment, useEffect, useMemo, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../api/client'
import type { Page, Publication, Reference, RelatedItem } from '../api/types'
import { useDocumentTitle, useFetch } from '../hooks/useFetch'
import { Icon } from '../components/Icons'
import { EmptyState, ErrorState, Skeleton, SourceBadge } from '../components/ui'

const TYPE_LABEL: Record<string, string> = {
  article: 'статья',
  book: 'книга',
  preprint: 'препринт',
  conference_paper: 'доклад на конференции',
}

const STATUS_LABEL: Record<string, string> = {
  indexed: 'проиндексирована',
  processed: 'обработана',
  pending: 'в обработке',
  failed: 'ошибка обработки',
}

export default function PublicationPage() {
  const { id = '' } = useParams()
  const { data, loading, error, reload } = useFetch<Publication>(() => api.publication(id), [id])
  const [direction, setDirection] = useState<'incoming' | 'outgoing'>('incoming')
  const [refsPage, setRefsPage] = useState(1)

  const { data: refs, error: refsError } = useFetch<Page<Reference>>(
    () => api.references(id, direction, refsPage, 10),
    [id, direction, refsPage],
  )
  const { data: related } = useFetch<{ items: RelatedItem[] }>(() => api.related(id, 6), [id])

  useDocumentTitle(data?.title ?? 'Публикация')

  useEffect(() => {
    if (!id) return
    api.view(id).catch(() => undefined)
  }, [id])

  useEffect(() => {
    setRefsPage(1)
  }, [direction])

  if (loading) {
    return (
      <div className="page" style={{ paddingTop: 40 }}>
        <Skeleton height={38} width="70%" />
        <div style={{ height: 12 }} />
        <Skeleton height={20} width="40%" />
        <div style={{ height: 24 }} />
        <Skeleton height={180} radius={16} />
        <div style={{ height: 16 }} />
        <Skeleton height={140} radius={16} />
      </div>
    )
  }

  if (error) {
    return (
      <div className="page" style={{ paddingTop: 40 }}>
        <ErrorState message={error} onRetry={reload} />
        <p style={{ marginTop: 16 }}>
          <Link to="/search" className="btn btn-sm">
            <Icon name="arrowLeft" size={14} /> К поиску
          </Link>
        </p>
      </div>
    )
  }

  if (!data) return null

  const year = data.publication_date?.slice(0, 4)
  const refItems = refs?.items ?? []

  return (
    <div className="page">
      <nav className="breadcrumbs" aria-label="Хлебные крошки">
        <Link to="/">Главная</Link>
        <span>/</span>
        <Link to="/search">Поиск</Link>
        {data.categories[0] && (
          <>
            <span>/</span>
            <Link to={`/search?category=${encodeURIComponent(data.categories[0].slug)}`}>
              {data.categories[0].name}
            </Link>
          </>
        )}
      </nav>

      <header className="pub-header">
        <div className="row wrap" style={{ gap: 8 }}>
          <span className="chip">{TYPE_LABEL[data.publication_type] ?? data.publication_type}</span>
          {data.processing_status && (
            <span className="chip">
              <Icon name="check" size={13} />
              {STATUS_LABEL[data.processing_status] ?? data.processing_status}
            </span>
          )}
          {year && <span className="chip">{year}</span>}
          {data.source && <SourceBadge source={data.source} />}
        </div>

        <h1>{data.title}</h1>

        {data.authors.length > 0 && (
          <div className="author-line">
            {data.authors.map((author, index) => (
              <span key={author.id}>
                <Link to={`/authors/${encodeURIComponent(author.id)}`}>{author.name}</Link>
                {author.affiliations.length > 0 && (
                  <span className="dim"> — {author.affiliations.map((item) => item.name).join(', ')}</span>
                )}
                {index < data.authors.length - 1 && <span className="dim">;</span>}
              </span>
            ))}
          </div>
        )}

        <div className="row wrap" style={{ gap: 8 }}>
          <MetricPill icon="quote" label="цитирований" value={data.metrics.citations} />
          <MetricPill icon="eye" label="просмотров" value={data.metrics.views} />
          <MetricPill icon="download" label="скачиваний" value={data.metrics.downloads} />
          <div className="spacer" />
          <LikeButton publication={data} />
        </div>
      </header>

      <div className="detail-layout">
        <div className="stack" style={{ gap: 'var(--space-5)' }}>
          {data.abstract && (
            <section className="panel">
              <div className="panel-title">Аннотация</div>
              <p style={{ margin: 0, color: 'var(--text-200)', lineHeight: 1.65 }}>{data.abstract}</p>
            </section>
          )}

          {data.keywords.length > 0 && (
            <section>
              <div className="panel-title">Ключевые слова</div>
              <div className="row wrap" style={{ gap: 8 }}>
                {data.keywords.map((keyword) => (
                  <Link key={keyword} to={`/search?q=${encodeURIComponent(keyword)}`} className="chip">
                    {keyword}
                  </Link>
                ))}
              </div>
            </section>
          )}

          {data.concepts.length > 0 && (
            <section>
              <div className="panel-title">Понятия предметной области</div>
              <div className="row wrap" style={{ gap: 8 }}>
                {data.concepts.map((concept) => (
                  <Link
                    key={concept.id}
                    to={`/search?q=${encodeURIComponent(concept.name)}`}
                    className="chip chip-accent"
                  >
                    {concept.name}
                  </Link>
                ))}
              </div>
            </section>
          )}

          <section>
            <div className="row" style={{ marginBottom: 10 }}>
              <div className="panel-title" style={{ margin: 0 }}>
                Библиография и цитирования
              </div>
              <div className="spacer" />
              <div className="chip" style={{ padding: 3, gap: 3 }}>
                {(
                  [
                    { value: 'incoming', label: 'Цитируют работу' },
                    { value: 'outgoing', label: 'Работа цитирует' },
                  ] as const
                ).map((option) => (
                  <button
                    key={option.value}
                    type="button"
                    className="chip"
                    aria-pressed={direction === option.value}
                    style={{ border: 'none', background: 'transparent' }}
                    onClick={() => setDirection(option.value)}
                  >
                    {option.label}
                  </button>
                ))}
              </div>
            </div>

            {refsError && <ErrorState message={refsError} />}
            {!refsError && refItems.length === 0 && (
              <EmptyState
                title="Связей не найдено"
                description="В корпусе пока нет библиографических связей этого типа."
              />
            )}
            {refItems.length > 0 && (
              <div>
                {refItems.map((reference, index) => (
                  <div className="ref-item" key={`${reference.citing_publication_id}-${reference.cited_publication_id}-${index}`}>
                    {reference.publication ? (
                      <Link
                        to={`/publications/${encodeURIComponent(reference.publication.id)}`}
                        style={{ fontWeight: 550 }}
                      >
                        {reference.publication.title}
                      </Link>
                    ) : (
                      <span className="muted" style={{ fontSize: 13.5 }}>
                        {reference.raw_reference}
                      </span>
                    )}
                    {reference.publication && (
                      <span className="dim" style={{ fontSize: 12.5 }}>
                        {reference.publication.authors.map((author) => author.name).join(', ')}
                        {reference.publication.publication_date &&
                          ` · ${reference.publication.publication_date.slice(0, 4)}`}
                      </span>
                    )}
                    {reference.publication && reference.raw_reference && (
                      <span className="dim mono" style={{ fontSize: 11.5 }}>
                        {reference.raw_reference}
                      </span>
                    )}
                  </div>
                ))}
                {refs && refs.total > refs.page_size && (
                  <div className="pager">
                    <button
                      type="button"
                      className="btn btn-sm"
                      disabled={refsPage <= 1}
                      onClick={() => setRefsPage((value) => Math.max(1, value - 1))}
                    >
                      <Icon name="arrowLeft" size={14} /> Назад
                    </button>
                    <span className="dim" style={{ fontSize: 13 }}>
                      {refsPage} из {Math.ceil(refs.total / refs.page_size)}
                    </span>
                    <button
                      type="button"
                      className="btn btn-sm"
                      disabled={refsPage >= Math.ceil(refs.total / refs.page_size)}
                      onClick={() => setRefsPage((value) => value + 1)}
                    >
                      Вперёд <Icon name="arrowRight" size={14} />
                    </button>
                  </div>
                )}
              </div>
            )}
          </section>

          {related && related.items.length > 0 && (
            <section>
              <div className="panel-title">Похожие публикации</div>
              <div className="stack">
                {related.items.map((item) => (
                  <Link
                    key={item.publication.id}
                    to={`/publications/${encodeURIComponent(item.publication.id)}`}
                    className="pub-row"
                  >
                    <span className="pub-title">{item.publication.title}</span>
                    <div className="pub-meta">
                      {item.shared.slice(0, 4).map((tag) => (
                        <span key={tag} className="chip">
                          {tag}
                        </span>
                      ))}
                      <span className="dim">близость {item.score.toFixed(2)}</span>
                    </div>
                  </Link>
                ))}
              </div>
            </section>
          )}
        </div>

        <aside className="sidebar">
          <div className="panel">
            <div className="panel-title">Библиографические данные</div>
            <dl className="definition">
              {data.journal && (
                <>
                  <dt>Журнал</dt>
                  <dd>{data.journal}</dd>
                </>
              )}
              {data.publisher && (
                <>
                  <dt>Издатель</dt>
                  <dd>{data.publisher}</dd>
                </>
              )}
              <dt>Тип</dt>
              <dd>{TYPE_LABEL[data.publication_type] ?? data.publication_type}</dd>
              <dt>Язык</dt>
              <dd>{data.language}</dd>
              {data.publication_date && (
                <>
                  <dt>Дата</dt>
                  <dd>{data.publication_date}</dd>
                </>
              )}
              {data.doi && (
                <>
                  <dt>DOI</dt>
                  <dd className="mono">{data.doi}</dd>
                </>
              )}
              {data.external_identifiers.map((identifier) => (
                <Fragment key={`${identifier.scheme}-${identifier.value}`}>
                  <dt>{identifier.scheme}</dt>
                  <dd className="mono">{identifier.value}</dd>
                </Fragment>
              ))}
              {data.categories.length > 0 && (
                <>
                  <dt>Разделы</dt>
                  <dd>{data.categories.map((category) => category.name).join(', ')}</dd>
                </>
              )}
            </dl>
          </div>

          <div className="panel">
            <div className="panel-title">Файлы</div>
            {data.files.length === 0 && <p className="muted" style={{ fontSize: 13 }}>Файлы не прикреплены.</p>}
            <div className="stack" style={{ gap: 8 }}>
              {data.files.map((file) => (
                <a
                  key={file.kind}
                  className="row"
                  style={{ gap: 8, fontSize: 13.5 }}
                  href={`/api/publications/${encodeURIComponent(data.id)}/files/${encodeURIComponent(file.kind)}`}
                  target="_blank"
                  rel="noreferrer"
                >
                  <Icon name="download" size={15} style={{ color: 'var(--text-400)' }} />
                  <span style={{ flex: 1 }}>{file.kind}</span>
                  <span className="dim mono">{formatBytes(file.size_bytes)}</span>
                </a>
              ))}
            </div>
          </div>

          <div className="panel">
            <div className="panel-title">Ссылка на карточку</div>
            <CopyLink />
          </div>
        </aside>
      </div>
    </div>
  )
}

function MetricPill({ icon, label, value }: { icon: string; label: string; value: number }) {
  return (
    <span className="chip" title={label}>
      <Icon name={icon} size={14} />
      <span className="mono">{value.toLocaleString('ru-RU')}</span>
      <span className="dim">{label}</span>
    </span>
  )
}

function LikeButton({ publication }: { publication: Publication }) {
  const [liked, setLiked] = useState(publication.liked)
  const [count, setCount] = useState(publication.likes_count)
  const [pending, setPending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const mounted = useRef(true)

  useEffect(() => {
    mounted.current = true
    return () => {
      mounted.current = false
    }
  }, [])

  const toggle = async () => {
    if (pending) return
    setPending(true)
    setError(null)
    try {
      const state = liked ? await api.unlike(publication.id) : await api.like(publication.id)
      if (!mounted.current) return
      setLiked(state.liked)
      setCount(state.likes_count)
    } catch {
      if (mounted.current) setError('Лайк не сохранился — состояние BFF не отвечает')
    } finally {
      if (mounted.current) setPending(false)
    }
  }

  return (
    <div className="row" style={{ gap: 8 }}>
      {error && (
        <span className="dim" style={{ fontSize: 12 }}>
          {error}
        </span>
      )}
      <button
        type="button"
        className={`btn btn-sm ${liked ? 'btn-primary' : ''}`}
        onClick={toggle}
        disabled={pending}
        aria-pressed={liked}
        title="Лайки хранятся в состоянии BFF в памяти и сбрасываются при перезапуске"
      >
        <Icon name="heart" size={15} />
        {liked ? 'В избранном' : 'Нравится'}
        <span className="mono">{count}</span>
      </button>
    </div>
  )
}

function CopyLink() {
  const [copied, setCopied] = useState(false)
  const url = useMemo(() => (typeof window === 'undefined' ? '' : window.location.href), [])

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(url)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 1600)
    } catch {
      setCopied(false)
    }
  }

  return (
    <button type="button" className="btn btn-sm" onClick={copy} style={{ width: '100%', justifyContent: 'center' }}>
      <Icon name={copied ? 'check' : 'file'} size={14} />
      {copied ? 'Скопировано' : 'Скопировать ссылку'}
    </button>
  )
}

function formatBytes(bytes: number) {
  if (!bytes) return '0 Б'
  if (bytes < 1024) return `${bytes} Б`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} КБ`
  return `${(bytes / (1024 * 1024)).toFixed(1)} МБ`
}
