package fuzzy

import (
	"fmt"
	"math"
	"sort"
)

// CurvePoint — точка кривой принадлежности (после нормализации μ̂ ∈ [0, 1]).
type CurvePoint struct {
	X      float64 `json:"x"`
	Mu     float64 `json:"mu"`
	MuNorm float64 `json:"muNorm"`
}

// SampleCurve строит μ(x) на [a−3σ, a+3σ]; в сетку всегда входит x=a (вершина).
func SampleCurve(f NormalFuzzy, n int) []CurvePoint {
	if n < 4 {
		n = 4
	}
	lo, hi := f.Support()
	maxMu := f.Membership(f.Mean)
	if maxMu <= 0 {
		maxMu = 1
	}

	xs := make([]float64, 0, n+1)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1)
		xs = append(xs, lo+t*(hi-lo))
	}
	if f.Mean >= lo && f.Mean <= hi {
		xs = append(xs, f.Mean)
	}
	sort.Float64s(xs)
	// убрать почти дубликаты, оставить точку с большим μ
	merged := make([]CurvePoint, 0, len(xs))
	const eps = 1e-9
	for _, x := range xs {
		mu := f.Membership(x)
		pt := CurvePoint{X: x, Mu: mu, MuNorm: mu / maxMu}
		if len(merged) > 0 && math.Abs(merged[len(merged)-1].X-x) < eps*(hi-lo+1) {
			if pt.MuNorm > merged[len(merged)-1].MuNorm {
				merged[len(merged)-1] = pt
			}
			continue
		}
		merged = append(merged, pt)
	}
	if len(merged) > 0 && merged[len(merged)-1].MuNorm < 0.999 {
		for i := range merged {
			if math.Abs(merged[i].X-f.Mean) < eps*(hi-lo+1) {
				merged[i].MuNorm = 1
				merged[i].Mu = maxMu
				break
			}
		}
	}
	return merged
}

// AlphaCutLevel — уровень α для отсечения: множество { x | μ(x) ≥ α }.
type AlphaCutLevel float64

func (a AlphaCutLevel) Validate() error {
	if a <= 0 || a > 1 || math.IsNaN(float64(a)) {
		return fmt.Errorf("критерий α должен быть в (0, 1], получено %.6f", a)
	}
	return nil
}
