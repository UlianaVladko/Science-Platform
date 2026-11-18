import { useEffect, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { api } from '../api/client'
import type { Discipline, Page as ApiPage, Question } from '../api/types'
import { useDocumentTitle, useFetch } from '../hooks/useFetch'
import { Icon } from '../components/Icons'
import { EmptyState, ErrorState, Skeleton, SourceBadge } from '../components/ui'

const SORTS = [
  { value: 'new', label: 'Новые' },
  { value: 'votes', label: 'По голосам' },
  { value: 'views', label: 'По просмотрам' },
] as const

function relativeTime(iso: string) {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return iso
  const diff = Date.now() - date.getTime()
  const minutes = Math.round(diff / 60000)
  if (minutes < 1) return 'только что'
  if (minutes < 60) return `${minutes} мин назад`
  const hours = Math.round(minutes / 60)
  if (hours < 24) return `${hours} ч назад`
  const days = Math.round(hours / 24)
  if (days < 31) return `${days} дн назад`
  return date.toLocaleDateString('ru-RU')
}

function QuestionList() {
  const [params, setParams] = useSearchParams()
  const sort = (params.get('sort') as string) ?? 'new'
  const category = params.get('category') ?? ''
  const page = Math.max(1, Number(params.get('page') ?? 1) || 1)

  const patch = (next: Record<string, string | number | undefined>) => {
    const merged = new URLSearchParams(params)
    Object.entries(next).forEach(([key, value]) => {
      if (value === undefined || value === '') merged.delete(key)
      else merged.set(key, String(value))
    })
    setParams(merged, { replace: true })
  }

  const { data, loading, error, reload } = useFetch<ApiPage<Question>>(
    () => api.questions({ page, page_size: 10, sort, category: category || undefined }),
    [page, sort, category],
  )
  const { data: disciplines } = useFetch<{ items: Discipline[] }>(() => api.disciplines(), [])

  const pages = data ? Math.max(1, Math.ceil(data.total / data.page_size)) : 1

  return (
    <>
      <div className="row wrap" style={{ gap: 10, marginBottom: 16 }}>
        <div className="chip" style={{ padding: 3, gap: 3 }}>
          {SORTS.map((item) => (
            <button
              key={item.value}
              type="button"
              className="chip"
              aria-pressed={sort === item.value}
              style={{ border: 'none', background: 'transparent' }}
              onClick={() => patch({ sort: item.value, page: undefined })}
            >
              {item.label}
            </button>
          ))}
        </div>
        <select
          className="input"
          style={{ maxWidth: 220 }}
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
        <div className="spacer" />
        {data && <span className="dim" style={{ fontSize: 13 }}>Вопросов: {data.total}</span>}
      </div>

      {loading && (
        <div className="stack">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} height={92} radius={14} />
          ))}
        </div>
      )}
      {error && <ErrorState message={error} onRetry={reload} />}
      {!loading && !error && data && data.items.length === 0 && (
        <EmptyState
          title="Вопросов пока нет"
          description="Задайте первый вопрос по научной литературе — ответы хранятся в состоянии BFF."
        />
      )}
      {!loading && !error && data && data.items.length > 0 && (
        <div className="stack">
          {data.items.map((question) => (
            <Link key={question.id} to={`/questions/${question.id}`} className="question-card">
              <div style={{ display: 'flex', flexDirection: 'column', gap: 4, alignItems: 'center', minWidth: 54 }}>
                <span className="mono" style={{ fontSize: 17, color: 'var(--accent-soft)' }}>
                  {question.votes}
                </span>
                <span className="dim" style={{ fontSize: 11 }}>
                  голосов
                </span>
                <span className="mono" style={{ fontSize: 13, color: 'var(--text-300)' }}>
                  {question.answers_count}
                </span>
                <span className="dim" style={{ fontSize: 11 }}>
                  ответов
                </span>
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 570, marginBottom: 4 }}>{question.title}</div>
                <div className="pub-meta">
                  <span className="chip">{question.category.name}</span>
                  <span>{question.author.name}</span>
                  <span>· {relativeTime(question.created_at)}</span>
                  <span>· {question.views_count} просмотров</span>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}

      {pages > 1 && (
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
    </>
  )
}

function QuestionForm({ onCreated }: { onCreated: (question: Question) => void }) {
  const { data: disciplines } = useFetch<{ items: Discipline[] }>(() => api.disciplines(), [])
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [slug, setSlug] = useState('')
  const [authorName, setAuthorName] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!slug && disciplines?.items.length) setSlug(disciplines.items[0].slug)
  }, [disciplines, slug])

  const titleValid = title.trim().length >= 5 && title.trim().length <= 200
  const bodyValid = body.trim().length >= 1 && body.trim().length <= 5000
  const canSubmit = titleValid && bodyValid && slug.length > 0 && !pending

  const submit = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!canSubmit) return
    setPending(true)
    setError(null)
    try {
      const created = await api.createQuestion({
        title: title.trim(),
        body: body.trim(),
        category_slug: slug,
        author_name: authorName.trim() || undefined,
      })
      setTitle('')
      setBody('')
      onCreated(created)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Не удалось создать вопрос')
    } finally {
      setPending(false)
    }
  }

  return (
    <form className="panel" onSubmit={submit} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div className="row" style={{ gap: 8 }}>
        <Icon name="quote" size={16} style={{ color: 'var(--accent-soft)' }} />
        <span className="widget-title">Задать вопрос</span>
        <div className="spacer" />
        <SourceBadge source="bff" />
      </div>

      <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        <span className="facet-title">Заголовок</span>
        <input
          className="input"
          value={title}
          onChange={(event) => setTitle(event.target.value)}
          placeholder="Например: как искать работы по семантике запроса?"
          maxLength={200}
          required
        />
        <span className="dim" style={{ fontSize: 11.5 }}>
          {title.trim().length}/200 · минимум 5 символов
        </span>
      </label>

      <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
        <span className="facet-title">Описание</span>
        <textarea
          className="input"
          value={body}
          onChange={(event) => setBody(event.target.value)}
          rows={4}
          maxLength={5000}
          placeholder="Контекст, что уже пробовали, какой результат нужен"
          required
        />
        <span className="dim" style={{ fontSize: 11.5 }}>
          {body.trim().length}/5000
        </span>
      </label>

      <div className="row wrap" style={{ gap: 10 }}>
        <select
          className="input"
          style={{ maxWidth: 240 }}
          value={slug}
          onChange={(event) => setSlug(event.target.value)}
          aria-label="Направление вопроса"
        >
          {(disciplines?.items ?? []).map((item) => (
            <option key={item.slug} value={item.slug}>
              {item.name}
            </option>
          ))}
        </select>
        <input
          className="input"
          style={{ maxWidth: 220 }}
          value={authorName}
          onChange={(event) => setAuthorName(event.target.value)}
          placeholder="Ваше имя (необязательно)"
          aria-label="Ваше имя"
        />
        <div className="spacer" />
        <button type="submit" className="btn btn-primary" disabled={!canSubmit}>
          <Icon name="plus" size={15} />
          {pending ? 'Отправляем…' : 'Опубликовать'}
        </button>
      </div>

      {error && <p style={{ color: 'var(--rose)', fontSize: 13, margin: 0 }}>{error}</p>}
    </form>
  )
}

export default function QuestionsPage() {
  useDocumentTitle('Вопросы')
  const [created, setCreated] = useState<Question | null>(null)

  return (
    <div className="page">
      <div className="section-head" style={{ paddingTop: 'clamp(24px, 3vw, 40px)' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: 'clamp(24px, 2.6vw, 34px)' }}>Вопросы и ответы</h1>
          <p>
            Социальный контур, которого нет в базовом контракте API: вопросы, ответы и голоса хранятся в
            памяти BFF и сбрасываются при перезапуске.
          </p>
        </div>
        <SourceBadge source="bff" />
      </div>

      <div style={{ marginBottom: 'var(--space-5)' }}>
        <QuestionForm
          onCreated={(question) => {
            setCreated(question)
          }}
        />
      </div>

      {created && (
        <div className="panel rise" style={{ marginBottom: 'var(--space-4)' }}>
          <div className="row" style={{ gap: 8 }}>
            <Icon name="check" size={16} style={{ color: 'var(--teal)' }} />
            <span style={{ flex: 1 }}>
              Вопрос создан: <Link to={`/questions/${created.id}`}>{created.title}</Link>
            </span>
            <button type="button" className="btn btn-ghost btn-sm" onClick={() => setCreated(null)}>
              Скрыть
            </button>
          </div>
        </div>
      )}

      <QuestionList />
    </div>
  )
}

export function QuestionDetailPage() {
  const { id = '' } = useParams()
  const numericId = Number(id)
  const valid = Number.isInteger(numericId) && numericId > 0

  const { data, loading, error, reload } = useFetch<Question>(
    () => (valid ? api.question(numericId) : Promise.reject(new Error('Некорректный идентификатор вопроса'))),
    [numericId, valid],
  )

  useDocumentTitle(data?.title ?? 'Вопрос')

  const [body, setBody] = useState('')
  const [authorName, setAuthorName] = useState('')
  const [pending, setPending] = useState(false)
  const [sendError, setSendError] = useState<string | null>(null)

  const submitAnswer = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!valid || body.trim().length === 0 || pending) return
    setPending(true)
    setSendError(null)
    try {
      await api.createAnswer(numericId, { body: body.trim(), author_name: authorName.trim() || undefined })
      setBody('')
      reload()
    } catch (reason) {
      setSendError(reason instanceof Error ? reason.message : 'Не удалось добавить ответ')
    } finally {
      setPending(false)
    }
  }

  if (loading) {
    return (
      <div className="page page-narrow" style={{ paddingTop: 40 }}>
        <Skeleton height={30} width="60%" />
        <div style={{ height: 12 }} />
        <Skeleton height={80} radius={14} />
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="page page-narrow" style={{ paddingTop: 40 }}>
        <ErrorState message={error ?? 'Вопрос не найден'} onRetry={valid ? reload : undefined} />
        <p style={{ marginTop: 16 }}>
          <Link to="/questions" className="btn btn-sm">
            <Icon name="arrowLeft" size={14} /> К списку вопросов
          </Link>
        </p>
      </div>
    )
  }

  const answers = data.answers ?? []

  return (
    <div className="page page-narrow">
      <nav className="breadcrumbs" aria-label="Хлебные крошки">
        <Link to="/">Главная</Link>
        <span>/</span>
        <Link to="/questions">Вопросы</Link>
        <span>/</span>
        <span>#{data.id}</span>
      </nav>

      <h1 style={{ marginTop: 0 }}>{data.title}</h1>
      <div className="row wrap" style={{ gap: 8, marginBottom: 16 }}>
        <span className="chip">{data.category.name}</span>
        <span className="dim" style={{ fontSize: 13 }}>
          {data.author.name} · {relativeTime(data.created_at)} · {data.views_count} просмотров
        </span>
        <div className="spacer" />
        <SourceBadge source={data.source} />
      </div>

      <p style={{ color: 'var(--text-200)', lineHeight: 1.65, whiteSpace: 'pre-wrap' }}>{data.body}</p>

      <h2 style={{ fontSize: 18, marginTop: 'var(--space-5)' }}>
        Ответы <span className="dim">({answers.length})</span>
      </h2>

      {answers.length === 0 && (
        <EmptyState title="Ответов пока нет" description="Ответьте первым — это помогает другим читателям." />
      )}

      <div className="stack">
        {answers.map((answer) => (
          <AnswerCard key={answer.id} answer={answer} />
        ))}
      </div>

      <form className="panel" onSubmit={submitAnswer} style={{ marginTop: 'var(--space-5)', display: 'flex', flexDirection: 'column', gap: 10 }}>
        <div className="row" style={{ gap: 8 }}>
          <Icon name="quote" size={16} style={{ color: 'var(--teal)' }} />
          <span className="widget-title">Ваш ответ</span>
        </div>
        <textarea
          className="input"
          rows={4}
          value={body}
          onChange={(event) => setBody(event.target.value)}
          placeholder="Ответ по существу, со ссылками на публикации корпуса"
        />
        <div className="row wrap" style={{ gap: 10 }}>
          <input
            className="input"
            style={{ maxWidth: 220 }}
            value={authorName}
            onChange={(event) => setAuthorName(event.target.value)}
            placeholder="Ваше имя (необязательно)"
            aria-label="Ваше имя"
          />
          <div className="spacer" />
          <button type="submit" className="btn btn-primary" disabled={pending || body.trim().length === 0}>
            <Icon name="check" size={15} />
            {pending ? 'Отправляем…' : 'Ответить'}
          </button>
        </div>
        {sendError && <p style={{ color: 'var(--rose)', fontSize: 13, margin: 0 }}>{sendError}</p>}
      </form>
    </div>
  )
}

function AnswerCard({ answer }: { answer: NonNullable<Question['answers']>[number] }) {
  const [votes, setVotes] = useState(answer.votes)
  const [voted, setVoted] = useState<0 | 1 | -1>(0)
  const [pending, setPending] = useState(false)

  const vote = async (value: 1 | -1) => {
    if (pending) return
    setPending(true)
    try {
      const result = await api.vote(answer.id, value)
      setVotes(result.votes)
      setVoted((current) => (current === value ? 0 : value))
    } catch {
      /* голоса живут в памяти BFF */
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="question-card" style={{ alignItems: 'flex-start' }}>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6, minWidth: 44 }}>
        <button
          type="button"
          className={`btn btn-sm ${voted === 1 ? 'btn-primary' : 'btn-ghost'}`}
          onClick={() => vote(1)}
          disabled={pending}
          aria-label="Полезный ответ"
        >
          +
        </button>
        <span className="mono" style={{ fontSize: 14 }}>
          {votes}
        </span>
        <button
          type="button"
          className={`btn btn-sm ${voted === -1 ? 'btn-primary' : 'btn-ghost'}`}
          onClick={() => vote(-1)}
          disabled={pending}
          aria-label="Бесполезный ответ"
        >
          −
        </button>
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <p style={{ margin: 0, color: 'var(--text-200)', lineHeight: 1.6, whiteSpace: 'pre-wrap' }}>{answer.body}</p>
        <div className="pub-meta" style={{ marginTop: 8 }}>
          <span>{answer.author.name}</span>
          <span>· {relativeTime(answer.created_at)}</span>
        </div>
      </div>
    </div>
  )
}
