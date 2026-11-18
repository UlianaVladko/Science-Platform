import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { Discipline, LookupResult } from '../api/types'
import { useDocumentTitle, useFetch } from '../hooks/useFetch'
import { Icon } from '../components/Icons'
import { SourceBadge } from '../components/ui'

type Form = {
  title: string
  abstract: string
  journal: string
  publisher: string
  publication_date: string
  language: string
  publication_type: string
  category_slug: string
  keywords: string
  doi: string
}

const EMPTY: Form = {
  title: '',
  abstract: '',
  journal: '',
  publisher: '',
  publication_date: '',
  language: 'ru',
  publication_type: 'article',
  category_slug: '',
  keywords: '',
  doi: '',
}

const TYPES = [
  { value: 'article', label: 'Статья' },
  { value: 'book', label: 'Книга' },
  { value: 'monograph', label: 'Монография' },
  { value: 'preprint', label: 'Препринт' },
  { value: 'conference_paper', label: 'Доклад на конференции' },
  { value: 'other', label: 'Другое' },
]

export default function AddPublicationPage() {
  useDocumentTitle('Добавить публикацию')
  const navigate = useNavigate()
  const { data: disciplines } = useFetch<{ items: Discipline[] }>(() => api.disciplines(), [])

  const [lookupQuery, setLookupQuery] = useState('')
  const [lookupKind, setLookupKind] = useState<'doi' | 'title'>('doi')
  const [lookupResult, setLookupResult] = useState<LookupResult | null>(null)
  const [lookupError, setLookupError] = useState<string | null>(null)
  const [lookingUp, setLookingUp] = useState(false)

  const [form, setForm] = useState<Form>(EMPTY)
  const [authorsRaw, setAuthorsRaw] = useState('')
  const [pending, setPending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const set = (key: keyof Form, value: string) => setForm((current) => ({ ...current, [key]: value }))

  const runLookup = async (event: React.FormEvent) => {
    event.preventDefault()
    const query = lookupQuery.trim()
    if (!query) return
    setLookingUp(true)
    setLookupError(null)
    setLookupResult(null)
    try {
      const result = await api.lookup(lookupKind === 'doi' ? { doi: query } : { title: query })
      setLookupResult(result)
      if (result.found) {
        setForm({
          title: result.title ?? '',
          abstract: result.abstract ?? '',
          journal: '',
          publisher: '',
          publication_date: result.year ? `${result.year}-01-01` : '',
          language: result.language ?? 'ru',
          publication_type: result.publication_type ?? 'article',
          category_slug: result.category_slug ?? '',
          keywords: (result.keywords ?? []).join(', '),
          doi: lookupKind === 'doi' ? query : '',
        })
        setAuthorsRaw((result.authors ?? []).map((author) => author.name).join('\n'))
      }
    } catch (reason) {
      setLookupError(reason instanceof Error ? reason.message : 'Поиск не удался')
    } finally {
      setLookingUp(false)
    }
  }

  const authors = authorsRaw
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .map((line) => {
      const [name, ...rest] = line.split('|')
      return { name: name.trim(), affiliation: rest.join('|').trim() }
    })

  const canSubmit = form.title.trim().length >= 3 && authors.length > 0 && form.category_slug.length > 0 && !pending

  const submit = async (event: React.FormEvent) => {
    event.preventDefault()
    if (!canSubmit) return
    setPending(true)
    setError(null)
    try {
      const created = await api.createPublication({
        title: form.title.trim(),
        abstract: form.abstract.trim() || null,
        journal: form.journal.trim() || null,
        publisher: form.publisher.trim() || null,
        publication_date: form.publication_date || null,
        language: form.language,
        publication_type: form.publication_type,
        category_slug: form.category_slug,
        keywords: form.keywords
          .split(',')
          .map((item) => item.trim())
          .filter(Boolean),
        doi: form.doi.trim() || null,
        authors: authors.map((author) => ({
          name: author.name,
          affiliation: author.affiliation,
        })),
      })
      navigate(`/publications/${encodeURIComponent(created.id)}`)
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : 'Не удалось сохранить публикацию')
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="page page-narrow">
      <div className="section-head" style={{ paddingTop: 'clamp(24px, 3vw, 40px)' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: 'clamp(24px, 2.6vw, 34px)' }}>Добавить публикацию</h1>
          <p>
            Подтяните метаданные по DOI или названию, проверьте поля и сохраните карточку. Демо-корпус
            позволяет создавать записи даже без согласованного API.
          </p>
        </div>
        <SourceBadge source="demo" />
      </div>

      <form className="panel" onSubmit={runLookup} style={{ marginBottom: 'var(--space-5)' }}>
        <div className="row wrap" style={{ gap: 10 }}>
          <div className="chip" style={{ padding: 3, gap: 3 }}>
            {(
              [
                { value: 'doi', label: 'DOI' },
                { value: 'title', label: 'Название' },
              ] as const
            ).map((option) => (
              <button
                key={option.value}
                type="button"
                className="chip"
                aria-pressed={lookupKind === option.value}
                style={{ border: 'none', background: 'transparent' }}
                onClick={() => setLookupKind(option.value)}
              >
                {option.label}
              </button>
            ))}
          </div>
          <input
            className="input"
            style={{ flex: 1, minWidth: 220 }}
            value={lookupQuery}
            onChange={(event) => setLookupQuery(event.target.value)}
            placeholder={lookupKind === 'doi' ? '10.1234/example.2026' : 'Часть названия публикации'}
            aria-label="Что искать в источниках"
          />
          <button type="submit" className="btn" disabled={lookingUp || !lookupQuery.trim()}>
            <Icon name="search" size={15} />
            {lookingUp ? 'Ищем…' : 'Заполнить'}
          </button>
        </div>

        {lookupError && <p style={{ color: 'var(--rose)', fontSize: 13, marginTop: 10 }}>{lookupError}</p>}
        {lookupResult && (
          <p className="dim" style={{ fontSize: 12.5, marginTop: 10, marginBottom: 0 }}>
            {lookupResult.found
              ? `Метаданные подставлены из источника «${lookupResult.source}». Проверьте поля перед сохранением.`
              : 'Совпадений не найдено — заполните карточку вручную.'}
          </p>
        )}
      </form>

      <form onSubmit={submit} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
        <div className="panel" style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div className="panel-title">Основные сведения</div>
          <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <span className="facet-title">Название</span>
            <input
              className="input"
              value={form.title}
              onChange={(event) => set('title', event.target.value)}
              placeholder="Полное название публикации"
              required
            />
          </label>

          <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <span className="facet-title">Аннотация</span>
            <textarea
              className="input"
              rows={4}
              value={form.abstract}
              onChange={(event) => set('abstract', event.target.value)}
              placeholder="Краткое содержание работы"
            />
          </label>

          <div className="row wrap" style={{ gap: 10 }}>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 6, flex: 1, minWidth: 180 }}>
              <span className="facet-title">Журнал</span>
              <input className="input" value={form.journal} onChange={(event) => set('journal', event.target.value)} />
            </label>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 6, flex: 1, minWidth: 180 }}>
              <span className="facet-title">Издатель</span>
              <input
                className="input"
                value={form.publisher}
                onChange={(event) => set('publisher', event.target.value)}
              />
            </label>
          </div>

          <div className="row wrap" style={{ gap: 10 }}>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 6, minWidth: 160 }}>
              <span className="facet-title">Дата</span>
              <input
                className="input"
                type="date"
                value={form.publication_date}
                onChange={(event) => set('publication_date', event.target.value)}
              />
            </label>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 6, minWidth: 180 }}>
              <span className="facet-title">Тип</span>
              <select
                className="input"
                value={form.publication_type}
                onChange={(event) => set('publication_type', event.target.value)}
              >
                {TYPES.map((item) => (
                  <option key={item.value} value={item.value}>
                    {item.label}
                  </option>
                ))}
              </select>
            </label>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 6, minWidth: 130 }}>
              <span className="facet-title">Язык</span>
              <select className="input" value={form.language} onChange={(event) => set('language', event.target.value)}>
                <option value="ru">Русский</option>
                <option value="en">Английский</option>
              </select>
            </label>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 6, minWidth: 200 }}>
              <span className="facet-title">Направление</span>
              <select
                className="input"
                value={form.category_slug}
                onChange={(event) => set('category_slug', event.target.value)}
                required
              >
                <option value="">Выберите…</option>
                {(disciplines?.items ?? []).map((item) => (
                  <option key={item.slug} value={item.slug}>
                    {item.name}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="row wrap" style={{ gap: 10 }}>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 6, flex: 1, minWidth: 220 }}>
              <span className="facet-title">Ключевые слова</span>
              <input
                className="input"
                value={form.keywords}
                onChange={(event) => set('keywords', event.target.value)}
                placeholder="через запятую"
              />
            </label>
            <label style={{ display: 'flex', flexDirection: 'column', gap: 6, minWidth: 200 }}>
              <span className="facet-title">DOI</span>
              <input className="input" value={form.doi} onChange={(event) => set('doi', event.target.value)} />
            </label>
          </div>
        </div>

        <div className="panel" style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div className="panel-title">Авторы</div>
          <textarea
            className="input"
            rows={4}
            value={authorsRaw}
            onChange={(event) => setAuthorsRaw(event.target.value)}
            placeholder={'Иванов Иван | МГУ\nПетрова Анна | Институт прикладной математики'}
          />
          <span className="dim" style={{ fontSize: 11.5 }}>
            По одному автору в строке, формат «ФИО | организация». Порядок строк = порядок авторства.
            Разобрано авторов: {authors.length}.
          </span>
        </div>

        {error && <p style={{ color: 'var(--rose)', fontSize: 13, margin: 0 }}>{error}</p>}

        <div className="row wrap" style={{ gap: 10 }}>
          <button type="submit" className="btn btn-primary" disabled={!canSubmit}>
            <Icon name="check" size={15} />
            {pending ? 'Сохраняем…' : 'Сохранить публикацию'}
          </button>
          <Link to="/search" className="btn btn-ghost">
            Отмена
          </Link>
          <div className="spacer" />
          <span className="dim" style={{ fontSize: 12 }}>
            Запись создаётся в демо-корпусе BFF и живёт до перезапуска сервиса.
          </span>
        </div>
      </form>
    </div>
  )
}
