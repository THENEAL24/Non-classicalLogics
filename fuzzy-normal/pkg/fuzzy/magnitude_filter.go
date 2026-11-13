package fuzzy

import (
	"fmt"
	"math"
	"sort"
)

// MagnitudeFilterOpts — β/γ по той же величине, что и критерий k-means.
type MagnitudeFilterOpts struct {
	Beta               float64
	Gamma              float64
	NormalCriterion    string // both | mean — как POST /api/cluster
	IntervalCriterion  string // length | ab | midlen
}

// MagnitudeFilterReport — цепочка β (порог) → γ (доля с max величиной).
type MagnitudeFilterReport struct {
	Beta         float64 `json:"beta"`
	Gamma        float64 `json:"gamma"`
	Metric       string  `json:"metric"`
	IndexName    string  `json:"indexName"`
	Max          float64 `json:"max"`
	Threshold    float64 `json:"threshold"`
	Before       int     `json:"countBefore"`
	AfterBeta    int     `json:"countAfterBeta,omitempty"`
	After        int     `json:"countAfter"`
	RhoMinKept   float64 `json:"rhoMinKept,omitempty"`
	RhoMaxKept   float64 `json:"rhoMaxKept,omitempty"`
}

func unitRatio(x float64) (float64, bool) {
	if math.IsNaN(x) || x <= 0 || x > 1 {
		return 1, false
	}
	if x >= 1-1e-12 {
		return 1, false
	}
	return x, true
}

func normalMagnitude(v NormalFuzzy, criterion string) (float64, string) {
	if criterion == "mean" {
		return v.Mean, "mean"
	}
	return v.Sigma, "sigma"
}

func intervalMagnitude(it IntervalFuzzy, criterion string) (float64, string) {
	switch criterion {
	case "ab":
		return math.Hypot(it.A, it.B), "ab"
	case "midlen":
		return math.Hypot(it.Mid, it.Length), "midlen"
	default:
		return it.Length, "length"
	}
}

func applyBetaGamma[T any](
	items []T,
	magnitude func(T) float64,
	metric string,
	beta, gamma float64,
) ([]T, MagnitudeFilterReport, error) {
	rep := MagnitudeFilterReport{
		Beta: beta, Gamma: gamma, Metric: metric, IndexName: metric,
		Before: len(items),
	}
	if len(items) == 0 {
		return nil, rep, fmt.Errorf("пустая выборка")
	}

	maxM := magnitude(items[0])
	for _, it := range items[1:] {
		m := magnitude(it)
		if m > maxM {
			maxM = m
		}
	}
	rep.Max = maxM

	betaR, betaOn := unitRatio(beta)
	rep.Beta = betaR
	if betaOn {
		rep.Threshold = betaR * maxM
	}

	current := items
	afterBeta := len(current)
	if betaOn {
		next := make([]T, 0, len(current))
		for _, it := range current {
			if magnitude(it)+1e-12 >= rep.Threshold {
				next = append(next, it)
			}
		}
		if len(next) == 0 {
			return nil, rep, fmt.Errorf(
				"после β-фильтра (%s)≥%.4f не осталось элементов (β=%.3f, max=%.4f)",
				metric, rep.Threshold, betaR, maxM,
			)
		}
		current = next
		afterBeta = len(current)
	}
	rep.AfterBeta = afterBeta

	gammaR, gammaOn := unitRatio(gamma)
	rep.Gamma = gammaR
	if gammaOn {
		type scored struct {
			item T
			m    float64
		}
		scoredItems := make([]scored, len(current))
		for i, it := range current {
			scoredItems[i] = scored{item: it, m: magnitude(it)}
		}
		sort.Slice(scoredItems, func(i, j int) bool { return scoredItems[i].m > scoredItems[j].m })
		keep := int(math.Ceil(gammaR * float64(len(scoredItems))))
		if keep < 1 {
			keep = 1
		}
		if keep > len(scoredItems) {
			keep = len(scoredItems)
		}
		current = make([]T, keep)
		for i := 0; i < keep; i++ {
			current[i] = scoredItems[i].item
		}
		rep.RhoMinKept = scoredItems[keep-1].m
		rep.RhoMaxKept = scoredItems[0].m
	}

	rep.After = len(current)
	if rep.After == 0 {
		return nil, rep, fmt.Errorf("после фильтров не осталось элементов")
	}
	if !betaOn && !gammaOn {
		rep.AfterBeta = rep.Before
	}
	return current, rep, nil
}

// FilterSetBySigmaBeta оставляет N(a, σ) с σ ≥ β·max(σ) (критерий «a и σ»).
func FilterSetBySigmaBeta(set Set, beta float64) (Set, MagnitudeFilterReport, error) {
	return FilterSetByMagnitude(set, MagnitudeFilterOpts{
		Beta: beta, Gamma: 1, NormalCriterion: "both",
	})
}

// FilterIntervalSetByLengthBeta оставляет интервалы с L ≥ β·max(L).
func FilterIntervalSetByLengthBeta(set IntervalSet, beta float64) (IntervalSet, MagnitudeFilterReport, error) {
	return FilterIntervalSetByMagnitude(set, MagnitudeFilterOpts{
		Beta: beta, Gamma: 1, IntervalCriterion: "length",
	})
}

// FilterSetByMagnitude — β и γ по a (критерий mean) или по σ (критерий both).
func FilterSetByMagnitude(set Set, opt MagnitudeFilterOpts) (Set, MagnitudeFilterReport, error) {
	crit := opt.NormalCriterion
	if crit == "" {
		crit = "both"
	}
	mag, metric := normalMagnitude(set.Values[0], crit)
	_ = mag
	filtered, rep, err := applyBetaGamma(set.Values, func(v NormalFuzzy) float64 {
		m, _ := normalMagnitude(v, crit)
		return m
	}, metric, opt.Beta, opt.Gamma)
	if err != nil {
		return Set{}, rep, err
	}
	return Set{Name: set.Name, Values: filtered}, rep, nil
}

// FilterIntervalSetByMagnitude — β и γ по L, √(A²+B²) или √(m²+L²).
func FilterIntervalSetByMagnitude(set IntervalSet, opt MagnitudeFilterOpts) (IntervalSet, MagnitudeFilterReport, error) {
	crit := opt.IntervalCriterion
	if crit == "" {
		crit = "length"
	}
	filtered, rep, err := applyBetaGamma(set.Items, func(it IntervalFuzzy) float64 {
		m, _ := intervalMagnitude(it, crit)
		return m
	}, intervalMetricLabel(crit), opt.Beta, opt.Gamma)
	if err != nil {
		return IntervalSet{}, rep, err
	}
	return IntervalSet{Alpha: set.Alpha, Items: filtered}, rep, nil
}

func intervalMetricLabel(criterion string) string {
	switch criterion {
	case "ab":
		return "ab"
	case "midlen":
		return "midlen"
	default:
		return "length"
	}
}

// Changed — отбор реально уменьшил выборку.
func (r MagnitudeFilterReport) Changed() bool {
	return r.After < r.Before
}
