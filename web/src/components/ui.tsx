import type { ReactNode } from 'react'
import { Icon } from './Icons'
import type { Source } from '../api/types'

export function Chip({
  children,
  tone,
}: {
  children: ReactNode
  tone?: 'accent' | 'teal' | 'demo'
}) {
  const cls = tone === 'accent' ? 'chip chip-accent' : tone === 'teal' ? 'chip chip-teal' : tone === 'demo' ? 'chip badge-demo' : 'chip'
  return <span className={cls}>{children}</span>
}

export function SourceBadge({ source }: { source?: Source }) {
  if (source === 'api') {
    return (
      <span className="chip badge-api" title="Данные из согласованного API команды API">
        API
      </span>
    )
  }
  if (source === 'demo') {
    return (
      <span className="chip badge-demo" title="Демонстрационные данные BFF: согласованный API этих сущностей не отдаёт">
        демо
      </span>
    )
  }
  if (source === 'bff') {
    return (
      <span className="chip chip-accent" title="Состояние, которое ведёт BFF в памяти">
        BFF
      </span>
    )
  }
  return null
}

export function Skeleton({ height = 16, width = '100%', radius = 8 }: { height?: number; width?: number | string; radius?: number }) {
  return <div className="skeleton" style={{ height, width, borderRadius: radius }} />
}

export function CardSkeleton() {
  return (
    <div className="card card-pad stack">
      <Skeleton height={14} width="40%" />
      <Skeleton height={22} width="85%" />
      <Skeleton height={14} width="70%" />
      <div className="row" style={{ gap: 8 }}>
        <Skeleton height={24} width={70} radius={999} />
        <Skeleton height={24} width={56} radius={999} />
      </div>
    </div>
  )
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string
  description?: string
  action?: ReactNode
}) {
  return (
    <div className="card card-pad stack" style={{ alignItems: 'center', textAlign: 'center', padding: 48 }}>
      <Icon name="spark" size={32} style={{ color: 'var(--accent-soft)' }} />
      <h3>{title}</h3>
      {description && <p className="muted" style={{ maxWidth: 460 }}>{description}</p>}
      {action}
    </div>
  )
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="card card-pad stack" style={{ alignItems: 'center', textAlign: 'center', padding: 40 }}>
      <Icon name="alert" size={28} style={{ color: 'var(--rose)' }} />
      <h3>Не удалось загрузить данные</h3>
      <p className="muted">{message}</p>
      {onRetry && (
        <button className="btn btn-sm" onClick={onRetry}>
          Повторить
        </button>
      )}
    </div>
  )
}

export function PendingData({ notice }: { notice?: string }) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 10,
        padding: '36px 16px',
        textAlign: 'center',
        color: 'var(--text-400)',
      }}
    >
      <Icon name="clock" size={26} style={{ color: 'var(--amber)' }} />
      <strong style={{ color: 'var(--text-200)' }}>Данные готовятся</strong>
      <span style={{ fontSize: 13.5, maxWidth: 320 }}>
        {notice ?? 'Витрина заполнится после завершения NLP-пайплайна'}
      </span>
    </div>
  )
}

export function Stat({ label, value, hint }: { label: string; value: ReactNode; hint?: string }) {
  return (
    <div className="stack" style={{ gap: 2 }}>
      <span className="dim" style={{ fontSize: 12.5, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
        {label}
      </span>
      <strong style={{ fontSize: 22, color: 'var(--text-100)', letterSpacing: '-0.02em' }}>{value}</strong>
      {hint && <span className="dim" style={{ fontSize: 12 }}>{hint}</span>}
    </div>
  )
}

export function MetricPill({ icon, value, label }: { icon: string; value: number; label: string }) {
  return (
    <span className="row" style={{ gap: 5 }} title={label}>
      <Icon name={icon} size={14} style={{ color: 'var(--text-400)' }} />
      <span className="mono" style={{ color: 'var(--text-300)' }}>
        {value.toLocaleString('ru-RU')}
      </span>
    </span>
  )
}
