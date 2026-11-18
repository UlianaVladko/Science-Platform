import { useEffect, useMemo, useRef, useState } from 'react'
import type { EChartsOption } from 'echarts'
import { api } from '../api/client'
import type { Discipline, Graph, Slice, TopItem, Widget } from '../api/types'
import { useDocumentTitle, useFetch, useIntersection } from '../hooks/useFetch'
import { ForceGraph } from '../components/charts/ForceGraph'
import { AXIS_STYLE, BASE_TEXT, Chart, PALETTE, TOOLTIP_STYLE } from '../components/charts/Chart'
import { ErrorState, PendingData, Skeleton, SourceBadge } from '../components/ui'

type TimelinePayload = { series: { year: number; count: number }[] }
type IntakePayload = { series: { date: string; count: number }[] }
type SlicesPayload = { slices: Slice[] }
type TopPayload = { items: TopItem[]; metric?: string }
type WordsPayload = { words: { text: string; weight: number }[] }
type LemmasPayload = { items: { lemma: string; count: number }[] }
type BucketsPayload = { buckets: { bucket: string; count: number }[] }
type AffiliationsPayload = { items: { name: string; count: number }[] }
type HeatmapPayload = { days: string[]; categories: string[]; values: [number, number, number][] }
type SummaryPayload = {
  total_publications: number
  total_authors: number
  total_citations: number
  total_downloads: number
  avg_citations_per_publication: number
  publications_this_period: number
  period: string
}

const PERIODS = ['2026', '2025', 'all'] as const
type Period = (typeof PERIODS)[number]

const WIDGET_TITLES: Record<string, string> = {
  'publications-timeline': 'Публикации по годам',
  'intake-timeline': 'Поступления в хранилище',
  'by-category': 'Распределение по направлениям',
  'top-authors': 'Самые продуктивные авторы',
  'keywords-cloud': 'Облако ключевых слов',
  'top-concepts': 'Ведущие понятия',
  'lemmas-frequency': 'Частотность лемм',
  'by-language': 'Языки публикаций',
  'by-type': 'Типы публикаций',
  'top-viewed': 'Самые просматриваемые',
  'top-cited': 'Самые цитируемые',
  'top-downloaded': 'Самые скачиваемые',
  'by-status': 'Статусы обработки',
  'by-source': 'Источники импорта',
  'coauthors-distribution': 'Размер авторского коллектива',
  'citation-graph': 'Граф цитирований',
  summary: 'Сводка по корпусу',
  'by-affiliation': 'Организации-лидеры',
  'activity-heatmap': 'Активность по направлениям',
  'citations-timeline': 'Накопленные цитирования',
  'top-publications': 'Топ публикаций',
}

const WIDGET_SIZE: Record<string, string> = {
  'publications-timeline': 'span-6',
  'intake-timeline': 'span-6',
  'by-category': 'span-4',
  'top-authors': 'span-4',
  'top-concepts': 'span-4',
  'keywords-cloud': 'span-6',
  'lemmas-frequency': 'span-6',
  'by-language': 'span-3',
  'by-type': 'span-3',
  'by-status': 'span-3',
  'by-source': 'span-3',
  'top-viewed': 'span-4',
  'top-cited': 'span-4',
  'top-downloaded': 'span-4',
  'coauthors-distribution': 'span-4',
  'by-affiliation': 'span-4',
  'activity-heatmap': 'span-4',
  'citations-timeline': 'span-6',
  'citation-graph': 'span-6',
  'top-publications': 'span-12',
}

const ORDER = [
  'summary',
  'publications-timeline',
  'citations-timeline',
  'by-category',
  'top-authors',
  'activity-heatmap',
  'intake-timeline',
  'by-language',
  'by-type',
  'by-status',
  'by-source',
  'coauthors-distribution',
  'by-affiliation',
  'keywords-cloud',
  'top-concepts',
  'lemmas-frequency',
  'top-cited',
  'top-viewed',
  'top-downloaded',
  'citation-graph',
  'top-publications',
]

export default function DashboardPage() {
  useDocumentTitle('Аналитика')
  const [period, setPeriod] = useState<Period>('all')
  const [category, setCategory] = useState('')
  const { data: disciplines } = useFetch<{ items: Discipline[] }>(() => api.disciplines(), [])

  return (
    <div className="page">
      <div className="section-head" style={{ paddingTop: 'clamp(24px, 3vw, 40px)' }}>
        <div>
          <h1 style={{ margin: 0, fontSize: 'clamp(24px, 2.6vw, 34px)' }}>Аналитика корпуса</h1>
          <p>
            Витрины поверх согласованного API и синтетического демо-корпуса. Виджеты грузятся по мере
            прокрутки — каждый помечен источником данных.
          </p>
        </div>
        <SourceBadge source="demo" />
      </div>

      <div className="filters-bar" style={{ marginBottom: 'var(--space-5)' }}>
        <span className="panel-title" style={{ margin: 0 }}>
          Период
        </span>
        <div className="chip" style={{ padding: 3, gap: 3 }}>
          {PERIODS.map((item) => (
            <button
              key={item}
              type="button"
              className="chip"
              aria-pressed={period === item}
              style={{ border: 'none', background: 'transparent' }}
              onClick={() => setPeriod(item)}
            >
              {item === 'all' ? 'Весь корпус' : item}
            </button>
          ))}
        </div>
        <div className="spacer" />
        <select
          className="input"
          style={{ maxWidth: 240 }}
          value={category}
          onChange={(event) => setCategory(event.target.value)}
          aria-label="Направление"
        >
          <option value="">Все направления</option>
          {(disciplines?.items ?? []).map((item) => (
            <option key={item.slug} value={item.slug}>
              {item.name}
            </option>
          ))}
        </select>
      </div>

      <div className="widget-grid">
        {ORDER.map((name) => (
          <DashboardWidget key={name} name={name} period={period} category={category} />
        ))}
      </div>
    </div>
  )
}

function DashboardWidget({ name, period, category }: { name: string; period: Period; category: string }) {
  const { ref, visible } = useIntersection<HTMLDivElement>()
  const [widget, setWidget] = useState<Widget | null>(null)
  const [error, setError] = useState<string | null>(null)

  const params = useMemo(
    () => ({
      category: category || undefined,
      period: period === 'all' ? undefined : period,
    }),
    [category, period],
  )

  useEffect(() => {
    if (!visible) return
    let cancelled = false
    setError(null)
    api
      .widget(name, params)
      .then((result) => {
        if (!cancelled) setWidget(result)
      })
      .catch((reason: Error) => {
        if (!cancelled) setError(reason.message)
      })
    return () => {
      cancelled = true
    }
  }, [visible, name, params])

  const title = WIDGET_TITLES[name] ?? name

  return (
    <section ref={ref} className={`widget ${WIDGET_SIZE[name] ?? 'span-4'}`} id={`widget-${name}`}>
      <div className="widget-head">
        <span className="widget-title">{title}</span>
        <div className="spacer" />
        {widget && <SourceBadge source={widget.source} />}
      </div>

      {!visible && <Skeleton height={name.startsWith('top-') || name === 'summary' ? 180 : 220} radius={12} />}
      {visible && !widget && !error && <Skeleton height={220} radius={12} />}
      {error && <ErrorState message={error} />}
      {widget && !widget.ready && <PendingData notice={widget.notice} />}
      {widget && widget.ready && <WidgetBody name={name} payload={widget.payload} />}

      {widget && (
        <span className="widget-sub">
          данные на {widget.data_as_of?.slice(0, 10)}
          {widget.source === 'demo' ? ' · демо-корпус BFF' : widget.source === 'bff' ? ' · расчёт BFF' : ''}
        </span>
      )}
    </section>
  )
}

function WidgetBody({ name, payload }: { name: string; payload: unknown }) {
  switch (name) {
    case 'summary':
      return <SummaryView payload={payload as SummaryPayload} />
    case 'publications-timeline':
      return <YearBars payload={payload as TimelinePayload} label="публикаций" />
    case 'citations-timeline':
      return <YearBars payload={payload as TimelinePayload} label="цитирований" cumulative color={PALETTE[1]} />
    case 'intake-timeline':
      return <IntakeView payload={payload as IntakePayload} />
    case 'by-category':
      return <SlicesView payload={payload as SlicesPayload} />
    case 'by-language':
    case 'by-type':
    case 'by-status':
    case 'by-source':
      return <SlicesDonut payload={payload as SlicesPayload} />
    case 'top-authors':
      return <TopListView payload={payload as TopPayload} link="author" />
    case 'top-concepts':
      return <TopListView payload={payload as TopPayload} />
    case 'top-cited':
    case 'top-viewed':
    case 'top-downloaded':
    case 'top-publications':
      return <TopListView payload={payload as TopPayload} link="publication" />
    case 'keywords-cloud':
      return <WordCloud payload={payload as WordsPayload} />
    case 'lemmas-frequency':
      return <LemmasView payload={payload as LemmasPayload} />
    case 'coauthors-distribution':
      return <CoauthorsView payload={payload as BucketsPayload} />
    case 'by-affiliation':
      return <AffiliationsView payload={payload as AffiliationsPayload} />
    case 'activity-heatmap':
      return <HeatmapView payload={payload as HeatmapPayload} />
    case 'citation-graph':
      return <GraphView payload={payload as Graph} />
    default:
      return <pre className="mono dim" style={{ fontSize: 11.5, overflowX: 'auto' }}>{JSON.stringify(payload, null, 2)}</pre>
  }
}

function SummaryView({ payload }: { payload: SummaryPayload }) {
  const items = [
    { label: 'публикаций', value: payload.total_publications },
    { label: 'авторов', value: payload.total_authors },
    { label: 'цитирований', value: payload.total_citations },
    { label: 'скачиваний', value: payload.total_downloads },
    { label: `за ${payload.period}`, value: payload.publications_this_period },
  ]
  return (
    <div className="kpi-grid" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))' }}>
      {items.map((item) => (
        <div className="kpi" key={item.label} style={{ ['--kpi-accent' as string]: 'var(--accent)' }}>
          <div className="kpi-value" style={{ fontSize: 24 }}>
            {item.value.toLocaleString('ru-RU')}
          </div>
          <div className="kpi-label">{item.label}</div>
        </div>
      ))}
    </div>
  )
}

function YearBars({
  payload,
  label,
  cumulative = false,
  color = PALETTE[0],
}: {
  payload: TimelinePayload
  label: string
  cumulative?: boolean
  color?: string
}) {
  const option = useMemo<EChartsOption | null>(() => {
    const series = payload.series ?? []
    if (series.length === 0) return null
    return {
      grid: { left: 6, right: 10, top: 18, bottom: 6, containLabel: true },
      tooltip: { trigger: 'axis', ...TOOLTIP_STYLE, valueFormatter: (value) => `${value} ${label}` },
      xAxis: { type: 'category', data: series.map((item) => String(item.year)), ...AXIS_STYLE },
      yAxis: { type: 'value', ...AXIS_STYLE },
      series: [
        cumulative
          ? {
              type: 'line',
              smooth: true,
              showSymbol: false,
              areaStyle: {
                color: {
                  type: 'linear',
                  x: 0,
                  y: 0,
                  x2: 0,
                  y2: 1,
                  colorStops: [
                    { offset: 0, color: color + '77' },
                    { offset: 1, color: color + '05' },
                  ],
                },
              },
              lineStyle: { color, width: 2 },
              itemStyle: { color },
              data: series.map((item) => item.count),
            }
          : {
              type: 'bar',
              data: series.map((item) => item.count),
              itemStyle: { color, borderRadius: [5, 5, 0, 0] },
              barMaxWidth: 26,
            },
      ],
    }
  }, [payload, label, cumulative, color])

  if (!option) return <PendingData notice="Нет данных за выбранный период" />
  return <Chart option={option} height={220} ariaLabel={`${label} по годам`} />
}

function IntakeView({ payload }: { payload: IntakePayload }) {
  const option = useMemo<EChartsOption | null>(() => {
    const series = payload.series ?? []
    if (series.length === 0) return null
    return {
      grid: { left: 6, right: 10, top: 18, bottom: 6, containLabel: true },
      tooltip: { trigger: 'axis', ...TOOLTIP_STYLE },
      xAxis: {
        type: 'category',
        data: series.map((item) => item.date),
        ...AXIS_STYLE,
        axisLabel: { ...AXIS_STYLE.axisLabel, formatter: (value: string) => value.slice(2) },
      },
      yAxis: { type: 'value', ...AXIS_STYLE },
      series: [
        {
          type: 'bar',
          data: series.map((item) => item.count),
          itemStyle: { color: PALETTE[2], borderRadius: [4, 4, 0, 0] },
          barMaxWidth: 20,
        },
      ],
    }
  }, [payload])

  if (!option) return <PendingData notice="Нет данных о поступлениях" />
  return <Chart option={option} height={220} ariaLabel="Поступления по месяцам" />
}

function SlicesView({ payload }: { payload: SlicesPayload }) {
  const option = useMemo<EChartsOption | null>(() => {
    const slices = (payload.slices ?? []).slice(0, 10)
    if (slices.length === 0) return null
    return {
      grid: { left: 6, right: 30, top: 8, bottom: 6, containLabel: true },
      tooltip: { trigger: 'item', ...TOOLTIP_STYLE },
      xAxis: { type: 'value', ...AXIS_STYLE },
      yAxis: {
        type: 'category',
        inverse: true,
        data: slices.map((item) => item.name),
        ...AXIS_STYLE,
        axisLabel: { ...AXIS_STYLE.axisLabel, width: 130, overflow: 'truncate' },
      },
      series: [
        {
          type: 'bar',
          data: slices.map((item, index) => ({
            value: item.count,
            itemStyle: { color: item.color ?? PALETTE[index % PALETTE.length], borderRadius: [0, 4, 4, 0] },
          })),
          barMaxWidth: 16,
          label: { show: true, position: 'right', color: '#6b7280', fontSize: 11 },
        },
      ],
    }
  }, [payload])

  if (!option) return <PendingData notice="Нет данных по направлениям" />
  return <Chart option={option} height={260} ariaLabel="Распределение по направлениям" />
}

function SlicesDonut({ payload }: { payload: SlicesPayload }) {
  const option = useMemo<EChartsOption | null>(() => {
    const slices = payload.slices ?? []
    if (slices.length === 0) return null
    return {
      tooltip: { trigger: 'item', ...TOOLTIP_STYLE, valueFormatter: (value) => `${value} шт.` },
      legend: { ...BASE_TEXT, bottom: 0, itemWidth: 9, itemHeight: 9, textStyle: { ...BASE_TEXT, fontSize: 11 } },
      series: [
        {
          type: 'pie',
          radius: ['52%', '76%'],
          center: ['50%', '44%'],
          avoidLabelOverlap: true,
          itemStyle: { borderColor: '#ffffff', borderWidth: 2 },
          label: { show: false },
          data: slices.map((item, index) => ({
            name: item.name,
            value: item.count,
            itemStyle: { color: item.color ?? PALETTE[index % PALETTE.length] },
          })),
        },
      ],
    }
  }, [payload])

  if (!option) return <PendingData notice="Нет данных" />
  return <Chart option={option} height={230} ariaLabel="Доля по категориям" />
}

function TopListView({ payload, link }: { payload: TopPayload; link?: 'publication' | 'author' }) {
  const items = payload.items ?? []
  if (items.length === 0) return <PendingData notice="Нет данных за выбранный период" />
  const max = Math.max(...items.map((item) => item.count || 1))
  return (
    <ol className="list-plain" style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      {items.slice(0, 8).map((item, index) => {
        const href =
          link === 'publication'
            ? `/publications/${encodeURIComponent(item.id)}`
            : link === 'author'
              ? `/authors/${encodeURIComponent(item.id)}`
              : undefined
        const body = (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4, width: '100%' }}>
            <div className="row" style={{ gap: 8 }}>
              <span className="mono dim" style={{ minWidth: 18, fontSize: 12 }}>
                {String(index + 1).padStart(2, '0')}
              </span>
              <span style={{ flex: 1, fontSize: 13.5, lineHeight: 1.35 }}>{item.label}</span>
              <span className="mono" style={{ fontSize: 12.5, color: 'var(--accent-soft)' }}>
                {item.count.toLocaleString('ru-RU')}
              </span>
            </div>
            <div
              style={{
                height: 3,
                borderRadius: 2,
                background: 'rgba(255,255,255,0.06)',
                marginLeft: 26,
                overflow: 'hidden',
              }}
            >
              <div
                style={{
                  width: `${Math.round((item.count / max) * 100)}%`,
                  height: '100%',
                  borderRadius: 2,
                  background: `linear-gradient(90deg, ${PALETTE[0]}, ${PALETTE[1]})`,
                }}
              />
            </div>
          </div>
        )
        return (
          <li key={`${item.id}-${index}`}>
            {href ? (
              <a href={href} style={{ display: 'block' }}>
                {body}
              </a>
            ) : (
              body
            )}
          </li>
        )
      })}
    </ol>
  )
}

function WordCloud({ payload }: { payload: WordsPayload }) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const [width, setWidth] = useState(520)

  useEffect(() => {
    const node = containerRef.current
    if (!node) return
    const observer = new ResizeObserver(() => setWidth(node.clientWidth || 520))
    observer.observe(node)
    return () => observer.disconnect()
  }, [])

  const words = (payload.words ?? []).slice(0, 42)
  const height = 240
  const placed = useMemo(() => layoutCloud(words, width, height), [words, width])

  if (words.length === 0) return <PendingData notice="Индексация ключевых слов ещё не завершена" />

  return (
    <div ref={containerRef} style={{ width: '100%' }}>
      <svg width={width} height={height} role="img" aria-label="Облако ключевых слов и понятий">
        {placed.map((word) => (
          <text
            key={word.text}
            x={word.x}
            y={word.y}
            fontSize={word.size}
            fill={PALETTE[word.colorIndex % PALETTE.length]}
            fillOpacity={0.55 + Math.min(0.45, word.size / 44)}
            textAnchor="middle"
            style={{ fontFamily: 'Inter, system-ui, sans-serif', fontWeight: 600 }}
          >
            {word.text}
          </text>
        ))}
      </svg>
    </div>
  )
}

function layoutCloud(words: { text: string; weight: number }[], width: number, height: number) {
  const maxWeight = Math.max(...words.map((word) => word.weight), 1)
  const minWeight = Math.min(...words.map((word) => word.weight))
  const span = Math.max(1, maxWeight - minWeight)
  const cx = width / 2
  const cy = height / 2
  return words.map((word, index) => {
    const ratio = (word.weight - minWeight) / span
    const size = 11 + ratio * 22
    const angle = index * 2.399963
    const radius = Math.sqrt(index / words.length) * Math.min(width, height) * 0.46
    return {
      text: word.text,
      size: Math.round(size * 10) / 10,
      x: cx + Math.cos(angle) * radius * 1.35,
      y: cy + Math.sin(angle) * radius * 0.85,
      colorIndex: index,
    }
  })
}

function LemmasView({ payload }: { payload: LemmasPayload }) {
  const items = payload.items ?? []
  const option = useMemo<EChartsOption | null>(() => {
    if (items.length === 0) return null
    return {
      grid: { left: 6, right: 12, top: 18, bottom: 6, containLabel: true },
      tooltip: { trigger: 'axis', ...TOOLTIP_STYLE },
      xAxis: {
        type: 'category',
        data: items.map((item) => item.lemma),
        ...AXIS_STYLE,
        axisLabel: { ...AXIS_STYLE.axisLabel, rotate: 40, fontSize: 10 },
      },
      yAxis: { type: 'value', ...AXIS_STYLE },
      series: [
        {
          type: 'bar',
          data: items.map((item) => item.count),
          itemStyle: { color: PALETTE[3], borderRadius: [4, 4, 0, 0] },
          barMaxWidth: 22,
        },
      ],
    }
  }, [items])

  if (!option) return <PendingData notice="Морфологический анализ ещё не подключён" />
  return <Chart option={option} height={230} ariaLabel="Частотность лемм" />
}

function CoauthorsView({ payload }: { payload: BucketsPayload }) {
  const option = useMemo<EChartsOption | null>(() => {
    const buckets = payload.buckets ?? []
    if (buckets.length === 0) return null
    return {
      grid: { left: 6, right: 10, top: 18, bottom: 6, containLabel: true },
      tooltip: { trigger: 'axis', ...TOOLTIP_STYLE },
      xAxis: {
        type: 'category',
        data: buckets.map((item) => item.bucket),
        ...AXIS_STYLE,
        axisLabel: { ...AXIS_STYLE.axisLabel, formatter: (value: string) => `${value} авт.` },
      },
      yAxis: { type: 'value', ...AXIS_STYLE },
      series: [
        {
          type: 'bar',
          data: buckets.map((item) => item.count),
          itemStyle: { color: PALETTE[4], borderRadius: [4, 4, 0, 0] },
          barMaxWidth: 24,
        },
      ],
    }
  }, [payload])

  if (!option) return <PendingData notice="Нет данных" />
  return <Chart option={option} height={210} ariaLabel="Распределение по числу соавторов" />
}

function AffiliationsView({ payload }: { payload: AffiliationsPayload }) {
  const items = (payload.items ?? []).slice(0, 9)
  const option = useMemo<EChartsOption | null>(() => {
    if (items.length === 0) return null
    return {
      grid: { left: 6, right: 12, top: 18, bottom: 6, containLabel: true },
      tooltip: { trigger: 'axis', ...TOOLTIP_STYLE },
      xAxis: {
        type: 'category',
        data: items.map((item) => item.name),
        ...AXIS_STYLE,
        axisLabel: { ...AXIS_STYLE.axisLabel, rotate: 32, fontSize: 10, width: 90, overflow: 'truncate' },
      },
      yAxis: { type: 'value', ...AXIS_STYLE },
      series: [
        {
          type: 'bar',
          data: items.map((item) => item.count),
          itemStyle: { color: PALETTE[5], borderRadius: [4, 4, 0, 0] },
          barMaxWidth: 22,
        },
      ],
    }
  }, [items])

  if (!option) return <PendingData notice="Нет данных об организациях" />
  return <Chart option={option} height={230} ariaLabel="Организации-лидеры по числу публикаций" />
}

function HeatmapView({ payload }: { payload: HeatmapPayload }) {
  const option = useMemo<EChartsOption | null>(() => {
    const days = payload.days ?? []
    const categories = payload.categories ?? []
    const values = payload.values ?? []
    if (days.length === 0 || categories.length === 0) return null
    return {
      grid: { left: 6, right: 12, top: 10, bottom: 40, containLabel: true },
      tooltip: {
        ...TOOLTIP_STYLE,
        formatter: (params: unknown) => {
          const item = params as { value?: [number, number, number] }
          const value = item.value
          if (!value) return ''
          const [d, c, v] = value
          return `${days[d]}<br/>${categories[c]}: <b>${v}</b>`
        },
      },
      xAxis: {
        type: 'category',
        data: days.map((day) => day.slice(5)),
        ...AXIS_STYLE,
        splitArea: { show: false },
        axisLabel: { ...AXIS_STYLE.axisLabel, interval: 4, fontSize: 9.5 },
      },
      yAxis: {
        type: 'category',
        data: categories,
        ...AXIS_STYLE,
        axisLabel: { ...AXIS_STYLE.axisLabel, fontSize: 10, width: 92, overflow: 'truncate' },
      },
      visualMap: {
        min: 0,
        max: Math.max(...values.map((item) => item[2]), 1),
        calculable: false,
        orient: 'horizontal',
        left: 'center',
        bottom: 0,
        itemWidth: 10,
        itemHeight: 80,
        textStyle: { ...BASE_TEXT, fontSize: 10 },
        inRange: { color: ['#eff6ff', '#93c5fd', '#3b82f6', '#1d4ed8'] },
      },
      series: [
        {
          type: 'heatmap',
          data: values,
          itemStyle: { borderRadius: 2, borderColor: '#ffffff', borderWidth: 1 },
          emphasis: { itemStyle: { borderColor: '#fff', borderWidth: 1 } },
        },
      ],
    }
  }, [payload])

  if (!option) return <PendingData notice="Нет данных об активности" />
  return <Chart option={option} height={250} ariaLabel="Тепловая карта активности по направлениям" />
}

function GraphView({ payload }: { payload: Graph }) {
  if (!payload || payload.nodes.length === 0) {
    return <PendingData notice="Для выбранного направления граф цитирований пуст" />
  }
  return <ForceGraph graph={payload} height={360} />
}
