import { useState } from 'react'
import { api } from '../api/client'
import type { Discipline, FunFact } from '../api/types'
import { Icon } from './Icons'

export function DisciplineCard({
  discipline,
  onSelect,
}: {
  discipline: Discipline
  onSelect: (slug: string) => void
}) {
  const [fact, setFact] = useState<FunFact | null>(null)
  const [loadingFact, setLoadingFact] = useState(false)
  const accent = discipline.accent || '#7c5cff'

  const prefetch = () => {
    if (fact || loadingFact) return
    setLoadingFact(true)
    api
      .funFact(discipline.slug)
      .then(setFact)
      .catch(() => setFact(null))
      .finally(() => setLoadingFact(false))
  }

  return (
    <button
      type="button"
      className="discipline-card"
      style={{ '--accent-fg': accent, '--accent-line': accent + '66' } as React.CSSProperties}
      onClick={() => onSelect(discipline.slug)}
      onMouseEnter={prefetch}
      onFocus={prefetch}
      aria-label={`${discipline.name}: ${discipline.publications_count} публикаций`}
    >
      <span className="discipline-icon">
        <Icon name={discipline.icon} size={21} />
      </span>
      <span className="discipline-name">{discipline.name}</span>
      {discipline.description && (
        <span className="muted" style={{ fontSize: 12.5, lineHeight: 1.4 }}>
          {discipline.description}
        </span>
      )}
      <span className="discipline-meta">
        <span>{discipline.publications_count.toLocaleString('ru-RU')} публикаций</span>
        <span>{discipline.subcategories_count} разделов</span>
      </span>

      <span className="fun-fact" aria-hidden="true">
        <strong>Интересный факт</strong>
        {fact ? fact.text : loadingFact ? 'Загружаем факт…' : 'Наведите, чтобы узнать факт'}
      </span>
    </button>
  )
}
