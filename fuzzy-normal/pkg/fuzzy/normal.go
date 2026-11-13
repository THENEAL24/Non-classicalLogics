package fuzzy

import (
	"errors"
	"fmt"
	"math"
)

// NormalFuzzy — нечёткая величина с нормальной функцией принадлежности
// μ(x) = exp(-((x-a)/σ)²), a — центр (мат. ожидание), σ — параметр размытости.
type NormalFuzzy struct {
	Mean  float64 `json:"mean"`  // a
	Sigma float64 `json:"sigma"` // σ > 0
}

var (
	ErrInvalidSigma = errors.New("sigma must be positive")
	ErrDivByZero    = errors.New("division by zero fuzzy value")
)

// New создаёт нечёткую величину с проверкой параметров.
func New(mean, sigma float64) (NormalFuzzy, error) {
	if sigma <= 0 || math.IsNaN(sigma) || math.IsInf(sigma, 0) {
		return NormalFuzzy{}, ErrInvalidSigma
	}
	if math.IsNaN(mean) || math.IsInf(mean, 0) {
		return NormalFuzzy{}, fmt.Errorf("invalid mean")
	}
	return NormalFuzzy{Mean: mean, Sigma: sigma}, nil
}

// MustNew как New, но паникует при ошибке (для тестов и генератора).
func MustNew(mean, sigma float64) NormalFuzzy {
	v, err := New(mean, sigma)
	if err != nil {
		panic(err)
	}
	return v
}

// Membership вычисляет степень принадлежности μ(x).
func (f NormalFuzzy) Membership(x float64) float64 {
	if f.Sigma <= 0 {
		return 0
	}
	z := (x - f.Mean) / f.Sigma
	return math.Exp(-z * z)
}

// Support возвращает практический носитель [a-3σ, a+3σ].
func (f NormalFuzzy) Support() (lo, hi float64) {
	return f.Mean - 3*f.Sigma, f.Mean + 3*f.Sigma
}

// Centroid — центр тяжести (для симметричной μ совпадает с Mean).
func (f NormalFuzzy) Centroid() float64 {
	return f.Mean
}

// String для вывода и журнала.
func (f NormalFuzzy) String() string {
	return fmt.Sprintf("N(a=%.4f, σ=%.4f)", f.Mean, f.Sigma)
}

// Add сложение (независимые величины): a=a1+a2, σ=√(σ1²+σ2²).
func Add(a, b NormalFuzzy) NormalFuzzy {
	return MustNew(a.Mean+b.Mean, math.Hypot(a.Sigma, b.Sigma))
}

// Sub вычитание: a=a1-a2, σ=√(σ1²+σ2²).
func Sub(a, b NormalFuzzy) NormalFuzzy {
	return MustNew(a.Mean-b.Mean, math.Hypot(a.Sigma, b.Sigma))
}

// Mul умножение (линейная аппроксимация extension principle).
func Mul(a, b NormalFuzzy) NormalFuzzy {
	mean := a.Mean * b.Mean
	sigma := math.Sqrt(math.Pow(b.Mean*a.Sigma, 2) + math.Pow(a.Mean*b.Sigma, 2))
	if sigma < 1e-12 {
		sigma = math.Min(a.Sigma, b.Sigma)
	}
	return MustNew(mean, sigma)
}

// Div деление.
func Div(a, b NormalFuzzy) (NormalFuzzy, error) {
	if math.Abs(b.Mean) < 1e-12 {
		return NormalFuzzy{}, ErrDivByZero
	}
	mean := a.Mean / b.Mean
	sigma := math.Sqrt(
		math.Pow(a.Sigma/b.Mean, 2) +
			math.Pow(a.Mean*b.Sigma/(b.Mean*b.Mean), 2),
	)
	return New(mean, sigma)
}

// CompareResult — результат сравнения.
type CompareResult int

const (
	Less CompareResult = iota - 1
	Equal
	Greater
)

func (c CompareResult) String() string {
	switch c {
	case Less:
		return "меньше"
	case Equal:
		return "равно"
	case Greater:
		return "больше"
	default:
		return "?"
	}
}

// CompareInfo — результат сравнения для UI и API.
type CompareInfo struct {
	Result      CompareResult `json:"result"`
	ResultText  string        `json:"resultText"`
	Relation    string        `json:"relation"`
	Title       string        `json:"title"`
	Explanation string        `json:"explanation"`
	Reason      string        `json:"reason"` // centroid | sigma | equal
	CentroidA   float64       `json:"centroidA"`
	CentroidB   float64       `json:"centroidB"`
	SigmaA      float64       `json:"sigmaA"`
	SigmaB      float64       `json:"sigmaB"`
	Rule        string        `json:"rule"`
	Subst       string        `json:"subst"`
}

// ExplainCompare возвращает подробное пояснение сравнения N₁ и N₂ (первый и второй операнд).
func ExplainCompare(a, b NormalFuzzy) CompareInfo {
	c := Compare(a, b)
	info := CompareInfo{
		Result:     c,
		ResultText: fmt.Sprintf("N₁ %s N₂", c.String()),
		CentroidA:  a.Mean,
		CentroidB:  b.Mean,
		SigmaA:     a.Sigma,
		SigmaB:     b.Sigma,
	}
	const eps = 1e-9
	centroidsEqual := math.Abs(a.Mean-b.Mean) <= eps
	info.Rule = "N₁ ≻ N₂ ⇔ a₁ > a₂  ∨  (a₁ = a₂ ∧ σ₁ < σ₂); N₁ ≺ N₂ — наоборот; при равных a и σ — эквивалентны"

	switch c {
	case Greater:
		info.Relation = "N₁ > N₂"
		if !centroidsEqual {
			info.Reason = "centroid"
			info.Title = "N₁ больше N₂ по центру"
			info.Subst = fmt.Sprintf("a₁ > a₂: %s > %s → N₁ > N₂", fmt.Sprintf("%.4f", a.Mean), fmt.Sprintf("%.4f", b.Mean))
			info.Explanation = fmt.Sprintf(
				"Сравнение идёт по центру (параметр a): у N₁ a=%.4f, у N₂ a=%.4f. "+
					"Так как %.4f > %.4f, N₁ расположена правее на оси и считается большей.",
				a.Mean, b.Mean, a.Mean, b.Mean,
			)
		} else {
			info.Reason = "sigma"
			info.Title = "N₁ больше N₂ по определённости"
			info.Subst = fmt.Sprintf(
				"a₁ = a₂ = %s; σ₁ < σ₂: %s < %s → N₁ > N₂",
				fmt.Sprintf("%.4f", a.Mean), fmt.Sprintf("%.4f", a.Sigma), fmt.Sprintf("%.4f", b.Sigma),
			)
			info.Explanation = fmt.Sprintf(
				"Центры совпадают (a=%.4f). При равных a сравнивается размытость σ: "+
					"меньшая σ означает более «узкую», определённую величину. "+
					"У N₁ σ=%.4f, у N₂ σ=%.4f — поэтому N₁ > N₂.",
				a.Mean, a.Sigma, b.Sigma,
			)
		}
	case Less:
		info.Relation = "N₁ < N₂"
		if !centroidsEqual {
			info.Reason = "centroid"
			info.Title = "N₁ меньше N₂ по центру"
			info.Subst = fmt.Sprintf("a₁ < a₂: %s < %s → N₁ < N₂", fmt.Sprintf("%.4f", a.Mean), fmt.Sprintf("%.4f", b.Mean))
			info.Explanation = fmt.Sprintf(
				"Сравнение по центру a: у N₁ a=%.4f, у N₂ a=%.4f. "+
					"Так как %.4f < %.4f, N₁ считается меньшей.",
				a.Mean, b.Mean, a.Mean, b.Mean,
			)
		} else {
			info.Reason = "sigma"
			info.Title = "N₁ меньше N₂ по определённости"
			info.Subst = fmt.Sprintf(
				"a₁ = a₂ = %s; σ₁ > σ₂: %s > %s → N₁ < N₂",
				fmt.Sprintf("%.4f", a.Mean), fmt.Sprintf("%.4f", a.Sigma), fmt.Sprintf("%.4f", b.Sigma),
			)
			info.Explanation = fmt.Sprintf(
				"Центры совпадают (a=%.4f). У N₁ σ=%.4f больше, чем у N₂ σ=%.4f — "+
					"N₁ более размыта, поэтому в порядке сравнения N₁ < N₂.",
				a.Mean, a.Sigma, b.Sigma,
			)
		}
	case Equal:
		info.Relation = "N₁ ≈ N₂"
		info.Reason = "equal"
		info.Title = "N₁ и N₂ эквивалентны"
		info.Subst = fmt.Sprintf("a₁ = a₂ = %s, σ₁ = σ₂ = %s", fmt.Sprintf("%.4f", a.Mean), fmt.Sprintf("%.4f", a.Sigma))
		info.Explanation = fmt.Sprintf(
			"Центры и размытость совпадают: a=%.4f, σ=%.4f у обеих величин (в пределах точности).",
			a.Mean, a.Sigma,
		)
	}
	return info
}

// Compare по центроидам с учётом перекрытия (при близких a — сравнение σ).
func Compare(a, b NormalFuzzy) CompareResult {
	const eps = 1e-9
	da, db := a.Centroid(), b.Centroid()
	if math.Abs(da-db) > eps {
		if da < db {
			return Less
		}
		return Greater
	}
	if math.Abs(a.Sigma-b.Sigma) <= eps {
		return Equal
	}
	if a.Sigma < b.Sigma {
		return Greater // меньше размытость — «определённее», условно больше
	}
	return Less
}

// Distance — метрика между величинами для кластеризации.
func Distance(a, b NormalFuzzy, sigmaWeight float64) float64 {
	if sigmaWeight <= 0 {
		sigmaWeight = 1
	}
	dm := a.Mean - b.Mean
	ds := a.Sigma - b.Sigma
	return math.Sqrt(dm*dm + sigmaWeight*sigmaWeight*ds*ds)
}
