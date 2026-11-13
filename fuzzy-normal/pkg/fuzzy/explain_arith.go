package fuzzy

import (
	"fmt"
	"math"
)

// ArithExplain — пошаговое пояснение арифметики N(a, σ).
type ArithExplain struct {
	Op            string  `json:"op"`
	Label         string  `json:"label"`
	RuleA         string  `json:"ruleA"`
	RuleSigma     string  `json:"ruleSigma"`
	SubstA        string  `json:"substA"`
	SubstSigma    string  `json:"substSigma"`
	Operand1Mean  float64 `json:"operand1Mean"`
	Operand1Sigma float64 `json:"operand1Sigma"`
	Operand2Mean  float64 `json:"operand2Mean"`
	Operand2Sigma float64 `json:"operand2Sigma"`
	ResultMean    float64 `json:"resultMean"`
	ResultSigma   float64 `json:"resultSigma"`
}

func fmt4(x float64) string {
	return fmt.Sprintf("%.4f", x)
}

// ExplainArithmetic строит общие формулы и подстановку чисел для результата res.
func ExplainArithmetic(op string, a, b, res NormalFuzzy) ArithExplain {
	ex := ArithExplain{
		Op:            op,
		Operand1Mean:  a.Mean,
		Operand1Sigma: a.Sigma,
		Operand2Mean:  b.Mean,
		Operand2Sigma: b.Sigma,
		ResultMean:    res.Mean,
		ResultSigma:   res.Sigma,
	}
	a1, s1, a2, s2 := a.Mean, a.Sigma, b.Mean, b.Sigma
	ra, rs := res.Mean, res.Sigma

	switch op {
	case "add":
		ex.Label = "Сложение"
		ex.RuleA = "a = a₁ + a₂"
		ex.RuleSigma = "σ = √(σ₁² + σ₂²)"
		ex.SubstA = fmt.Sprintf("a = %s + %s = %s", fmt4(a1), fmt4(a2), fmt4(ra))
		ex.SubstSigma = fmt.Sprintf(
			"σ = √(%s² + %s²) = √(%s + %s) = %s",
			fmt4(s1), fmt4(s2),
			fmt4(s1*s1), fmt4(s2*s2),
			fmt4(rs),
		)
	case "sub":
		ex.Label = "Вычитание"
		ex.RuleA = "a = a₁ − a₂"
		ex.RuleSigma = "σ = √(σ₁² + σ₂²)"
		ex.SubstA = fmt.Sprintf("a = %s − %s = %s", fmt4(a1), fmt4(a2), fmt4(ra))
		ex.SubstSigma = fmt.Sprintf(
			"σ = √(%s² + %s²) = √(%s + %s) = %s",
			fmt4(s1), fmt4(s2),
			fmt4(s1*s1), fmt4(s2*s2),
			fmt4(rs),
		)
	case "mul":
		ex.Label = "Умножение"
		ex.RuleA = "a = a₁ · a₂"
		ex.RuleSigma = "σ = √((a₂·σ₁)² + (a₁·σ₂)²)"
		t1 := a2 * s1
		t2 := a1 * s2
		ex.SubstA = fmt.Sprintf("a = %s · %s = %s", fmt4(a1), fmt4(a2), fmt4(ra))
		ex.SubstSigma = fmt.Sprintf(
			"σ = √((%s·%s)² + (%s·%s)²) = √(%s + %s) = %s",
			fmt4(a2), fmt4(s1), fmt4(a1), fmt4(s2),
			fmt4(t1*t1), fmt4(t2*t2),
			fmt4(rs),
		)
	case "div":
		ex.Label = "Деление"
		ex.RuleA = "a = a₁ / a₂  (при a₂ ≠ 0)"
		ex.RuleSigma = "σ = √((σ₁/a₂)² + (a₁·σ₂/a₂²)²)"
		t1 := s1 / a2
		t2 := a1 * s2 / (a2 * a2)
		ex.SubstA = fmt.Sprintf("a = %s / %s = %s", fmt4(a1), fmt4(a2), fmt4(ra))
		ex.SubstSigma = fmt.Sprintf(
			"σ = √((%s/%s)² + (%s·%s/%s²)²) = √(%s + %s) = %s",
			fmt4(s1), fmt4(a2), fmt4(a1), fmt4(s2), fmt4(a2),
			fmt4(t1*t1), fmt4(t2*t2),
			fmt4(rs),
		)
	default:
		ex.Label = op
		ex.RuleA = "a — см. описание операции"
		ex.RuleSigma = "σ — см. описание операции"
		ex.SubstA = fmt.Sprintf("a = %s", fmt4(ra))
		ex.SubstSigma = fmt.Sprintf("σ = %s", fmt4(rs))
	}
	if math.IsNaN(ra) || math.IsNaN(rs) {
		ex.SubstA += " (некорректные операнды)"
	}
	return ex
}
