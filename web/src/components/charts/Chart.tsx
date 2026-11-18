import { useEffect, useRef } from 'react'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

// Distinct, restrained colors are reserved for data visualization; the app chrome stays neutral.
export const PALETTE = [
  '#2563eb', // blue
  '#0f9f8f', // teal
  '#e09a18', // amber
  '#8055d8', // violet
  '#e34d72', // rose
  '#079fca', // cyan
  '#70a92e', // lime
  '#e76b2f', // orange
  '#5356d8', // indigo
  '#159b68', // emerald
  '#c83d91', // pink
  '#64748b', // slate
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
    // SVG rendering keeps chart colors crisp and consistent across browsers and DPR settings.
    const chart = echarts.init(node, undefined, { renderer: 'svg' })
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
      chartRef.current.setOption(
        {
          // Explicit global palette is a fallback for every series that does not define itemStyle.
          color: PALETTE,
          animation: true,
          animationDuration: 450,
          animationEasing: 'cubicOut',
          ...option,
        },
        { notMerge: true, lazyUpdate: false },
      )
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
