import { useEffect, useMemo, useRef, useState } from 'react'
import {
  forceCenter,
  forceCollide,
  forceLink,
  forceManyBody,
  forceSimulation,
  type Simulation,
  type SimulationLinkDatum,
  type SimulationNodeDatum,
} from 'd3-force'
import { Link } from 'react-router-dom'
import type { Graph } from '../../api/types'
import { PALETTE } from './Chart'

type NodeDatum = SimulationNodeDatum & {
  id: string
  label: string
  type: string
  category: string
  radius: number
  color: string
}

type LinkDatum = SimulationLinkDatum<NodeDatum> & { type: string; weight: number; color: string }

const COLORS = PALETTE.slice(0, 6)

function colorFor(category: string, index: number) {
  if (!category) return COLORS[index % COLORS.length]
  let hash = 0
  for (const char of category) hash = (hash * 31 + char.codePointAt(0)!) % 9973
  return COLORS[hash % COLORS.length]
}

export function ForceGraph({
  graph,
  height = 420,
  focusId,
}: {
  graph: Graph
  height?: number
  focusId?: string
}) {
  const [size, setSize] = useState({ width: 720, height })
  const [hovered, setHovered] = useState<NodeDatum | null>(null)
  const [tick, setTick] = useState(0)
  const wrapRef = useRef<HTMLDivElement | null>(null)
  const simulationRef = useRef<Simulation<NodeDatum, LinkDatum> | null>(null)
  const nodesRef = useRef<NodeDatum[]>([])
  const linksRef = useRef<LinkDatum[]>([])

  useEffect(() => {
    const node = wrapRef.current
    if (!node) return
    const observer = new ResizeObserver(() => {
      const width = node.clientWidth || 720
      setSize({ width, height })
    })
    observer.observe(node)
    return () => observer.disconnect()
  }, [height])

  const nodes = useMemo<NodeDatum[]>(
    () =>
      graph.nodes.map((item, index) => ({
        id: item.id,
        label: item.label,
        type: item.type,
        category: item.category ?? '',
        radius: Math.min(26, 7 + Math.sqrt(item.weight ?? 1) * 3.2),
        color: colorFor(item.category ?? '', index),
      })),
    [graph],
  )

  useEffect(() => {
    const byId = new Map(nodes.map((item) => [item.id, item]))
    const links: LinkDatum[] = []
    for (const edge of graph.edges) {
      const source = byId.get(edge.source)
      const target = byId.get(edge.target)
      if (!source || !target) continue
      links.push({
        source,
        target,
        type: edge.type,
        weight: edge.weight ?? 0,
        color: colorFor(edge.type, links.length),
      })
    }

    nodesRef.current = nodes
    linksRef.current = links

    const simulation = forceSimulation<NodeDatum, LinkDatum>(nodes)
      .force('charge', forceManyBody().strength(-170))
      .force('center', forceCenter(size.width / 2, size.height / 2))
      .force('collide', forceCollide<NodeDatum>().radius((item) => item.radius + 4))
      .force(
        'link',
        forceLink<NodeDatum, LinkDatum>(links)
          .id((item) => item.id)
          .distance(80)
          .strength(0.6),
      )
      .alphaDecay(0.055)
      .on('tick', () => setTick((value) => (value + 1) % 100000))

    simulationRef.current = simulation
    return () => {
      simulation.stop()
      simulationRef.current = null
    }
  }, [graph, nodes, size.width, size.height])

  useEffect(() => {
    simulationRef.current?.force('center', forceCenter(size.width / 2, size.height / 2))
    simulationRef.current?.alpha(0.35).restart()
  }, [size.width, size.height])

  const focusNode = hovered ?? nodesRef.current.find((item) => item.id === focusId) ?? null
  const neighbourIds = useMemo(() => {
    if (!focusNode) return new Set<string>()
    const ids = new Set<string>([focusNode.id])
    linksRef.current.forEach((link) => {
      const source = typeof link.source === 'string' ? link.source : (link.source as NodeDatum).id
      const target = typeof link.target === 'string' ? link.target : (link.target as NodeDatum).id
      if (source === focusNode.id) ids.add(target)
      if (target === focusNode.id) ids.add(source)
    })
    return ids
  }, [focusNode, tick])

  if (graph.nodes.length === 0) return null

  return (
    <div className="graph-wrap">
      <div ref={wrapRef} className="graph-canvas" style={{ height }}>
        <svg width={size.width} height={size.height} role="img" aria-label={`Граф: ${graph.nodes.length} узлов, ${graph.edges.length} связей`}>
          <g>
            {linksRef.current.map((link, index) => {
              const source = link.source as NodeDatum
              const target = link.target as NodeDatum
              if (source.x == null || target.x == null) return null
              const active =
                focusNode && (source.id === focusNode.id || target.id === focusNode.id)
              return (
                <line
                  key={index}
                  x1={source.x}
                  y1={source.y}
                  x2={target.x}
                  y2={target.y}
                  stroke={active ? (focusNode?.color ?? link.color) : link.color}
                  strokeOpacity={active ? 0.92 : 0.42}
                  strokeWidth={active ? 2.6 : Math.min(1.8, 0.8 + link.weight * 0.08)}
                />
              )
            })}
            {nodesRef.current.map((node) => {
              const dim = focusNode ? !neighbourIds.has(node.id) : false
              return (
                <g
                  key={node.id}
                  transform={`translate(${node.x ?? 0},${node.y ?? 0})`}
                  onMouseEnter={() => setHovered(node)}
                  onMouseLeave={() => setHovered(null)}
                  style={{ cursor: 'pointer', opacity: dim ? 0.3 : 1 }}
                >
                  <circle r={node.radius} fill={node.color} fillOpacity={0.9} stroke="#ffffff" strokeWidth={1.5} />
                  {(node.radius > 12 || node.id === focusId) && (
                    <text
                      y={node.radius + 13}
                      textAnchor="middle"
                      fontSize={10.5}
                      fill="#4b5563"
                      style={{ pointerEvents: 'none' }}
                    >
                      {node.label.length > 22 ? `${node.label.slice(0, 21)}…` : node.label}
                    </text>
                  )}
                </g>
              )
            })}
          </g>
        </svg>
      </div>

      <div className="graph-legend">
        {hovered ? (
          <div className="row" style={{ gap: 8 }}>
            <span className="chip chip-accent">{hovered.type}</span>
            {hovered.type === 'author' ? (
              <Link to={`/authors/${encodeURIComponent(hovered.id)}`}>{hovered.label}</Link>
            ) : hovered.type === 'publication' ? (
              <Link to={`/publications/${encodeURIComponent(hovered.id)}`}>{hovered.label}</Link>
            ) : (
              <span>{hovered.label}</span>
            )}
            {hovered.category && <span className="dim">· {hovered.category}</span>}
          </div>
        ) : (
          <span className="dim" style={{ fontSize: 12.5 }}>
            {graph.nodes.length} узлов · {graph.edges.length} связей
            {graph.truncated ? ' · граф усечён по лимиту глубины' : ''}
            {' · '}наведите на узел, чтобы подсветить окружение
          </span>
        )}
      </div>
    </div>
  )
}
