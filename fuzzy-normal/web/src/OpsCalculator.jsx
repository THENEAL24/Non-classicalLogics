import { useMemo } from 'react'

const ARITH_OPS = {
  add: { sym: '+', label: 'Сложение' },
  sub: { sym: '−', label: 'Вычитание' },
  mul: { sym: '×', label: 'Умножение' },
  div: { sym: '÷', label: 'Деление' },
}

function fmtFuzzy(f) {
  if (!f) return '—'
  return `N(a=${f.mean.toFixed(4)}, σ=${f.sigma.toFixed(4)})`
}

export function FormulaBlock({ detail }) {
  if (!detail) return null
  return (
    <div className="formula-block" aria-label="Формулы операции">
      <p className="formula-block-title">{detail.label}</p>
      <dl className="formula-list">
        <div className="formula-row">
          <dt>Общий вид</dt>
          <dd>
            <code className="formula-expr math-sym">{detail.ruleA}</code>
            <br />
            <code className="formula-expr math-sym">{detail.ruleSigma}</code>
          </dd>
        </div>
        <div className="formula-row">
          <dt>Подстановка</dt>
          <dd>
            <code className="formula-expr formula-subst">{detail.substA}</code>
            <br />
            <code className="formula-expr formula-subst">{detail.substSigma}</code>
          </dd>
        </div>
      </dl>
    </div>
  )
}

export function CompareBlock({ detail, caption }) {
  if (!detail) return null
  return (
    <div className="compare-inner">
      {caption && <p className="note">{caption}</p>}
      {detail.rule && (
        <p className="formula-compare-rule">
          <span className="formula-compare-label">Критерий:</span>{' '}
          <code className="formula-expr">{detail.rule}</code>
        </p>
      )}
      {detail.subst && (
        <p className="formula-compare-subst">
          <span className="formula-compare-label">Подстановка:</span>{' '}
          <code className="formula-expr formula-subst">{detail.subst}</code>
        </p>
      )}
      <p>
        <strong>{detail.relation}</strong> — {detail.title}
      </p>
      <p>{detail.explanation}</p>
      <table className="data table-narrow">
        <thead>
          <tr>
            <th />
            <th className="num math-sym">a</th>
            <th className="num math-sym">σ</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>N₁</td>
            <td className="num">{detail.centroidA?.toFixed(4)}</td>
            <td className="num">{detail.sigmaA?.toFixed(4)}</td>
          </tr>
          <tr>
            <td>N₂</td>
            <td className="num">{detail.centroidB?.toFixed(4)}</td>
            <td className="num">{detail.sigmaB?.toFixed(4)}</td>
          </tr>
        </tbody>
      </table>
    </div>
  )
}

function OperandCard({ label, a, sigma, onA, onS, pick }) {
  return (
    <div className="calc-param-card">
      <div className="calc-param-head">
        <span className="calc-slot-label">{label}</span>
        {pick}
      </div>
      <div className="calc-param-fields">
        <label className="calc-param-lbl">
          <span className="math-sym">a</span>
          <input
            type="text"
            inputMode="decimal"
            className="calc-field-input"
            value={a}
            onChange={(e) => onA(e.target.value)}
          />
        </label>
        <label className="calc-param-lbl">
          <span className="math-sym">σ</span>
          <input
            type="text"
            inputMode="decimal"
            className="calc-field-input"
            value={sigma}
            onChange={(e) => onS(e.target.value)}
          />
        </label>
      </div>
      <p className="calc-param-preview">{fmtFuzzy({ mean: parseFloat(a), sigma: parseFloat(sigma) })}</p>
    </div>
  )
}

export default function OpsCalculator({
  busy,
  aMean,
  aSigma,
  bMean,
  bSigma,
  setAMean,
  setASigma,
  setBMean,
  setBSigma,
  opResult,
  lastOp,
  onCompute,
  setItems,
}) {
  const n1 = useMemo(() => ({ mean: parseFloat(aMean), sigma: parseFloat(aSigma) }), [aMean, aSigma])
  const n2 = useMemo(() => ({ mean: parseFloat(bMean), sigma: parseFloat(bSigma) }), [bMean, bSigma])

  const swapOperands = () => {
    setAMean(bMean)
    setASigma(bSigma)
    setBMean(aMean)
    setBSigma(aSigma)
  }

  const resultText =
    lastOp && lastOp !== 'compare' && opResult?.result ? fmtFuzzy(opResult.result) : null

  const expression = `${fmtFuzzy(n1)} ${lastOp && ARITH_OPS[lastOp] ? ARITH_OPS[lastOp].sym : '…'} ${fmtFuzzy(n2)}`

  const pickControl = (slot) =>
    setItems?.length > 0 ? (
      <label className="calc-pick-inline">
        <select
          value=""
          onChange={(e) => {
            const i = parseInt(e.target.value, 10)
            if (Number.isNaN(i)) return
            const it = setItems[i]
            if (!it) return
            if (slot === 1) {
              setAMean(String(it.mean))
              setASigma(String(it.sigma))
            } else {
              setBMean(String(it.mean))
              setBSigma(String(it.sigma))
            }
          }}
        >
          <option value="">из выборки…</option>
          {setItems.slice(0, 200).map((it, i) => (
            <option key={i} value={i}>
              №{i + 1}
            </option>
          ))}
        </select>
      </label>
    ) : null

  return (
    <div className="calc-layout">
      <div className="calc">
        <div className="calc-display">
          <div className="calc-expression">{expression}</div>
          <div className="calc-result" aria-live="polite">
            {resultText ? (
              <>
                <span className="calc-result-eq">=</span> {resultText}
              </>
            ) : (
              <span className="calc-result-placeholder">Введите a, σ с клавиатуры и нажмите операцию</span>
            )}
          </div>
          {lastOp && ARITH_OPS[lastOp] && <div className="calc-op-hint">{ARITH_OPS[lastOp].label}</div>}
        </div>

        <div className="calc-param-stack">
          <OperandCard
            label="N₁"
            a={aMean}
            sigma={aSigma}
            onA={setAMean}
            onS={setASigma}
            pick={pickControl(1)}
          />
          <OperandCard
            label="N₂"
            a={bMean}
            sigma={bSigma}
            onA={setBMean}
            onS={setBSigma}
            pick={pickControl(2)}
          />
        </div>

        <div className="calc-op-bar" role="group" aria-label="Арифметика">
          {Object.entries(ARITH_OPS).map(([op, meta]) => (
            <button
              key={op}
              type="button"
              className="calc-op-btn"
              disabled={busy}
              onClick={() => onCompute(op)}
              title={meta.label}
            >
              {meta.sym}
            </button>
          ))}
          <button type="button" className="calc-op-btn calc-op-btn-muted" disabled={busy} onClick={swapOperands} title="Поменять N₁ и N₂">
            ⇄
          </button>
        </div>

        {opResult?.arithDetail && lastOp && lastOp !== 'compare' && (
          <FormulaBlock detail={opResult.arithDetail} />
        )}
      </div>

      <aside className="calc-compare-panel panel">
        <div className="panel-title">Сравнение</div>
        <div className="panel-inner">
          <p className="note">Сначала сравнивается a; при равных a — меньшая σ.</p>
          <button type="button" className="btn btn-wide" disabled={busy} onClick={() => onCompute('compare')}>
            ≶ Сравнить N₁ и N₂
          </button>
          {lastOp === 'compare' && opResult?.compareDetail?.resultText && (
            <p className="stat compare-verdict">{opResult.compareDetail.resultText}</p>
          )}
          <CompareBlock detail={opResult?.compareDetail} caption={null} />
        </div>
      </aside>
    </div>
  )
}
