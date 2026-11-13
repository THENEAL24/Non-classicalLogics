package fuzzy

import "fmt"

// CurveSample — одна кривая с полной сеткой точек (для просмотра α-отсечения).
type CurveSample struct {
	Index  int           `json:"index"`
	Source NormalFuzzy   `json:"source"`
	Points []CurvePoint  `json:"points"`
	Cut    IntervalFuzzy `json:"cut"`
}

// AnalyzeOptions — построение всех кривых и α-отсечение.
type AnalyzeOptions struct {
	Alpha          float64
	PointsPerCurve int // точек на каждой из N кривых
	PreviewPoints  int // точек в bank для общего графика
}

func (o *AnalyzeOptions) defaults() {
	if o.Alpha <= 0 {
		o.Alpha = 0.5
	}
	if o.PointsPerCurve <= 0 {
		o.PointsPerCurve = 40
	}
	if o.PreviewPoints <= 0 {
		o.PreviewPoints = 64
	}
}

// AnalyzeReport: N нечётких величин → N кривых → N интервалов [A,B].
type AnalyzeReport struct {
	Alpha         float64         `json:"alpha"`
	CurveCount    int             `json:"curveCount"`
	IntervalCount int             `json:"intervalCount"`
	CurveBank     CurveBank       `json:"curveBank"`
	Intervals     []IntervalFuzzy `json:"intervals"`
	Curves        []CurveSample   `json:"curves"` // все N кривых с отсечением
	Summary       string          `json:"summary"`
}

// AnalyzeSet: для каждой нечёткой величины-параметра (a, σ) строится кривая, затем α-cut.
func AnalyzeSet(set Set, opt AnalyzeOptions) (AnalyzeReport, error) {
	opt.defaults()
	n := set.Len()
	if n == 0 {
		return AnalyzeReport{}, fmt.Errorf("множество пусто: сначала сгенерируйте нечёткие величины-параметры кривых")
	}

	bank := BuildCurveBank(set, opt.PointsPerCurve, opt.PreviewPoints)

	iv, err := IntervalsFromSet(set, opt.Alpha)
	if err != nil {
		return AnalyzeReport{}, err
	}
	curves := make([]CurveSample, n)
	for i := 0; i < n; i++ {
		curves[i] = CurveSample{
			Index:  i,
			Source: set.Values[i],
			Points: CurveFromParams(set.Values[i], opt.PointsPerCurve),
			Cut:    iv.Items[i],
		}
	}

	summary := fmt.Sprintf(
		"%d кривых μ(x) (по %d точек). α=%.3f → %d интервалов [A,B].",
		n, opt.PointsPerCurve, opt.Alpha, iv.Len(),
	)

	return AnalyzeReport{
		Alpha:         opt.Alpha,
		CurveCount:    n,
		IntervalCount: iv.Len(),
		CurveBank:     bank,
		Intervals:     iv.Items,
		Curves:        curves,
		Summary:       summary,
	}, nil
}
