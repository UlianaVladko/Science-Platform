import { useEffect, useRef } from 'react'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

export const PALETTE = [
  '#374151', '#64748b', '#9ca3af', '#4b5563',
  '#6b7280', '#a1a1aa', '#52525b', '#d1d5db',
  '#71717a', '#9ca3af', '#475569', '#cbd5e1',
]

export const AXIS_STYLE = {
  axisLine: { lineStyle: { color: 'rgba(31,41,55,0.16)' } },
  axisTick: { show: false },
  axisLabel: { color: '#6b7280', fontSize: 11 },
  splitLine: { lineStyle: { color: 'rgba(31,41,55,0.08)' } },
}

export const TOOLTIP_STYLE = {
  backgroundColor: '#ffffff',
  borderColor: 'rgba(31,41,55,0.16)',
  borderWidth: 1,
  textStyle: { color: '#374151', fontSize: 12 },
  extraCssText: 'border-radius:10px;box-shadow:0 10px 30px -12px rgba(15,23,42,.12);',
}

export const BASE_TEXT = { color: '#4b5563', fontFamily: 'Inter, system-ui, sans-serif' }

export function useChart(option: EChartsOption | null, height = 260) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const chartRef = useRef<echarts.ECharts | null>(null)

  useEffect(() => {
    const node = containerRef.current
    if (!node) return
    const chart = echarts.init(node, undefined, { renderer: 'canvas' })
    chartRef.current = chart
    const observer = new ResizeObserver(() => chart.resize())
    observer.observe(node)
    return () => {
      observer.disconnect()
      chart.dispose()
      chartRef.current = null
    }
  }, [])

  useEffect(() => {
    if (chartRef.current && option) {
      chartRef.current.setOption(option, true)
    }
  }, [option])

  return { containerRef, height }
}

export function Chart({ option, height = 260, ariaLabel }: { option: EChartsOption | null; height?: number; ariaLabel?: string }) {
  const { containerRef } = useChart(option)
  return (
    <div
      ref={containerRef}
      className="chart"
      style={{ height }}
      role="img"
      aria-label={ariaLabel ?? 'График аналитики'}
    />
  )
}
