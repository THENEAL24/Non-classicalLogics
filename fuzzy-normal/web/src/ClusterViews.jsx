import { useEffect, useMemo, useState } from 'react'
import { api } from './api'
import ClusterPlot from './ClusterPlot.jsx'
import { IntervalClusterFormula, NormalClusterFormula } from './ClusterFormula.jsx'

function normalFilterCopy(criterionId) {
  if (criterionId === 'mean') {
    return { metricLabel: 'a', metricHint: 'a ≥ β·max a', rhoHint: 'a' }
  }
  return { metricLabel: 'σ', metricHint: 'σ ≥ β·max σ', rhoHint: 'σ' }
}

function intervalFilterCopy(criterionId) {
  switch (criterionId) {
    case 'ab':
      return {
        metricLabel: '‖(A,B)‖',
        metricHint: '‖(A,B)‖ ≥ β·max',
        rhoHint: '‖(A,B)‖',
      }
    case 'midlen':
      return {
        metricLabel: '‖(m,L)‖',
        metricHint: '‖(m,L)‖ ≥ β·max',
        rhoHint: '‖(m,L)‖',
      }
    default:
      return { metricLabel: 'L', metricHint: 'L ≥ β·max L', rhoHint: 'L' }
  }
}

function ClusterHead({ data }) {
  if (!data) return null
  return (
    <p className="stat">
      K = {data.k}, {data.criterion}, итераций: {data.iterations}
      {data.alpha != null && <> · α = {data.alpha.toFixed(3)}</>}
    </p>
  )
}

function MagnitudeFilterBanner({ filter, metricLabel, rhoHint }) {
  if (!filter) return null
  const before = filter.countBefore ?? 0
  const after = filter.countAfter ?? before
  const mid = filter.countAfterBeta ?? after
  if (after >= before) {
    return (
      <p className="filter-banner filter-banner-idle">
        β = {filter.beta.toFixed(2)}, γ = {filter.gamma.toFixed(2)}: отбор не уменьшает выборку ({before} эл., max{' '}
        {metricLabel} = {filter.max.toFixed(4)}).
      </p>
    )
  }
  const pct = (filter.beta * 100).toFixed(0)
  return (
    <p className="filter-banner">
      <strong>β</strong>: {metricLabel} ≥ {filter.threshold.toFixed(4)} ({pct}% от max {filter.max.toFixed(4)}).
      {mid !== before && mid !== after && (
        <>
          {' '}
          После β: {before} → {mid}.
        </>
      )}
      {filter.gamma < 1 - 1e-6 && (
        <>
          {' '}
          <strong>γ</strong>: топ-{(filter.gamma * 100).toFixed(0)}% по {rhoHint}
          {filter.rhoMinKept != null && (
            <>
              {' '}
              ({rhoHint} от {filter.rhoMinKept.toFixed(3)} до {filter.rhoMaxKept.toFixed(3)})
            </>
          )}
          .
        </>
      )}
      {' '}
      Итого:{' '}
      <strong>
        {before} → {after}
      </strong>
      .
    </p>
  )
}

function MagnitudeFilterControls({
  beta,
  onBetaChange,
  gamma,
  onGammaChange,
  filter,
  metricLabel,
  metricHint,
  rhoHint,
  refining,
  error,
}) {
  return (
    <div className="filter-panel">
      <label className="lbl lbl-range filter-panel-range">
        <span>
          β — порог ({metricHint}) <em className="lbl-hint">{parseFloat(beta).toFixed(2)}</em>
        </span>
        <input
          type="range"
          min="0.05"
          max="1"
          step="0.05"
          value={beta}
          onChange={(e) => onBetaChange(e.target.value)}
        />
      </label>
      <label className="lbl lbl-range filter-panel-range">
        <span>
          γ — доля лидеров по {rhoHint} <em className="lbl-hint">{parseFloat(gamma).toFixed(2)}</em>
        </span>
        <input
          type="range"
          min="0.05"
          max="1"
          step="0.05"
          value={gamma}
          onChange={(e) => onGammaChange(e.target.value)}
        />
      </label>
      <MagnitudeFilterBanner filter={filter} metricLabel={metricLabel} rhoHint={rhoHint} />
      {refining && <p className="note filter-panel-busy">Пересчёт кластеров…</p>}
      {error && <p className="error-inline">{error}</p>}
    </div>
  )
}

function useLiveRefine({ baselineKey, params, fetchRefine }) {
  const [beta, setBeta] = useState('0.8')
  const [gamma, setGamma] = useState('1')
  const [filter, setFilter] = useState(null)
  const [refined, setRefined] = useState(null)
  const [refining, setRefining] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    setFilter(null)
    setRefined(null)
    setError('')
  }, [baselineKey])

  const paramKey = useMemo(() => JSON.stringify(params), [params])

  useEffect(() => {
    if (!baselineKey) return undefined
    const body = JSON.parse(paramKey)
    let cancelled = false
    const timer = setTimeout(async () => {
      setRefining(true)
      setError('')
      try {
        const part = await fetchRefine({ ...body, beta: parseFloat(beta), gamma: parseFloat(gamma) })
        if (cancelled) return
        setFilter(part.filter ?? null)
        setRefined(part.refined ?? null)
      } catch (e) {
        if (!cancelled) {
          setError(e.message || 'Ошибка фильтра')
          setRefined(null)
        }
      } finally {
        if (!cancelled) setRefining(false)
      }
    }, 100)
    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [beta, gamma, baselineKey, paramKey, fetchRefine])

  return { beta, setBeta, gamma, setGamma, filter, refined, refining, error }
}

function NormalClustersBody({ data, criterionId, compact }) {
  if (!data?.clusters?.length) return null
  const h = compact ? 220 : 280
  return (
    <>
      <ClusterHead data={data} />
      {!compact && <NormalClusterFormula data={data} criterionId={criterionId} />}
      <ClusterPlot clusters={data.clusters} variant="normal" criterion={data.criterion} height={h} />
      <div className="cluster-results-scroll">
        {data.clusters.map((c) => (
          <div key={c.index} className="cluster-block">
            <h4 className="cluster-title">
              Кластер {c.index} · {c.size} эл.
            </h4>
            <p className="cluster-centroid">
              Центроид: a={c.centroid.mean.toFixed(4)}, σ={c.centroid.sigma.toFixed(4)}
            </p>
            <div className="scroll-table cluster-table">
              <table className="data">
                <thead>
                  <tr>
                    <th>№</th>
                    <th className="num math-sym">a центр</th>
                    <th className="num math-sym">σ</th>
                  </tr>
                </thead>
                <tbody>
                  {(c.items || []).map((it, j) => (
                    <tr key={j}>
                      <td>{j + 1}</td>
                      <td className="num">{it.mean.toFixed(4)}</td>
                      <td className="num">{it.sigma.toFixed(4)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))}
      </div>
    </>
  )
}

export function NormalClusters({ data, criterionId, clusterParams }) {
  const baseline = data?.baseline ?? data
  const baselineKey = baseline?.report || baseline?.iterations
  const params = useMemo(
    () => ({
      k: clusterParams.k,
      criterion: clusterParams.criterion,
    }),
    [clusterParams.k, clusterParams.criterion],
  )
  const live = useLiveRefine({
    baselineKey,
    params,
    fetchRefine: api.clusterRefine,
  })

  if (!baseline?.clusters?.length) return null
  const nTotal = baseline.clusters.reduce((s, c) => s + (c.size || 0), 0)
  const afterCount = live.filter?.countAfter
  const showRefined = live.refined?.clusters?.length > 0
  const filterUi = normalFilterCopy(criterionId)

  return (
    <div className="cluster-results">
      <section className="cluster-stage">
        <h4 className="cluster-stage-title">1. Исходная кластеризация · N = {nTotal}</h4>
        <NormalClustersBody data={baseline} criterionId={criterionId} compact />
      </section>
      <section className="cluster-stage cluster-stage-filtered">
        <h4 className="cluster-stage-title">
          2. После фильтра по величине
          {afterCount != null && <> · N = {afterCount}</>}
        </h4>
        <MagnitudeFilterControls
          beta={live.beta}
          onBetaChange={live.setBeta}
          gamma={live.gamma}
          onGammaChange={live.setGamma}
          filter={live.filter}
          metricLabel={filterUi.metricLabel}
          metricHint={filterUi.metricHint}
          rhoHint={filterUi.rhoHint}
          refining={live.refining}
          error={live.error}
        />
        {showRefined ? (
          <NormalClustersBody data={live.refined} criterionId={criterionId} />
        ) : (
          !live.refining &&
          !live.error && (
            <p className="note">Уменьшите β, чтобы отсечь часть величин и пересчитать кластеры.</p>
          )
        )}
      </section>
    </div>
  )
}

function IntervalClustersBody({ data, criterionId, compact }) {
  if (!data?.clusters?.length) return null
  const h = compact ? 220 : 280
  return (
    <>
      <ClusterHead data={data} />
      {!compact && <IntervalClusterFormula data={data} criterionId={criterionId} />}
      <ClusterPlot clusters={data.clusters} variant="interval" criterion={data.criterion} height={h} />
      <div className="cluster-results-scroll">
        {data.clusters.map((c) => (
          <div key={c.index} className="cluster-block">
            <h4 className="cluster-title">
              Кластер {c.index} · {c.size} интерв.
            </h4>
            <p className="cluster-centroid">
              Центроид: A={c.centroid.a.toFixed(4)}, B={c.centroid.b.toFixed(4)}, L={c.centroid.length.toFixed(4)}
            </p>
            <div className="scroll-table cluster-table">
              <table className="data cluster-wide">
                <thead>
                  <tr>
                    <th>№</th>
                    <th className="num">A лев.</th>
                    <th className="num">B прав.</th>
                    <th className="num">L</th>
                    <th className="num math-sym">a центр</th>
                    <th className="num math-sym">σ исх.</th>
                  </tr>
                </thead>
                <tbody>
                  {(c.items || []).map((it, j) => (
                    <tr key={j}>
                      <td>{j + 1}</td>
                      <td className="num">{it.a.toFixed(4)}</td>
                      <td className="num">{it.b.toFixed(4)}</td>
                      <td className="num">{it.length.toFixed(4)}</td>
                      <td className="num">{it.source?.mean?.toFixed(4)}</td>
                      <td className="num">{it.source?.sigma?.toFixed(4)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        ))}
      </div>
    </>
  )
}

export function IntervalClusters({ data, criterionId, clusterParams }) {
  const baseline = data?.baseline ?? data
  const baselineKey = baseline?.report || baseline?.iterations
  const params = useMemo(
    () => ({
      k: clusterParams.k,
      criterion: clusterParams.criterion,
    }),
    [clusterParams.k, clusterParams.criterion],
  )
  const live = useLiveRefine({
    baselineKey,
    params,
    fetchRefine: api.intervalClusterRefine,
  })

  if (!baseline?.clusters?.length) return null
  const nTotal = baseline.clusters.reduce((s, c) => s + (c.size || 0), 0)
  const baselineView = { ...baseline, alpha: data?.alpha ?? clusterParams.alpha }
  const refinedView =
    live.refined?.clusters?.length > 0
      ? { ...live.refined, alpha: data?.alpha ?? clusterParams.alpha }
      : null
  const afterCount = live.filter?.countAfter
  const filterUi = intervalFilterCopy(criterionId)

  return (
    <div className="cluster-results">
      <section className="cluster-stage">
        <h4 className="cluster-stage-title">1. Исходная кластеризация · N = {nTotal}</h4>
        <IntervalClustersBody data={baselineView} criterionId={criterionId} compact />
      </section>
      <section className="cluster-stage cluster-stage-filtered">
        <h4 className="cluster-stage-title">
          2. После фильтра по величине
          {afterCount != null && <> · N = {afterCount}</>}
        </h4>
        <MagnitudeFilterControls
          beta={live.beta}
          onBetaChange={live.setBeta}
          gamma={live.gamma}
          onGammaChange={live.setGamma}
          filter={live.filter}
          metricLabel={filterUi.metricLabel}
          metricHint={filterUi.metricHint}
          rhoHint={filterUi.rhoHint}
          refining={live.refining}
          error={live.error}
        />
        {refinedView ? (
          <IntervalClustersBody data={refinedView} criterionId={criterionId} />
        ) : (
          !live.refining &&
          !live.error && (
            <p className="note">Уменьшите β, чтобы отсечь короткие интервалы и пересчитать кластеры.</p>
          )
        )}
      </section>
    </div>
  )
}
