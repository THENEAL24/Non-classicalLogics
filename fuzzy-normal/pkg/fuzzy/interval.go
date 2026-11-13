package fuzzy

import (
	"fmt"
	"math"
)

// IntervalFuzzy — нечёткая величина-интервал [A, B]: μ(x)=1 на [A,B], иначе 0
// (прямоугольная / «плато» после α-отсечения исходной кривой).
type IntervalFuzzy struct {
	A           float64     `json:"a"`
	B           float64     `json:"b"`
	Alpha       float64     `json:"alpha"`
	Length      float64     `json:"length"`
	Mid         float64     `json:"mid"`
	SourceIndex int         `json:"sourceIndex"`
	Source      NormalFuzzy `json:"source"`
}

func (i IntervalFuzzy) String() string {
	return fmt.Sprintf("[%.4f, %.4f]_α=%.3f (L=%.4f)", i.A, i.B, i.Alpha, i.Length)
}

// Membership для интервальной нечёткой величины.
func (i IntervalFuzzy) Membership(x float64) float64 {
	if x >= i.A && x <= i.B {
		return 1
	}
	return 0
}

// AlphaCut строит интервал [A, B] = { x | μ(x) ≥ α } для N(a, σ).
func AlphaCut(f NormalFuzzy, alpha float64) (IntervalFuzzy, error) {
	if err := (AlphaCutLevel(alpha)).Validate(); err != nil {
		return IntervalFuzzy{}, err
	}
	if alpha >= 1 {
		return IntervalFuzzy{
			A: f.Mean, B: f.Mean, Alpha: alpha,
			Length: 0, Mid: f.Mean, Source: f,
		}, nil
	}
	delta := f.Sigma * math.Sqrt(-math.Log(alpha))
	lo, hi := f.Support()
	a := f.Mean - delta
	b := f.Mean + delta
	if a < lo {
		a = lo
	}
	if b > hi {
		b = hi
	}
	if b < a {
		a, b = b, a
	}
	return IntervalFuzzy{
		A: a, B: b, Alpha: alpha,
		Length: b - a, Mid: (a + b) / 2, Source: f,
	}, nil
}

// IntervalSet — множество интервальных нечётких величин после отсечения.
type IntervalSet struct {
	Alpha float64         `json:"alpha"`
	Items []IntervalFuzzy `json:"items"`
}

func (s IntervalSet) Len() int { return len(s.Items) }

// IntervalsFromSet для каждой N(a, σ) строит кривую (для отчёта) и α-интервал [A, B].
func IntervalsFromSet(set Set, alpha float64) (IntervalSet, error) {
	if err := (AlphaCutLevel(alpha)).Validate(); err != nil {
		return IntervalSet{}, err
	}
	out := IntervalSet{Alpha: alpha, Items: make([]IntervalFuzzy, 0, set.Len())}
	for i, v := range set.Values {
		iv, err := AlphaCut(v, alpha)
		if err != nil {
			return IntervalSet{}, err
		}
		iv.SourceIndex = i
		out.Items = append(out.Items, iv)
	}
	return out, nil
}
