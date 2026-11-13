import { useMemo } from 'react'

const PALETTE = [
  '#6366f1',
  '#ec4899',
  '#10b981',
  '#f59e0b',
  '#8b5cf6',
  '#06b6d4',
  '#ef4444',
  '#84cc16',
  '#f97316',
  '#14b8a6',
]

function padExtent(min, max, ratio = 0.08) {
  if (!Number.isFinite(min) || !Number.isFinite(max)) return [0, 1]
  if (Math.abs(max - min) < 1e-9) {
    const m = Math.max(Math.abs(min), 1)
    return [min - m * ratio, max + m * ratio]
  }
  const d = (max - min) * ratio
  return [min - d, max + d]
}

function pickIntervalAxes(criterion) {
  const s = (criterion || '').toLowerCase()
  if (s.includes('границ') || s.includes('a и b')) {
    return {
      xLabel: 'A',
      yLabel: 'B',
      point: (it) => ({ x: it.a, y: it.b }),
      centroid: (c) => ({ x: c.a, y: c.b }),
    }
  }
  if (s.includes('середин') || s.includes('midlen')) {
    return {
      xLabel: 'середина',
      yLabel: 'L',
      point: (it) => ({ x: it.mid ?? (it.a + it.b) / 2, y: it.length }),
      centroid: (c) => ({ x: c.mid ?? (c.a + c.b) / 2, y: c.length }),
    }
  }
  return {
    xLabel: 'середина',
    yLabel: 'L',
    point: (it) => ({ x: it.mid ?? (it.a + it.b) / 2, y: it.length }),
    centroid: (c) => ({ x: c.mid ?? (c.a + c.b) / 2, y: c.length }),
  }
}

export default function ClusterPlot({ clusters, variant = 'normal', criterion, width = 520, height = 300 }) {
  const layout = useMemo(() => {
    if (!clusters?.length) return null

    const axes =
      variant === 'interval'
        ? pickIntervalAxes(criterion)
        : {
            xLabel: 'a',
            yLabel: 'σ',
            point: (it) => ({ x: it.mean, y: it.sigma }),
            centroid: (c) => ({ x: c.mean, y: c.sigma }),
          }

    const series = []
    const xs = []
    const ys = []
    clusters.forEach((cl, i) => {
      const color = PALETTE[i % PALETTE.length]
      const pts = (cl.items || []).map((it, j) => {
        const { x, y } = axes.point(it)
        xs.push(x)
        ys.push(y)
        return { x, y, key: `${cl.index}-${j}` }
      })
      const cen = axes.centroid(cl.centroid || {})
      if (Number.isFinite(cen.x) && Number.isFinite(cen.y)) {
        xs.push(cen.x)
        ys.push(cen.y)
      }
      series.push({
        index: cl.index,
        color,
        points: pts,
        centroid: cen,
        size: cl.size ?? pts.length,
      })
    })

    const [xMin, xMax] = padExtent(Math.min(...xs), Math.max(...xs))
    const [yMin, yMax] = padExtent(Math.min(...ys), Math.max(...ys))

    const pad = { l: 48, r: 16, t: 20, b: 40 }
    const w = width - pad.l - pad.r
    const h = height - pad.t - pad.b
    const sx = (x) => pad.l + ((x - xMin) / (xMax - xMin)) * w
    const sy = (y) => pad.t + h - ((y - yMin) / (yMax - yMin)) * h

    const xTicks = 5
    const yTicks = 5
    const xTickVals = Array.from({ length: xTicks }, (_, i) => xMin + (i / (xTicks - 1)) * (xMax - xMin))
    const yTickVals = Array.from({ length: yTicks }, (_, i) => yMin + (i / (yTicks - 1)) * (yMax - yMin))

    return {
      axes,
      series,
      pad,
      w,
      h,
      sx,
      sy,
      xMin,
      xMax,
      yMin,
      yMax,
      xTickVals,
      yTickVals,
    }
  }, [clusters, variant, criterion, width, height])

  if (!layout) return null

  const { axes, series, pad, w, h, sx, sy, xTickVals, yTickVals } = layout

  return (
    <div className="cluster-plot-wrap">
      <svg
        viewBox={`0 0 ${width} ${height}`}
        width="100%"
        height={height}
        className="cluster-plot-svg"
        role="img"
        aria-label="График кластеризации"
      >
        <rect width={width} height={height} fill="#faf9fc" rx="8" />
        {yTickVals.map((yv) => (
          <line
            key={`g-${yv}`}
            x1={pad.l}
            y1={sy(yv)}
            x2={pad.l + w}
            y2={sy(yv)}
            stroke="#eeeaf5"
          />
        ))}
        <line x1={pad.l} y1={pad.t + h} x2={pad.l + w} y2={pad.t + h} stroke="#c4bcd4" />
        <line x1={pad.l} y1={pad.t} x2={pad.l} y2={pad.t + h} stroke="#c4bcd4" />

        {xTickVals.map((xv) => (
          <text key={`xt-${xv}`} x={sx(xv)} y={pad.t + h + 16} textAnchor="middle" className="chart-tick">
            {xv.toFixed(xv >= 100 || Math.abs(xv) < 0.01 ? 1 : 2)}
          </text>
        ))}
        {yTickVals.map((yv) => (
          <text key={`yt-${yv}`} x={pad.l - 6} y={sy(yv) + 3} textAnchor="end" className="chart-tick">
            {yv.toFixed(yv >= 100 || Math.abs(yv) < 0.01 ? 1 : 2)}
          </text>
        ))}

        <text x={pad.l + w / 2} y={height - 6} textAnchor="middle" className="cluster-axis-label">
          {axes.xLabel}
        </text>
        <text
          x={14}
          y={pad.t + h / 2}
          textAnchor="middle"
          transform={`rotate(-90 14 ${pad.t + h / 2})`}
          className="cluster-axis-label"
        >
          {axes.yLabel}
        </text>

        {series.map((s) => (
          <g key={s.index}>
            {s.points.map((p) => (
              <circle
                key={p.key}
                cx={sx(p.x)}
                cy={sy(p.y)}
                r={4}
                fill={s.color}
                fillOpacity={0.78}
                stroke="#fff"
                strokeWidth={1}
              />
            ))}
            {Number.isFinite(s.centroid.x) && Number.isFinite(s.centroid.y) && (
              <g>
                <circle
                  cx={sx(s.centroid.x)}
                  cy={sy(s.centroid.y)}
                  r={8}
                  fill={s.color}
                  stroke="#1e1b2e"
                  strokeWidth={1.5}
                />
                <line
                  x1={sx(s.centroid.x) - 5}
                  y1={sy(s.centroid.y)}
                  x2={sx(s.centroid.x) + 5}
                  y2={sy(s.centroid.y)}
                  stroke="#1e1b2e"
                  strokeWidth={1.2}
                />
                <line
                  x1={sx(s.centroid.x)}
                  y1={sy(s.centroid.y) - 5}
                  x2={sx(s.centroid.x)}
                  y2={sy(s.centroid.y) + 5}
                  stroke="#1e1b2e"
                  strokeWidth={1.2}
                />
              </g>
            )}
          </g>
        ))}
      </svg>
      <div className="cluster-legend">
        {series.map((s) => (
          <span key={s.index} className="cluster-legend-item">
            <span className="cluster-legend-swatch" style={{ background: s.color }} />
            Кластер {s.index} · {s.size}
          </span>
        ))}
        <span className="cluster-legend-note">Крест на цветном круге — центроид k-means</span>
      </div>
    </div>
  )
}
