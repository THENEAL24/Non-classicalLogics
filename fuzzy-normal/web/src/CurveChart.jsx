import { useMemo } from 'react'

/** Одна кривая с α-отсечением. xMin/xMax — общая шкала (чтобы при листании было видно смещение). */
export default function CurveChart({
  curve,
  alpha: alphaProp,
  width = 520,
  height = 160,
  xMin: xMinProp,
  xMax: xMaxProp,
  curveIndex,
}) {
  const geom = useMemo(() => {
    if (!curve?.points?.length) return null
    const pts = curve.points
    const cut = curve.cut
    const alpha = alphaProp ?? cut?.alpha ?? 0.5
    const xs = pts.map((p) => p.x)
    let minX = xMinProp ?? Math.min(...xs, cut.a)
    let maxX = xMaxProp ?? Math.max(...xs, cut.b)
    if (maxX - minX < 1e-6) maxX = minX + 1

    const pad = { l: 12, r: 12, t: 22, b: 24 }
    const w = width - pad.l - pad.r
    const h = height - pad.t - pad.b
    const sx = (x) => pad.l + ((x - minX) / (maxX - minX)) * w
    const sy = (mu) => pad.t + h - mu * h

    const line = pts
      .map((p, i) => `${i === 0 ? 'M' : 'L'} ${sx(p.x).toFixed(2)} ${sy(p.muNorm).toFixed(2)}`)
      .join(' ')

    const mean = curve.source?.mean ?? (cut.a + cut.b) / 2
    const peak = { sx: sx(mean), sy: sy(1) }

    return {
      cut,
      alpha,
      line,
      sx,
      sy,
      pad,
      w,
      h,
      minX,
      maxX,
      xA: sx(cut.a),
      xB: sx(cut.b),
      yAlpha: sy(alpha),
      peak,
      mean,
    }
  }, [curve, alphaProp, width, height, xMinProp, xMaxProp])

  if (!geom) return null

  const { cut, alpha, line, pad, w, h, minX, maxX, xA, xB, yAlpha, peak, mean } = geom
  const ticks = 5
  const xTickVals = Array.from({ length: ticks }, (_, i) => minX + (i / (ticks - 1)) * (maxX - minX))

  return (
    <svg
      key={`${curveIndex}-${mean}-${cut.a}-${cut.b}`}
      viewBox={`0 0 ${width} ${height}`}
      width="100%"
      height={height}
      className="curve-detail-svg"
      role="img"
      aria-label={`Кривая ${(curveIndex ?? 0) + 1}`}
    >
      <rect width={width} height={height} fill="#faf9fc" rx="8" />
      {[0, 0.25, 0.5, 0.75, 1].map((g) => (
        <line
          key={g}
          x1={pad.l}
          y1={pad.t + h * (1 - g)}
          x2={pad.l + w}
          y2={pad.t + h * (1 - g)}
          stroke="#eeeaf5"
        />
      ))}
      <line x1={pad.l} y1={pad.t + h} x2={pad.l + w} y2={pad.t + h} stroke="#ccc4d8" />
      {xTickVals.map((x) => {
        const sx = pad.l + ((x - minX) / (maxX - minX)) * w
        return (
          <g key={x}>
            <line x1={sx} y1={pad.t + h} x2={sx} y2={pad.t + h + 4} stroke="#aaa" />
            <text x={sx} y={height - 6} textAnchor="middle" className="chart-tick">
              {x.toFixed(x >= 10 ? 1 : 2)}
            </text>
          </g>
        )
      })}
      {curveIndex != null && (
        <text x={pad.l + 4} y={14} className="chart-badge">
          № {(curveIndex + 1).toFixed(0)}
        </text>
      )}
      <rect
        x={xA}
        y={pad.t + h - h}
        width={Math.max(xB - xA, 0)}
        height={h}
        fill="#6366f1"
        opacity="0.12"
      />
      <line x1={xA} y1={pad.t} x2={xA} y2={pad.t + h} stroke="#059669" strokeWidth="2" strokeDasharray="4 3" />
      <line x1={xB} y1={pad.t} x2={xB} y2={pad.t + h} stroke="#059669" strokeWidth="2" strokeDasharray="4 3" />
      <line
        x1={peak.sx}
        y1={pad.t}
        x2={peak.sx}
        y2={pad.t + h}
        stroke="#94a3b8"
        strokeWidth="1"
        strokeDasharray="2 4"
      />
      <circle cx={peak.sx} cy={peak.sy} r="4" fill="#312e81" />
      <path d={line} fill="none" stroke="#312e81" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" />
      <line
        x1={pad.l}
        y1={yAlpha}
        x2={pad.l + w}
        y2={yAlpha}
        stroke="#ea580c"
        strokeWidth="1.5"
        strokeDasharray="5 4"
      />
      <text x={pad.l + w - 4} y={yAlpha - 4} textAnchor="end" className="chart-tick alpha">
        α={alpha.toFixed(2)}
      </text>
    </svg>
  )
}
