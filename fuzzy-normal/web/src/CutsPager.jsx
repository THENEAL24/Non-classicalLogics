import { useEffect, useMemo, useState } from 'react'
import CurveChart from './CurveChart.jsx'

function xDomainForCurves(curves) {
  let minX = Infinity
  let maxX = -Infinity
  for (const c of curves || []) {
    for (const p of c.points || []) {
      if (p.x < minX) minX = p.x
      if (p.x > maxX) maxX = p.x
    }
    if (c.cut) {
      minX = Math.min(minX, c.cut.a)
      maxX = Math.max(maxX, c.cut.b)
    }
  }
  if (!Number.isFinite(minX)) return { xMin: 0, xMax: 1 }
  const m = (maxX - minX) * 0.02 || 0.5
  return { xMin: minX - m, xMax: maxX + m }
}

export default function CutsPager({ curves, alpha }) {
  const total = curves?.length || 0
  const [idx, setIdx] = useState(0)

  const { xMin, xMax } = useMemo(() => xDomainForCurves(curves), [curves])

  useEffect(() => {
    setIdx(0)
  }, [curves])

  useEffect(() => {
    const onKey = (e) => {
      if (e.key === 'ArrowLeft') setIdx((i) => Math.max(0, i - 1))
      if (e.key === 'ArrowRight') setIdx((i) => Math.min(total - 1, i + 1))
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [total])

  if (!total) return null

  const c = curves[idx]
  const cut = c.cut

  return (
    <div className="pager">
      <div className="pager-nav">
        <button type="button" className="btn btn-sm btn-ghost" disabled={idx <= 0} onClick={() => setIdx((i) => i - 1)}>
          ← Назад
        </button>
        <label className="pager-jump lbl">
          <span>Кривая</span>
          <input
            type="number"
            min={1}
            max={total}
            value={idx + 1}
            onChange={(e) => {
              const v = parseInt(e.target.value, 10)
              if (!Number.isNaN(v)) setIdx(Math.min(total - 1, Math.max(0, v - 1)))
            }}
          />
        </label>
        <span className="pager-of">из {total}</span>
        <button
          type="button"
          className="btn btn-sm btn-ghost"
          disabled={idx >= total - 1}
          onClick={() => setIdx((i) => i + 1)}
        >
          Вперёд →
        </button>
      </div>
      <div className="pager-chart" key={`chart-wrap-${idx}`}>
        <CurveChart
          curve={c}
          alpha={alpha}
          width={520}
          height={160}
          xMin={xMin}
          xMax={xMax}
          curveIndex={idx}
        />
      </div>
      <table className="data pager-meta">
        <tbody>
          <tr>
            <th>Параметры</th>
            <td className="num">
              a = {c.source.mean.toFixed(4)}, σ = {c.source.sigma.toFixed(4)}
            </td>
          </tr>
          <tr>
            <th>Отсечение [A, B]</th>
            <td className="num">
              A = {cut.a.toFixed(4)}, B = {cut.b.toFixed(4)}, L = {cut.length.toFixed(4)}
            </td>
          </tr>
        </tbody>
      </table>
      <p className="note">
        Ось x общая для всех {total} кривых — при переключении видно смещение «колокола». Точка — вершина (a, 1).
      </p>
    </div>
  )
}
