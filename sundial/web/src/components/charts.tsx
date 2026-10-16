import { scaleBand, scaleLinear } from 'd3-scale'
import { area, line, curveMonotoneX } from 'd3-shape'
import { useEffect, useMemo, useRef, useState, type ReactNode, type RefObject } from 'react'
import { compact, usd } from '../lib/format'

/** Categorical series in fixed order (validated palette slots 1–5); colours live in CSS vars. */
export const SERIES = ['var(--s1)', 'var(--s2)', 'var(--s3)', 'var(--s4)', 'var(--s5)']

export function Legend({ items }: { items: { label: string; color: string; dashed?: boolean }[] }) {
  return (
    <ul className="legend">
      {items.map((i) => (
        <li key={i.label}>
          <i style={i.dashed ? { background: 'none', borderTop: `2px dashed ${i.color}`, height: 0, width: 16 } : { background: i.color }} />
          {i.label}
        </li>
      ))}
    </ul>
  )
}

/** Render charts at their real pixel width so text stays readable at every size. */
function useWidth(ref: RefObject<HTMLDivElement>, fallback = 760) {
  const [w, setW] = useState(fallback)
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const ro = new ResizeObserver(([e]) => setW(Math.max(280, Math.round(e.contentRect.width))))
    ro.observe(el)
    return () => ro.disconnect()
  }, [ref])
  return w
}

function useHover<T>() {
  const ref = useRef<HTMLDivElement>(null)
  const [hover, setHover] = useState<{ x: number; y: number; datum: T } | null>(null)
  return { ref, hover, setHover }
}

function Tip({ x, y, children, width }: { x: number; y: number; children: ReactNode; width: number }) {
  const left = Math.min(Math.max(8, x + 14), width - 230)
  return <div className="tip" style={{ left, top: Math.max(0, y - 10) }} role="status">{children}</div>
}

export interface AreaSeries { key: string; label: string; color: string; values: number[] }

/** Stacked area of income by source, with a dashed spending line and a crosshair tooltip. */
export function IncomeArea({ years, series, spending, height = 300 }:
  { years: number[]; series: AreaSeries[]; spending?: number[]; height?: number }) {
  const { ref, hover, setHover } = useHover<number>()
  const width = useWidth(ref), m = { t: 12, r: 12, b: 28, l: 56 }
  const stacks = useMemo(() => {
    const acc = years.map(() => 0)
    return series.map((s) => {
      const lo = acc.slice()
      s.values.forEach((v, i) => { acc[i] += Math.max(0, v) })
      return { ...s, lo, hi: acc.slice() }
    })
  }, [years, series])
  const max = Math.max(1, ...stacks.flatMap((s) => s.hi), ...(spending ?? []))
  const x = scaleLinear().domain([years[0], years[years.length - 1]]).range([m.l, width - m.r])
  const y = scaleLinear().domain([0, max * 1.08]).nice().range([height - m.b, m.t])
  const ticks = y.ticks(5)
  const every = width < 520 ? 10 : 5
  const xt = years.filter((yr) => yr % every === 0)
  const onMove = (e: React.MouseEvent<SVGSVGElement>) => {
    const box = e.currentTarget.getBoundingClientRect()
    const px = ((e.clientX - box.left) / box.width) * width
    const i = Math.max(0, Math.min(years.length - 1, Math.round(x.invert(px)) - years[0]))
    setHover({ x: e.clientX - box.left, y: e.clientY - box.top, datum: i })
  }
  return (
    <div className="chart-wrap" ref={ref}>
      <svg viewBox={`0 0 ${width} ${height}`} width={width} height={height} className="chart" role="img" aria-label="Income by source over time"
        onMouseMove={onMove} onMouseLeave={() => setHover(null)}>
        {ticks.map((t) => (
          <g key={t}><line className="grid" x1={m.l} x2={width - m.r} y1={y(t)} y2={y(t)} />
            <text className="axis" x={m.l - 8} y={y(t) + 4} textAnchor="end">{compact(t)}</text></g>
        ))}
        {xt.map((t) => <text key={t} className="axis" x={x(t)} y={height - 8} textAnchor="middle">{t}</text>)}
        {stacks.map((s) => {
          const gen = area<number>().x((_, i) => x(years[i])).y0((_, i) => y(s.lo[i])).y1((_, i) => y(s.hi[i])).curve(curveMonotoneX)
          return <path key={s.key} d={gen(years) ?? ''} fill={s.color} stroke="var(--surface)" strokeWidth={1.5} />
        })}
        {spending && (
          <path d={line<number>().x((_, i) => x(years[i])).y((v) => y(v)).curve(curveMonotoneX)(spending) ?? ''}
            fill="none" stroke="var(--ink)" strokeWidth={2} strokeDasharray="5 4" />
        )}
        {hover && <line className="crosshair" x1={x(years[hover.datum])} x2={x(years[hover.datum])} y1={m.t} y2={height - m.b} />}
      </svg>
      {hover && (
        <Tip x={hover.x} y={hover.y} width={ref.current?.clientWidth ?? 600}>
          <strong>{years[hover.datum]}</strong>
          {stacks.map((s) => s.values[hover.datum] > 0 && (
            <div key={s.key} className="tip-row"><i style={{ background: s.color }} />{s.label}<b>{usd(s.values[hover.datum])}</b></div>
          ))}
          {spending && <div className="tip-row"><i className="dash" />Spending<b>{usd(spending[hover.datum])}</b></div>}
        </Tip>
      )}
    </div>
  )
}

/** Monte Carlo fan: 10th–90th percentile band and the median path. */
export function FanChart({ years, p10, p50, p90, height = 280 }: { years: number[]; p10: number[]; p50: number[]; p90: number[]; height?: number }) {
  const { ref, hover, setHover } = useHover<number>()
  const width = useWidth(ref), m = { t: 12, r: 12, b: 28, l: 64 }
  const x = scaleLinear().domain([years[0], years[years.length - 1]]).range([m.l, width - m.r])
  const y = scaleLinear().domain([0, Math.max(1, ...p90) * 1.05]).nice().range([height - m.b, m.t])
  const band = area<number>().x((_, i) => x(years[i])).y0((_, i) => y(p10[i])).y1((_, i) => y(p90[i])).curve(curveMonotoneX)
  const mid = line<number>().x((_, i) => x(years[i])).y((v) => y(v)).curve(curveMonotoneX)
  return (
    <div className="chart-wrap" ref={ref}>
      <svg viewBox={`0 0 ${width} ${height}`} width={width} height={height} className="chart" role="img" aria-label="Range of portfolio outcomes"
        onMouseMove={(e) => {
          const box = e.currentTarget.getBoundingClientRect()
          const i = Math.max(0, Math.min(years.length - 1, Math.round(x.invert(((e.clientX - box.left) / box.width) * width)) - years[0]))
          setHover({ x: e.clientX - box.left, y: e.clientY - box.top, datum: i })
        }} onMouseLeave={() => setHover(null)}>
        {y.ticks(5).map((t) => (
          <g key={t}><line className="grid" x1={m.l} x2={width - m.r} y1={y(t)} y2={y(t)} />
            <text className="axis" x={m.l - 8} y={y(t) + 4} textAnchor="end">{compact(t)}</text></g>
        ))}
        {years.filter((yr) => yr % (width < 520 ? 10 : 5) === 0).map((t) => <text key={t} className="axis" x={x(t)} y={height - 8} textAnchor="middle">{t}</text>)}
        <path d={band(years) ?? ''} fill="var(--s1)" opacity={0.22} />
        <path d={mid(p50) ?? ''} fill="none" stroke="var(--s1)" strokeWidth={2.5} />
        {hover && <line className="crosshair" x1={x(years[hover.datum])} x2={x(years[hover.datum])} y1={m.t} y2={height - m.b} />}
      </svg>
      {hover && (
        <Tip x={hover.x} y={hover.y} width={ref.current?.clientWidth ?? 600}>
          <strong>{years[hover.datum]}</strong>
          <div className="tip-row">Good markets (90th)<b>{usd(p90[hover.datum])}</b></div>
          <div className="tip-row">Typical (median)<b>{usd(p50[hover.datum])}</b></div>
          <div className="tip-row">Poor markets (10th)<b>{usd(p10[hover.datum])}</b></div>
        </Tip>
      )}
    </div>
  )
}

/** Horizontal bars for one measure across categories (no legend needed). */
export function HBars({ items, format = compact, highlight }: { items: { label: string; value: number }[]; format?: (v: number) => string; highlight?: number }) {
  const ref = useRef<HTMLDivElement>(null)
  const width = useWidth(ref), rowH = 40, barH = 22, labelW = Math.min(300, Math.round(width * 0.45)), right = 84
  const height = items.length * rowH + 4
  const max = Math.max(1, ...items.map((i) => i.value))
  const x = scaleLinear().domain([0, max]).range([0, width - labelW - right])
  const yb = scaleBand<number>().domain(items.map((_, i) => i)).range([0, items.length * rowH])
  const [hover, setHover] = useState<number | null>(null)
  return (
    <div ref={ref}><svg viewBox={`0 0 ${width} ${height}`} width={width} height={height} className="chart" role="img" aria-label="Comparison">
      {items.map((it, i) => {
        const y0 = (yb(i) ?? 0) + (rowH - barH) / 2
        const w = Math.max(2, x(Math.max(0, it.value)))
        const r = Math.min(4, w / 2)
        return (
          <g key={it.label} onMouseEnter={() => setHover(i)} onMouseLeave={() => setHover(null)}>
            <rect x={0} y={yb(i)} width={width} height={rowH} fill={hover === i ? 'var(--hover)' : 'transparent'} />
            <text className={`rowlabel${highlight === i ? ' strong' : ''}`} x={labelW - 12} y={y0 + barH / 2 + 5} textAnchor="end">{it.label}</text>
            <path d={`M${labelW},${y0} H${labelW + w - r} Q${labelW + w},${y0} ${labelW + w},${y0 + r} V${y0 + barH - r} Q${labelW + w},${y0 + barH} ${labelW + w - r},${y0 + barH} H${labelW} Z`}
              fill={highlight === i ? 'var(--s1)' : 'var(--s1-soft)'}><title>{`${it.label}: ${format(it.value)}`}</title></path>
            <text className="value" x={labelW + w + 8} y={y0 + barH / 2 + 5}>{format(it.value)}</text>
          </g>
        )
      })}
    </svg></div>
  )
}

/** Heat map of expected lifetime benefits by claiming age (sequential, one hue). */
export function ClaimHeat({ grid, start, best, step }: { grid: number[][]; start: number[]; best: number[]; step: number }) {
  const rows = grid.length, cols = grid[0]?.length ?? 0
  const cell = 18, left = 56, top = 34
  const values = grid.flat()
  const lo = Math.min(...values), hi = Math.max(...values)
  const shade = scaleLinear().domain([lo, hi]).range([0.08, 1])
  const [hover, setHover] = useState<[number, number] | null>(null)
  const ageLabel = (m: number) => `${Math.floor(m / 12)}${m % 12 ? `·${m % 12}` : ''}`
  return (
    <div className="heat">
      <svg viewBox={`0 0 ${left + cols * cell + 10} ${top + rows * cell + 10}`} className="chart" style={{ maxWidth: 680 }} role="img" aria-label="Expected lifetime benefits by claiming ages">
        {Array.from({ length: cols }, (_, j) => j % 4 === 0 && (
          <text key={j} className="axis" x={left + j * cell + cell / 2} y={top - 8} textAnchor="middle">{ageLabel(start[1] + j * step)}</text>))}
        {Array.from({ length: rows }, (_, i) => i % 4 === 0 && (
          <text key={i} className="axis" x={left - 8} y={top + i * cell + cell / 2 + 4} textAnchor="end">{ageLabel(start[0] + i * step)}</text>))}
        {grid.map((row, i) => row.map((v, j) => {
          const isBest = Math.abs(start[0] + i * step - best[0]) < step && Math.abs(start[1] + j * step - best[1]) < step
          return (
            <rect key={`${i}-${j}`} x={left + j * cell + 1} y={top + i * cell + 1} width={cell - 2} height={cell - 2} rx={3}
              fill="var(--seq-9)" fillOpacity={shade(v)} stroke={isBest ? 'var(--ink)' : 'none'} strokeWidth={isBest ? 2.5 : 0}
              onMouseEnter={() => setHover([i, j])} onMouseLeave={() => setHover(null)}>
              <title>{`${ageLabel(start[0] + i * step)} / ${ageLabel(start[1] + j * step)}: ${usd(v)}`}</title>
            </rect>
          )
        }))}
      </svg>
      <p className="hint">{hover ? `First person claims at ${ageLabel(start[0] + hover[0] * step)}, second at ${ageLabel(start[1] + hover[1] * step)}: ${usd(grid[hover[0]][hover[1]])} expected lifetime benefits (today's dollars).`
        : 'Rows: first person’s claiming age. Columns: second person’s. Darker = more expected lifetime benefits. Outlined = best.'}</p>
    </div>
  )
}

export function Meter({ value, label }: { value: number; label: string }) {
  const v = Math.max(0, Math.min(1, value))
  return (
    <div className="meter" role="meter" aria-valuenow={Math.round(v * 100)} aria-valuemin={0} aria-valuemax={100} aria-label={label}>
      <div className="meter-fill" style={{ width: `${v * 100}%` }} />
    </div>
  )
}
