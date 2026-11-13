function spanFromItems(items, pick) {
  if (!items.length) return null
  let min = Infinity
  let max = -Infinity
  for (const it of items) {
    const v = pick(it)
    if (!Number.isFinite(v)) continue
    if (v < min) min = v
    if (v > max) max = v
  }
  if (!Number.isFinite(min) || !Number.isFinite(max)) return null
  return max - min
}

function allClusterItems(clusters) {
  return (clusters || []).flatMap((c) => c.items || [])
}

function FormulaPanel({ title, rows, note }) {
  return (
    <div className="formula-block cluster-formula" aria-label="Формулы кластеризации">
      <p className="formula-block-title">{title}</p>
      <dl className="formula-list">
        {rows.map(({ label, expr, subst }) => (
          <div key={label} className="formula-row">
            <dt>{label}</dt>
            <dd>
              <code className="formula-expr math-sym">{expr}</code>
              {subst != null && (
                <>
                  <br />
                  <code className="formula-expr formula-subst">{subst}</code>
                </>
              )}
            </dd>
          </div>
        ))}
      </dl>
      {note && <p className="note cluster-formula-note">{note}</p>}
    </div>
  )
}

export function NormalClusterFormula({ data, criterionId }) {
  if (!data?.clusters?.length) return null

  const id = criterionId === 'mean' ? 'mean' : 'both'
  const items = allClusterItems(data.clusters)
  const spanA = spanFromItems(items, (it) => it.mean)
  const spanS = spanFromItems(items, (it) => it.sigma)
  const w =
    spanA != null && spanS != null && spanS > 1e-12 ? spanA / spanS : null

  if (id === 'mean') {
    return (
      <FormulaPanel
        title="Метрика (критерий: только a)"
        rows={[
          {
            label: 'd(N₁, N₂)',
            expr: 'd = |a₁ − a₂|',
            subst: 'σ не участвует.',
          },
        ]}
      />
    )
  }

  const wSubst =
    w != null
      ? `w = (max a − min a) / (max σ − min σ) ≈ ${w.toFixed(2)} (по этой выборке)`
      : 'w — вес σ в метрике (масштаб выборки).'

  return (
    <FormulaPanel
      title="Метрика (критерий: a и σ)"
      rows={[
        {
          label: 'd(N₁, N₂)',
          expr: 'd = √((a₁ − a₂)² + w²(σ₁ − σ₂)²)',
          subst: wSubst,
        },
      ]}
      note="w подбирается по размаху a и σ в выборке, чтобы обе координаты сопоставимо влияли на d."
    />
  )
}

export function IntervalClusterFormula({ data, criterionId }) {
  if (!data?.clusters?.length) return null

  const id = criterionId || 'length'

  if (id === 'length') {
    return (
      <FormulaPanel
        title="Метрика (критерий: длина L)"
        rows={[
          {
            label: 'd',
            expr: 'd = |L₁ − L₂|,   L = B − A',
          },
        ]}
      />
    )
  }

  if (id === 'ab') {
    return (
      <FormulaPanel
        title="Метрика (критерий: A и B)"
        rows={[
          {
            label: 'd',
            expr: 'd = √((A₁ − A₂)² + (B₁ − B₂)²)',
          },
        ]}
      />
    )
  }

  const items = allClusterItems(data.clusters)
  const spanM = spanFromItems(items, (it) => it.mid ?? (it.a + it.b) / 2)
  const spanL = spanFromItems(items, (it) => it.length)

  let note = null
  if (spanM != null && spanL != null && spanL > 1e-9 && spanM / spanL > 8) {
    note =
      'Размах m заметно больше размаха L; при этой формуле d чаще определяется положением середины, чем длиной.'
  }

  return (
    <FormulaPanel
      title="Метрика (критерий: середина и L)"
      rows={[
        {
          label: 'd',
          expr: 'd = √((m₁ − m₂)² + (L₁ − L₂)²),   m = (A + B) / 2',
        },
      ]}
      note={note}
    />
  )
}
