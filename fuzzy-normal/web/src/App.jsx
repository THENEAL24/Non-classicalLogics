import { useCallback, useEffect, useState } from 'react'
import { api } from './api'
import CurvesExplorer from './CurvesExplorer.jsx'
import CutsPager from './CutsPager.jsx'
import { IntervalClusters, NormalClusters } from './ClusterViews.jsx'
import OpsCalculator from './OpsCalculator.jsx'

const PAGES = [
  { id: 'data', label: 'Генерация' },
  { id: 'intervals', label: 'Кривые и α' },
  { id: 'cluster', label: 'Кластеры' },
  { id: 'ops', label: 'Операции' },
  { id: 'mu', label: 'μ(x)' },
  { id: 'log', label: 'Журнал' },
]

function Panel({ title, children }) {
  return (
    <div className="panel">
      {title && <div className="panel-title">{title}</div>}
      <div className="panel-inner">{children}</div>
    </div>
  )
}

function PageHead({ title, lead }) {
  return (
    <header className="page-head">
      <h2>{title}</h2>
      {lead && <p>{lead}</p>}
    </header>
  )
}

function Lbl({ label, value, onChange, wide, type = 'number', step = 'any' }) {
  return (
    <label className={`lbl ${wide ? 'wide' : ''}`}>
      <span>{label}</span>
      <input type={type} step={step} value={value} onChange={(e) => onChange(e.target.value)} />
    </label>
  )
}

export default function App() {
  const [page, setPage] = useState('data')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const [aMean, setAMean] = useState('10')
  const [aSigma, setASigma] = useState('2')
  const [bMean, setBMean] = useState('5')
  const [bSigma, setBSigma] = useState('1.5')
  const [opResult, setOpResult] = useState(null)
  const [lastOp, setLastOp] = useState('')

  const [count, setCount] = useState('150')
  const [setData, setSetData] = useState({ count: 0, items: [] })

  const [k, setK] = useState('3')
  const [criterion, setCriterion] = useState('both')
  const [clusterResult, setClusterResult] = useState(null)

  const [x, setX] = useState('10')
  const [muResult, setMuResult] = useState(null)
  const [journal, setJournal] = useState('')

  const [alpha, setAlpha] = useState('0.5')
  const [intervalReport, setIntervalReport] = useState(null)
  const [ivK, setIvK] = useState('3')
  const [ivCrit, setIvCrit] = useState('length')
  const [ivCluster, setIvCluster] = useState(null)

  const n = setData.count ?? 0
  const fuzzyA = () => ({ mean: parseFloat(aMean), sigma: parseFloat(aSigma) })
  const fuzzyB = () => ({ mean: parseFloat(bMean), sigma: parseFloat(bSigma) })
  const refreshJournal = useCallback(async () => {
    const j = await api.journal()
    setJournal(j.text || '')
  }, [])

  const refreshSet = useCallback(async () => {
    setSetData(await api.getSet())
  }, [])

  useEffect(() => {
    refreshJournal().catch(() => {})
    refreshSet().catch(() => {})
  }, [refreshJournal, refreshSet])

  const run = async (fn) => {
    setError('')
    setBusy(true)
    try {
      await fn()
      await refreshJournal()
    } catch (e) {
      setError(e.message)
    } finally {
      setBusy(false)
    }
  }

  const doOp = (op) =>
    run(async () => {
      const data = await api.ops({ op, a: fuzzyA(), b: fuzzyB() })
      setLastOp(op)
      setOpResult(data)
    })

  const statusText = busy
    ? 'Идёт расчёт…'
    : n > 0
      ? `В памяти ${n} величин (параметров кривых).`
      : 'Данные не сгенерированы — начните с п. 1.'

  return (
    <div className={`layout${page === 'cluster' ? ' layout-wide' : ''}`}>
      <aside className="side">
        <div className="brand">
          <h1>Неклассические логики</h1>
          <p>Нормальное распределение · N(a, σ)</p>
          <p className="brand-formula math-sym" title="Степень принадлежности">
            μ(x) = exp(−((x − a) / σ)²)
          </p>
        </div>
        <nav>
          {PAGES.map((p) => (
            <button
              key={p.id}
              type="button"
              className={`nav-item ${page === p.id ? 'active' : ''}`}
              onClick={() => setPage(p.id)}
            >
              {p.label}
            </button>
          ))}
        </nav>
        <div className="side-foot">{statusText}</div>
      </aside>

      <div className="main">
        {error && <div className="alert alert-err">{error}</div>}

        {page === 'data' && (
          <>
            <PageHead
              title="Генерация"
            />
            <p className="gen-brief math-sym">
              <strong className="gen-brief-label">Как генерируется:</strong> для каждой из N величин свой ξ и η —
              равномерно на [0; 1].{' '}
              <code className="gen-formula">a = a_min + ξ·(a_max − a_min)</code>,{' '}
              <code className="gen-formula">σ = σ_min + η·(σ_max − σ_min)</code>. По умолчанию a ∈ [0; 100],
              σ ∈ [0,5; 8]. Дальше для каждой пары (a, σ) строится μ(x); при повторной генерации числа
              новые.
            </p>
            <Panel title="Параметры">
              <div className="form-row">
                <Lbl label="Количество N" value={count} onChange={setCount} />
                <button
                  type="button"
                  className="btn"
                  disabled={busy}
                  onClick={() =>
                    run(async () => {
                      setSetData(await api.generate({ count: parseInt(count, 10) }))
                      setIntervalReport(null)
                      setClusterResult(null)
                      setIvCluster(null)
                    })
                  }
                >
                  Сгенерировать
                </button>
              </div>
              <p className="stat">
                В памяти: <strong>{n}</strong>
              </p>
            </Panel>
            <Panel title="Таблица сгенерированных">
              {n === 0 ? (
                <p className="note">Нажмите «Сгенерировать».</p>
              ) : (
                <div className="scroll-table scroll-table-centered">
                  <table className="data data-centered">
                    <thead>
                      <tr>
                        <th>№</th>
                        <th className="num math-sym">a</th>
                        <th className="num math-sym">σ</th>
                      </tr>
                    </thead>
                    <tbody>
                      {(setData.items || []).map((it, i) => (
                        <tr key={i}>
                          <td>{i + 1}</td>
                          <td className="num">{it.mean.toFixed(4)}</td>
                          <td className="num">{it.sigma.toFixed(4)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </Panel>
          </>
        )}

        {page === 'intervals' && (
          <>
            <PageHead title="Кривые и α-отсечение" />
            <ol className="steps">
              <li>По каждой из N величин строится кривая μ(x), нормализация μ̂=μ/max μ.</li>
              <li>Задаётся уровень α; множество {'{x | μ(x) ≥ α}'} — отрезок [A, B].</li>
              <li>[A, B] — интервальная нечёткая величина (μ=1 на отрезке).</li>
              <li>Длины L=B−A; интервалы можно кластеризовать (п. 3).</li>
            </ol>
            <Panel title="Расчёт">
              <div className="form-row">
                <Lbl label="α" value={alpha} onChange={setAlpha} step="0.01" />
                <button
                  type="button"
                  className="btn"
                  disabled={busy || n === 0}
                  onClick={() =>
                    run(async () => {
                      const data = await api.intervalAnalyze({
                        alpha: parseFloat(alpha),
                      })
                      setIntervalReport(data)
                      setIvCluster(null)
                    })
                  }
                >
                  Построить кривые и интервалы
                </button>
              </div>
              {n === 0 && <p className="note">Сначала п. 1 — генерация.</p>}
            </Panel>

            {intervalReport && (
              <>
                <Panel title="Итог">
                  <p className="stat">{intervalReport.summary}</p>
                </Panel>
                <Panel title={`Общий вид — ${intervalReport.curveCount} кривых`}>
                  <CurvesExplorer bank={intervalReport.curveBank} alpha={parseFloat(alpha)} />
                </Panel>
                <Panel
                  title={
                    <>
                      <span className="math-sym">α</span>-отсечение — все {intervalReport.curveCount} кривых
                    </>
                  }
                >
                  <CutsPager curves={intervalReport.curves} alpha={parseFloat(alpha)} />
                </Panel>
                <Panel title={`Таблица интервалов — ${intervalReport.intervalCount} строк`}>
                  <div className="scroll-table">
                    <table className="data">
                      <thead>
                        <tr>
                          <th>№</th>
                          <th className="num">A</th>
                          <th className="num">B</th>
                          <th className="num">L</th>
                          <th>
                            исходные a, <span className="math-sym">σ</span>
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {(intervalReport.intervals || []).map((it, i) => (
                          <tr key={i}>
                            <td>{i + 1}</td>
                            <td className="num">{it.a.toFixed(4)}</td>
                            <td className="num">{it.b.toFixed(4)}</td>
                            <td className="num">{it.length.toFixed(4)}</td>
                            <td>
                              {it.source.mean.toFixed(2)}; {it.source.sigma.toFixed(2)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </Panel>
              </>
            )}
          </>
        )}

        {page === 'cluster' && (
          <>
            <PageHead
              title="Кластеризация"
              lead="Сначала k-means по всей выборке; ниже в блоке «После фильтра» ползунок β пересчитывает кластеры в реальном времени."
            />
            <div className="two-col cluster-cols">
              <Panel
                title={
                  <>
                    Исходные <span className="math-sym">N(a, σ)</span>
                  </>
                }
              >
                <div className="form-row">
                  <Lbl label="K" value={k} onChange={setK} />
                  <label className="lbl">
                    <span>Критерий</span>
                    <select value={criterion} onChange={(e) => setCriterion(e.target.value)}>
                      <option value="both">a и σ</option>
                      <option value="mean">только a</option>
                    </select>
                  </label>
                  <button
                    type="button"
                    className="btn"
                    disabled={busy || n === 0}
                    onClick={() =>
                      run(async () => {
                        setClusterResult(
                          await api.cluster({
                            k: parseInt(k, 10),
                            criterion,
                          }),
                        )
                      })
                    }
                  >
                    k-means
                  </button>
                </div>
                {clusterResult && (
                  <NormalClusters
                    data={clusterResult}
                    criterionId={criterion}
                    clusterParams={{ k: parseInt(k, 10), criterion }}
                  />
                )}
              </Panel>
              <Panel title="Интервалы [A, B]">
                <div className="form-row">
                  <Lbl label="K" value={ivK} onChange={setIvK} />
                  <label className="lbl">
                    <span>Критерий</span>
                    <select value={ivCrit} onChange={(e) => setIvCrit(e.target.value)}>
                      <option value="length">длина L</option>
                      <option value="ab">A и B</option>
                      <option value="midlen">середина и L</option>
                    </select>
                  </label>
                  <button
                    type="button"
                    className="btn"
                    disabled={busy || !intervalReport}
                    onClick={() =>
                      run(async () => {
                        setIvCluster(
                          await api.intervalCluster({
                            alpha: parseFloat(alpha),
                            k: parseInt(ivK, 10),
                            criterion: ivCrit,
                          }),
                        )
                      })
                    }
                  >
                    k-means
                  </button>
                </div>
                {!intervalReport && <p className="note">Нужен расчёт в п. 2.</p>}
                {ivCluster && (
                  <IntervalClusters
                    data={ivCluster}
                    criterionId={ivCrit}
                    clusterParams={{
                      k: parseInt(ivK, 10),
                      criterion: ivCrit,
                      alpha: parseFloat(alpha),
                    }}
                  />
                )}
              </Panel>
            </div>
          </>
        )}

        {page === 'ops' && (
          <>
            <PageHead
              title="Операции"
              lead="Калькулятор N(a, σ): введите N₁ и N₂, нажмите операцию. Можно подставить величину из сгенерированной выборки."
            />
            <Panel title="Калькулятор и сравнение">
              <OpsCalculator
                busy={busy}
                aMean={aMean}
                aSigma={aSigma}
                bMean={bMean}
                bSigma={bSigma}
                setAMean={setAMean}
                setASigma={setASigma}
                setBMean={setBMean}
                setBSigma={setBSigma}
                opResult={opResult}
                lastOp={lastOp}
                onCompute={doOp}
                setItems={setData.items}
              />
            </Panel>
          </>
        )}

        {page === 'mu' && (
          <>
            <PageHead title="μ(x)" lead="Степень принадлежности в заданной точке." />
            <Panel title="Ввод">
              <div className="form-row">
                <Lbl label="a" value={aMean} onChange={setAMean} />
                <Lbl label="σ" value={aSigma} onChange={setASigma} />
                <Lbl label="x" value={x} onChange={setX} />
                <button
                  type="button"
                  className="btn"
                  disabled={busy}
                  onClick={() =>
                    run(async () => {
                      setMuResult(await api.membership({ fuzzy: fuzzyA(), x: parseFloat(x) }))
                    })
                  }
                >
                  Вычислить
                </button>
              </div>
              {muResult && (
                <p className="stat">
                  μ({muResult.x}) = <strong>{muResult.mu.toFixed(6)}</strong>
                </p>
              )}
            </Panel>
          </>
        )}

        {page === 'log' && (
          <>
            <PageHead title="Журнал" lead="Запись действий в fuzzy_journal.log" />
            <Panel title="События">
              <button type="button" className="btn btn-ghost" disabled={busy} onClick={() => refreshJournal()}>
                Обновить
              </button>
              <pre className="log">{journal || '(пусто)'}</pre>
            </Panel>
          </>
        )}
      </div>
    </div>
  )
}
