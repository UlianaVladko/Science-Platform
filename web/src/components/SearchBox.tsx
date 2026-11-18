import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { Suggestion } from '../api/types'
import { useDebounced } from '../hooks/useFetch'
import { Icon } from './Icons'

const TYPE_LABEL: Record<Suggestion['type'], string> = {
  publication: 'публикация',
  author: 'автор',
  term: 'понятие',
}

const MODES = [
  { value: 'hybrid', label: 'Гибрид' },
  { value: 'lexical', label: 'Точно' },
  { value: 'semantic', label: 'Смысл' },
] as const

export type SearchMode = (typeof MODES)[number]['value']

export function SearchBox({
  initialQuery = '',
  initialMode = 'hybrid',
  size = 'lg',
  autoFocus = false,
}: {
  initialQuery?: string
  initialMode?: SearchMode
  size?: 'lg' | 'md'
  autoFocus?: boolean
}) {
  const navigate = useNavigate()
  const [query, setQuery] = useState(initialQuery)
  const [mode, setMode] = useState<SearchMode>(initialMode)
  const [suggestions, setSuggestions] = useState<Suggestion[]>([])
  const [open, setOpen] = useState(false)
  const [active, setActive] = useState(-1)
  const containerRef = useRef<HTMLDivElement | null>(null)
  const debounced = useDebounced(query, 220)

  useEffect(() => {
    if (debounced.trim().length < 2) {
      setSuggestions([])
      return
    }
    let cancelled = false
    api
      .suggest(debounced.trim(), 7)
      .then((result) => {
        if (!cancelled) {
          setSuggestions(result.items)
          setActive(-1)
        }
      })
      .catch(() => {
        if (!cancelled) setSuggestions([])
      })
    return () => {
      cancelled = true
    }
  }, [debounced])

  useEffect(() => {
    const onClick = (event: MouseEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onClick)
    return () => document.removeEventListener('mousedown', onClick)
  }, [])

  const submit = (text: string, nextMode: SearchMode = mode) => {
    const trimmed = text.trim()
    setOpen(false)
    const params = new URLSearchParams()
    if (trimmed) params.set('q', trimmed)
    params.set('mode', nextMode)
    navigate(`/search?${params.toString()}`)
  }

  const pick = (item: Suggestion) => {
    setOpen(false)
    if (item.type === 'publication' && item.id) {
      navigate(`/publications/${encodeURIComponent(item.id)}`)
      return
    }
    if (item.type === 'author' && item.id) {
      navigate(`/authors/${encodeURIComponent(item.id)}`)
      return
    }
    submit(item.text)
  }

  const onKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'ArrowDown' && suggestions.length > 0) {
      event.preventDefault()
      setOpen(true)
      setActive((i) => (i + 1) % suggestions.length)
      return
    }
    if (event.key === 'ArrowUp' && suggestions.length > 0) {
      event.preventDefault()
      setActive((i) => (i - 1 + suggestions.length) % suggestions.length)
      return
    }
    if (event.key === 'Enter') {
      if (open && active >= 0 && suggestions[active]) {
        event.preventDefault()
        pick(suggestions[active])
      } else {
        submit(query)
      }
      return
    }
    if (event.key === 'Escape') {
      setOpen(false)
    }
  }

  const showPop = open && suggestions.length > 0
  const height = size === 'lg' ? 60 : 48
  const placeholder = useMemo(() => 'Найдите публикацию, автора или научное понятие…', [])

  return (
    <div className="search-shell" ref={containerRef}>
      <div className="search-input-wrap" style={{ paddingLeft: size === 'lg' ? 18 : 14 }}>
        <Icon name="search" size={size === 'lg' ? 20 : 17} style={{ color: 'var(--text-400)', flexShrink: 0 }} />
        <input
          className="search-input"
          style={{ height: height - 24, fontSize: size === 'lg' ? 16 : 14.5 }}
          value={query}
          autoFocus={autoFocus}
          placeholder={placeholder}
          onChange={(event) => {
            setQuery(event.target.value)
            setOpen(true)
          }}
          onFocus={() => setOpen(true)}
          onKeyDown={onKeyDown}
          aria-label="Поиск по научной литературе"
          role="combobox"
          aria-expanded={showPop}
          aria-controls="search-suggestions"
          aria-autocomplete="list"
        />
        <div className="search-modes" role="group" aria-label="Режим поиска">
          {MODES.map((item) => (
            <button
              key={item.value}
              type="button"
              className="search-mode"
              aria-pressed={mode === item.value}
              onClick={() => setMode(item.value)}
            >
              {item.label}
            </button>
          ))}
        </div>
        <button
          type="button"
          className="btn btn-primary"
          style={{ borderRadius: 999, paddingInline: 20 }}
          onClick={() => submit(query)}
        >
          Найти
        </button>
      </div>

      {showPop && (
        <div className="suggest-pop" id="search-suggestions" role="listbox">
          {suggestions.map((item, index) => (
            <button
              key={`${item.type}-${item.text}-${index}`}
              type="button"
              className="suggest-item"
              role="option"
              aria-selected={index === active}
              onMouseEnter={() => setActive(index)}
              onClick={() => pick(item)}
            >
              <span className="suggest-type">{TYPE_LABEL[item.type]}</span>
              <span style={{ flex: 1 }}>{item.text}</span>
              {item.matches_count != null && (
                <span className="dim mono">{item.matches_count}</span>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
